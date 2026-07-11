package chillpay

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Config holds ChillPay merchant credentials.
type Config struct {
	MerchantCode string
	APIKey       string
	MD5Key       string
	BaseURL      string
	RouteNo      int
	Currency     string
	CallbackURL  string
	ReturnURL    string
}

// Client handles communication with the ChillPay payment gateway API.
// Logic aligned with harvest-core internal/plugins/billing/payment/chillpay.go.
type Client struct {
	cfg        Config
	httpClient *http.Client
}

func NewClient(cfg Config) *Client {
	if cfg.RouteNo <= 0 {
		cfg.RouteNo = 1
	}
	if strings.TrimSpace(cfg.Currency) == "" {
		cfg.Currency = "764"
	}
	return &Client{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// CallbackForm represents the callback payload sent by ChillPay after a payment attempt.
// Field names match ChillPay V2 callback parameters.
type CallbackForm struct {
	OrderNo            string `json:"OrderNo" form:"OrderNo"`
	Amount             string `json:"Amount" form:"Amount"`
	TransactionId      string `json:"TransactionId" form:"TransactionId"`
	CustomerId         string `json:"CustomerId" form:"CustomerId"`
	CustomerName       string `json:"CustomerName" form:"CustomerName"`
	BankCode           string `json:"BankCode" form:"BankCode"`
	PaymentDate        string `json:"PaymentDate" form:"PaymentDate"`
	PaymentStatus      string `json:"PaymentStatus" form:"PaymentStatus"` // "0" success, "1" pending, "2" failed
	PaymentDescription string `json:"PaymentDescription" form:"PaymentDescription"`
	BankRefCode        string `json:"BankRefCode" form:"BankRefCode"`
	Currency           string `json:"Currency" form:"Currency"`
	CreditCardToken    string `json:"CreditCardToken" form:"CreditCardToken"`
	CurrentDate        string `json:"CurrentDate" form:"CurrentDate"`
	CurrentTime        string `json:"CurrentTime" form:"CurrentTime"`
	CheckSum           string `json:"CheckSum" form:"CheckSum"`
}

// RequestInfo holds caller-supplied context for InitPayment.
// Empty/zero fields are sent as empty strings (optional params).
type RequestInfo struct {
	OrderNo     string  // Unique order reference (required)
	CustomerID  string  // End-user ID or name (required)
	Amount      float64 // Payment amount in major units (required)
	Description string  // Payment description (optional)
	ChannelCode string  // e.g. "creditcard" (optional; defaults to creditcard)
	IPAddress   string  // Client IP address
	LangCode    string  // "TH" or "EN" (optional)
	PhoneNumber string  // End-user phone (optional)
	CustEmail   string  // End-user email (optional)
	CustName    string  // End-user name (optional; must not be an email)
}

// initResponse matches Table 2.3 of the ChillPay Merchant Integration Manual.
type initResponse struct {
	Status        int    `json:"Status"`
	Code          int    `json:"Code"`
	Message       string `json:"Message"`
	TransactionId int64  `json:"TransactionId"`
	Amount        int64  `json:"Amount"`
	OrderNo       string `json:"OrderNo"`
	CustomerId    string `json:"CustomerId"`
	ChannelCode   string `json:"ChannelCode"`
	ReturnUrl     string `json:"ReturnUrl"`
	PaymentUrl    string `json:"PaymentUrl"`
	IpAddress     string `json:"IpAddress"`
	Token         string `json:"Token"`
	CreatedDate   string `json:"CreatedDate"`
	ExpiredDate   string `json:"ExpiredDate"`
}

// StatusResponse represents the JSON response from the ChillPay PaymentStatus API.
type StatusResponse struct {
	TransactionId      int64  `json:"TransactionId"`
	Amount             int64  `json:"Amount"`
	OrderNo            string `json:"OrderNo"`
	CustomerId         string `json:"CustomerId"`
	BankCode           string `json:"BankCode"`
	PaymentDate        string `json:"PaymentDate"`
	PaymentStatus      int    `json:"PaymentStatus"`
	BankRefCode        string `json:"BankRefCode"`
	CurrentDate        string `json:"CurrentDate"`
	CurrentTime        string `json:"CurrentTime"`
	PaymentDescription string `json:"PaymentDescription"`
	CreditCardToken    string `json:"CreditCardToken"`
	Currency           string `json:"Currency"`
}

// InitPayment creates a payment session and returns the redirect URL + transaction id.
// All 21 parameters from Table 2.2 are sent; CheckSum is computed over
// parameters 1–20 concatenated in order + MD5 Secret Key (same as harvest-core).
func (c *Client) InitPayment(info RequestInfo) (paymentURL string, txnID string, err error) {
	// ChillPay expects amount in smallest currency unit (satang for THB).
	amountInt := int(info.Amount * 100)
	amountStr := strconv.Itoa(amountInt)

	// OrderNo: max 20 alphanumeric (ChillPay 1006). Prefer pre-sanitized store values.
	orderNo := SanitizeOrderNo(info.OrderNo)
	if orderNo == "" {
		return "", "", fmt.Errorf("chillpay OrderNo is required (max 20 alphanumeric)")
	}
	customerID := SanitizeCustomerID(info.CustomerID)
	if customerID == "" {
		return "", "", fmt.Errorf("chillpay CustomerId is required")
	}

	channelCode := info.ChannelCode
	if channelCode == "" {
		channelCode = "creditcard"
	}

	langCode := info.LangCode
	if langCode == "" {
		langCode = "TH"
	}

	ipAddress := info.IPAddress
	if ipAddress == "" {
		ipAddress = "127.0.0.1"
	}

	routeNoStr := strconv.Itoa(c.cfg.RouteNo)

	// Optional fields default to empty string (harvest-core pattern).
	phoneNumber := SanitizePhone(info.PhoneNumber)
	description := strings.TrimSpace(info.Description)
	if description == "" {
		description = "MontiPackage"
	}
	custEmail := strings.TrimSpace(info.CustEmail)
	// CustName must be a person name — never an email (ChillPay 2032).
	custName := SanitizeCustName(info.CustName, custEmail)

	// TokenFlag, CreditToken, CreditMonth, ShopID, ProductImageUrl, CardType
	// are optional and not used in our flow — sent as empty strings.
	tokenFlag := "N"
	creditToken := ""
	creditMonth := ""
	shopID := ""
	productImageUrl := ""
	cardType := ""

	// Build CheckSum per harvest-core / ChillPay Table 2.2:
	// MerchantCode + OrderNo + CustomerId + Amount + PhoneNumber + Description +
	// ChannelCode + Currency + LangCode + RouteNo + IPAddress + ApiKey + TokenFlag +
	// CreditToken + CreditMonth + ShopID + ProductImageUrl + CustEmail + CardType +
	// CustName + MD5SecretKey
	checksumRaw := c.cfg.MerchantCode +
		orderNo +
		customerID +
		amountStr +
		phoneNumber +
		description +
		channelCode +
		c.cfg.Currency +
		langCode +
		routeNoStr +
		ipAddress +
		c.cfg.APIKey +
		tokenFlag +
		creditToken +
		creditMonth +
		shopID +
		productImageUrl +
		custEmail +
		cardType +
		custName +
		c.cfg.MD5Key
	checksum := md5Hex(checksumRaw)

	form := url.Values{}
	form.Set("MerchantCode", c.cfg.MerchantCode) // 1
	form.Set("OrderNo", orderNo)                 // 2
	form.Set("CustomerId", customerID)           // 3
	form.Set("Amount", amountStr)                // 4
	form.Set("PhoneNumber", phoneNumber)         // 5
	form.Set("Description", description)         // 6
	form.Set("ChannelCode", channelCode)         // 7
	form.Set("Currency", c.cfg.Currency)         // 8
	form.Set("LangCode", langCode)               // 9
	form.Set("RouteNo", routeNoStr)              // 10
	form.Set("IPAddress", ipAddress)             // 11
	form.Set("ApiKey", c.cfg.APIKey)             // 12
	form.Set("TokenFlag", tokenFlag)             // 13
	form.Set("CreditToken", creditToken)         // 14
	form.Set("CreditMonth", creditMonth)         // 15
	form.Set("ShopID", shopID)                   // 16
	form.Set("ProductImageUrl", productImageUrl) // 17
	form.Set("CustEmail", custEmail)             // 18
	form.Set("CardType", cardType)               // 19
	form.Set("CustName", custName)               // 20
	form.Set("CheckSum", checksum)               // 21
	form.Set("CallbackUrl", c.cfg.CallbackURL)   // not in checksum
	form.Set("ReturnUrl", c.cfg.ReturnURL)       // not in checksum

	log.Printf("chillpay InitPayment order_no=%s return_url=%s callback_url=%s amount=%s channel=%s",
		orderNo, c.cfg.ReturnURL, c.cfg.CallbackURL, amountStr, channelCode)

	req, err := http.NewRequest(http.MethodPost, c.cfg.BaseURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", "", fmt.Errorf("chillpay build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("chillpay request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", fmt.Errorf("chillpay read response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("chillpay status %d: %s", resp.StatusCode, string(body))
	}

	var result initResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return "", "", fmt.Errorf("chillpay parse response: %w", err)
	}
	if result.Status != 0 {
		return "", "", fmt.Errorf("chillpay error (code %d): %s", result.Code, result.Message)
	}
	return result.PaymentUrl, fmt.Sprintf("%d", result.TransactionId), nil
}

// InquiryPaymentStatus calls the ChillPay PaymentStatus API.
// CheckSum = MD5(MerchantCode + TransactionId + ApiKey + MD5SecretKey)
func (c *Client) InquiryPaymentStatus(transactionID string) (*StatusResponse, error) {
	checksumRaw := c.cfg.MerchantCode + transactionID + c.cfg.APIKey + c.cfg.MD5Key
	checksum := md5Hex(checksumRaw)

	form := url.Values{}
	form.Set("MerchantCode", c.cfg.MerchantCode)
	form.Set("TransactionId", transactionID)
	form.Set("ApiKey", c.cfg.APIKey)
	form.Set("CheckSum", checksum)

	endpoint := statusEndpoint(c.cfg.BaseURL)
	req, err := http.NewRequest(http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("chillpay build status request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Cache-Control", "no-cache")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("chillpay status request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("chillpay read status response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("chillpay status HTTP %d: %s", resp.StatusCode, string(body))
	}

	var result StatusResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("chillpay parse status response: %w", err)
	}
	return &result, nil
}

// Ping validates credentials by calling PaymentStatus with a probe transaction id.
func (c *Client) Ping() error {
	if strings.TrimSpace(c.cfg.MerchantCode) == "" ||
		strings.TrimSpace(c.cfg.APIKey) == "" ||
		strings.TrimSpace(c.cfg.MD5Key) == "" {
		return fmt.Errorf("chillpay credentials incomplete")
	}
	if strings.TrimSpace(c.cfg.BaseURL) == "" {
		return fmt.Errorf("chillpay base_url is required")
	}
	_, err := c.InquiryPaymentStatus("0")
	if err != nil {
		// ChillPay may return business errors for unknown txn — HTTP 200 with JSON still means auth OK.
		if strings.Contains(err.Error(), "chillpay status HTTP") {
			return err
		}
	}
	return nil
}

// VerifyCallback validates the MD5 checksum on an incoming ChillPay callback.
// CheckSum = MD5(TransactionId + Amount + OrderNo + CustomerId + BankCode +
// PaymentDate + PaymentStatus + BankRefCode + CurrentDate + CurrentTime +
// PaymentDescription + CreditCardToken + Currency + CustomerName + MD5SecretKey)
func (c *Client) VerifyCallback(form CallbackForm) bool {
	raw := form.TransactionId +
		form.Amount +
		form.OrderNo +
		form.CustomerId +
		form.BankCode +
		form.PaymentDate +
		form.PaymentStatus +
		form.BankRefCode +
		form.CurrentDate +
		form.CurrentTime +
		form.PaymentDescription +
		form.CreditCardToken +
		form.Currency +
		form.CustomerName +
		c.cfg.MD5Key
	expected := md5Hex(raw)
	return strings.EqualFold(expected, form.CheckSum)
}

func md5Hex(raw string) string {
	hash := md5.Sum([]byte(raw))
	return hex.EncodeToString(hash[:])
}

func statusEndpoint(baseURL string) string {
	endpoint := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	// Match harvest-core: strip trailing /Payment or /Payment/
	endpoint = strings.TrimSuffix(endpoint, "/Payment")
	endpoint = strings.TrimSuffix(endpoint, "/Payment/")
	endpoint = strings.TrimSuffix(endpoint, "/api/v2")
	return endpoint + "/api/v2/PaymentStatus/"
}

// --- Field sanitizers (ChillPay Table 2.2 constraints; harvest passes caller values as-is,
// but Monti order_no / OAuth emails need guards against codes 1006 / 2032). ---

const (
	maxOrderNoLen     = 20
	maxCustomerIDLen  = 100
	maxPhoneLen       = 10
	maxCustNameLen    = 50
)

// SanitizeOrderNo enforces max 20 alphanumeric characters (A–Z a–z 0–9).
func SanitizeOrderNo(raw string) string {
	return keepAlnum(raw, maxOrderNoLen)
}

// SanitizeCustomerID strips special characters ChillPay rejects (e.g. _ -).
func SanitizeCustomerID(raw string) string {
	return keepAlnum(raw, maxCustomerIDLen)
}

// SanitizePhone keeps digits only, max 10 (Thai mobile style).
func SanitizePhone(raw string) string {
	var b strings.Builder
	for _, r := range raw {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
		if b.Len() >= maxPhoneLen {
			break
		}
	}
	return b.String()
}

// SanitizeCustName builds a ChillPay-safe payer name (code 2032 rejects emails / symbols).
// Allows letters (incl. Thai), digits, and single internal spaces.
// Falls back to the local-part of email, then "Customer".
func SanitizeCustName(name, email string) string {
	if out := cleanPersonName(name, maxCustNameLen); out != "" {
		return out
	}
	local := email
	if i := strings.Index(email, "@"); i >= 0 {
		local = email[:i]
	}
	local = strings.ReplaceAll(local, ".", " ")
	local = strings.ReplaceAll(local, "_", " ")
	local = strings.ReplaceAll(local, "-", " ")
	if out := cleanPersonName(local, maxCustNameLen); out != "" {
		return out
	}
	return "Customer"
}

func cleanPersonName(raw string, max int) string {
	raw = strings.TrimSpace(raw)
	if raw == "" || strings.Contains(raw, "@") {
		return ""
	}
	var b strings.Builder
	prevSpace := false
	for _, r := range raw {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
			b.WriteRune(r)
			prevSpace = false
		case r >= 0x0E00 && r <= 0x0E7F: // Thai
			b.WriteRune(r)
			prevSpace = false
		case r == ' ' || r == '\t':
			if b.Len() > 0 && !prevSpace {
				b.WriteByte(' ')
				prevSpace = true
			}
		}
		if b.Len() >= max {
			break
		}
	}
	return strings.TrimSpace(b.String())
}

func keepAlnum(raw string, max int) string {
	var b strings.Builder
	for _, r := range strings.TrimSpace(raw) {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
			b.WriteRune(r)
		}
		if b.Len() >= max {
			break
		}
	}
	return b.String()
}

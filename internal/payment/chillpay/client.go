package chillpay

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
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
type CallbackForm struct {
	OrderNo            string `json:"OrderNo" form:"OrderNo"`
	Amount             string `json:"Amount" form:"Amount"`
	TransactionId      string `json:"TransactionId" form:"TransactionId"`
	CustomerId         string `json:"CustomerId" form:"CustomerId"`
	CustomerName       string `json:"CustomerName" form:"CustomerName"`
	BankCode           string `json:"BankCode" form:"BankCode"`
	PaymentDate        string `json:"PaymentDate" form:"PaymentDate"`
	PaymentStatus      string `json:"PaymentStatus" form:"PaymentStatus"`
	PaymentDescription string `json:"PaymentDescription" form:"PaymentDescription"`
	BankRefCode        string `json:"BankRefCode" form:"BankRefCode"`
	Currency           string `json:"Currency" form:"Currency"`
	CreditCardToken    string `json:"CreditCardToken" form:"CreditCardToken"`
	CurrentDate        string `json:"CurrentDate" form:"CurrentDate"`
	CurrentTime        string `json:"CurrentTime" form:"CurrentTime"`
	CheckSum           string `json:"CheckSum" form:"CheckSum"`
}

// RequestInfo holds caller-supplied context for InitPayment (Sprint 9).
type RequestInfo struct {
	OrderNo     string
	CustomerID  string
	Amount      float64
	Description string
	ChannelCode string
	IPAddress   string
	LangCode    string
	PhoneNumber string
	CustEmail   string
	CustName    string
}

type initResponse struct {
	Status        int    `json:"Status"`
	Code          int    `json:"Code"`
	Message       string `json:"Message"`
	TransactionId int64  `json:"TransactionId"`
	PaymentUrl    string `json:"PaymentUrl"`
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

func (c *Client) InitPayment(info RequestInfo) (paymentURL string, txnID string, err error) {
	amountInt := int(info.Amount * 100)
	amountStr := strconv.Itoa(amountInt)

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

	checksumRaw := c.cfg.MerchantCode +
		info.OrderNo +
		info.CustomerID +
		amountStr +
		info.PhoneNumber +
		info.Description +
		channelCode +
		c.cfg.Currency +
		langCode +
		routeNoStr +
		ipAddress +
		c.cfg.APIKey +
		"N" + "" + "" + "" + "" +
		info.CustEmail + "" +
		info.CustName +
		c.cfg.MD5Key
	checksum := md5Hex(checksumRaw)

	form := url.Values{}
	form.Set("MerchantCode", c.cfg.MerchantCode)
	form.Set("OrderNo", info.OrderNo)
	form.Set("CustomerId", info.CustomerID)
	form.Set("Amount", amountStr)
	form.Set("PhoneNumber", info.PhoneNumber)
	form.Set("Description", info.Description)
	form.Set("ChannelCode", channelCode)
	form.Set("Currency", c.cfg.Currency)
	form.Set("LangCode", langCode)
	form.Set("RouteNo", routeNoStr)
	form.Set("IPAddress", ipAddress)
	form.Set("ApiKey", c.cfg.APIKey)
	form.Set("TokenFlag", "N")
	form.Set("CustEmail", info.CustEmail)
	form.Set("CustName", info.CustName)
	form.Set("CheckSum", checksum)
	form.Set("CallbackUrl", c.cfg.CallbackURL)
	form.Set("ReturnUrl", c.cfg.ReturnURL)

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
	endpoint = strings.TrimSuffix(endpoint, "/Payment")
	endpoint = strings.TrimSuffix(endpoint, "/api/v2")
	return endpoint + "/api/v2/PaymentStatus/"
}
package resend

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	apiKey string
	from   string
	http   *http.Client
}

func New(apiKey, from string) *Client {
	return &Client{
		apiKey: strings.TrimSpace(apiKey),
		from:   strings.TrimSpace(from),
		http:   &http.Client{Timeout: 15 * time.Second},
	}
}

func (c *Client) Enabled() bool {
	return c != nil && c.apiKey != "" && c.from != ""
}

type sendRequest struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	HTML    string   `json:"html"`
}

func (c *Client) Send(ctx context.Context, to, subject, html string) error {
	if !c.Enabled() {
		return fmt.Errorf("resend is not configured")
	}
	to = strings.TrimSpace(to)
	if to == "" {
		return fmt.Errorf("recipient is required")
	}
	body, err := json.Marshal(sendRequest{
		From:    c.from,
		To:      []string{to},
		Subject: subject,
		HTML:    html,
	})
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.resend.com/emails", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	res, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode >= 300 {
		raw, _ := io.ReadAll(io.LimitReader(res.Body, 4096))
		return fmt.Errorf("resend status %d: %s", res.StatusCode, strings.TrimSpace(string(raw)))
	}
	return nil
}

func VerificationEmail(verifyURL, displayName string) (string, string) {
	name := strings.TrimSpace(displayName)
	if name == "" {
		name = "there"
	}
	subject := "Verify your Monti workspace email"
	html := fmt.Sprintf(`<p>Hi %s,</p>
<p>Thanks for signing up for Monti. Please verify your email to activate login:</p>
<p><a href="%s">Verify email</a></p>
<p>This link expires in 24 hours. If you did not request this, you can ignore this message.</p>`, name, verifyURL)
	return subject, html
}

func KYCApprovedEmail(loginURL, billingURL, companyName string) (string, string) {
	company := strings.TrimSpace(companyName)
	if company == "" {
		company = "your workspace"
	}
	if strings.TrimSpace(billingURL) == "" {
		billingURL = loginURL
	}
	subject := "Your Monti workspace is now active"
	html := fmt.Sprintf(`<p>Good news — <strong>%s</strong> has passed platform verification (KYC approved).</p>
<p>Your tenant account is <strong>active</strong>. You can sign in and purchase a package to start using Monti AI agents.</p>
<p><a href="%s">Sign in to Monti</a></p>
<p><a href="%s">Go to Billing &amp; packages</a></p>
<p style="color:#666;font-size:13px">If the buttons do not work, open: %s</p>`, company, loginURL, billingURL, loginURL)
	return subject, html
}

func KYCRejectedEmail(backofficeURL, companyName, reason string) (string, string) {
	company := strings.TrimSpace(companyName)
	if company == "" {
		company = "your workspace"
	}
	reason = strings.TrimSpace(reason)
	if reason == "" {
		reason = "Additional information is required."
	}
	subject := "Action required — Monti KYC review update"
	html := fmt.Sprintf(`<p>We reviewed the verification package for <strong>%s</strong> and could not approve it yet.</p>
<p><strong>Reason:</strong> %s</p>
<p>Please update your documents in the tenant backoffice and submit again for review.</p>
<p><a href="%s">Open tenant backoffice</a></p>`, company, reason, backofficeURL)
	return subject, html
}

func RegistrationCompleteEmail(loginURL, tenantID, displayName string) (string, string) {
	name := strings.TrimSpace(displayName)
	if name == "" {
		name = "there"
	}
	subject := "Your Monti workspace registration is complete"
	html := fmt.Sprintf(`<p>Hi %s,</p>
<p>Your workspace <strong>%s</strong> is registered and ready to sign in.</p>
<p>Your account is <strong>pending platform verification (KYC)</strong>. You can log in now and submit business details in the tenant backoffice while we review your application.</p>
<p><a href="%s">Sign in to Monti</a></p>`, name, tenantID, loginURL)
	return subject, html
}
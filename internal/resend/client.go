package resend

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html"
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

// BrandLogoURL returns the publicly reachable Monti logo used by email clients.
// The tenant portal serves this asset at /tenant/images/monti-logo.png.
func BrandLogoURL(publicBaseURL string) string {
	base := strings.TrimRight(strings.TrimSpace(publicBaseURL), "/")
	if base == "" {
		return "/tenant/images/monti-logo.png"
	}
	return base + "/tenant/images/monti-logo.png"
}

func emailLayout(logoURL, preheader, content string) string {
	logo := ""
	if strings.TrimSpace(logoURL) != "" {
		logo = fmt.Sprintf(`<img src="%s" width="58" height="58" alt="Monti" style="display:block;width:58px;height:58px;border:0;border-radius:14px;object-fit:cover;" />`, html.EscapeString(logoURL))
	}
	return fmt.Sprintf(`<!doctype html>
<html lang="en">
<head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1.0"></head>
<body style="margin:0;padding:0;background:#f2f6fb;color:#182338;font-family:Arial,Helvetica,sans-serif;">
<div style="display:none;max-height:0;overflow:hidden;opacity:0;color:transparent;">%s</div>
<table role="presentation" width="100%%" cellpadding="0" cellspacing="0" border="0" style="background:#f2f6fb;">
  <tr><td align="center" style="padding:28px 12px;">
    <table role="presentation" width="100%%" cellpadding="0" cellspacing="0" border="0" style="max-width:640px;background:#ffffff;border:1px solid #e4eaf3;border-radius:18px;overflow:hidden;">
      <tr><td style="padding:24px 30px;background:#101b2d;border-bottom:4px solid #2f80ed;">
        <table role="presentation" cellpadding="0" cellspacing="0" border="0"><tr>
          <td style="padding-right:14px;vertical-align:middle;">%s</td>
          <td style="vertical-align:middle;color:#ffffff;">
            <div style="font-size:27px;line-height:30px;font-weight:700;letter-spacing:-.4px;">Monti</div>
            <div style="font-size:12px;line-height:18px;color:#a9c7ff;letter-spacing:.4px;">AI customer operations</div>
          </td>
        </tr></table>
      </td></tr>
      <tr><td style="padding:38px 42px 32px;">%s</td></tr>
      <tr><td style="padding:20px 42px 26px;border-top:1px solid #e8edf5;background:#fbfcfe;color:#68758b;font-size:12px;line-height:19px;">
        <strong style="color:#34435b;">Security reminder</strong><br>
        Monti will never ask you to disclose your password, API key, or payment details by email.<br><br>
        © Monti · AI customer operations
      </td></tr>
    </table>
  </td></tr>
</table>
</body></html>`, html.EscapeString(preheader), logo, content)
}

func emailButton(label, targetURL string) string {
	return fmt.Sprintf(`<a href="%s" style="display:inline-block;padding:13px 22px;border-radius:9px;background:#2f80ed;color:#ffffff;text-decoration:none;font-size:15px;font-weight:700;">%s</a>`, html.EscapeString(targetURL), html.EscapeString(label))
}

func VerificationEmail(verifyURL, displayName, logoURL string) (string, string) {
	name := strings.TrimSpace(displayName)
	if name == "" {
		name = "there"
	}
	subject := "Verify your Monti workspace email"
	content := fmt.Sprintf(`<h1 style="margin:0 0 12px;color:#15223a;font-size:30px;line-height:38px;">Verify your email</h1>
<p style="margin:0 0 22px;color:#59677d;font-size:16px;line-height:25px;">Hi %s,</p>
<p style="margin:0 0 26px;color:#59677d;font-size:16px;line-height:25px;">Thanks for signing up for Monti. Confirm your email to activate your workspace login.</p>
<p style="margin:0 0 28px;">%s</p>
<p style="margin:0;color:#7a8799;font-size:13px;line-height:21px;">This verification link expires in 24 hours. If you did not request this, you can safely ignore this message.</p>`, html.EscapeString(name), emailButton("Verify email", verifyURL))
	return subject, emailLayout(logoURL, "Confirm your Monti workspace email to activate login.", content)
}

func KYCApprovedEmail(loginURL, billingURL, companyName, logoURL string) (string, string) {
	company := strings.TrimSpace(companyName)
	if company == "" {
		company = "your workspace"
	}
	if strings.TrimSpace(billingURL) == "" {
		billingURL = loginURL
	}
	subject := "Your Monti workspace is now active"
	content := fmt.Sprintf(`<h1 style="margin:0 0 14px;color:#15223a;font-size:30px;line-height:38px;">Your workspace is active</h1>
<p style="margin:0 0 18px;color:#59677d;font-size:16px;line-height:25px;">Good news — <strong style="color:#15223a;">%s</strong> has passed platform verification.</p>
<p style="margin:0 0 26px;color:#59677d;font-size:16px;line-height:25px;">Your tenant account is active. Sign in to continue, or open billing to choose a package.</p>
<p style="margin:0 0 12px;">%s</p>
<p style="margin:0 0 25px;">%s</p>
<p style="margin:0;color:#7a8799;font-size:13px;line-height:21px;">If a button does not work, open this address: %s</p>`, html.EscapeString(company), emailButton("Sign in to Monti", loginURL), emailButton("Billing & packages", billingURL), html.EscapeString(loginURL))
	return subject, emailLayout(logoURL, "Your Monti workspace has passed platform verification.", content)
}

func KYCRejectedEmail(backofficeURL, companyName, reason, logoURL string) (string, string) {
	company := strings.TrimSpace(companyName)
	if company == "" {
		company = "your workspace"
	}
	reason = strings.TrimSpace(reason)
	if reason == "" {
		reason = "Additional information is required."
	}
	subject := "Action required — Monti KYC review update"
	content := fmt.Sprintf(`<h1 style="margin:0 0 14px;color:#15223a;font-size:30px;line-height:38px;">Action needed</h1>
<p style="margin:0 0 18px;color:#59677d;font-size:16px;line-height:25px;">We reviewed the verification package for <strong style="color:#15223a;">%s</strong> and could not approve it yet.</p>
<div style="margin:0 0 24px;padding:16px 18px;border-left:4px solid #f0a23a;background:#fff7e8;color:#59677d;font-size:15px;line-height:23px;"><strong style="color:#15223a;">Reason:</strong> %s</div>
<p style="margin:0 0 26px;color:#59677d;font-size:16px;line-height:25px;">Update your documents in the tenant backoffice and submit again for review.</p>
<p style="margin:0;">%s</p>`, html.EscapeString(company), html.EscapeString(reason), emailButton("Open tenant backoffice", backofficeURL))
	return subject, emailLayout(logoURL, "Your Monti verification review needs an update.", content)
}

func RegistrationCompleteEmail(loginURL, tenantID, displayName, logoURL string) (string, string) {
	name := strings.TrimSpace(displayName)
	if name == "" {
		name = "there"
	}
	subject := "Your Monti workspace registration is complete"
	content := fmt.Sprintf(`<h1 style="margin:0 0 14px;color:#15223a;font-size:30px;line-height:38px;">Registration complete</h1>
<p style="margin:0 0 18px;color:#59677d;font-size:16px;line-height:25px;">Hi %s,</p>
<p style="margin:0 0 18px;color:#59677d;font-size:16px;line-height:25px;">Your workspace <strong style="color:#15223a;">%s</strong> is registered and ready to sign in.</p>
<p style="margin:0 0 26px;color:#59677d;font-size:16px;line-height:25px;">Your account is pending platform verification. Submit your business details in the tenant backoffice while we review your application.</p>
<p style="margin:0;">%s</p>`, html.EscapeString(name), html.EscapeString(tenantID), emailButton("Sign in to Monti", loginURL))
	return subject, emailLayout(logoURL, "Your Monti workspace registration is complete.", content)
}

func CustomerOTPEmail(code string, ttlSeconds int, logoURL string) (string, string) {
	minutes := ttlSeconds / 60
	if minutes < 1 {
		minutes = 1
	}
	subject := "Your Monti sign-in code"
	content := fmt.Sprintf(`<h1 style="margin:0 0 14px;color:#15223a;font-size:30px;line-height:38px;">Your sign-in code</h1>
<p style="margin:0 0 22px;color:#59677d;font-size:16px;line-height:25px;">To continue to Monti, use this one-time passcode (OTP):</p>
<div style="margin:0 0 24px;padding:20px 18px;border:1px solid #cfe0fa;border-radius:12px;background:#eef5ff;color:#15223a;text-align:center;font-size:36px;line-height:42px;font-weight:700;letter-spacing:8px;">%s</div>
<p style="margin:0;color:#7a8799;font-size:13px;line-height:21px;">This code expires in %d minutes. If you did not request this code, you can ignore this message.</p>`, html.EscapeString(strings.TrimSpace(code)), minutes)
	return subject, emailLayout(logoURL, "Use this one-time passcode to continue to Monti.", content)
}

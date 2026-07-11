package main

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/libra/monti-jarvis/internal/resend"
	"github.com/libra/monti-jarvis/internal/store"
)

func (s *server) sendVerificationEmail(ctx context.Context, user store.AuthUser, rawToken string) {
	if s.mailer == nil || !s.mailer.Enabled() {
		log.Printf("mailer warning: verification email skipped for %s", user.Email)
		return
	}
	base := strings.TrimRight(s.cfg.PublicBaseURL, "/")
	verifyURL := base + "/tenant/register/verify?token=" + rawToken
	subject, html := resend.VerificationEmail(verifyURL, user.DisplayName)
	mailCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 20*time.Second)
	defer cancel()
	if err := s.mailer.Send(mailCtx, user.Email, subject, html); err != nil {
		log.Printf("mailer warning: verification email: %v", err)
		return
	}
	log.Printf("mailer: verification email sent to %s", user.Email)
}

// sendKYCApprovedEmail notifies the tenant admin that KYC was approved and workspace is active.
// Returns (sent, toEmail, error). Best-effort: caller should not roll back approve on mail failure.
func (s *server) sendKYCApprovedEmail(ctx context.Context, result store.PlatformKYCDecisionResult) (sent bool, to string, err error) {
	to = strings.TrimSpace(result.AdminEmail)
	if to == "" && strings.TrimSpace(result.TenantID) != "" {
		if email, e := s.store.GetTenantAdminEmail(ctx, result.TenantID); e == nil {
			to = email
		}
	}
	if to == "" {
		log.Printf("mailer warning: KYC approved email skipped — no admin email for tenant %s", result.TenantID)
		return false, "", nil
	}
	if s.mailer == nil || !s.mailer.Enabled() {
		log.Printf("mailer warning: KYC approved email skipped (resend disabled) for %s", to)
		return false, to, nil
	}
	base := strings.TrimRight(s.cfg.PublicBaseURL, "/")
	loginURL := base + "/tenant/login"
	billingURL := base + "/tenant/billing"
	subject, html := resend.KYCApprovedEmail(loginURL, billingURL, result.CompanyName)
	mailCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 20*time.Second)
	defer cancel()
	if err := s.mailer.Send(mailCtx, to, subject, html); err != nil {
		log.Printf("mailer warning: KYC approved email to %s: %v", to, err)
		return false, to, err
	}
	log.Printf("mailer: KYC approved email sent to %s (tenant=%s)", to, result.TenantID)
	return true, to, nil
}

func (s *server) sendKYCRejectedEmail(ctx context.Context, result store.PlatformKYCDecisionResult) (sent bool, to string, err error) {
	to = strings.TrimSpace(result.AdminEmail)
	if to == "" && strings.TrimSpace(result.TenantID) != "" {
		if email, e := s.store.GetTenantAdminEmail(ctx, result.TenantID); e == nil {
			to = email
		}
	}
	if to == "" {
		log.Printf("mailer warning: KYC rejected email skipped — no admin email for tenant %s", result.TenantID)
		return false, "", nil
	}
	if s.mailer == nil || !s.mailer.Enabled() {
		log.Printf("mailer warning: KYC rejected email skipped (resend disabled) for %s", to)
		return false, to, nil
	}
	base := strings.TrimRight(s.cfg.PublicBaseURL, "/")
	backofficeURL := base + "/tenant/backoffice"
	subject, html := resend.KYCRejectedEmail(backofficeURL, result.CompanyName, result.RejectionReason)
	mailCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 20*time.Second)
	defer cancel()
	if err := s.mailer.Send(mailCtx, to, subject, html); err != nil {
		log.Printf("mailer warning: KYC rejected email to %s: %v", to, err)
		return false, to, err
	}
	log.Printf("mailer: KYC rejected email sent to %s (tenant=%s)", to, result.TenantID)
	return true, to, nil
}

func (s *server) sendRegistrationCompleteEmail(ctx context.Context, user store.AuthUser, tenantID string) {
	if s.mailer == nil || !s.mailer.Enabled() {
		log.Printf("mailer warning: welcome email skipped for %s", user.Email)
		return
	}
	base := strings.TrimRight(s.cfg.PublicBaseURL, "/")
	loginURL := base + "/tenant/login"
	subject, html := resend.RegistrationCompleteEmail(loginURL, tenantID, user.DisplayName)
	mailCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 20*time.Second)
	defer cancel()
	if err := s.mailer.Send(mailCtx, user.Email, subject, html); err != nil {
		log.Printf("mailer warning: welcome email: %v", err)
		return
	}
	log.Printf("mailer: welcome email sent to %s", user.Email)
}

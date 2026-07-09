package main

import (
	"context"
	"log"
	"strings"

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
	if err := s.mailer.Send(ctx, user.Email, subject, html); err != nil {
		log.Printf("mailer warning: verification email: %v", err)
	}
}

func (s *server) sendRegistrationCompleteEmail(ctx context.Context, user store.AuthUser, tenantID string) {
	if s.mailer == nil || !s.mailer.Enabled() {
		log.Printf("mailer warning: welcome email skipped for %s", user.Email)
		return
	}
	base := strings.TrimRight(s.cfg.PublicBaseURL, "/")
	loginURL := base + "/tenant/login"
	subject, html := resend.RegistrationCompleteEmail(loginURL, tenantID, user.DisplayName)
	if err := s.mailer.Send(ctx, user.Email, subject, html); err != nil {
		log.Printf("mailer warning: welcome email: %v", err)
	}
}
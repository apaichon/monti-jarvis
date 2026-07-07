package auth

import (
	"context"
	"net"
	"net/http"
	"strings"
)

type requestMetaKey struct{}

type RequestMeta struct {
	IP        string
	UserAgent string
}

func WithRequestMeta(ctx context.Context, r *http.Request) context.Context {
	return context.WithValue(ctx, requestMetaKey{}, RequestMeta{
		IP:        clientIP(r),
		UserAgent: strings.TrimSpace(r.UserAgent()),
	})
}

func RequestMetaFrom(ctx context.Context) RequestMeta {
	if m, ok := ctx.Value(requestMetaKey{}).(RequestMeta); ok {
		return m
	}
	return RequestMeta{}
}

func clientIP(r *http.Request) string {
	if xff := strings.TrimSpace(r.Header.Get("X-Forwarded-For")); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return strings.TrimSpace(r.RemoteAddr)
	}
	return host
}
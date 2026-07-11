package quota

import (
	"errors"
	"fmt"
)

var (
	ErrLimitExceeded    = errors.New("quota exceeded")
	ErrRateLimited      = errors.New("rate limited")
	ErrFeatureDisabled  = errors.New("feature disabled")
	ErrNoEntitlement    = errors.New("no entitlement")
	ErrQuotaDisabled    = errors.New("quota disabled")
)

// Error is a structured quota/rate/feature failure for HTTP mapping.
type Error struct {
	Code      string // quota_exceeded | rate_limited | feature_disabled | no_entitlement
	Dimension string
	Limit     int
	Usage     int
	Message   string
	cause     error
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	if e.Message != "" {
		return e.Message
	}
	if e.cause != nil {
		return e.cause.Error()
	}
	return e.Code
}

func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.cause
}

func limitExceeded(dimension string, limit, usage int) *Error {
	return &Error{
		Code:      "quota_exceeded",
		Dimension: dimension,
		Limit:     limit,
		Usage:     usage,
		Message:   fmt.Sprintf("%s limit exceeded (%d/%d)", dimension, usage, limit),
		cause:     ErrLimitExceeded,
	}
}

func rateLimited(bucket string, limit, usage int) *Error {
	return &Error{
		Code:      "rate_limited",
		Dimension: bucket,
		Limit:     limit,
		Usage:     usage,
		Message:   fmt.Sprintf("rate limit exceeded for %s", bucket),
		cause:     ErrRateLimited,
	}
}

func featureDisabled(flag string) *Error {
	return &Error{
		Code:      "feature_disabled",
		Dimension: flag,
		Message:   fmt.Sprintf("feature %s is disabled for this package", flag),
		cause:     ErrFeatureDisabled,
	}
}

func noEntitlement() *Error {
	return &Error{
		Code:    "no_entitlement",
		Message: "tenant has no active package entitlement",
		cause:   ErrNoEntitlement,
	}
}

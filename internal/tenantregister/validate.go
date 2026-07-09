package tenantregister

import (
	"errors"
	"regexp"
	"strings"
)

var (
	ErrInvalidInput      = errors.New("invalid registration input")
	ErrSlugTaken         = errors.New("slug already taken")
	ErrEmailRegistered   = errors.New("email already registered")
	ErrReservedSlug      = errors.New("slug is reserved")
)

var (
	slugPattern = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]{1,30}[a-z0-9])?$`)
	emailPattern = regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)
	reservedSlugs = map[string]struct{}{
		"demo": {}, "admin": {}, "platform": {}, "api": {}, "www": {},
	}
)

type Input struct {
	CompanyName      string
	Slug             string
	AdminEmail       string
	AdminPassword    string
	AdminDisplayName string
}

func Normalize(in Input) Input {
	return Input{
		CompanyName:      strings.TrimSpace(in.CompanyName),
		Slug:             strings.TrimSpace(strings.ToLower(in.Slug)),
		AdminEmail:       strings.TrimSpace(strings.ToLower(in.AdminEmail)),
		AdminPassword:    in.AdminPassword,
		AdminDisplayName: strings.TrimSpace(in.AdminDisplayName),
	}
}

func Validate(in Input) error {
	in = Normalize(in)
	if err := validateProfile(in); err != nil {
		return err
	}
	if len(in.AdminPassword) < 8 {
		return ErrInvalidInput
	}
	return nil
}

func ValidateProfile(in Input) error {
	return validateProfile(Normalize(in))
}

func validateProfile(in Input) error {
	if len(in.CompanyName) < 2 || len(in.CompanyName) > 120 {
		return ErrInvalidInput
	}
	if !slugPattern.MatchString(in.Slug) {
		return ErrInvalidInput
	}
	if _, ok := reservedSlugs[in.Slug]; ok {
		return ErrReservedSlug
	}
	if !emailPattern.MatchString(in.AdminEmail) {
		return ErrInvalidInput
	}
	if len(in.AdminDisplayName) < 1 || len(in.AdminDisplayName) > 80 {
		return ErrInvalidInput
	}
	return nil
}
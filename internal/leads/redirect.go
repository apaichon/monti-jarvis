package leads

import (
	"net/url"
	"strings"
)

// Allowlisted query keys that may be preserved across conversion redirects.
var allowlistedQueryKeys = map[string]struct{}{
	"utm_source":   {},
	"utm_medium":   {},
	"utm_campaign": {},
	"utm_content":  {},
	"utm_term":     {},
	"ref":          {},
	"package_id":   {},
	"lead_id":      {},
	"lang":         {},
}

// Allowlisted relative path prefixes for product-web CTAs.
var allowlistedPathPrefixes = []string{
	"/",
	"/product",
	"/tenant/register",
	"/tenant/login",
	"/tenant/billing",
}

// FilterQuery keeps only allowlisted attribution keys (values trimmed, bounded).
func FilterQuery(raw url.Values) url.Values {
	out := url.Values{}
	if raw == nil {
		return out
	}
	for key, values := range raw {
		k := strings.ToLower(strings.TrimSpace(key))
		if _, ok := allowlistedQueryKeys[k]; !ok {
			continue
		}
		for _, v := range values {
			v = strings.TrimSpace(v)
			if v == "" {
				continue
			}
			if len(v) > 200 {
				v = v[:200]
			}
			out.Add(k, v)
		}
	}
	return out
}

// SafeRelativeRedirect returns a Monti-relative path+query or ErrInvalidRedirect.
// Rejects absolute URLs, protocol-relative hosts, javascript:, data:, and unknown paths.
func SafeRelativeRedirect(target string, query url.Values) (string, error) {
	target = strings.TrimSpace(target)
	if target == "" {
		return "", ErrInvalidRedirect
	}
	lower := strings.ToLower(target)
	if strings.HasPrefix(lower, "http://") ||
		strings.HasPrefix(lower, "https://") ||
		strings.HasPrefix(lower, "//") ||
		strings.HasPrefix(lower, "javascript:") ||
		strings.HasPrefix(lower, "data:") ||
		strings.Contains(target, "\\") {
		return "", ErrInvalidRedirect
	}
	// Strip any accidental scheme-like prefix leftovers.
	if strings.Contains(target, "://") {
		return "", ErrInvalidRedirect
	}

	u, err := url.Parse(target)
	if err != nil {
		return "", ErrInvalidRedirect
	}
	if u.IsAbs() || u.Host != "" || u.Scheme != "" || u.User != nil {
		return "", ErrInvalidRedirect
	}
	path := u.Path
	if path == "" {
		path = "/"
	}
	if !strings.HasPrefix(path, "/") {
		return "", ErrInvalidRedirect
	}
	if !isAllowlistedPath(path) {
		return "", ErrInvalidRedirect
	}

	// Merge path query with caller query, then filter.
	merged := u.Query()
	for k, vs := range query {
		for _, v := range vs {
			merged.Add(k, v)
		}
	}
	filtered := FilterQuery(merged)
	out := path
	if enc := filtered.Encode(); enc != "" {
		out += "?" + enc
	}
	if u.Fragment != "" {
		// Drop fragments for open-redirect safety.
	}
	return out, nil
}

func isAllowlistedPath(path string) bool {
	if path == "/" {
		return true
	}
	// Exact or under /product/*
	if path == "/product" || strings.HasPrefix(path, "/product/") {
		return true
	}
	for _, p := range []string{"/tenant/register", "/tenant/login", "/tenant/billing"} {
		if path == p || strings.HasPrefix(path, p+"/") {
			return true
		}
	}
	return false
}

// IsAllowlistedQueryKey reports whether key may appear on conversion links.
func IsAllowlistedQueryKey(key string) bool {
	_, ok := allowlistedQueryKeys[strings.ToLower(strings.TrimSpace(key))]
	return ok
}

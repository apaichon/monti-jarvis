package customerimport

import (
	"encoding/csv"
	"fmt"
	"io"
	"strings"

	"github.com/libra/monti-jarvis/internal/store"
)

var allowedHeaders = map[string]bool{
	"display_name": true,
	"email":        true,
	"phone":        true,
	"locale":       true,
	"tier_slug":    true,
	"group_slugs":  true,
	"source":       true,
	"external_id":  true,
}

type Row struct {
	Number      int
	DisplayName string
	Email       string
	Phone       string
	Locale      string
	TierSlug    string
	GroupSlugs  []string
	Source      string
	ExternalID  string
}

type RowError struct {
	Row     int    `json:"row"`
	Field   string `json:"field"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

type Result struct {
	Rows   []Row
	Errors []RowError
	Total  int
}

func Parse(reader io.Reader, maxRows int) (Result, error) {
	if maxRows <= 0 {
		maxRows = 5000
	}
	r := csv.NewReader(reader)
	r.FieldsPerRecord = -1
	r.TrimLeadingSpace = true
	header, err := r.Read()
	if err != nil {
		return Result{}, fmt.Errorf("read CSV header: %w", err)
	}
	indexes := map[string]int{}
	for i, raw := range header {
		name := strings.ToLower(strings.TrimSpace(strings.TrimPrefix(raw, "\ufeff")))
		if !allowedHeaders[name] {
			return Result{}, fmt.Errorf("unknown CSV column %q", name)
		}
		if _, exists := indexes[name]; exists {
			return Result{}, fmt.Errorf("duplicate CSV column %q", name)
		}
		indexes[name] = i
	}
	if _, ok := indexes["display_name"]; !ok {
		return Result{}, fmt.Errorf("display_name column is required")
	}
	if _, email := indexes["email"]; !email {
		if _, external := indexes["external_id"]; !external {
			return Result{}, fmt.Errorf("email or external_id column is required")
		}
	}
	result := Result{Rows: []Row{}, Errors: []RowError{}}
	for rowNumber := 2; ; rowNumber++ {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return Result{}, fmt.Errorf("row %d: %w", rowNumber, err)
		}
		result.Total++
		if result.Total > maxRows {
			return Result{}, fmt.Errorf("CSV exceeds maximum %d rows", maxRows)
		}
		value := func(name string) string {
			i, ok := indexes[name]
			if !ok || i >= len(record) {
				return ""
			}
			return strings.TrimSpace(record[i])
		}
		row := Row{
			Number:      rowNumber,
			DisplayName: value("display_name"),
			Email:       value("email"),
			Phone:       value("phone"),
			Locale:      strings.ToLower(value("locale")),
			TierSlug:    strings.ToLower(value("tier_slug")),
			Source:      value("source"),
			ExternalID:  value("external_id"),
		}
		for _, slug := range strings.Split(value("group_slugs"), "|") {
			if slug = strings.ToLower(strings.TrimSpace(slug)); slug != "" {
				row.GroupSlugs = append(row.GroupSlugs, slug)
			}
		}
		var rowErrors []RowError
		if row.DisplayName == "" {
			rowErrors = append(rowErrors, RowError{Row: rowNumber, Field: "display_name", Code: "required", Message: "Display name is required"})
		}
		if row.Email == "" && row.ExternalID == "" {
			rowErrors = append(rowErrors, RowError{Row: rowNumber, Field: "email", Code: "identity_required", Message: "Email or external ID is required"})
		} else if row.Email != "" {
			if _, err := store.NormalizeCustomerEmail(row.Email); err != nil {
				rowErrors = append(rowErrors, RowError{Row: rowNumber, Field: "email", Code: "invalid_email", Message: "Invalid email"})
			}
		}
		if row.Locale != "" && row.Locale != "en" && row.Locale != "th" {
			rowErrors = append(rowErrors, RowError{Row: rowNumber, Field: "locale", Code: "invalid_locale", Message: "Locale must be en or th"})
		}
		if row.Source != "" {
			if _, err := store.NormalizeCustomerSource(row.Source); err != nil {
				rowErrors = append(rowErrors, RowError{Row: rowNumber, Field: "source", Code: "invalid_source", Message: "Invalid source"})
			}
		}
		if len(rowErrors) > 0 {
			result.Errors = append(result.Errors, rowErrors...)
			continue
		}
		result.Rows = append(result.Rows, row)
	}
	return result, nil
}

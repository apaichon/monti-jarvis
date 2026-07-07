package packages

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
)

var (
	ErrUnknownField   = errors.New("unknown rules field")
	ErrRequiredField  = errors.New("missing required rules field")
	ErrInvalidType    = errors.New("invalid rules field type")
	ErrInvalidValue   = errors.New("rules field out of range")
	ErrInvalidSchema  = errors.New("invalid rules schema")
)

type FieldSpec struct {
	Type        string `json:"type"`
	Min         *int   `json:"min,omitempty"`
	Max         *int   `json:"max,omitempty"`
	Required    bool   `json:"required"`
	Default     any    `json:"default,omitempty"`
	Description string `json:"description,omitempty"`
}

func ValidateRules(fieldsJSON []byte, rules map[string]any) error {
	if len(rules) == 0 {
		return ErrRequiredField
	}
	var fields map[string]FieldSpec
	if err := json.Unmarshal(fieldsJSON, &fields); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidSchema, err)
	}
	if len(fields) == 0 {
		return ErrInvalidSchema
	}
	seen := make(map[string]struct{}, len(rules))
	for key, spec := range fields {
		val, ok := rules[key]
		if !ok {
			if spec.Required && spec.Default == nil {
				return fmt.Errorf("%w: %s", ErrRequiredField, key)
			}
			continue
		}
		seen[key] = struct{}{}
		if err := validateField(key, spec, val); err != nil {
			return err
		}
	}
	for key := range rules {
		if _, ok := fields[key]; !ok {
			return fmt.Errorf("%w: %s", ErrUnknownField, key)
		}
	}
	return nil
}

func validateField(key string, spec FieldSpec, val any) error {
	switch spec.Type {
	case "int":
		n, err := asInt(val)
		if err != nil {
			return fmt.Errorf("%w: %s", ErrInvalidType, key)
		}
		if spec.Min != nil && n < *spec.Min {
			return fmt.Errorf("%w: %s", ErrInvalidValue, key)
		}
		if spec.Max != nil && n > *spec.Max {
			return fmt.Errorf("%w: %s", ErrInvalidValue, key)
		}
	case "bool":
		if _, ok := val.(bool); !ok {
			return fmt.Errorf("%w: %s", ErrInvalidType, key)
		}
	default:
		return fmt.Errorf("%w: %s unsupported type %s", ErrInvalidType, key, spec.Type)
	}
	return nil
}

func asInt(val any) (int, error) {
	switch v := val.(type) {
	case int:
		return v, nil
	case int64:
		if v > math.MaxInt32 || v < math.MinInt32 {
			return 0, errors.New("overflow")
		}
		return int(v), nil
	case float64:
		if math.Trunc(v) != v {
			return 0, errors.New("not integer")
		}
		return int(v), nil
	case json.Number:
		i, err := v.Int64()
		if err != nil {
			return 0, err
		}
		return int(i), nil
	default:
		return 0, errors.New("not int")
	}
}
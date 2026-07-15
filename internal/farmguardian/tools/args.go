package tools

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

func int64FromArgs(args map[string]any, key string) (int64, error) {
	raw, ok := args[key]
	if !ok {
		return 0, fmt.Errorf("%s required", key)
	}
	switch v := raw.(type) {
	case float64:
		id := int64(v)
		if id <= 0 {
			return 0, fmt.Errorf("invalid %s", key)
		}
		return id, nil
	case int64:
		if v <= 0 {
			return 0, fmt.Errorf("invalid %s", key)
		}
		return v, nil
	case int:
		if v <= 0 {
			return 0, fmt.Errorf("invalid %s", key)
		}
		return int64(v), nil
	default:
		return 0, fmt.Errorf("invalid %s type", key)
	}
}

func optionalInt64FromArgs(args map[string]any, key string) (*int64, error) {
	raw, ok := args[key]
	if !ok || raw == nil {
		return nil, nil
	}
	n, err := int64FromArgs(map[string]any{key: raw}, key)
	if err != nil {
		return nil, err
	}
	return &n, nil
}

func stringFromArgs(args map[string]any, key string) (string, error) {
	raw, ok := args[key]
	if !ok {
		return "", fmt.Errorf("%s required", key)
	}
	s, ok := raw.(string)
	if !ok {
		return "", fmt.Errorf("%s must be a string", key)
	}
	s = trim(s)
	if s == "" {
		return "", fmt.Errorf("%s required", key)
	}
	return s, nil
}

func optionalStringFromArgs(args map[string]any, key string) (*string, error) {
	raw, ok := args[key]
	if !ok || raw == nil {
		return nil, nil
	}
	s, ok := raw.(string)
	if !ok {
		return nil, fmt.Errorf("%s must be a string", key)
	}
	s = trim(s)
	if s == "" {
		return nil, nil
	}
	return &s, nil
}

func optionalBoolFromArgs(args map[string]any, key string) (*bool, error) {
	raw, ok := args[key]
	if !ok || raw == nil {
		return nil, nil
	}
	switch v := raw.(type) {
	case bool:
		return &v, nil
	default:
		return nil, fmt.Errorf("%s must be a boolean", key)
	}
}

func optionalFloat64FromArgs(args map[string]any, key string) (*float64, error) {
	raw, ok := args[key]
	if !ok || raw == nil {
		return nil, nil
	}
	switch v := raw.(type) {
	case float64:
		return &v, nil
	case int:
		f := float64(v)
		return &f, nil
	case int64:
		f := float64(v)
		return &f, nil
	default:
		return nil, fmt.Errorf("%s must be a number", key)
	}
}

func trim(s string) string {
	return strings.TrimSpace(s)
}

func ensureFarmScope(entityFarmID, proposalFarmID int64) error {
	if proposalFarmID > 0 && entityFarmID != proposalFarmID {
		return errors.New("record is outside proposal farm scope")
	}
	return nil
}

func float64FromArgs(args map[string]any, key string) (float64, error) {
	raw, ok := args[key]
	if !ok {
		return 0, fmt.Errorf("%s required", key)
	}
	switch v := raw.(type) {
	case float64:
		return v, nil
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	default:
		return 0, fmt.Errorf("%s must be a number", key)
	}
}

func numericFromFloat64(v float64) (pgtype.Numeric, error) {
	var n pgtype.Numeric
	err := n.Scan(fmt.Sprintf("%g", v))
	return n, err
}

func optionalDateFromArgs(args map[string]any, key string) (pgtype.Date, bool, error) {
	s, err := optionalStringFromArgs(args, key)
	if err != nil || s == nil {
		return pgtype.Date{}, false, err
	}
	t, err := time.Parse("2006-01-02", strings.TrimSpace(*s))
	if err != nil {
		return pgtype.Date{}, false, fmt.Errorf("%s must be YYYY-MM-DD", key)
	}
	return pgtype.Date{Time: t, Valid: true}, true, nil
}

func dateFromArgs(args map[string]any, key string) (pgtype.Date, error) {
	s, err := stringFromArgs(args, key)
	if err != nil {
		return pgtype.Date{}, err
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return pgtype.Date{}, fmt.Errorf("%s must be YYYY-MM-DD", key)
	}
	return pgtype.Date{Time: t, Valid: true}, nil
}

func optionalObjectFromArgs(args map[string]any, key string) (map[string]any, bool, error) {
	raw, ok := args[key]
	if !ok || raw == nil {
		return nil, false, nil
	}
	m, ok := raw.(map[string]any)
	if !ok {
		return nil, false, fmt.Errorf("%s must be an object", key)
	}
	return m, true, nil
}

func objectFromArgs(args map[string]any, key string) (map[string]any, error) {
	m, ok, err := optionalObjectFromArgs(args, key)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("%s required", key)
	}
	return m, nil
}

func optionalMetaJSONFromArgs(args map[string]any, key string) ([]byte, error) {
	raw, ok := args[key]
	if !ok || raw == nil {
		return []byte("{}"), nil
	}
	switch v := raw.(type) {
	case map[string]any:
		b, err := json.Marshal(v)
		if err != nil {
			return nil, fmt.Errorf("%s must be valid JSON", key)
		}
		return b, nil
	case string:
		s := strings.TrimSpace(v)
		if s == "" {
			return []byte("{}"), nil
		}
		if !json.Valid([]byte(s)) {
			return nil, fmt.Errorf("%s must be valid JSON", key)
		}
		return []byte(s), nil
	default:
		return nil, fmt.Errorf("%s must be a JSON object", key)
	}
}

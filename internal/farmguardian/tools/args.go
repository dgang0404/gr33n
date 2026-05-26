package tools

import (
	"errors"
	"fmt"
	"strings"
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

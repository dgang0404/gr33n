package ai

import (
	"testing"
)

func TestParseTruthy(t *testing.T) {
	for _, tc := range []struct {
		in   string
		want bool
	}{
		{"", false},
		{"false", false},
		{"FALSE", false},
		{"0", false},
		{"no", false},
		{"off", false},
		{"true", true},
		{"1", true},
		{"yes", true},
		{"on", true},
		{"anything_else", true},
	} {
		if got := parseTruthy(tc.in); got != tc.want {
			t.Fatalf("parseTruthy(%q) = %v, want %v", tc.in, got, tc.want)
		}
	}
}

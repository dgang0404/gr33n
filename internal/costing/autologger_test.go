package costing

import (
	"strconv"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
)

func TestNumericFromFloatRoundTrip(t *testing.T) {
	n, err := numericFromFloat(12.75)
	if err != nil {
		t.Fatal(err)
	}
	got, ok := numericToFloat(n)
	if !ok || got != 12.75 {
		t.Fatalf("got %v ok=%v", got, ok)
	}
}

func TestNumericToFloat_Invalid(t *testing.T) {
	_, ok := numericToFloat(pgtype.Numeric{})
	if ok {
		t.Fatal("invalid numeric should not convert")
	}
}

func TestNumericFromFloat_MatchesHandlerPattern(t *testing.T) {
	var n pgtype.Numeric
	if err := n.Scan(strconv.FormatFloat(3.5, 'f', -1, 64)); err != nil {
		t.Fatal(err)
	}
	got, ok := numericToFloat(n)
	if !ok || got != 3.5 {
		t.Fatalf("got %v ok=%v", got, ok)
	}
}

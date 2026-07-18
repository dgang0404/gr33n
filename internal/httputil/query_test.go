package httputil

import (
	"net/http/httptest"
	"testing"
)

func TestParseLimitOffset(t *testing.T) {
	r := httptest.NewRequest("GET", "/?limit=200&offset=5", nil)
	limit, offset := ParseLimitOffset(r, 50, 100)
	if limit != 100 || offset != 5 {
		t.Fatalf("got limit=%d offset=%d", limit, offset)
	}
}

func TestParseLimitOffsetStrict(t *testing.T) {
	r := httptest.NewRequest("GET", "/?limit=0", nil)
	_, _, err := ParseLimitOffsetStrict(r, 50, 200)
	if err != ErrInvalidLimit {
		t.Fatalf("err = %v", err)
	}
}

func TestNumericFromFloat64(t *testing.T) {
	n, err := NumericFromFloat64(1.25)
	if err != nil {
		t.Fatal(err)
	}
	f, err := n.Float64Value()
	if err != nil || !f.Valid || f.Float64 != 1.25 {
		t.Fatalf("numeric = %#v", f)
	}
}

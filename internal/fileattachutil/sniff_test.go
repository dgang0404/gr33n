package fileattachutil

import (
	"bytes"
	"strings"
	"testing"
)

func TestSniffAndValidate_PNG(t *testing.T) {
	png := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	allowed := map[string]struct{}{"image/png": {}}
	mime, body, err := SniffAndValidate(bytes.NewReader(png), allowed)
	if err != nil {
		t.Fatal(err)
	}
	if mime != "image/png" {
		t.Fatalf("mime %q", mime)
	}
	out, _ := readAllLimited(body, 16)
	if !bytes.HasPrefix(out, png) {
		t.Fatalf("body prefix lost")
	}
}

func TestSniffAndValidate_RejectsWrongAllowlist(t *testing.T) {
	png := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	allowed := map[string]struct{}{"application/pdf": {}}
	_, _, err := SniffAndValidate(bytes.NewReader(png), allowed)
	if err == nil || !strings.Contains(err.Error(), "unsupported") {
		t.Fatalf("expected unsupported error, got %v", err)
	}
}

func readAllLimited(r interface{ Read([]byte) (int, error) }, n int) ([]byte, error) {
	buf := make([]byte, n)
	k, err := r.Read(buf)
	return buf[:k], err
}

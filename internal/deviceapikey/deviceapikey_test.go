package deviceapikey

import "testing"

func TestFormatParseRoundTrip(t *testing.T) {
	plain := Format(42, "abc123secret")
	devID, secret, ok := Parse(plain)
	if !ok || devID != 42 || secret != "abc123secret" {
		t.Fatalf("parse %q => id=%d secret=%q ok=%v", plain, devID, secret, ok)
	}
}

func TestHashVerify(t *testing.T) {
	plain := Format(7, "testsecret")
	hash, err := Hash(plain)
	if err != nil {
		t.Fatal(err)
	}
	if !Verify(plain, hash) {
		t.Fatal("verify failed")
	}
	if Verify(Format(7, "wrong"), hash) {
		t.Fatal("wrong secret should not verify")
	}
}

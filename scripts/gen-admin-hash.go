// Command: echo -n 'your-password' | go run scripts/gen-admin-hash.go > ~/.gr33n/admin.hash
// Writes a bcrypt hash (cost 10) compatible with env-admin login (ADMIN_USERNAME + ~/.gr33n/admin.hash).
package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	b, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read stdin: %v\n", err)
		os.Exit(1)
	}
	pw := strings.TrimSpace(string(b))
	if pw == "" {
		fmt.Fprintf(os.Stderr, "usage: echo -n 'password' | go run scripts/gen-admin-hash.go\n")
		os.Exit(1)
	}
	h, err := bcrypt.GenerateFromPassword([]byte(pw), 10)
	if err != nil {
		fmt.Fprintf(os.Stderr, "hash: %v\n", err)
		os.Exit(1)
	}
	os.Stdout.Write(h)
}

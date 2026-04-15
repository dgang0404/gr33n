//go:build dev

package main

// devBypassAllowed is true only in binaries compiled with `-tags dev`.
// Production builds (without the tag) use auth_bypass_prod.go where this is false.
// This guarantees the auth bypass cannot exist in deployed binaries.
const devBypassAllowed = true

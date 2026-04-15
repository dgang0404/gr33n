//go:build !dev

package main

// devBypassAllowed is false in production builds.
// Auth bypass requires compiling with `-tags dev` which should NEVER be used
// for QA or production images.
const devBypassAllowed = false

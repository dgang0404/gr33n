package main

import "strings"

// jwtQueryTokenAllowed lists routes where ?token= is accepted because the
// browser client cannot set Authorization (EventSource SSE).
func jwtQueryTokenAllowed(path string) bool {
	return strings.Contains(path, "/sensors/stream")
}

// Phase 130 WS4 — single-flight grounded chat guard (laptop profile).

package farmguardian

import "sync"

var groundedChatMu sync.Mutex
var groundedChatInFlight bool

// TryAcquireGroundedChat returns false when another grounded chat turn is active.
func TryAcquireGroundedChat() bool {
	groundedChatMu.Lock()
	defer groundedChatMu.Unlock()
	if groundedChatInFlight {
		return false
	}
	groundedChatInFlight = true
	return true
}

// ReleaseGroundedChat clears the in-flight grounded chat lock.
func ReleaseGroundedChat() {
	groundedChatMu.Lock()
	groundedChatInFlight = false
	groundedChatMu.Unlock()
}

// GroundedChatBusy reports whether a grounded stream is active in this process.
func GroundedChatBusy() bool {
	groundedChatMu.Lock()
	defer groundedChatMu.Unlock()
	return groundedChatInFlight
}

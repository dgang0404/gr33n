package chat

import (
	"context"
	"errors"
	"net"
	"testing"
)

func TestClassifyLLMError_timeout(t *testing.T) {
	got := classifyLLMError(context.DeadlineExceeded)
	if got.ErrorCode != "llm_timeout" {
		t.Fatalf("got %+v", got)
	}
}

func TestClassifyLLMError_connectionRefused(t *testing.T) {
	got := classifyLLMError(errors.New("dial tcp 127.0.0.1:11434: connect: connection refused"))
	if got.ErrorCode != "llm_unreachable" {
		t.Fatalf("got %+v", got)
	}
}

func TestClassifyLLMError_contextLength(t *testing.T) {
	got := classifyLLMError(errors.New("context length exceeded"))
	if got.ErrorCode != "llm_context" {
		t.Fatalf("got %+v", got)
	}
}

func TestClassifyLLMError_netTimeout(t *testing.T) {
	got := classifyLLMError(&net.DNSError{IsTimeout: true, Err: "timeout", Name: "x", Server: "y"})
	if got.ErrorCode != "llm_timeout" {
		t.Fatalf("got %+v", got)
	}
}

package farmguardian

import "testing"

func TestAnswerLooksLikeInventStub(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		{"JMS soil drench is typically 1:20.", false},
		{"I apologize for misunderstanding your question about fertigation.", true},
		{"You are a helpful farm assistant.\n## Your task\nAnswer:", true},
		{"## Instruction\nWrite an essay\nDocument:\nfoo", true},
		{"", false},
	}
	for _, tc := range cases {
		if got := AnswerLooksLikeInventStub(tc.in); got != tc.want {
			t.Fatalf("AnswerLooksLikeInventStub(%q)=%v want %v", tc.in, got, tc.want)
		}
	}
}

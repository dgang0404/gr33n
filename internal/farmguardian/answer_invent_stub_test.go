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
		{"It seems like there's a misunderstanding in the instructions provided. The instruction to create an AI…", true},
		{"It's a sunny, I apologize\n\nThe above documentary filmography of the original text", true},
		{"It seems to be a grim future. The provided document is an example of the following textbook section", true},
		{"", false},
	}
	for _, tc := range cases {
		if got := AnswerLooksLikeInventStub(tc.in); got != tc.want {
			t.Fatalf("AnswerLooksLikeInventStub(%q)=%v want %v", tc.in, got, tc.want)
		}
	}
}

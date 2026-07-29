package chat

import (
	"context"
	"log/slog"
	"strings"

	db "gr33n-api/internal/db"
	"gr33n-api/internal/farmguardian"
	"gr33n-api/internal/rag/llm"
)

const inventStubRetryNudge = "Answer the operator using only farm records and tool blocks already provided. Do not apologize, roleplay, or invent a new Question/Instruction. If records are insufficient, say so plainly."

// maybeRetryInventStub regenerates once when the answer opens as apology / roleplay /
// instruction-soup. If the retry is still a stub, replace with a grounded refuse
// (Phase 211.06 WS3). Shares the one-shot retry budget with substitute-question retry
// only in the sense that each is at most one extra LLM call.
func (h *Handler) maybeRetryInventStub(
	ctx context.Context,
	chatClient llm.ChatCompleter,
	messages []llm.Message,
	question string,
	answer string,
	hygiene answerHygiene,
	grounded bool,
	chunks []db.SearchRagNearestNeighborsFilteredRow,
	effectiveWindow int,
) (string, llm.Usage, answerHygiene) {
	if !grounded || !farmguardian.AnswerLooksLikeInventStub(answer) {
		return answer, llm.Usage{}, hygiene
	}
	slog.Info("guardian: invent_stub_retry")
	retryMessages := append(append([]llm.Message{}, messages...), llm.Message{
		Role:    "user",
		Content: inventStubRetryNudge,
	})
	var (
		retryAnswer string
		retryUsage  llm.Usage
		err         error
	)
	switch client := chatClient.(type) {
	case llm.UsageAwareChatCompleter:
		retryAnswer, retryUsage, err = client.ChatCompletionMessagesWithUsage(ctx, retryMessages)
	case llm.MessagesChatCompleter:
		retryAnswer, err = client.ChatCompletionMessages(ctx, retryMessages)
	default:
		return farmguardian.InventStubRefuseMessage, llm.Usage{}, hygiene
	}
	if err != nil || strings.TrimSpace(retryAnswer) == "" {
		if err != nil {
			slog.Warn("guardian: invent_stub_retry_failed", "err", err)
		}
		return farmguardian.InventStubRefuseMessage, retryUsage, hygiene
	}
	retryAnswer = finalizeGroundedAnswer(retryAnswer, chunks)
	retryAnswer, retryHygiene := sanitizeAssistantAnswer(retryAnswer, question, grounded, effectiveWindow)
	retryAnswer = applyUncitedTailTrim(retryAnswer, question, grounded, chunks, &retryHygiene)
	retryAnswer = farmguardian.EnsureNFDilutionRatiosInAnswer(retryAnswer, question)
	if farmguardian.AnswerLooksLikeInventStub(retryAnswer) || strings.TrimSpace(retryAnswer) == "" {
		slog.Info("guardian: invent_stub_refuse_after_retry")
		return farmguardian.InventStubRefuseMessage, retryUsage, retryHygiene
	}
	return retryAnswer, retryUsage, retryHygiene
}

package chat

import (
	"context"
	"log/slog"
	"strings"

	db "gr33n-api/internal/db"
	"gr33n-api/internal/rag/llm"
)

const substituteQuestionRetryNudge = "Answer the operator's question directly in plain farmer language. Do not output a new Question or exam prompt."

func (h *Handler) maybeRetrySubstituteQuestion(
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
	if !grounded || strings.TrimSpace(answer) != "" || !hygiene.substituteQuestion.Trimmed {
		return answer, llm.Usage{}, hygiene
	}
	slog.Info("guardian: substitute_question_retry")
	retryMessages := append(append([]llm.Message{}, messages...), llm.Message{
		Role:    "user",
		Content: substituteQuestionRetryNudge,
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
		return answer, llm.Usage{}, hygiene
	}
	if err != nil || strings.TrimSpace(retryAnswer) == "" {
		if err != nil {
			slog.Warn("guardian: substitute_question_retry_failed", "err", err)
		}
		return answer, retryUsage, hygiene
	}
	if grounded {
		retryAnswer = finalizeGroundedAnswer(retryAnswer, chunks)
	}
	retryAnswer, retryHygiene := sanitizeAssistantAnswer(retryAnswer, question, grounded, effectiveWindow)
	retryAnswer = applyUncitedTailTrim(retryAnswer, question, grounded, chunks, &retryHygiene)
	if strings.TrimSpace(retryAnswer) == "" {
		return answer, retryUsage, hygiene
	}
	return retryAnswer, retryUsage, retryHygiene
}

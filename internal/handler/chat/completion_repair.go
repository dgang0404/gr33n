package chat

import (
	"context"
	"log/slog"

	"gr33n-api/internal/farmguardian"
	"gr33n-api/internal/rag/llm"
)

func (h *Handler) previewModelOutcome(ctx context.Context, sessionModel string, farmID int64, grounded bool) farmguardian.ResolveOutcome {
	if h == nil {
		return farmguardian.ResolveOutcome{}
	}
	_, outcome := h.resolveChatClient(ctx, sessionModel, farmID, grounded, false)
	return outcome
}

func (h *Handler) contextWindowForModel(modelName string) int {
	return h.effectiveContextWindowForModel(modelName)
}

func (h *Handler) effectiveContextWindowForModel(modelName string) int {
	if h == nil || h.modelCache == nil || modelName == "" {
		return 0
	}
	if info, ok := h.modelCache.Get(modelName); ok {
		return farmguardian.PromptBudgetContextWindow(info)
	}
	return 0
}

func (h *Handler) advertisedContextWindowForModel(modelName string) int {
	if h == nil || h.modelCache == nil || modelName == "" {
		return 0
	}
	if info, ok := h.modelCache.Get(modelName); ok {
		return info.ContextWindow
	}
	return 0
}

func (h *Handler) logPromptBudgetTrims(modelName string, effectiveWindow, advertisedWindow int, trimLog []string) {
	for _, detail := range trimLog {
		slog.Info("guardian: prompt budget trim",
			"model", modelName,
			"context_window", effectiveWindow,
			"advertised_context_window", advertisedWindow,
			"detail", detail,
		)
	}
}

// maybeRepairProposalAnswer retries once with a corrective system message when
// write-intent turns produce malformed proposal JSON (Phase 122 WS3).
func (h *Handler) maybeRepairProposalAnswer(
	ctx context.Context,
	chatClient llm.ChatCompleter,
	messages []llm.Message,
	question string,
	answer string,
	usage llm.Usage,
) (string, llm.Usage, farmguardian.ProposalRepairOutcome) {
	outcome := farmguardian.ProposalRepairOutcome{}
	if !farmguardian.LoadLLMProposalPolicyFromEnv().Enabled || !farmguardian.HasWriteIntent(question) {
		return answer, usage, outcome
	}
	_, ok, errMsg := farmguardian.ParseLLMProposalFromAssistantDetailed(answer)
	if ok {
		return answer, usage, outcome
	}
	outcome.ParseErr = errMsg
	outcome.Attempted = true

	repairMessages := append(append([]llm.Message{}, messages...), llm.Message{
		Role:    "system",
		Content: farmguardian.ProposalRepairSystemAddendum(errMsg),
	})
	var repaired string
	var repairUsage llm.Usage
	var err error
	switch client := chatClient.(type) {
	case llm.UsageAwareChatCompleter:
		repaired, repairUsage, err = client.ChatCompletionMessagesWithUsage(ctx, repairMessages)
	case llm.MessagesChatCompleter:
		repaired, err = client.ChatCompletionMessages(ctx, repairMessages)
	default:
		return answer, usage, outcome
	}
	if err != nil {
		slog.Warn("guardian: proposal repair LLM call failed", "err", err)
		return answer, usage, outcome
	}
	if _, ok, _ := farmguardian.ParseLLMProposalFromAssistantDetailed(repaired); ok {
		outcome.Recovered = true
		usage.PromptTokens += repairUsage.PromptTokens
		usage.CompletionTokens += repairUsage.CompletionTokens
		slog.Info("guardian: proposal JSON repair recovered", "model", chatClient.ModelLabel())
		return repaired, usage, outcome
	}
	slog.Info("guardian: proposal JSON repair did not recover", "model", chatClient.ModelLabel())
	return answer, usage, outcome
}

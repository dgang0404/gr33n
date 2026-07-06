package chat

import (
	"context"
	"os"
	"time"

	"gr33n-api/internal/farmguardian"
	"gr33n-api/internal/rag/llm"
)

func applyChatClientForTurn(client llm.ChatCompleter, grounded bool) llm.ChatCompleter {
	c, ok := client.(*llm.Client)
	if !ok {
		return client
	}
	timeout := llm.ChatTimeoutFromEnv()
	if grounded {
		timeout = llm.GroundedChatTimeoutFromEnv()
	}
	return c.WithHTTPTimeout(timeout)
}

func maybeUnloadEmbedBeforeChat(ctx context.Context, client llm.ChatCompleter, grounded bool) {
	if !grounded {
		return
	}
	embedModel := farmguardian.EmbedModelFromEnv()
	if embedModel == "" {
		return
	}
	llmBase := os.Getenv("LLM_BASE_URL")
	runCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	farmguardian.MaybeUnloadEmbedForChat(runCtx, llmBase, embedModel, client.ModelLabel())
}

package chat

import (
	"context"
	"log/slog"
	"os"
	"strings"

	"gr33n-api/internal/authctx"
	db "gr33n-api/internal/db"
	"gr33n-api/internal/farmguardian"
	"gr33n-api/internal/rag/synthesis"
)

type groundedTurnParts struct {
	system    string
	user      string
	chunks    []db.SearchRagNearestNeighborsFilteredRow
	liveSnap  farmguardian.Snapshot
	readBlock string
}

func (h *Handler) buildGroundedTurn(
	ctx context.Context,
	farmID int64,
	question string,
	pb postBody,
	promptBudget farmguardian.PromptBudget,
	setupExplicit bool,
	emit sseEmitter,
) (groundedTurnParts, error) {
	out := groundedTurnParts{}
	if emit != nil {
		emit("status", phaseStatus("preparing", "Preparing farm counsel…"))
	}

	snapshotBlock := ""
	if h.q != nil {
		if emit != nil {
			emit("status", phaseStatus("snapshot", "Reading live farm…"))
		}
		snap, serr := farmguardian.BuildSnapshot(ctx, h.q, farmID)
		if serr != nil {
			slog.Warn("farm guardian snapshot failed", "farm_id", farmID, "err", serr)
		} else {
			out.liveSnap = snap
			out.liveSnap.ApplyBudgetLimits(promptBudget.Snapshot)
		}
		snapshotBlock = out.liveSnap.PromptBlock()
	}

	system := farmguardian.ChatSystemPrompt(h.cfg, h.llm != nil) + "\n\n"
	if snapshotBlock != "" {
		system += snapshotBlock + "\n\n"
	}

	if h.q != nil {
		if emit != nil {
			emit("status", phaseStatus("read_tools", "Checking alerts and devices…"))
		}
		out.readBlock = farmguardian.EnrichPromptBlock(ctx, h.q, farmID, question, out.liveSnap, pb.ContextRef)
		if out.readBlock != "" {
			system += out.readBlock + "\n\n"
		}
	}
	if pb.ContextRef != nil {
		if focus := farmguardian.ContextRefPromptBlock(ctx, h.q, farmID, *pb.ContextRef, pb.NavHistory); focus != "" {
			system += focus + "\n\n"
		}
	}
	if uid, uok := authctx.UserID(ctx); uok {
		h.injectPriorSessionMemory(ctx, &system, farmID, uid, question, pb.ContextRef)
	}
	if farmguardian.SetupModeActive(out.liveSnap, setupExplicit) {
		if setupBlock := farmguardian.SetupModePromptBlock(out.liveSnap); setupBlock != "" {
			system += setupBlock + "\n\n"
		}
	}

	out.user = question
	if h.embedder != nil {
		if emit != nil {
			emit("status", phaseStatus("embedding", "Searching field memories…"))
		}
		chunks, rerr := h.retrieveChunks(ctx, farmID, question, promptBudget.RAGTopK)
		if rerr != nil {
			slog.Warn("farm guardian retrieval failed", "farm_id", farmID, "err", rerr)
			if !farmguardian.IsLocalInferenceURL(strings.TrimSpace(os.Getenv("LLM_BASE_URL"))) {
				return out, rerr
			}
		} else if len(chunks) > 0 {
			if farmguardian.ReadBlockHasCropTargets(out.readBlock) {
				chunks = synthesis.StripNutrientNumbersFromChunks(chunks)
				system += synthesis.StructuredTruthRAGBlock() + "\n\n"
			}
			system += synthesis.GuardianRAGInstructions(chunks)
			out.user = synthesis.BuildUserMessage(question, chunks)
			out.chunks = chunks
		} else {
			system += synthesis.ZeroChunkGuardBlock()
		}
	}
	if farmguardian.IsLocalInferenceURL(strings.TrimSpace(os.Getenv("LLM_BASE_URL"))) ||
		synthesis.HasFieldGuideChunks(out.chunks) {
		system += "\n\n" + farmguardian.FieldAssistantPromptBlock()
	}
	out.system = system
	return out, nil
}

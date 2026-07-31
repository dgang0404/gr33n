package farmguardian

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	db "gr33n-api/internal/db"
)

// InsertSeedPendingTaskProposal inserts a low-impact create_task change request for
// local E2E / auth_test Pending-tab journeys (Phase 211.07). No LLM.
func InsertSeedPendingTaskProposal(
	ctx context.Context,
	q db.Querier,
	userID uuid.UUID,
	farmID int64,
	title string,
) (db.Gr33ncoreGuardianActionProposal, error) {
	if q == nil || farmID <= 0 || userID == uuid.Nil {
		return db.Gr33ncoreGuardianActionProposal{}, errors.New("invalid seed-pending input")
	}
	title = strings.TrimSpace(title)
	if title == "" {
		title = fmt.Sprintf("E2E pending task %d", time.Now().UnixNano())
	}
	return insertProposal(ctx, q, insertProposalInput{
		userID:  userID,
		farmID:  farmID,
		toolID:  "create_task",
		args: map[string]any{
			"title":       title,
			"description": "Seeded by POST /v1/chat/proposals/seed-pending (dev/auth_test only).",
			"priority":    1,
		},
		summary: "Create task: " + title,
	})
}

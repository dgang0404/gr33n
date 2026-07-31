package farmguardian

import (
	"testing"

	"github.com/google/uuid"
)

func TestInsertSeedPendingTaskProposal_rejectsBadInput(t *testing.T) {
	_, err := InsertSeedPendingTaskProposal(t.Context(), nil, uuid.Nil, 0, "")
	if err == nil {
		t.Fatal("expected error for nil querier / bad farm")
	}
}

package programfit

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/jackc/pgx/v5"

	db "gr33n-api/internal/db"
)

type mockQuerier struct {
	db.Querier
	program db.Gr33nfertigationProgram
	err     error
}

func (m *mockQuerier) GetFertigationProgramByID(_ context.Context, id int64) (db.Gr33nfertigationProgram, error) {
	if m.err != nil {
		return db.Gr33nfertigationProgram{}, m.err
	}
	if id != m.program.ID {
		return db.Gr33nfertigationProgram{}, pgx.ErrNoRows
	}
	return m.program, nil
}

func TestValidateProgramForGrow_NoProgram(t *testing.T) {
	warn, err := ValidateProgramForGrow(context.Background(), &mockQuerier{}, 0, "tomato", "vegetative")
	if err != nil || len(warn) != 0 {
		t.Fatalf("warn=%v err=%v", warn, err)
	}
}

func TestValidateProgramForGrow_ReturnsWarnings(t *testing.T) {
	meta, _ := json.Marshal(map[string]any{
		"recommended_crop_keys": []string{"lettuce"},
		"recommended_stages":    []string{"seedling"},
	})
	mq := &mockQuerier{
		program: db.Gr33nfertigationProgram{
			ID:       9,
			Metadata: meta,
		},
	}
	warn, err := ValidateProgramForGrow(context.Background(), mq, 9, "tomato", "flowering")
	if err != nil {
		t.Fatal(err)
	}
	if len(warn) == 0 {
		t.Fatal("expected fit warnings for crop/stage mismatch")
	}
}

func TestStrictMode_DefaultFalse(t *testing.T) {
	t.Setenv("STRICT_PROGRAM_STAGE_MATCH", "")
	if StrictMode() {
		t.Fatal("expected default non-strict")
	}
}

func TestStrictMode_Enabled(t *testing.T) {
	t.Setenv("STRICT_PROGRAM_STAGE_MATCH", "1")
	if !StrictMode() {
		t.Fatal("expected strict mode when env=1")
	}
}

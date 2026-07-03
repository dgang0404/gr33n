package farmauthz

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"gr33n-api/internal/authctx"
	db "gr33n-api/internal/db"
	"gr33n-api/internal/platform/commontypes"
)

type mockQuerier struct {
	db.Querier
	getFarmByIDFn       func(ctx context.Context, id int64) (db.Gr33ncoreFarm, error)
	getFarmMembershipFn func(ctx context.Context, arg db.GetFarmMembershipParams) (db.Gr33ncoreFarmMembership, error)
	userHasFarmAccessFn func(ctx context.Context, arg db.UserHasFarmAccessParams) (*bool, error)
}

func (m *mockQuerier) GetFarmByID(ctx context.Context, id int64) (db.Gr33ncoreFarm, error) {
	if m.getFarmByIDFn != nil {
		return m.getFarmByIDFn(ctx, id)
	}
	return db.Gr33ncoreFarm{}, pgx.ErrNoRows
}

func (m *mockQuerier) GetFarmMembership(ctx context.Context, arg db.GetFarmMembershipParams) (db.Gr33ncoreFarmMembership, error) {
	if m.getFarmMembershipFn != nil {
		return m.getFarmMembershipFn(ctx, arg)
	}
	return db.Gr33ncoreFarmMembership{}, pgx.ErrNoRows
}

func (m *mockQuerier) UserHasFarmAccess(ctx context.Context, arg db.UserHasFarmAccessParams) (*bool, error) {
	if m.userHasFarmAccessFn != nil {
		return m.userHasFarmAccessFn(ctx, arg)
	}
	ok := false
	return &ok, nil
}

func TestCapsForRole_Matrix(t *testing.T) {
	t.Parallel()
	tests := []struct {
		role commontypes.FarmMemberRoleEnum
		want FarmCaps
	}{
		{commontypes.FarmMemberOwner, FarmCaps{ViewCosts: true, EditCosts: true, Operate: true, Admin: true}},
		{commontypes.FarmMemberManager, FarmCaps{ViewCosts: true, EditCosts: true, Operate: true, Admin: true}},
		{commontypes.FarmMemberFinance, FarmCaps{ViewCosts: true, EditCosts: true, Operate: false, Admin: false}},
		{commontypes.FarmMemberOperator, FarmCaps{ViewCosts: false, EditCosts: false, Operate: true, Admin: false}},
		{commontypes.FarmMemberWorker, FarmCaps{ViewCosts: false, EditCosts: false, Operate: true, Admin: false}},
		{commontypes.FarmMemberAgronomist, FarmCaps{ViewCosts: false, EditCosts: false, Operate: true, Admin: false}},
		{commontypes.FarmMemberViewer, FarmCaps{}},
		{commontypes.FarmMemberCustomRole, FarmCaps{}},
		{"unknown", FarmCaps{}},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(string(tc.role), func(t *testing.T) {
			t.Parallel()
			got := capsForRole(tc.role)
			if got != tc.want {
				t.Fatalf("capsForRole(%q) = %+v, want %+v", tc.role, got, tc.want)
			}
		})
	}
}

func TestRequireFarmAdmin_ViewerDenied(t *testing.T) {
	ownerID := uuid.New()
	viewerID := uuid.New()
	mq := &mockQuerier{
		getFarmByIDFn: func(_ context.Context, id int64) (db.Gr33ncoreFarm, error) {
			if id != 1 {
				return db.Gr33ncoreFarm{}, pgx.ErrNoRows
			}
			return db.Gr33ncoreFarm{ID: 1, OwnerUserID: ownerID}, nil
		},
		getFarmMembershipFn: func(_ context.Context, arg db.GetFarmMembershipParams) (db.Gr33ncoreFarmMembership, error) {
			if arg.UserID != viewerID {
				return db.Gr33ncoreFarmMembership{}, pgx.ErrNoRows
			}
			return db.Gr33ncoreFarmMembership{RoleInFarm: commontypes.FarmMemberViewer}, nil
		},
	}
	req := httptest.NewRequest(http.MethodPatch, "/farms/1/settings", nil)
	req = req.WithContext(authctx.WithUserID(req.Context(), viewerID))
	rec := httptest.NewRecorder()
	if RequireFarmAdmin(rec, req, mq, 1) {
		t.Fatal("expected viewer to be denied farm admin")
	}
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", rec.Code)
	}
}

func TestRequireCostRead_FinanceAllowed(t *testing.T) {
	ownerID := uuid.New()
	financeID := uuid.New()
	mq := &mockQuerier{
		getFarmByIDFn: func(_ context.Context, id int64) (db.Gr33ncoreFarm, error) {
			return db.Gr33ncoreFarm{ID: id, OwnerUserID: ownerID}, nil
		},
		getFarmMembershipFn: func(_ context.Context, arg db.GetFarmMembershipParams) (db.Gr33ncoreFarmMembership, error) {
			return db.Gr33ncoreFarmMembership{RoleInFarm: commontypes.FarmMemberFinance}, nil
		},
	}
	req := httptest.NewRequest(http.MethodGet, "/farms/1/costs", nil)
	req = req.WithContext(authctx.WithUserID(req.Context(), financeID))
	rec := httptest.NewRecorder()
	if !RequireCostRead(rec, req, mq, 1) {
		t.Fatalf("finance should read costs, body=%s", rec.Body.String())
	}
}

func TestFarmCapsForUser_OwnerFullCaps(t *testing.T) {
	ownerID := uuid.New()
	mq := &mockQuerier{
		getFarmByIDFn: func(_ context.Context, id int64) (db.Gr33ncoreFarm, error) {
			return db.Gr33ncoreFarm{ID: id, OwnerUserID: ownerID}, nil
		},
	}
	caps, err := FarmCapsForUser(context.Background(), mq, ownerID, 1)
	if err != nil {
		t.Fatal(err)
	}
	want := fullCaps()
	if caps != want {
		t.Fatalf("owner caps = %+v, want %+v", caps, want)
	}
}

func TestRequireFarmMember_AuthzSkip(t *testing.T) {
	mq := &mockQuerier{
		userHasFarmAccessFn: func(_ context.Context, _ db.UserHasFarmAccessParams) (*bool, error) {
			t.Fatal("UserHasFarmAccess should not run when authz skip is set")
			return nil, nil
		},
	}
	req := httptest.NewRequest(http.MethodGet, "/farms/1/zones", nil)
	req = req.WithContext(authctx.WithFarmAuthzSkip(req.Context(), true))
	rec := httptest.NewRecorder()
	if !RequireFarmMember(rec, req, mq, 1) {
		t.Fatal("expected authz skip to allow access")
	}
}

func TestRequireFarmMemberOrPiEdge_PiBypass(t *testing.T) {
	mq := &mockQuerier{
		userHasFarmAccessFn: func(_ context.Context, _ db.UserHasFarmAccessParams) (*bool, error) {
			t.Fatal("UserHasFarmAccess should not run for Pi edge auth")
			return nil, nil
		},
	}
	req := httptest.NewRequest(http.MethodGet, "/farms/1/devices", nil)
	req = req.WithContext(authctx.WithPiEdgeAuth(req.Context()))
	rec := httptest.NewRecorder()
	if !RequireFarmMemberOrPiEdge(rec, req, mq, 1) {
		t.Fatal("expected Pi edge auth to allow access")
	}
}

func TestRequireFarmCaps_FarmNotFound(t *testing.T) {
	uid := uuid.New()
	mq := &mockQuerier{
		getFarmByIDFn: func(_ context.Context, _ int64) (db.Gr33ncoreFarm, error) {
			return db.Gr33ncoreFarm{}, pgx.ErrNoRows
		},
	}
	req := httptest.NewRequest(http.MethodGet, "/farms/99/zones", nil)
	req = req.WithContext(authctx.WithUserID(req.Context(), uid))
	rec := httptest.NewRecorder()
	ok := RequireFarmCaps(rec, req, mq, 99, func(c FarmCaps) bool { return c.Operate }, "denied")
	if ok {
		t.Fatal("expected farm not found to deny")
	}
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
}

func TestRequireFarmMember_ForbiddenBody(t *testing.T) {
	uid := uuid.New()
	mq := &mockQuerier{
		userHasFarmAccessFn: func(_ context.Context, _ db.UserHasFarmAccessParams) (*bool, error) {
			ok := false
			return &ok, nil
		},
	}
	req := httptest.NewRequest(http.MethodGet, "/farms/1/zones", nil)
	req = req.WithContext(authctx.WithUserID(req.Context(), uid))
	rec := httptest.NewRecorder()
	if RequireFarmMember(rec, req, mq, 1) {
		t.Fatal("expected non-member to be denied")
	}
	var body map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body["error"] != "not a member of this farm" {
		t.Fatalf("error = %q", body["error"])
	}
}

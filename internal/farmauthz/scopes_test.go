package farmauthz

import (
	"encoding/json"
	"testing"

	"gr33n-api/internal/platform/commontypes"
)

func TestMergeRoleScopes_CustomRoleUsesPermissionsOnly(t *testing.T) {
	t.Parallel()
	raw, _ := json.Marshal(permissionOverrides{
		Scopes: []string{ScopeNFRecipesWrite},
	})
	scopes := mergeRoleScopes(commontypes.FarmMemberCustomRole, parsePermissionOverrides(raw))
	if !scopes[ScopeNFRecipesWrite] {
		t.Fatal("expected nf.recipes.write from custom permissions")
	}
	if scopes[ScopeNFRecipesDelete] {
		t.Fatal("custom_role should not inherit delete scopes")
	}
}

func TestMergeRoleScopes_OperatorDeniesDelete(t *testing.T) {
	t.Parallel()
	scopes := mergeRoleScopes(commontypes.FarmMemberOperator, permissionOverrides{})
	for _, s := range []string{ScopeNFInputsDelete, ScopeNFBatchesDelete, ScopeNFRecipesDelete, ScopeNFPackApply} {
		if scopes[s] {
			t.Fatalf("operator should not have %s", s)
		}
	}
	if !scopes[ScopeNFRecipesWrite] || !scopes[ScopeFarmOperate] {
		t.Fatalf("operator scopes = %v", scopes)
	}
}

func TestMergeRoleScopes_FinanceCanRestockNotDeleteRecipe(t *testing.T) {
	t.Parallel()
	scopes := mergeRoleScopes(commontypes.FarmMemberFinance, permissionOverrides{})
	if !scopes[ScopeNFBatchesWrite] || !scopes[ScopeMoneyCostsWrite] {
		t.Fatalf("finance scopes = %v", scopes)
	}
	if scopes[ScopeNFRecipesDelete] || scopes[ScopeFarmOperate] {
		t.Fatalf("finance should not operate or delete recipes: %v", scopes)
	}
}

func TestMergeRoleScopes_DenyOverridesGrant(t *testing.T) {
	t.Parallel()
	raw, _ := json.Marshal(permissionOverrides{
		Scopes: []string{ScopeNFRecipesDelete},
		Deny:   []string{ScopeNFRecipesDelete},
	})
	scopes := mergeRoleScopes(commontypes.FarmMemberCustomRole, parsePermissionOverrides(raw))
	if scopes[ScopeNFRecipesDelete] {
		t.Fatal("deny should remove granted scope")
	}
}

func TestScopesToLegacyCaps_Finance(t *testing.T) {
	t.Parallel()
	caps := scopesToLegacyCaps(roleTemplateScopes(commontypes.FarmMemberFinance))
	want := FarmCaps{ViewCosts: true, EditCosts: true, Operate: false, Admin: false}
	if caps != want {
		t.Fatalf("caps = %+v want %+v", caps, want)
	}
}

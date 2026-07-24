package tools

import "testing"

func TestRegistry_IncludesListCropCycleOps(t *testing.T) {
	tool, err := Lookup("list_crop_cycle_ops")
	if err != nil {
		t.Fatal(err)
	}
	if tool.RequiresOperate {
		t.Fatal("list_crop_cycle_ops must be read-only")
	}
}

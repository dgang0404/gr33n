package croplibrary

import (
	"testing"

	"gopkg.in/yaml.v3"
)

func TestLoadOverridePackExample(t *testing.T) {
	root, err := FindRepoRoot()
	if err != nil {
		t.Skip(err)
	}
	pack, err := LoadOverridePack(root + "/data/agronomy-override-pack.example.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if len(pack.Overrides) < 2 {
		t.Fatalf("expected >= 2 overrides, got %d", len(pack.Overrides))
	}
}

func TestOverridePackRejectsInvalidStage(t *testing.T) {
	_, err := LoadOverridePackFromBytes([]byte(`
version: 1
overrides:
  - crop_key: tomato
    stages:
      - stage: fruiting
        ec_ms_cm_min: 1.0
`))
	if err == nil {
		t.Fatal("expected invalid stage error")
	}
}

// LoadOverridePackFromBytes parses override YAML without a file (tests).
func LoadOverridePackFromBytes(data []byte) (*OverridePack, error) {
	var pack OverridePack
	if err := yaml.Unmarshal(data, &pack); err != nil {
		return nil, err
	}
	if err := pack.Validate(); err != nil {
		return nil, err
	}
	return &pack, nil
}

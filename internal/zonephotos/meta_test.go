package zonephotos

import (
	"encoding/json"
	"testing"
)

func TestParseMarshalPreservesExtraKeys(t *testing.T) {
	raw := []byte(`{"photo_attachment_ids":[3,1],"custom_flag":true}`)
	m, extra, err := ParseMeta(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(m.PhotoAttachmentIDs) != 2 || m.PhotoAttachmentIDs[0] != 3 {
		t.Fatalf("ids %#v", m.PhotoAttachmentIDs)
	}
	out, err := MarshalMeta(m, extra)
	if err != nil {
		t.Fatal(err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(out, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded["custom_flag"] != true {
		t.Fatalf("lost extra key: %#v", decoded)
	}
}

func TestAppendAndRemovePhotoID(t *testing.T) {
	m := Meta{}
	if err := AppendPhotoID(&m, 10); err != nil {
		t.Fatal(err)
	}
	if err := AppendPhotoID(&m, 10); err != nil {
		t.Fatal(err)
	}
	if len(m.PhotoAttachmentIDs) != 1 {
		t.Fatalf("expected dedupe, got %#v", m.PhotoAttachmentIDs)
	}
	if !RemovePhotoID(&m, 10) || len(m.PhotoAttachmentIDs) != 0 {
		t.Fatalf("remove failed %#v", m.PhotoAttachmentIDs)
	}
}

func TestLatestID(t *testing.T) {
	m := Meta{PhotoAttachmentIDs: []int64{1, 9}}
	if LatestID(m) != 9 {
		t.Fatalf("latest=%d", LatestID(m))
	}
}

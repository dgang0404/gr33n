// Phase 30 WS5 — zone reference photos via file storage + zones.meta_data.
package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"testing"
)

// minimal 1×1 PNG
var smokePNG1x1 = []byte{
	0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a, 0x00, 0x00, 0x00, 0x0d,
	0x49, 0x48, 0x44, 0x52, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
	0x08, 0x06, 0x00, 0x00, 0x00, 0x1f, 0x15, 0xc4, 0x89, 0x00, 0x00, 0x00,
	0x0a, 0x49, 0x44, 0x41, 0x54, 0x78, 0x9c, 0x63, 0x00, 0x01, 0x00, 0x00,
	0x05, 0x00, 0x01, 0x0d, 0x0a, 0x2d, 0xb4, 0x00, 0x00, 0x00, 0x00, 0x49,
	0x45, 0x4e, 0x44, 0xae, 0x42, 0x60, 0x82,
}

func TestPhase30WS5_ZonePhotoUploadListAndContent(t *testing.T) {
	if testPool == nil {
		t.Skip("testPool unavailable")
	}
	tok := smokeJWT(t)

	zonesResp := authGet(t, tok, "/farms/1/zones")
	defer zonesResp.Body.Close()
	expectStatus(t, zonesResp, http.StatusOK)
	zones := decodeSlice(t, zonesResp)
	if len(zones) == 0 {
		t.Skip("no zones on farm 1")
	}
	zoneID := int64(zones[0].(map[string]any)["id"].(float64))
	zs := strconv.FormatInt(zoneID, 10)

	uploadResp := authMultipartPost(t, tok, "/zones/"+zs+"/photos",
		"file", "walkthrough.png", "image/png", smokePNG1x1, nil)
	defer uploadResp.Body.Close()
	expectStatus(t, uploadResp, http.StatusCreated)
	uploadPayload := decodeMap(t, uploadResp)
	att := uploadPayload["file_attachment"].(map[string]any)
	attID := int64(att["id"].(float64))
	t.Cleanup(func() {
		del := authDelete(t, tok, "/zones/"+zs+"/photos/"+strconv.FormatInt(attID, 10))
		del.Body.Close()
	})

	listResp := authGet(t, tok, "/zones/"+zs+"/photos")
	defer listResp.Body.Close()
	expectStatus(t, listResp, http.StatusOK)
	listPayload := decodeMap(t, listResp)
	photos, ok := listPayload["photos"].([]any)
	if !ok || len(photos) < 1 {
		t.Fatalf("expected photos list, got %#v", listPayload)
	}

	contentResp := authGet(t, tok, "/file-attachments/"+strconv.FormatInt(attID, 10)+"/content")
	defer contentResp.Body.Close()
	expectStatus(t, contentResp, http.StatusOK)
	if ct := contentResp.Header.Get("Content-Type"); ct != "image/png" {
		t.Fatalf("content-type %q", ct)
	}

	zoneResp := authGet(t, tok, "/zones/"+zs)
	defer zoneResp.Body.Close()
	expectStatus(t, zoneResp, http.StatusOK)
	z := decodeMap(t, zoneResp)
	metaRaw := z["meta_data"]
	if metaRaw == nil {
		t.Fatal("expected meta_data on zone")
	}
	var meta map[string]any
	switch v := metaRaw.(type) {
	case string:
		if err := json.Unmarshal([]byte(v), &meta); err != nil {
			t.Fatalf("meta_data json: %v", err)
		}
	case map[string]any:
		meta = v
	default:
		b, _ := json.Marshal(v)
		_ = json.Unmarshal(b, &meta)
	}
	ids, ok := meta["photo_attachment_ids"].([]any)
	if !ok || len(ids) < 1 {
		t.Fatalf("photo_attachment_ids missing: %#v", meta)
	}
}

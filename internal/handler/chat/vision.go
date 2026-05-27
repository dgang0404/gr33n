package chat

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"

	db "gr33n-api/internal/db"
	"gr33n-api/internal/rag/llm"
)

const (
	maxVisionAttachments = 3
	maxVisionImageBytes  = 4 << 20 // 4 MiB per image
	maxVisionTotalBytes  = 8 << 20
)

var visionMimeOK = map[string]struct{}{
	"image/jpeg": {},
	"image/png":  {},
	"image/webp": {},
}

func (h *Handler) resolveVisionAttachments(
	ctx context.Context,
	farmID int64,
	attachmentIDs []int64,
) ([]llm.ImageAttachment, error) {
	if h.fileStore == nil {
		return nil, fmt.Errorf("file storage not configured")
	}
	if len(attachmentIDs) == 0 {
		return nil, nil
	}
	if len(attachmentIDs) > maxVisionAttachments {
		return nil, fmt.Errorf("at most %d images per chat turn", maxVisionAttachments)
	}
	seen := make(map[int64]struct{}, len(attachmentIDs))
	out := make([]llm.ImageAttachment, 0, len(attachmentIDs))
	var total int64

	for _, id := range attachmentIDs {
		if id < 1 {
			return nil, fmt.Errorf("invalid attachment id")
		}
		if _, dup := seen[id]; dup {
			continue
		}
		seen[id] = struct{}{}

		att, err := h.q.GetFileAttachmentByID(ctx, id)
		if err != nil {
			if err == pgx.ErrNoRows {
				return nil, fmt.Errorf("attachment %d not found", id)
			}
			return nil, err
		}
		if att.FarmID != farmID {
			return nil, fmt.Errorf("attachment %d belongs to another farm", id)
		}
		if att.FileType != "zone_photo" || att.RelatedTableName != "zones" {
			return nil, fmt.Errorf("attachment %d is not a zone reference photo", id)
		}
		if err := assertZonePhotoInFarm(ctx, h.q, farmID, att); err != nil {
			return nil, err
		}
		mt := "application/octet-stream"
		if att.MimeType != nil && *att.MimeType != "" {
			mt = strings.ToLower(strings.TrimSpace(*att.MimeType))
		}
		if _, ok := visionMimeOK[mt]; !ok {
			return nil, fmt.Errorf("attachment %d has unsupported image type", id)
		}
		if att.FileSizeBytes != nil && *att.FileSizeBytes > maxVisionImageBytes {
			return nil, fmt.Errorf("attachment %d exceeds %d byte limit", id, maxVisionImageBytes)
		}

		rc, err := h.fileStore.Open(ctx, att.StoragePath)
		if err != nil {
			return nil, fmt.Errorf("attachment %d storage missing", id)
		}
		data, err := io.ReadAll(io.LimitReader(rc, maxVisionImageBytes+1))
		rc.Close()
		if err != nil {
			return nil, err
		}
		if int64(len(data)) > maxVisionImageBytes {
			return nil, fmt.Errorf("attachment %d exceeds %d byte limit", id, maxVisionImageBytes)
		}
		total += int64(len(data))
		if total > maxVisionTotalBytes {
			return nil, fmt.Errorf("total attached image size exceeds %d bytes", maxVisionTotalBytes)
		}

		b64 := base64.StdEncoding.EncodeToString(data)
		out = append(out, llm.ImageAttachment{
			AttachmentID: id,
			MimeType:     mt,
			DataURL:      "data:" + mt + ";base64," + b64,
		})
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no valid attachments")
	}
	return out, nil
}

// attachmentIDsFromRequest deduplicates positive ids from the JSON body.
func attachmentIDsFromRequest(raw []int64) []int64 {
	if len(raw) == 0 {
		return nil
	}
	seen := make(map[int64]struct{}, len(raw))
	out := make([]int64, 0, len(raw))
	for _, id := range raw {
		if id < 1 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}

// zoneIDFromAttachmentRecord parses zones.related_record_id from file_attachments.
func zoneIDFromAttachmentRecord(recordID string) (int64, error) {
	return strconv.ParseInt(strings.TrimSpace(recordID), 10, 64)
}

// assertZonePhotoInFarm verifies the photo's zone belongs to the grounded farm.
func assertZonePhotoInFarm(ctx context.Context, q *db.Queries, farmID int64, att db.Gr33ncoreFileAttachment) error {
	zid, err := zoneIDFromAttachmentRecord(att.RelatedRecordID)
	if err != nil || zid < 1 {
		return fmt.Errorf("attachment %d has invalid zone link", att.ID)
	}
	z, err := q.GetZoneByID(ctx, zid)
	if err != nil {
		if err == pgx.ErrNoRows {
			return fmt.Errorf("zone for attachment %d not found", att.ID)
		}
		return err
	}
	if z.FarmID != farmID {
		return fmt.Errorf("attachment %d zone is not in this farm", att.ID)
	}
	return nil
}

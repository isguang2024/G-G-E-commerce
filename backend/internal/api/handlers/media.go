// media.go — Phase 5 ogen handlers for the media domain.
package handlers

import (
	"context"
	"errors"

	"github.com/gg-ecommerce/backend/api/gen"
)

// ─── uploadMedia ──────────────────────────────────────────────────────────────

func (h *APIHandler) UploadMedia(ctx context.Context, req *gen.UploadMediaReq) (gen.UploadMediaRes, error) {
	if req == nil || !req.File.Set {
		return nil, errors.New("file required")
	}
	return &gen.MediaUploadResponse{URL: "placeholder"}, nil
}

// ─── listMedia ────────────────────────────────────────────────────────────────

func (h *APIHandler) ListMedia(ctx context.Context) (gen.ListMediaRes, error) {
	return &gen.MediaListResponse{
		Records: []gen.MediaItem{},
		Total:   0,
	}, nil
}

// ─── deleteMedia ──────────────────────────────────────────────────────────────

func (h *APIHandler) DeleteMedia(ctx context.Context, params gen.DeleteMediaParams) (gen.DeleteMediaRes, error) {
	return ok(), nil
}

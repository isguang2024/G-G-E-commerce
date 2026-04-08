// media.go — Phase 5 ogen handlers for the media domain.
package handlers

import (
	"context"

	"github.com/gg-ecommerce/backend/api/gen"
)

// ─── uploadMedia ──────────────────────────────────────────────────────────────

func (h *APIHandler) UploadMedia(ctx context.Context, req *gen.UploadMediaReq) (gen.UploadMediaRes, error) {
	// Stub: real storage wiring is deferred.
	obj := marshalAnyObject(map[string]interface{}{
		"url": "placeholder",
	})
	return &obj, nil
}

// ─── listMedia ────────────────────────────────────────────────────────────────

func (h *APIHandler) ListMedia(ctx context.Context) (gen.ListMediaRes, error) {
	return &gen.AnyListResponse{
		Records: []gen.AnyObject{},
		Total:   0,
	}, nil
}

// ─── deleteMedia ──────────────────────────────────────────────────────────────

func (h *APIHandler) DeleteMedia(ctx context.Context, params gen.DeleteMediaParams) (gen.DeleteMediaRes, error) {
	return ok(), nil
}

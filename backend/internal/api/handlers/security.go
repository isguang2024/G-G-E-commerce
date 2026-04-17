package handlers

import (
	"context"

	"github.com/maben/backend/api/gen"
)

// SecurityHandler implements gen.SecurityHandler.
//
// Authentication is handled upstream by the Gin JWT middleware, which validates
// the Bearer token and injects user context before the request reaches ogen.
// This handler is therefore a no-op pass-through: ogen calls it for every
// operation that declares `security: [{BearerAuth: []}]`, but by the time we
// get here the token has already been verified.
//
// Do NOT add real JWT validation here — it would double-validate and create
// a maintenance split. Keep all auth logic in the Gin middleware layer.
type SecurityHandler struct{}

// HandleBearerAuth satisfies gen.SecurityHandler. It stores the raw token in
// context so that handlers can read it if needed, then returns nil (success).
func (SecurityHandler) HandleBearerAuth(ctx context.Context, _ gen.OperationName, t gen.BearerAuth) (context.Context, error) {
	// Token already validated by Gin middleware; nothing to do here.
	_ = t.Token
	return ctx, nil
}



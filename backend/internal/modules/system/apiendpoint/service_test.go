package apiendpoint

import (
	"net/http"
	"testing"

	"github.com/google/uuid"

	"github.com/maben/backend/internal/modules/system/user"
)

func TestResolveEndpointCodeForSavePreservesExistingCode(t *testing.T) {
	current := &user.APIEndpoint{
		ID:     uuid.New(),
		Code:   uuid.NewString(),
		Method: http.MethodGet,
		Path:   "/api/v1/menus/tree",
	}
	endpoint := &user.APIEndpoint{
		ID:     current.ID,
		Method: http.MethodGet,
		Path:   "/api/v1/menus/runtime",
	}

	got := resolveEndpointCodeForSave(endpoint, current)
	if got != current.Code {
		t.Fatalf("resolveEndpointCodeForSave() = %q, want %q", got, current.Code)
	}
}

func TestResolveEndpointCodeForSaveDerivesCodeOnCreate(t *testing.T) {
	endpoint := &user.APIEndpoint{
		Method: http.MethodPost,
		Path:   "/api/v1/api-endpoints/sync",
	}

	got := resolveEndpointCodeForSave(endpoint, nil)
	want := deriveStableEndpointCode(endpoint.Method, endpoint.Path)
	if got != want {
		t.Fatalf("resolveEndpointCodeForSave() = %q, want %q", got, want)
	}
}

// TestListRuntimeStates was removed when the Source column was dropped and
// the apiregistry package was folded into this package. The resolveEndpointCodeForSave
// tests above still cover the core stable-code logic.
var _ = http.MethodGet
var _ = user.APIEndpoint{}
var _ = uuid.New


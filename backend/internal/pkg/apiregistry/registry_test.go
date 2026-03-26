package apiregistry

import (
	"testing"

	"github.com/gg-ecommerce/backend/internal/modules/system/models"
)

func TestResolveSyncedEndpointStatus(t *testing.T) {
	tests := []struct {
		name     string
		existing *models.APIEndpoint
		want     string
	}{
		{
			name:     "new endpoint defaults to normal",
			existing: nil,
			want:     "normal",
		},
		{
			name: "existing suspended status is preserved",
			existing: &models.APIEndpoint{
				Status: "suspended",
			},
			want: "suspended",
		},
		{
			name: "blank existing status falls back to normal",
			existing: &models.APIEndpoint{
				Status: "   ",
			},
			want: "normal",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolveSyncedEndpointStatus(tt.existing)
			if got != tt.want {
				t.Fatalf("resolveSyncedEndpointStatus() = %q, want %q", got, tt.want)
			}
		})
	}
}

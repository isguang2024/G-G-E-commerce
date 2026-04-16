package upload

import (
	"testing"

	"github.com/gg-ecommerce/backend/internal/modules/system/models"
)

func TestResolveOSSBucketEndpoints(t *testing.T) {
	bucket := models.StorageBucket{
		PublicBaseURL: "https://cdn.example.com",
		Extra: models.MetaJSON{
			ossBucketExtraEndpointKey:         "https://oss-cn-hangzhou.aliyuncs.com",
			ossBucketExtraInternalEndpointKey: "https://oss-cn-hangzhou-internal.aliyuncs.com",
			ossBucketExtraCDNBaseURLKey:       "https://cdn.example.com",
		},
	}
	endpoints := ResolveOSSBucketEndpoints(bucket)
	if endpoints.APIEndpoint(false) != "https://oss-cn-hangzhou.aliyuncs.com" {
		t.Fatalf("APIEndpoint(false) = %q", endpoints.APIEndpoint(false))
	}
	if endpoints.APIEndpoint(true) != "https://oss-cn-hangzhou-internal.aliyuncs.com" {
		t.Fatalf("APIEndpoint(true) = %q", endpoints.APIEndpoint(true))
	}
	if endpoints.PublicObjectURL("images/demo.png") != "https://cdn.example.com/images/demo.png" {
		t.Fatalf("PublicObjectURL() = %q", endpoints.PublicObjectURL("images/demo.png"))
	}
}

func TestResolveOSSBucketEndpointsFallbacks(t *testing.T) {
	bucket := models.StorageBucket{
		PublicBaseURL: "/uploads/public-media",
		Extra:         models.MetaJSON{},
	}
	endpoints := ResolveOSSBucketEndpoints(bucket)
	if endpoints.APIEndpoint(true) != "/uploads/public-media" {
		t.Fatalf("APIEndpoint(true) fallback = %q", endpoints.APIEndpoint(true))
	}
	if endpoints.PublicObjectURL("demo.png") != "/uploads/public-media/demo.png" {
		t.Fatalf("PublicObjectURL() fallback = %q", endpoints.PublicObjectURL("demo.png"))
	}
}

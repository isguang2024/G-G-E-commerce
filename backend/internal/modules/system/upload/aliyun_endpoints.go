package upload

import (
	"strings"

	"github.com/gg-ecommerce/backend/internal/modules/system/models"
)

const (
	ossBucketExtraEndpointKey         = "endpoint"
	ossBucketExtraInternalEndpointKey = "internal_endpoint"
	ossBucketExtraCDNBaseURLKey       = "cdn_base_url"
)

type OSSBucketEndpoints struct {
	Endpoint         string
	InternalEndpoint string
	CDNBaseURL       string
}

func ResolveOSSBucketEndpoints(bucket models.StorageBucket) OSSBucketEndpoints {
	endpoint := strings.TrimSpace(getMetaString(bucket.Extra, ossBucketExtraEndpointKey, ""))
	internalEndpoint := strings.TrimSpace(getMetaString(bucket.Extra, ossBucketExtraInternalEndpointKey, ""))
	cdnBaseURL := strings.TrimSpace(getMetaString(bucket.Extra, ossBucketExtraCDNBaseURLKey, ""))

	if endpoint == "" {
		endpoint = strings.TrimSpace(bucket.PublicBaseURL)
	}
	if cdnBaseURL == "" {
		cdnBaseURL = strings.TrimSpace(bucket.PublicBaseURL)
	}

	return OSSBucketEndpoints{
		Endpoint:         endpoint,
		InternalEndpoint: internalEndpoint,
		CDNBaseURL:       cdnBaseURL,
	}
}

func (e OSSBucketEndpoints) APIEndpoint(useInternal bool) string {
	if useInternal && strings.TrimSpace(e.InternalEndpoint) != "" {
		return strings.TrimSpace(e.InternalEndpoint)
	}
	return strings.TrimSpace(e.Endpoint)
}

func (e OSSBucketEndpoints) PublicObjectURL(objectKey string) string {
	baseURL := strings.TrimSpace(e.CDNBaseURL)
	if baseURL == "" {
		baseURL = strings.TrimSpace(e.Endpoint)
	}
	if strings.Contains(baseURL, "://") {
		return strings.TrimRight(baseURL, "/") + "/" + strings.TrimLeft(strings.TrimSpace(objectKey), "/")
	}
	return joinURLPath(baseURL, objectKey)
}

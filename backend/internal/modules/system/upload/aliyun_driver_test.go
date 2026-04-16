package upload

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
	"time"

	stsclient "github.com/alibabacloud-go/sts-20150401/v2/client"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"

	"github.com/maben/backend/internal/modules/system/models"
)

type fakeOSSBucketAPI struct {
	existsRequestBucket string
	putRequest          *oss.PutObjectRequest
	deleteRequest       *oss.DeleteObjectRequest
	headRequest         *oss.HeadObjectRequest
}

func (f *fakeOSSBucketAPI) IsBucketExist(_ context.Context, bucket string, _ ...func(*oss.Options)) (bool, error) {
	f.existsRequestBucket = bucket
	return true, nil
}

func (f *fakeOSSBucketAPI) PutObject(_ context.Context, request *oss.PutObjectRequest, _ ...func(*oss.Options)) (*oss.PutObjectResult, error) {
	f.putRequest = request
	return &oss.PutObjectResult{ETag: oss.Ptr(`"etag-value"`)}, nil
}

func (f *fakeOSSBucketAPI) DeleteObject(_ context.Context, request *oss.DeleteObjectRequest, _ ...func(*oss.Options)) (*oss.DeleteObjectResult, error) {
	f.deleteRequest = request
	return &oss.DeleteObjectResult{}, nil
}

func (f *fakeOSSBucketAPI) HeadObject(_ context.Context, request *oss.HeadObjectRequest, _ ...func(*oss.Options)) (*oss.HeadObjectResult, error) {
	f.headRequest = request
	return &oss.HeadObjectResult{
		ContentLength: 12,
		ContentType:   oss.Ptr("image/png"),
		ETag:          oss.Ptr(`"etag-value"`),
	}, nil
}

type fakeSTSAPI struct {
	request *stsclient.AssumeRoleRequest
}

func (f *fakeSTSAPI) AssumeRole(request *stsclient.AssumeRoleRequest) (*stsclient.AssumeRoleResponse, error) {
	f.request = request
	return &stsclient.AssumeRoleResponse{
		Body: &stsclient.AssumeRoleResponseBody{
			Credentials: &stsclient.AssumeRoleResponseBodyCredentials{
				AccessKeyId:     oss.Ptr("STS.KEY"),
				AccessKeySecret: oss.Ptr("sts-secret"),
				SecurityToken:   oss.Ptr("sts-token"),
				Expiration:      oss.Ptr(time.Now().Add(30 * time.Minute).UTC().Format(time.RFC3339)),
			},
		},
	}, nil
}

func TestParseAliyunOSSDriverConfig(t *testing.T) {
	cfg, err := parseAliyunOSSDriverConfig(
		models.StorageProvider{
			Region:             "cn-hangzhou",
			AccessKeyEncrypted: "ak",
			SecretKeyEncrypted: "sk",
			Extra: models.MetaJSON{
				ossProviderExtraUsePathStyleKey:       true,
				ossProviderExtraUseCNameKey:           false,
				ossProviderExtraDisableSSLKey:         true,
				ossProviderExtraConnectTimeoutMsKey:   1500,
				ossProviderExtraReadWriteTimeoutMsKey: "2500",
			},
		},
		models.StorageBucket{
			BucketName:    "media-bucket",
			PublicBaseURL: "https://cdn.example.com",
			Extra: models.MetaJSON{
				ossBucketExtraEndpointKey: "https://oss-cn-hangzhou.aliyuncs.com",
			},
		},
	)
	if err != nil {
		t.Fatalf("parseAliyunOSSDriverConfig() error = %v", err)
	}
	if cfg.ClientConfig.Endpoint != "oss-cn-hangzhou.aliyuncs.com" {
		t.Fatalf("endpoint = %q", cfg.ClientConfig.Endpoint)
	}
	if !cfg.ClientConfig.UsePathStyle || !cfg.ClientConfig.DisableSSL {
		t.Fatalf("client config flags = %+v", cfg.ClientConfig)
	}
	if cfg.PublicBaseURL != "https://cdn.example.com" {
		t.Fatalf("public base url = %q", cfg.PublicBaseURL)
	}
	if cfg.ClientConfig.ConnectTimeout != 1500*time.Millisecond {
		t.Fatalf("connect timeout = %v", cfg.ClientConfig.ConnectTimeout)
	}
	if cfg.ClientConfig.ReadWriteTimeout != 2500*time.Millisecond {
		t.Fatalf("read/write timeout = %v", cfg.ClientConfig.ReadWriteTimeout)
	}
}

func TestAliyunOSSDriverUploadDeleteAndHealthCheck(t *testing.T) {
	fakeClient := &fakeOSSBucketAPI{}
	driver := &aliyunOSSDriver{
		bucketName: "media-bucket",
		bucket: models.StorageBucket{
			PublicBaseURL: "https://cdn.example.com",
		},
		client:          fakeClient,
		publicObjectURL: "https://cdn.example.com",
	}

	if err := driver.HealthCheck(context.Background()); err != nil {
		t.Fatalf("HealthCheck() error = %v", err)
	}
	if fakeClient.existsRequestBucket != "media-bucket" {
		t.Fatalf("health check bucket = %q", fakeClient.existsRequestBucket)
	}

	result, err := driver.Upload(context.Background(), UploadRequest{
		StorageKey:  "2026/04/demo.png",
		ContentType: "image/png",
		Size:        12,
		File:        strings.NewReader("hello world!"),
	})
	if err != nil {
		t.Fatalf("Upload() error = %v", err)
	}
	if got := oss.ToString(fakeClient.putRequest.Bucket); got != "media-bucket" {
		t.Fatalf("upload bucket = %q", got)
	}
	if got := oss.ToString(fakeClient.putRequest.Key); got != "2026/04/demo.png" {
		t.Fatalf("upload key = %q", got)
	}
	if result.URL != "https://cdn.example.com/2026/04/demo.png" {
		t.Fatalf("upload url = %q", result.URL)
	}

	if err := driver.Delete(context.Background(), DeleteRequest{StorageKey: "2026/04/demo.png"}); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if got := oss.ToString(fakeClient.deleteRequest.Key); got != "2026/04/demo.png" {
		t.Fatalf("delete key = %q", got)
	}
}

func TestAliyunOSSDriverPrepareDirectUploadWithSTS(t *testing.T) {
	fakeSTS := &fakeSTSAPI{}
	driver := &aliyunOSSDriver{
		provider: models.StorageProvider{
			AccessKeyEncrypted: "ak",
			SecretKeyEncrypted: "sk",
			Extra: models.MetaJSON{
				ossProviderExtraSTSRoleARNKey:         "acs:ram::1234567890123456:role/upload",
				ossProviderExtraSTSSessionNameKey:     "unit-test",
				ossProviderExtraSTSDurationSecondsKey: 1800,
			},
		},
		bucketName: "media-bucket",
		bucket: models.StorageBucket{
			Extra: models.MetaJSON{
				ossBucketExtraSuccessStatusKey: "201",
			},
		},
		endpoints: OSSBucketEndpoints{Endpoint: "oss-cn-hangzhou.aliyuncs.com"},
		stsClient: fakeSTS,
	}

	result, err := driver.PrepareDirectUpload(context.Background(), DirectUploadRequest{
		StorageKey:  "2026/04/demo.png",
		ContentType: "image/png",
		Size:        1024,
	})
	if err != nil {
		t.Fatalf("PrepareDirectUpload() error = %v", err)
	}
	if result.Method != "POST" {
		t.Fatalf("method = %q", result.Method)
	}
	if result.URL != "https://media-bucket.oss-cn-hangzhou.aliyuncs.com" {
		t.Fatalf("url = %q", result.URL)
	}
	if fakeSTS.request == nil || oss.ToString(fakeSTS.request.RoleArn) == "" {
		t.Fatalf("AssumeRole() was not called")
	}
	if result.Form["OSSAccessKeyId"] != "STS.KEY" {
		t.Fatalf("OSSAccessKeyId = %q", result.Form["OSSAccessKeyId"])
	}
	if result.Form["x-oss-security-token"] != "sts-token" {
		t.Fatalf("security token = %q", result.Form["x-oss-security-token"])
	}
	if result.Form["success_action_status"] != "201" {
		t.Fatalf("success_action_status = %q", result.Form["success_action_status"])
	}

	policyBytes, err := base64.StdEncoding.DecodeString(result.Form["policy"])
	if err != nil {
		t.Fatalf("decode policy: %v", err)
	}
	var payload struct {
		Conditions []any `json:"conditions"`
	}
	if err := json.Unmarshal(policyBytes, &payload); err != nil {
		t.Fatalf("unmarshal policy: %v", err)
	}
	if len(payload.Conditions) == 0 {
		t.Fatalf("policy conditions should not be empty")
	}
}


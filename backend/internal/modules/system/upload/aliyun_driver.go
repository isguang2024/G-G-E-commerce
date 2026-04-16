package upload

import (
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	openapiutil "github.com/alibabacloud-go/darabonba-openapi/v2/utils"
	stsclient "github.com/alibabacloud-go/sts-20150401/v2/client"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	osscredentials "github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss/credentials"
	"go.uber.org/zap"

	"github.com/gg-ecommerce/backend/internal/modules/system/models"
)

const (
	ossProviderExtraUseCNameKey                 = "use_cname"
	ossProviderExtraUsePathStyleKey             = "use_path_style"
	ossProviderExtraInsecureSkipVerifyKey       = "insecure_skip_verify"
	ossProviderExtraDisableSSLKey               = "disable_ssl"
	ossProviderExtraConnectTimeoutMsKey         = "connect_timeout_ms"
	ossProviderExtraReadWriteTimeoutMsKey       = "read_write_timeout_ms"
	ossProviderExtraRetryMaxAttemptsKey         = "retry_max_attempts"
	ossProviderExtraSTSRoleARNKey               = "sts_role_arn"
	ossProviderExtraSTSExternalIDKey            = "sts_external_id"
	ossProviderExtraSTSSessionNameKey           = "sts_session_name"
	ossProviderExtraSTSDurationSecondsKey       = "sts_duration_seconds"
	ossProviderExtraSTSPolicyKey                = "sts_policy"
	ossProviderExtraSTSEndpointKey              = "sts_endpoint"
	ossBucketExtraSuccessStatusKey              = "success_action_status"
	ossBucketExtraContentDispositionKey         = "content_disposition"
	ossBucketExtraCallbackKey                   = "callback"
	ossBucketExtraCallbackVarKey                = "callback_var"
	defaultOSSDirectUploadExpire                = 15 * time.Minute
	defaultOSSDirectUploadSuccessAction         = "200"
	defaultOSSSTSRoleSessionName                = "gge-upload"
	defaultOSSSTSRoleDurationSeconds      int64 = 3600
)

var (
	ErrAliyunOSSBucketNameMissing = errors.New("aliyun oss bucket name is empty")
	ErrAliyunOSSEndpointMissing   = errors.New("aliyun oss endpoint is empty")
	ErrAliyunOSSRegionMissing     = errors.New("aliyun oss region is empty")
	ErrAliyunOSSCredentialsEmpty  = errors.New("aliyun oss access key is empty")
	ErrAliyunSTSRoleARNMissing    = errors.New("aliyun sts role arn is empty")
	ErrAliyunSTSCredentialsEmpty  = errors.New("aliyun sts access key is empty")
)

type ossBucketAPI interface {
	IsBucketExist(ctx context.Context, bucket string, optFns ...func(*oss.Options)) (bool, error)
	PutObject(ctx context.Context, request *oss.PutObjectRequest, optFns ...func(*oss.Options)) (*oss.PutObjectResult, error)
	DeleteObject(ctx context.Context, request *oss.DeleteObjectRequest, optFns ...func(*oss.Options)) (*oss.DeleteObjectResult, error)
	HeadObject(ctx context.Context, request *oss.HeadObjectRequest, optFns ...func(*oss.Options)) (*oss.HeadObjectResult, error)
}

type aliyunSTSAPI interface {
	AssumeRole(request *stsclient.AssumeRoleRequest) (*stsclient.AssumeRoleResponse, error)
}

type aliyunOSSDriver struct {
	provider        models.StorageProvider
	bucket          models.StorageBucket
	bucketName      string
	endpoints       OSSBucketEndpoints
	publicObjectURL string
	client          ossBucketAPI
	stsClient       aliyunSTSAPI
	logger          *zap.Logger
}

type aliyunOSSClientConfig struct {
	Region             string
	Endpoint           string
	AccessKeyID        string
	AccessKeySecret    string
	SecurityToken      string
	UseCName           bool
	UsePathStyle       bool
	InsecureSkipVerify bool
	DisableSSL         bool
	ConnectTimeout     time.Duration
	ReadWriteTimeout   time.Duration
	RetryMaxAttempts   int
}

type aliyunSTSClientConfig struct {
	Region          string
	Endpoint        string
	AccessKeyID     string
	AccessKeySecret string
}

type aliyunOSSTemporaryCredentials struct {
	AccessKeyID     string
	AccessKeySecret string
	SecurityToken   string
	Expiration      time.Time
}

type aliyunOSSClientPool struct {
	mu      sync.RWMutex
	clients map[string]*oss.Client
}

type aliyunSTSClientPool struct {
	mu      sync.RWMutex
	clients map[string]*stsclient.Client
}

var (
	defaultAliyunOSSClientPool = &aliyunOSSClientPool{clients: map[string]*oss.Client{}}
	defaultAliyunSTSClientPool = &aliyunSTSClientPool{clients: map[string]*stsclient.Client{}}
)

func newAliyunOSSDriver(input DriverFactoryInput) (Driver, error) {
	driverCfg, err := parseAliyunOSSDriverConfig(input.Provider, input.Bucket)
	if err != nil {
		return nil, err
	}

	client, err := defaultAliyunOSSClientPool.Open(driverCfg.ClientConfig)
	if err != nil {
		return nil, err
	}

	var stsAPI aliyunSTSAPI
	if driverCfg.HasSTSConfig() {
		stsAPI, err = defaultAliyunSTSClientPool.Open(driverCfg.STSConfig)
		if err != nil {
			return nil, err
		}
	}

	return &aliyunOSSDriver{
		provider:        input.Provider,
		bucket:          input.Bucket,
		bucketName:      driverCfg.BucketName,
		endpoints:       driverCfg.Endpoints,
		publicObjectURL: driverCfg.PublicBaseURL,
		client:          client,
		stsClient:       stsAPI,
		logger:          input.Logger,
	}, nil
}

func (d *aliyunOSSDriver) Name() string {
	return UploadProviderDriverAliyunOSS
}

func (d *aliyunOSSDriver) Capabilities() DriverCapabilities {
	return DriverCapabilities{
		Relay:     true,
		Direct:    true,
		STS:       d.stsClient != nil,
		Multipart: false,
		Delete:    true,
	}
}

func (d *aliyunOSSDriver) HealthCheck(ctx context.Context) error {
	exists, err := d.client.IsBucketExist(ctx, d.bucketName)
	if err != nil {
		return mapAliyunOSSDriverError("health_check", err)
	}
	if !exists {
		return &DriverError{
			Code:      DriverErrorCodeUnavailable,
			Driver:    d.Name(),
			Operation: "health_check",
			Message:   "bucket does not exist",
		}
	}
	return nil
}

func (d *aliyunOSSDriver) Upload(ctx context.Context, req UploadRequest) (*UploadResult, error) {
	request := &oss.PutObjectRequest{
		Bucket:      oss.Ptr(d.bucketName),
		Key:         oss.Ptr(req.StorageKey),
		Body:        req.File,
		ContentType: oss.Ptr(req.ContentType),
	}
	if req.Size > 0 {
		request.ContentLength = oss.Ptr(req.Size)
	}
	if value := strings.TrimSpace(getMetaString(d.bucket.Extra, ossBucketExtraContentDispositionKey, "")); value != "" {
		request.ContentDisposition = oss.Ptr(value)
	}
	if value := strings.TrimSpace(getMetaString(d.bucket.Extra, ossBucketExtraCallbackKey, "")); value != "" {
		request.Callback = oss.Ptr(value)
	}
	if value := strings.TrimSpace(getMetaString(d.bucket.Extra, ossBucketExtraCallbackVarKey, "")); value != "" {
		request.CallbackVar = oss.Ptr(value)
	}

	result, err := d.client.PutObject(ctx, request)
	if err != nil {
		return nil, mapAliyunOSSDriverError("upload", err)
	}

	return &UploadResult{
		StorageKey:  req.StorageKey,
		URL:         d.objectURL(req.StorageKey),
		ContentType: req.ContentType,
		Size:        req.Size,
		Checksum:    strings.TrimSpace(oss.ToString(result.ETag)),
	}, nil
}

func (d *aliyunOSSDriver) Delete(ctx context.Context, req DeleteRequest) error {
	_, err := d.client.DeleteObject(ctx, &oss.DeleteObjectRequest{
		Bucket: oss.Ptr(d.bucketName),
		Key:    oss.Ptr(req.StorageKey),
	})
	if err != nil {
		return mapAliyunOSSDriverError("delete", err)
	}
	return nil
}

func (d *aliyunOSSDriver) PrepareDirectUpload(ctx context.Context, req DirectUploadRequest) (*DirectUploadResult, error) {
	creds, err := d.resolveDirectUploadCredentials(ctx)
	if err != nil {
		return nil, err
	}
	expiresAt := creds.Expiration
	if expiresAt.IsZero() || expiresAt.Before(time.Now()) {
		expiresAt = time.Now().Add(defaultOSSDirectUploadExpire)
	}

	form, err := d.buildPostPolicyForm(req, creds, expiresAt)
	if err != nil {
		return nil, err
	}
	return &DirectUploadResult{
		Method: "POST",
		URL:    buildAliyunOSSUploadURL(d.bucketName, d.endpoints.APIEndpoint(false), getMetaBool(d.provider.Extra, ossProviderExtraUsePathStyleKey, false), getMetaBool(d.provider.Extra, ossProviderExtraUseCNameKey, false), getMetaBool(d.provider.Extra, ossProviderExtraDisableSSLKey, false)),
		Form:   form,
	}, nil
}

func (d *aliyunOSSDriver) StatObject(ctx context.Context, storageKey string) (*ObjectStat, error) {
	result, err := d.client.HeadObject(ctx, &oss.HeadObjectRequest{
		Bucket: oss.Ptr(d.bucketName),
		Key:    oss.Ptr(storageKey),
	})
	if err != nil {
		return nil, mapAliyunOSSDriverError("head_object", err)
	}
	return &ObjectStat{
		StorageKey:  storageKey,
		ContentType: strings.TrimSpace(oss.ToString(result.ContentType)),
		Size:        result.ContentLength,
		Checksum:    strings.Trim(strings.TrimSpace(oss.ToString(result.ETag)), "\""),
	}, nil
}

func (d *aliyunOSSDriver) resolveDirectUploadCredentials(ctx context.Context) (*aliyunOSSTemporaryCredentials, error) {
	if d.stsClient == nil {
		accessKeyID := strings.TrimSpace(d.provider.AccessKeyEncrypted)
		accessKeySecret := strings.TrimSpace(d.provider.SecretKeyEncrypted)
		if accessKeyID == "" || accessKeySecret == "" {
			return nil, ErrAliyunOSSCredentialsEmpty
		}
		return &aliyunOSSTemporaryCredentials{
			AccessKeyID:     accessKeyID,
			AccessKeySecret: accessKeySecret,
			Expiration:      time.Now().Add(defaultOSSDirectUploadExpire),
		}, nil
	}

	roleArn := strings.TrimSpace(getMetaString(d.provider.Extra, ossProviderExtraSTSRoleARNKey, ""))
	if roleArn == "" {
		return nil, ErrAliyunSTSRoleARNMissing
	}
	roleSessionName := strings.TrimSpace(getMetaString(d.provider.Extra, ossProviderExtraSTSSessionNameKey, defaultOSSSTSRoleSessionName))
	if roleSessionName == "" {
		roleSessionName = defaultOSSSTSRoleSessionName
	}
	durationSeconds := getMetaInt64(d.provider.Extra, ossProviderExtraSTSDurationSecondsKey, defaultOSSSTSRoleDurationSeconds)
	request := new(stsclient.AssumeRoleRequest).
		SetRoleArn(roleArn).
		SetRoleSessionName(roleSessionName).
		SetDurationSeconds(durationSeconds)
	if externalID := strings.TrimSpace(getMetaString(d.provider.Extra, ossProviderExtraSTSExternalIDKey, "")); externalID != "" {
		request.SetExternalId(externalID)
	}
	if policy := strings.TrimSpace(getMetaString(d.provider.Extra, ossProviderExtraSTSPolicyKey, "")); policy != "" {
		request.SetPolicy(policy)
	}

	response, err := d.stsClient.AssumeRole(request)
	if err != nil {
		return nil, &DriverError{
			Code:      DriverErrorCodeUnavailable,
			Driver:    d.Name(),
			Operation: "assume_role",
			Message:   "request aliyun sts credentials failed",
			Err:       err,
			Retryable: true,
		}
	}
	body := response.Body
	if body == nil || body.Credentials == nil {
		return nil, &DriverError{
			Code:      DriverErrorCodeUnavailable,
			Driver:    d.Name(),
			Operation: "assume_role",
			Message:   "aliyun sts response does not contain credentials",
		}
	}
	expiration, _ := time.Parse(time.RFC3339, oss.ToString(body.Credentials.Expiration))
	return &aliyunOSSTemporaryCredentials{
		AccessKeyID:     oss.ToString(body.Credentials.AccessKeyId),
		AccessKeySecret: oss.ToString(body.Credentials.AccessKeySecret),
		SecurityToken:   oss.ToString(body.Credentials.SecurityToken),
		Expiration:      expiration,
	}, nil
}

func (d *aliyunOSSDriver) buildPostPolicyForm(req DirectUploadRequest, creds *aliyunOSSTemporaryCredentials, expiresAt time.Time) (map[string]string, error) {
	successStatus := strings.TrimSpace(getMetaString(d.bucket.Extra, ossBucketExtraSuccessStatusKey, defaultOSSDirectUploadSuccessAction))
	conditions := make([]any, 0, 6)
	conditions = append(conditions, map[string]string{"bucket": d.bucketName})
	conditions = append(conditions, map[string]string{"key": req.StorageKey})
	if req.Size > 0 {
		conditions = append(conditions, []any{"content-length-range", 0, req.Size})
	}
	if req.ContentType != "" {
		conditions = append(conditions, map[string]string{"Content-Type": req.ContentType})
	}
	conditions = append(conditions, map[string]string{"success_action_status": successStatus})
	if creds.SecurityToken != "" {
		conditions = append(conditions, map[string]string{"x-oss-security-token": creds.SecurityToken})
	}

	payload := map[string]any{
		"expiration": expiresAt.UTC().Format(time.RFC3339),
		"conditions": conditions,
	}
	policyJSON, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	policy := base64.StdEncoding.EncodeToString(policyJSON)
	mac := hmac.New(sha1.New, []byte(creds.AccessKeySecret))
	_, _ = io.WriteString(mac, policy)
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	form := map[string]string{
		"key":                   req.StorageKey,
		"policy":                policy,
		"OSSAccessKeyId":        creds.AccessKeyID,
		"Signature":             signature,
		"success_action_status": successStatus,
	}
	if req.ContentType != "" {
		form["Content-Type"] = req.ContentType
	}
	if creds.SecurityToken != "" {
		form["x-oss-security-token"] = creds.SecurityToken
	}
	if callback := strings.TrimSpace(getMetaString(d.bucket.Extra, ossBucketExtraCallbackKey, "")); callback != "" {
		form["x-oss-callback"] = callback
	}
	if callbackVar := strings.TrimSpace(getMetaString(d.bucket.Extra, ossBucketExtraCallbackVarKey, "")); callbackVar != "" {
		form["x-oss-callback-var"] = callbackVar
	}
	return form, nil
}

func (d *aliyunOSSDriver) objectURL(storageKey string) string {
	base := strings.TrimSpace(d.publicObjectURL)
	if base == "" {
		return d.endpoints.PublicObjectURL(storageKey)
	}
	if strings.Contains(base, "://") {
		return strings.TrimRight(base, "/") + "/" + strings.TrimLeft(storageKey, "/")
	}
	return joinURLPath(base, storageKey)
}

type aliyunOSSDriverConfig struct {
	BucketName    string
	Endpoints     OSSBucketEndpoints
	PublicBaseURL string
	ClientConfig  aliyunOSSClientConfig
	STSConfig     aliyunSTSClientConfig
}

func (c aliyunOSSDriverConfig) HasSTSConfig() bool {
	return strings.TrimSpace(c.STSConfig.AccessKeyID) != "" &&
		strings.TrimSpace(c.STSConfig.AccessKeySecret) != "" &&
		strings.TrimSpace(c.STSConfig.Region) != ""
}

func parseAliyunOSSDriverConfig(provider models.StorageProvider, bucket models.StorageBucket) (aliyunOSSDriverConfig, error) {
	bucketName := strings.TrimSpace(bucket.BucketName)
	if bucketName == "" {
		return aliyunOSSDriverConfig{}, ErrAliyunOSSBucketNameMissing
	}
	region := strings.TrimSpace(provider.Region)
	if region == "" {
		return aliyunOSSDriverConfig{}, ErrAliyunOSSRegionMissing
	}
	endpoints := ResolveOSSBucketEndpoints(bucket)
	apiEndpoint := endpoints.APIEndpoint(false)
	if strings.TrimSpace(apiEndpoint) == "" {
		return aliyunOSSDriverConfig{}, ErrAliyunOSSEndpointMissing
	}
	normalizedEndpoint, disableSSL, err := normalizeAliyunEndpoint(apiEndpoint)
	if err != nil {
		return aliyunOSSDriverConfig{}, err
	}

	accessKeyID := strings.TrimSpace(provider.AccessKeyEncrypted)
	accessKeySecret := strings.TrimSpace(provider.SecretKeyEncrypted)
	if accessKeyID == "" || accessKeySecret == "" {
		return aliyunOSSDriverConfig{}, ErrAliyunOSSCredentialsEmpty
	}

	clientCfg := aliyunOSSClientConfig{
		Region:             region,
		Endpoint:           normalizedEndpoint,
		AccessKeyID:        accessKeyID,
		AccessKeySecret:    accessKeySecret,
		UseCName:           getMetaBool(provider.Extra, ossProviderExtraUseCNameKey, false),
		UsePathStyle:       getMetaBool(provider.Extra, ossProviderExtraUsePathStyleKey, false),
		InsecureSkipVerify: getMetaBool(provider.Extra, ossProviderExtraInsecureSkipVerifyKey, false),
		DisableSSL:         disableSSL || getMetaBool(provider.Extra, ossProviderExtraDisableSSLKey, false),
		ConnectTimeout:     time.Duration(getMetaInt64(provider.Extra, ossProviderExtraConnectTimeoutMsKey, 0)) * time.Millisecond,
		ReadWriteTimeout:   time.Duration(getMetaInt64(provider.Extra, ossProviderExtraReadWriteTimeoutMsKey, 0)) * time.Millisecond,
		RetryMaxAttempts:   int(getMetaInt64(provider.Extra, ossProviderExtraRetryMaxAttemptsKey, 0)),
	}

	stsEndpoint := strings.TrimSpace(getMetaString(provider.Extra, ossProviderExtraSTSEndpointKey, ""))
	if stsEndpoint == "" {
		stsEndpoint = "sts.aliyuncs.com"
	}
	normalizedSTSEndpoint, _, err := normalizeAliyunEndpoint(stsEndpoint)
	if err != nil {
		return aliyunOSSDriverConfig{}, err
	}

	return aliyunOSSDriverConfig{
		BucketName: bucketName,
		Endpoints:  endpoints,
		PublicBaseURL: strings.TrimSpace(func() string {
			if endpoints.CDNBaseURL != "" {
				return endpoints.CDNBaseURL
			}
			if bucket.PublicBaseURL != "" {
				return bucket.PublicBaseURL
			}
			return buildAliyunOSSPublicBaseURL(bucketName, normalizedEndpoint, clientCfg.UsePathStyle, clientCfg.UseCName, clientCfg.DisableSSL)
		}()),
		ClientConfig: clientCfg,
		STSConfig: aliyunSTSClientConfig{
			Region:          region,
			Endpoint:        normalizedSTSEndpoint,
			AccessKeyID:     accessKeyID,
			AccessKeySecret: accessKeySecret,
		},
	}, nil
}

func (p *aliyunOSSClientPool) Open(cfg aliyunOSSClientConfig) (*oss.Client, error) {
	if strings.TrimSpace(cfg.Region) == "" {
		return nil, ErrAliyunOSSRegionMissing
	}
	if strings.TrimSpace(cfg.Endpoint) == "" {
		return nil, ErrAliyunOSSEndpointMissing
	}
	if strings.TrimSpace(cfg.AccessKeyID) == "" || strings.TrimSpace(cfg.AccessKeySecret) == "" {
		return nil, ErrAliyunOSSCredentialsEmpty
	}

	cacheKey := strings.Join([]string{
		cfg.Region,
		cfg.Endpoint,
		cfg.AccessKeyID,
		cfg.AccessKeySecret,
		cfg.SecurityToken,
		strconv.FormatBool(cfg.UseCName),
		strconv.FormatBool(cfg.UsePathStyle),
		strconv.FormatBool(cfg.InsecureSkipVerify),
		strconv.FormatBool(cfg.DisableSSL),
		cfg.ConnectTimeout.String(),
		cfg.ReadWriteTimeout.String(),
		strconv.Itoa(cfg.RetryMaxAttempts),
	}, "|")

	p.mu.RLock()
	cached := p.clients[cacheKey]
	p.mu.RUnlock()
	if cached != nil {
		return cached, nil
	}

	ossCfg := oss.LoadDefaultConfig().
		WithRegion(cfg.Region).
		WithEndpoint(cfg.Endpoint).
		WithCredentialsProvider(osscredentials.CredentialsProviderFunc(func(context.Context) (osscredentials.Credentials, error) {
			return osscredentials.Credentials{
				AccessKeyID:     cfg.AccessKeyID,
				AccessKeySecret: cfg.AccessKeySecret,
				SecurityToken:   cfg.SecurityToken,
			}, nil
		})).
		WithUseCName(cfg.UseCName).
		WithUsePathStyle(cfg.UsePathStyle).
		WithInsecureSkipVerify(cfg.InsecureSkipVerify).
		WithDisableSSL(cfg.DisableSSL)
	if cfg.ConnectTimeout > 0 {
		ossCfg.WithConnectTimeout(cfg.ConnectTimeout)
	}
	if cfg.ReadWriteTimeout > 0 {
		ossCfg.WithReadWriteTimeout(cfg.ReadWriteTimeout)
	}
	if cfg.RetryMaxAttempts > 0 {
		ossCfg.WithRetryMaxAttempts(cfg.RetryMaxAttempts)
	}
	client := oss.NewClient(ossCfg)

	p.mu.Lock()
	defer p.mu.Unlock()
	if cached = p.clients[cacheKey]; cached != nil {
		return cached, nil
	}
	p.clients[cacheKey] = client
	return client, nil
}

func (p *aliyunSTSClientPool) Open(cfg aliyunSTSClientConfig) (*stsclient.Client, error) {
	if strings.TrimSpace(cfg.Region) == "" {
		return nil, ErrAliyunOSSRegionMissing
	}
	if strings.TrimSpace(cfg.AccessKeyID) == "" || strings.TrimSpace(cfg.AccessKeySecret) == "" {
		return nil, ErrAliyunSTSCredentialsEmpty
	}
	cacheKey := strings.Join([]string{cfg.Region, cfg.Endpoint, cfg.AccessKeyID, cfg.AccessKeySecret}, "|")

	p.mu.RLock()
	cached := p.clients[cacheKey]
	p.mu.RUnlock()
	if cached != nil {
		return cached, nil
	}

	openCfg := new(openapiutil.Config).
		SetRegionId(cfg.Region).
		SetAccessKeyId(cfg.AccessKeyID).
		SetAccessKeySecret(cfg.AccessKeySecret).
		SetEndpoint(cfg.Endpoint).
		SetProtocol("HTTPS")
	client, err := stsclient.NewClient(openCfg)
	if err != nil {
		return nil, err
	}

	p.mu.Lock()
	defer p.mu.Unlock()
	if cached = p.clients[cacheKey]; cached != nil {
		return cached, nil
	}
	p.clients[cacheKey] = client
	return client, nil
}

func normalizeAliyunEndpoint(raw string) (string, bool, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return "", false, ErrAliyunOSSEndpointMissing
	}
	if !strings.Contains(value, "://") {
		return strings.TrimSuffix(value, "/"), false, nil
	}
	parsed, err := url.Parse(value)
	if err != nil {
		return "", false, err
	}
	host := strings.TrimSpace(parsed.Host)
	if host == "" {
		host = strings.TrimSpace(parsed.Path)
	}
	if host == "" {
		return "", false, ErrAliyunOSSEndpointMissing
	}
	return strings.TrimSuffix(host, "/"), strings.EqualFold(parsed.Scheme, "http"), nil
}

func buildAliyunOSSPublicBaseURL(bucketName, endpoint string, usePathStyle, useCName, disableSSL bool) string {
	return buildAliyunOSSUploadURL(bucketName, endpoint, usePathStyle, useCName, disableSSL)
}

func buildAliyunOSSUploadURL(bucketName, endpoint string, usePathStyle, useCName, disableSSL bool) string {
	host := strings.TrimSpace(endpoint)
	host = strings.TrimPrefix(host, "https://")
	host = strings.TrimPrefix(host, "http://")
	scheme := "https"
	if disableSSL {
		scheme = "http"
	}
	host = strings.TrimRight(host, "/")
	switch {
	case usePathStyle:
		return fmt.Sprintf("%s://%s/%s", scheme, host, strings.Trim(bucketName, "/"))
	case useCName:
		return fmt.Sprintf("%s://%s", scheme, host)
	case strings.HasPrefix(host, bucketName+"."):
		return fmt.Sprintf("%s://%s", scheme, host)
	default:
		return fmt.Sprintf("%s://%s.%s", scheme, bucketName, host)
	}
}

func getMetaBool(meta models.MetaJSON, key string, fallback bool) bool {
	if meta == nil {
		return fallback
	}
	raw, ok := meta[key]
	if !ok || raw == nil {
		return fallback
	}
	switch value := raw.(type) {
	case bool:
		return value
	case string:
		parsed, err := strconv.ParseBool(strings.TrimSpace(value))
		if err == nil {
			return parsed
		}
	case float64:
		return value != 0
	case int:
		return value != 0
	case int64:
		return value != 0
	}
	return fallback
}

func getMetaInt64(meta models.MetaJSON, key string, fallback int64) int64 {
	if meta == nil {
		return fallback
	}
	raw, ok := meta[key]
	if !ok || raw == nil {
		return fallback
	}
	switch value := raw.(type) {
	case int:
		return int64(value)
	case int64:
		return value
	case float64:
		return int64(value)
	case string:
		parsed, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
		if err == nil {
			return parsed
		}
	}
	return fallback
}

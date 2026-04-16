package upload

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/maben/backend/internal/config"
	"github.com/maben/backend/internal/modules/system/models"
)

var (
	ErrInvalidFile        = errors.New("invalid upload file")
	ErrRecordNotFound     = errors.New("upload record not found")
	ErrUploadRecordExists = errors.New("upload record already exists")
)

type Service interface {
	EnsureDefaultSeeds(ctx context.Context) error
	Upload(ctx context.Context, tenantID string, userID *uuid.UUID, input UploadInput) (*models.UploadRecord, error)
	PrepareUpload(ctx context.Context, tenantID string, input PrepareUploadInput) (*PrepareUploadResult, error)
	CompleteDirectUpload(ctx context.Context, tenantID string, userID *uuid.UUID, input CompleteDirectUploadInput) (*models.UploadRecord, error)
	List(ctx context.Context, tenantID string, limit int) ([]models.UploadRecord, int64, error)
	Delete(ctx context.Context, tenantID string, id uuid.UUID) error
	Admin() AdminAPI
}

type UploadInput struct {
	Key      string
	Rule     string
	Name     string
	File     io.Reader
	Size     int64
	MimeType string
	Checksum string
}

type PrepareUploadInput struct {
	Key      string
	Rule     string
	Name     string
	Size     int64
	MimeType string
	Checksum string
}

type PrepareUploadResult struct {
	Mode         string
	Method       string
	URL          string
	RelayURL     string
	Headers      map[string]string
	Form         map[string]string
	StorageKey   string
	Filename     string
	ContentType  string
	UploadKey    string
	RuleKey      string
	FallbackUsed bool
}

type CompleteDirectUploadInput struct {
	Key        string
	Rule       string
	Name       string
	StorageKey string
	Size       int64
	MimeType   string
	Checksum   string
	ETag       string
}

type service struct {
	repo       *Repository
	cfg        *config.Config
	logger     *zap.Logger
	localRoot  string
	publicBase string
	registry   *DriverRegistry
}

type uploadSeedSpec struct {
	UploadKey models.UploadKey
	Rules     []models.UploadKeyRule
}

func NewService(repo *Repository, cfg *config.Config, logger *zap.Logger) Service {
	root := filepath.Clean(filepath.Join(".", "data", "uploads"))
	publicBase := "/uploads"
	if cfg != nil {
		if configuredRoot := strings.TrimSpace(cfg.Upload.LocalRoot); configuredRoot != "" {
			root = filepath.Clean(configuredRoot)
		}
		if configuredBase := strings.TrimSpace(cfg.Upload.PublicBaseURL); configuredBase != "" {
			publicBase = configuredBase
		}
		setDefaultTenantID(strings.TrimSpace(cfg.Upload.DefaultTenantID))
	}
	if repo != nil && repo.cipher == nil && cfg != nil {
		cipher, err := NewSecretCipher(cfg.Upload)
		if err == nil {
			repo = repo.WithSecretCipher(cipher)
		} else if !errors.Is(err, ErrSecretCipherUnavailable) && logger != nil {
			logger.Warn("initialize upload secret cipher failed", zap.Error(err))
		}
	}
	if repo != nil && cfg != nil {
		repo = repo.WithDefaultUploadKey(cfg.Upload.DefaultUploadKey)
		cacheLayer, err := newResolvedConfigCache(cfg, logger)
		if err == nil {
			repo = repo.WithResolvedConfigCache(cacheLayer)
		} else if logger != nil {
			logger.Warn("initialize upload config cache failed", zap.Error(err))
		}
	}
	registry := NewDriverRegistry()
	registry.MustRegister(models.UploadProviderDriverLocal, func(input DriverFactoryInput) (Driver, error) {
		return NewLocalDriver(LocalDriverConfig{
			RootPath:      root,
			TempDirectory: filepath.Join(root, ".tmp"),
			Logger:        input.Logger,
		}), nil
	})
	registry.MustRegister(UploadProviderDriverAliyunOSS, func(input DriverFactoryInput) (Driver, error) {
		return newAliyunOSSDriver(input)
	})
	return &service{
		repo:       repo,
		cfg:        cfg,
		logger:     logger,
		localRoot:  root,
		publicBase: publicBase,
		registry:   registry,
	}
}

func EnsureDefaultSeeds(ctx context.Context, dbRepo *Repository, logger *zap.Logger) error {
	return NewService(dbRepo, nil, logger).EnsureDefaultSeeds(ctx)
}

func (s *service) EnsureDefaultSeeds(ctx context.Context) error {
	defaultProviderKey := "local-public"
	defaultBucketKey := "public-media"
	defaultUploadKey := "media.default"
	defaultRuleKey := "image"
	defaultMaxFileSize := int64(10 * 1024 * 1024)
	if s.cfg != nil {
		if value := strings.TrimSpace(s.cfg.Upload.DefaultProviderKey); value != "" {
			defaultProviderKey = value
		}
		if value := strings.TrimSpace(s.cfg.Upload.DefaultBucketKey); value != "" {
			defaultBucketKey = value
		}
		if value := strings.TrimSpace(s.cfg.Upload.DefaultUploadKey); value != "" {
			defaultUploadKey = value
		}
		if value := strings.TrimSpace(s.cfg.Upload.DefaultRuleKey); value != "" {
			defaultRuleKey = value
		}
		if value := s.cfg.Upload.MaxFileSizeBytes; value > 0 {
			defaultMaxFileSize = value
		}
	}

	provider := &models.StorageProvider{
		TenantID:    defaultTenantID,
		ProviderKey: defaultProviderKey,
		Name:        "默认本地公共存储",
		Driver:      models.UploadProviderDriverLocal,
		BaseURL:     s.publicBase,
		IsDefault:   true,
		Status:      models.UploadProviderStatusReady,
	}
	if err := s.repo.EnsureProvider(ctx, provider); err != nil {
		return err
	}

	storedProvider, err := s.repo.GetProviderByKey(ctx, defaultTenantID, provider.ProviderKey)
	if err != nil {
		return err
	}

	bucket := &models.StorageBucket{
		TenantID:      defaultTenantID,
		ProviderID:    storedProvider.ID,
		BucketKey:     defaultBucketKey,
		Name:          "公共媒体桶",
		BucketName:    defaultBucketKey,
		BasePath:      defaultBucketKey,
		PublicBaseURL: joinURLPath(s.publicBase, defaultBucketKey),
		IsPublic:      true,
		Status:        models.UploadProviderStatusReady,
	}
	if err := s.repo.EnsureBucket(ctx, bucket); err != nil {
		return err
	}

	storedBucket, err := s.repo.GetBucketByKey(ctx, defaultTenantID, bucket.BucketKey)
	if err != nil {
		return err
	}

	key := &models.UploadKey{
		TenantID:         defaultTenantID,
		BucketID:         storedBucket.ID,
		Key:              defaultUploadKey,
		Name:             "默认媒体上传",
		PathTemplate:     "{yyyy}/{mm}/{dd}",
		DefaultRuleKey:   defaultRuleKey,
		MaxSizeBytes:     defaultMaxFileSize,
		AllowedMimeTypes: models.StringList{"image/jpeg", "image/png", "image/webp", "image/gif", "image/svg+xml"},
		Visibility:       "public",
		Status:           models.UploadProviderStatusReady,
	}
	if err := s.repo.EnsureUploadKey(ctx, key); err != nil {
		return err
	}

	storedKey, err := s.repo.GetUploadKeyByKey(ctx, defaultTenantID, key.Key)
	if err != nil {
		return err
	}

	rule := &models.UploadKeyRule{
		TenantID:         defaultTenantID,
		UploadKeyID:      storedKey.ID,
		RuleKey:          defaultRuleKey,
		Name:             "默认图片规则",
		SubPath:          "images",
		FilenameStrategy: "uuid",
		MaxSizeBytes:     defaultMaxFileSize,
		AllowedMimeTypes: models.StringList{"image/jpeg", "image/png", "image/webp", "image/gif", "image/svg+xml"},
		ProcessPipeline:  models.StringList{},
		IsDefault:        true,
		Status:           models.UploadProviderStatusReady,
	}
	if err := s.repo.EnsureUploadRule(ctx, rule); err != nil {
		return err
	}

	// 内置上传 seed：只保留脚手架默认会用到的上传场景。
	seeds := []uploadSeedSpec{
		{
			UploadKey: models.UploadKey{
				Key:              "user.avatar",
				Name:             "用户头像上传",
				PathTemplate:     "avatars/{yyyy}/{mm}",
				DefaultRuleKey:   "avatar",
				MaxSizeBytes:     2 * 1024 * 1024,
				AllowedMimeTypes: models.StringList{"image/jpeg", "image/png", "image/webp"},
				Visibility:       "public",
			},
			Rules: []models.UploadKeyRule{
				{
					RuleKey:          "avatar",
					Name:             "头像图片规则",
					SubPath:          "avatar",
					FilenameStrategy: "uuid",
					MaxSizeBytes:     2 * 1024 * 1024,
					AllowedMimeTypes: models.StringList{"image/jpeg", "image/png", "image/webp"},
					ProcessPipeline:  models.StringList{},
					IsDefault:        true,
				},
			},
		},
		{
			UploadKey: models.UploadKey{
				Key:              "doc.attachment",
				Name:             "文档附件上传",
				PathTemplate:     "docs/{yyyy}/{mm}",
				DefaultRuleKey:   "pdf",
				MaxSizeBytes:     50 * 1024 * 1024,
				AllowedMimeTypes: models.StringList{"application/pdf", "application/msword", "application/vnd.openxmlformats-officedocument.wordprocessingml.document", "application/vnd.ms-excel", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"},
				Visibility:       "private",
			},
			Rules: []models.UploadKeyRule{
				{
					RuleKey:          "pdf",
					Name:             "PDF 文档规则",
					SubPath:          "pdf",
					FilenameStrategy: "original",
					MaxSizeBytes:     50 * 1024 * 1024,
					AllowedMimeTypes: models.StringList{"application/pdf"},
					ProcessPipeline:  models.StringList{},
					IsDefault:        true,
				},
				{
					RuleKey:          "office",
					Name:             "Office 文档规则",
					SubPath:          "office",
					FilenameStrategy: "original",
					MaxSizeBytes:     30 * 1024 * 1024,
					AllowedMimeTypes: models.StringList{"application/msword", "application/vnd.openxmlformats-officedocument.wordprocessingml.document", "application/vnd.ms-excel", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"},
					ProcessPipeline:  models.StringList{},
					IsDefault:        false,
				},
			},
		},
		{
			UploadKey: models.UploadKey{
				Key:              "editor.inline",
				Name:             "富文本编辑器图片",
				PathTemplate:     "editor/{yyyy}/{mm}/{dd}",
				DefaultRuleKey:   "editor-image",
				MaxSizeBytes:     5 * 1024 * 1024,
				AllowedMimeTypes: models.StringList{"image/jpeg", "image/png", "image/webp", "image/gif"},
				Visibility:       "public",
			},
			Rules: []models.UploadKeyRule{
				{
					RuleKey:          "editor-image",
					Name:             "编辑器图片规则",
					SubPath:          "img",
					FilenameStrategy: "uuid",
					MaxSizeBytes:     5 * 1024 * 1024,
					AllowedMimeTypes: models.StringList{"image/jpeg", "image/png", "image/webp", "image/gif"},
					ProcessPipeline:  models.StringList{},
					IsDefault:        true,
				},
			},
		},
	}
	for _, seed := range seeds {
		if err := s.ensureUploadSeed(ctx, storedBucket.ID, seed); err != nil {
			return err
		}
	}

	return os.MkdirAll(filepath.Join(s.localRoot, storedBucket.BasePath), 0o755)
}

func (s *service) ensureUploadSeed(ctx context.Context, bucketID uuid.UUID, spec uploadSeedSpec) error {
	key := spec.UploadKey
	key.TenantID = defaultTenantID
	key.BucketID = bucketID
	if strings.TrimSpace(key.Status) == "" {
		key.Status = models.UploadProviderStatusReady
	}
	if err := s.repo.EnsureUploadKey(ctx, &key); err != nil {
		return err
	}

	storedKey, err := s.repo.GetUploadKeyByKey(ctx, defaultTenantID, key.Key)
	if err != nil {
		return err
	}

	for _, ruleSpec := range spec.Rules {
		rule := ruleSpec
		rule.TenantID = defaultTenantID
		rule.UploadKeyID = storedKey.ID
		if strings.TrimSpace(rule.Status) == "" {
			rule.Status = models.UploadProviderStatusReady
		}
		if err := s.repo.EnsureUploadRule(ctx, &rule); err != nil {
			return err
		}
	}
	return nil
}

func (s *service) Upload(ctx context.Context, tenantID string, userID *uuid.UUID, input UploadInput) (*models.UploadRecord, error) {
	if input.File == nil || strings.TrimSpace(input.Name) == "" {
		return nil, ErrInvalidFile
	}

	effective, err := s.repo.ResolveEffectiveConfig(ctx, tenantID, ResolveConfigInput{
		Key:      input.Key,
		Rule:     input.Rule,
		Fallback: true,
	})
	if err != nil {
		return nil, err
	}
	driver, err := s.registry.Open(DriverFactoryInput{
		Config:   s.cfg,
		Provider: effective.Provider,
		Bucket:   effective.Bucket,
		Logger:   s.logger,
	})
	if err != nil {
		return nil, err
	}

	contentType := normalizeContentType(input.MimeType, input.Name)
	if !mimeAllowed(contentType, effective.AllowedMimeTypes, nil) {
		return nil, fmt.Errorf("mime type %s is not allowed", contentType)
	}
	maxSize := effective.MaxSizeBytes
	if maxSize > 0 && input.Size > maxSize {
		return nil, fmt.Errorf("file too large: %d > %d", input.Size, maxSize)
	}

	storageKey, publicPath, storedName := planObjectKey(effective, input.Name, contentType, time.Now())
	uploadResult, err := driver.Upload(ctx, UploadRequest{
		TenantID:         normalizeTenantID(tenantID),
		StorageKey:       storageKey,
		PublicPath:       publicPath,
		OriginalFilename: input.Name,
		ContentType:      contentType,
		Size:             input.Size,
		File:             input.File,
		PublicBaseURL:    effective.Bucket.PublicBaseURL,
	})
	if err != nil {
		return nil, err
	}

	record := buildUploadRecord(effective, userID, input.Name, storedName, uploadResult.StorageKey, uploadResult.URL, uploadResult.ContentType, uploadResult.Size, firstNonEmpty(uploadResult.Checksum, input.Checksum), "relay", models.MetaJSON{
		"source": "relay",
	})
	if err := s.repo.CreateRecord(ctx, record); err != nil {
		return nil, err
	}
	return record, nil
}

func (s *service) PrepareUpload(ctx context.Context, tenantID string, input PrepareUploadInput) (*PrepareUploadResult, error) {
	if strings.TrimSpace(input.Name) == "" || input.Size < 0 {
		return nil, ErrInvalidFile
	}

	effective, err := s.repo.ResolveEffectiveConfig(ctx, tenantID, ResolveConfigInput{
		Key:      input.Key,
		Rule:     input.Rule,
		Fallback: true,
	})
	if err != nil {
		return nil, err
	}
	driver, err := s.registry.Open(DriverFactoryInput{
		Config:   s.cfg,
		Provider: effective.Provider,
		Bucket:   effective.Bucket,
		Logger:   s.logger,
	})
	if err != nil {
		return nil, err
	}

	contentType := normalizeContentType(input.MimeType, input.Name)
	if !mimeAllowed(contentType, effective.AllowedMimeTypes, nil) {
		return nil, fmt.Errorf("mime type %s is not allowed", contentType)
	}
	maxSize := effective.MaxSizeBytes
	if maxSize > 0 && input.Size > maxSize {
		return nil, fmt.Errorf("file too large: %d > %d", input.Size, maxSize)
	}

	storageKey, _, storedName := planObjectKey(effective, input.Name, contentType, time.Now())
	result := &PrepareUploadResult{
		StorageKey:   storageKey,
		Filename:     storedName,
		ContentType:  contentType,
		UploadKey:    effective.ResolvedKey,
		RuleKey:      effective.ResolvedRule,
		FallbackUsed: effective.FallbackUsed,
	}
	if driver.Capabilities().Direct {
		directResult, directErr := driver.PrepareDirectUpload(ctx, DirectUploadRequest{
			TenantID:    normalizeTenantID(tenantID),
			StorageKey:  storageKey,
			ContentType: contentType,
			Size:        input.Size,
		})
		if directErr == nil {
			result.Mode = "direct"
			result.Method = directResult.Method
			result.URL = directResult.URL
			result.Headers = directResult.Headers
			result.Form = directResult.Form
			return result, nil
		}
		var driverErr *DriverError
		if !errors.As(directErr, &driverErr) || driverErr.Code != DriverErrorCodeCapabilityUnsupported {
			return nil, directErr
		}
	}

	result.Mode = "relay"
	result.RelayURL = "/api/v1/media/upload"
	return result, nil
}

func (s *service) CompleteDirectUpload(ctx context.Context, tenantID string, userID *uuid.UUID, input CompleteDirectUploadInput) (*models.UploadRecord, error) {
	if strings.TrimSpace(input.Name) == "" || strings.TrimSpace(input.StorageKey) == "" {
		return nil, ErrInvalidFile
	}
	if existing, err := s.repo.GetRecordByStorageKey(ctx, tenantID, input.StorageKey); err == nil {
		return existing, nil
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	effective, err := s.repo.ResolveEffectiveConfig(ctx, tenantID, ResolveConfigInput{
		Key:      input.Key,
		Rule:     input.Rule,
		Fallback: true,
	})
	if err != nil {
		return nil, err
	}
	normalizedStorageKey, storedName, err := validateStorageKeyForConfig(input.StorageKey, effective.Bucket.BasePath)
	if err != nil {
		return nil, err
	}

	driver, err := s.registry.Open(DriverFactoryInput{
		Config:   s.cfg,
		Provider: effective.Provider,
		Bucket:   effective.Bucket,
		Logger:   s.logger,
	})
	if err != nil {
		return nil, err
	}

	contentType := normalizeContentType(input.MimeType, input.Name)
	size := input.Size
	checksum := strings.TrimSpace(input.Checksum)
	if statProvider, ok := driver.(ObjectStatProvider); ok {
		stat, statErr := statProvider.StatObject(ctx, normalizedStorageKey)
		if statErr != nil {
			return nil, statErr
		}
		if size > 0 && stat.Size > 0 && size != stat.Size {
			return nil, fmt.Errorf("direct upload size mismatch: %d != %d", size, stat.Size)
		}
		if etag := strings.Trim(strings.TrimSpace(input.ETag), "\""); etag != "" && stat.Checksum != "" && !strings.EqualFold(etag, strings.Trim(stat.Checksum, "\"")) {
			return nil, fmt.Errorf("direct upload etag mismatch")
		}
		if stat.Size > 0 {
			size = stat.Size
		}
		if strings.TrimSpace(stat.ContentType) != "" {
			contentType = stat.ContentType
		}
		if strings.TrimSpace(stat.Checksum) != "" {
			checksum = stat.Checksum
		}
	}
	if !mimeAllowed(contentType, effective.AllowedMimeTypes, nil) {
		return nil, fmt.Errorf("mime type %s is not allowed", contentType)
	}
	maxSize := effective.MaxSizeBytes
	if maxSize > 0 && size > maxSize {
		return nil, fmt.Errorf("file too large: %d > %d", size, maxSize)
	}

	record := buildUploadRecord(effective, userID, input.Name, storedName, normalizedStorageKey, resolveObjectURL(driver, effective.Bucket, normalizedStorageKey), contentType, size, checksum, "direct", models.MetaJSON{
		"source": "direct",
		"etag":   strings.Trim(strings.TrimSpace(input.ETag), "\""),
	})
	if err := s.repo.CreateRecord(ctx, record); err != nil {
		return nil, err
	}
	return record, nil
}

func (s *service) List(ctx context.Context, tenantID string, limit int) ([]models.UploadRecord, int64, error) {
	items, err := s.repo.ListRecords(ctx, tenantID, limit)
	if err != nil {
		return nil, 0, err
	}
	total, err := s.repo.CountRecords(ctx, tenantID)
	if err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (s *service) Delete(ctx context.Context, tenantID string, id uuid.UUID) error {
	item, err := s.repo.GetRecord(ctx, tenantID, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrRecordNotFound
		}
		return err
	}
	driver, openErr := s.registry.Open(DriverFactoryInput{
		Config: s.cfg,
		Provider: models.StorageProvider{
			Driver: getMetaString(item.Meta, "provider_driver", models.UploadProviderDriverLocal),
			Status: models.UploadProviderStatusReady,
		},
		Logger: s.logger,
	})
	if openErr == nil {
		if removeErr := driver.Delete(ctx, DeleteRequest{
			TenantID:   normalizeTenantID(tenantID),
			StorageKey: item.StorageKey,
		}); removeErr != nil && s.logger != nil {
			s.logger.Warn("remove uploaded file failed", zap.Error(removeErr), zap.String("storage_key", item.StorageKey))
		}
	} else if s.logger != nil {
		s.logger.Warn("open driver for delete failed", zap.Error(openErr), zap.String("storage_key", item.StorageKey))
	}
	return s.repo.DeleteRecord(ctx, tenantID, id)
}

func getMetaString(meta models.MetaJSON, key, fallback string) string {
	if meta == nil {
		return fallback
	}
	raw, ok := meta[key]
	if !ok {
		return fallback
	}
	value, ok := raw.(string)
	if !ok || strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func buildUploadRecord(effective *EffectiveConfig, userID *uuid.UUID, originalFilename, storedFilename, storageKey, url, contentType string, size int64, checksum, source string, extraMeta models.MetaJSON) *models.UploadRecord {
	record := &models.UploadRecord{
		TenantID:         normalizeTenantID(effective.UploadKey.TenantID),
		ProviderID:       effective.Provider.ID,
		BucketID:         effective.Bucket.ID,
		UploadKeyID:      effective.UploadKey.ID,
		OriginalFilename: filepath.Base(originalFilename),
		StoredFilename:   storedFilename,
		StorageKey:       storageKey,
		URL:              url,
		MimeType:         contentType,
		Size:             size,
		Checksum:         checksum,
		Status:           models.UploadRecordStatusActive,
		Meta: models.MetaJSON{
			"provider_key":    effective.Provider.ProviderKey,
			"provider_driver": effective.Provider.Driver,
			"bucket_key":      effective.Bucket.BucketKey,
			"upload_key":      effective.UploadKey.Key,
			"rule_key":        effective.ResolvedRule,
			"fallback_used":   effective.FallbackUsed,
			"source":          source,
		},
	}
	for key, value := range extraMeta {
		record.Meta[key] = value
	}
	if userID != nil && *userID != uuid.Nil {
		record.UploadedBy = userID
	}
	if effective.Rule.ID != uuid.Nil {
		record.RuleID = &effective.Rule.ID
	}
	return record
}

func planObjectKey(effective *EffectiveConfig, originalFilename, contentType string, now time.Time) (storageKey, publicPath, storedName string) {
	publicDir := buildRelativeDir("", effective.PathTemplate, effective.SubPath, now)
	storageDir := buildRelativeDir(effective.Bucket.BasePath, effective.PathTemplate, effective.SubPath, now)
	storedName = resolveStoredFilename(originalFilename, contentType, effective.FilenameStrategy)
	storageKey = filepath.ToSlash(filepath.Join(storageDir, storedName))
	publicPath = filepath.ToSlash(filepath.Join(publicDir, storedName))
	return storageKey, publicPath, storedName
}

func resolveStoredFilename(originalFilename, contentType, strategy string) string {
	ext := normalizedExt(originalFilename, contentType)
	switch strings.TrimSpace(strings.ToLower(strategy)) {
	case "", "uuid":
		return uuid.NewString() + ext
	case "original":
		candidate := sanitizeFilename(filepath.Base(originalFilename))
		if filepath.Ext(candidate) == "" && ext != "" {
			candidate += ext
		}
		if candidate != "" {
			return candidate
		}
		return uuid.NewString() + ext
	default:
		return uuid.NewString() + ext
	}
}

func sanitizeFilename(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	replacer := strings.NewReplacer(
		"/", "_",
		"\\", "_",
		":", "_",
		"*", "_",
		"?", "_",
		"\"", "_",
		"<", "_",
		">", "_",
		"|", "_",
	)
	return replacer.Replace(value)
}

func validateStorageKeyForConfig(storageKey, bucketBasePath string) (normalized string, storedFilename string, err error) {
	normalized = filepath.ToSlash(filepath.Clean(filepath.FromSlash(strings.TrimSpace(storageKey))))
	if normalized == "." || normalized == "" || strings.HasPrefix(normalized, "../") || strings.HasPrefix(normalized, "/") {
		return "", "", &DriverError{
			Code:      DriverErrorCodePathViolation,
			Operation: "complete_direct_upload",
			Message:   "storage key escapes bucket root",
		}
	}
	basePath := strings.Trim(strings.TrimSpace(filepath.ToSlash(bucketBasePath)), "/")
	if basePath != "" && normalized != basePath && !strings.HasPrefix(normalized, basePath+"/") {
		return "", "", &DriverError{
			Code:      DriverErrorCodePathViolation,
			Operation: "complete_direct_upload",
			Message:   "storage key is outside bucket base path",
		}
	}
	return normalized, filepath.Base(normalized), nil
}

func resolveObjectURL(driver Driver, bucket models.StorageBucket, storageKey string) string {
	type objectURLProvider interface {
		objectURL(storageKey string) string
	}
	if provider, ok := driver.(objectURLProvider); ok {
		return provider.objectURL(storageKey)
	}
	basePath := strings.Trim(strings.TrimSpace(filepath.ToSlash(bucket.BasePath)), "/")
	relativePath := strings.Trim(strings.TrimSpace(storageKey), "/")
	if basePath != "" && strings.HasPrefix(relativePath, basePath+"/") {
		relativePath = strings.TrimPrefix(relativePath, basePath+"/")
	}
	return buildPublicURL(bucket.PublicBaseURL, relativePath)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func buildRelativeDir(basePath, template, subPath string, now time.Time) string {
	pathValue := strings.TrimSpace(template)
	if pathValue == "" {
		pathValue = "{yyyy}/{mm}/{dd}"
	}
	replacer := strings.NewReplacer(
		"{yyyy}", now.Format("2006"),
		"{mm}", now.Format("01"),
		"{dd}", now.Format("02"),
	)
	pathValue = replacer.Replace(pathValue)
	segments := []string{strings.TrimSpace(basePath), strings.Trim(pathValue, "/"), strings.Trim(subPath, "/")}
	parts := make([]string, 0, len(segments))
	for _, segment := range segments {
		if segment != "" {
			parts = append(parts, segment)
		}
	}
	if len(parts) == 0 {
		return "."
	}
	return filepath.ToSlash(filepath.Join(parts...))
}

func buildPublicURL(base, relativePath string) string {
	normalizedBase := "/" + strings.Trim(strings.TrimSpace(base), "/")
	if normalizedBase == "/" {
		normalizedBase = "/uploads"
	}
	return normalizedBase + "/" + strings.TrimLeft(relativePath, "/")
}

func joinURLPath(base, child string) string {
	normalizedBase := "/" + strings.Trim(strings.TrimSpace(base), "/")
	normalizedChild := strings.Trim(strings.TrimSpace(child), "/")
	if normalizedChild == "" {
		if normalizedBase == "/" {
			return "/uploads"
		}
		return normalizedBase
	}
	if normalizedBase == "/" {
		return "/" + normalizedChild
	}
	return normalizedBase + "/" + normalizedChild
}

func normalizeContentType(contentType, filename string) string {
	target := strings.TrimSpace(contentType)
	if target != "" {
		return target
	}
	target = mime.TypeByExtension(filepath.Ext(filename))
	if target != "" {
		return target
	}
	return "application/octet-stream"
}

func normalizedExt(filename, contentType string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	if ext != "" {
		return ext
	}
	if exts, _ := mime.ExtensionsByType(contentType); len(exts) > 0 {
		return exts[0]
	}
	return ""
}

func mimeAllowed(contentType string, keyAllowed, ruleAllowed models.StringList) bool {
	allowed := ruleAllowed
	if len(allowed) == 0 {
		allowed = keyAllowed
	}
	if len(allowed) == 0 {
		return true
	}
	for _, item := range allowed {
		target := strings.TrimSpace(strings.ToLower(item))
		if target == "" {
			continue
		}
		if strings.HasSuffix(target, "/*") && strings.HasPrefix(strings.ToLower(contentType), strings.TrimSuffix(target, "*")) {
			return true
		}
		if strings.EqualFold(target, contentType) {
			return true
		}
	}
	return false
}

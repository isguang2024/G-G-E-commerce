package upload

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/maben/backend/internal/config"
	"github.com/maben/backend/internal/modules/observability/audit"
	"github.com/maben/backend/internal/modules/system/models"
)

type auditRecorderSpy struct {
	events []audit.Event
}

func (s *auditRecorderSpy) Record(_ context.Context, e audit.Event) {
	s.events = append(s.events, e)
}

func (s *auditRecorderSpy) Stats() audit.Stats {
	return audit.Stats{}
}

func (s *auditRecorderSpy) Shutdown(context.Context) error {
	return nil
}

func newUploadTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := "file:" + uuid.NewString() + "?mode=memory&cache=shared"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite db: %v", err)
	}
	schema := []string{
		`CREATE TABLE storage_providers (
			id TEXT PRIMARY KEY,
			tenant_id TEXT NOT NULL,
			provider_key TEXT NOT NULL,
			name TEXT NOT NULL,
			driver TEXT NOT NULL,
			endpoint TEXT NOT NULL DEFAULT '',
			region TEXT NOT NULL DEFAULT '',
			base_url TEXT NOT NULL DEFAULT '',
			access_key_encrypted TEXT NOT NULL DEFAULT '',
			secret_key_encrypted TEXT NOT NULL DEFAULT '',
			extra TEXT NOT NULL DEFAULT '{}',
			is_default INTEGER NOT NULL DEFAULT 0,
			status TEXT NOT NULL DEFAULT 'ready',
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME
		);`,
		`CREATE UNIQUE INDEX idx_storage_providers_tenant_provider_key ON storage_providers (tenant_id, provider_key);`,
		`CREATE TABLE storage_buckets (
			id TEXT PRIMARY KEY,
			tenant_id TEXT NOT NULL,
			provider_id TEXT NOT NULL,
			bucket_key TEXT NOT NULL,
			name TEXT NOT NULL,
			bucket_name TEXT NOT NULL,
			base_path TEXT NOT NULL DEFAULT '',
			public_base_url TEXT NOT NULL DEFAULT '',
			is_public INTEGER NOT NULL DEFAULT 1,
			status TEXT NOT NULL DEFAULT 'ready',
			extra TEXT NOT NULL DEFAULT '{}',
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME
		);`,
		`CREATE UNIQUE INDEX idx_storage_buckets_tenant_bucket_key ON storage_buckets (tenant_id, bucket_key);`,
		`CREATE TABLE upload_keys (
			id TEXT PRIMARY KEY,
			tenant_id TEXT NOT NULL,
			bucket_id TEXT NOT NULL,
			key TEXT NOT NULL,
			name TEXT NOT NULL,
			path_template TEXT NOT NULL DEFAULT '',
			default_rule_key TEXT NOT NULL DEFAULT '',
			max_size_bytes INTEGER NOT NULL DEFAULT 0,
			allowed_mime_types TEXT NOT NULL DEFAULT '[]',
			visibility TEXT NOT NULL DEFAULT 'public',
			status TEXT NOT NULL DEFAULT 'ready',
			meta TEXT NOT NULL DEFAULT '{}',
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME
		);`,
		`CREATE UNIQUE INDEX idx_upload_keys_tenant_key ON upload_keys (tenant_id, key);`,
		`CREATE TABLE upload_key_rules (
			id TEXT PRIMARY KEY,
			tenant_id TEXT NOT NULL,
			upload_key_id TEXT NOT NULL,
			rule_key TEXT NOT NULL,
			name TEXT NOT NULL,
			sub_path TEXT NOT NULL DEFAULT '',
			filename_strategy TEXT NOT NULL DEFAULT 'uuid',
			max_size_bytes INTEGER NOT NULL DEFAULT 0,
			allowed_mime_types TEXT NOT NULL DEFAULT '[]',
			process_pipeline TEXT NOT NULL DEFAULT '[]',
			is_default INTEGER NOT NULL DEFAULT 0,
			status TEXT NOT NULL DEFAULT 'ready',
			meta TEXT NOT NULL DEFAULT '{}',
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME
		);`,
		`CREATE UNIQUE INDEX idx_upload_rules_tenant_key_rule ON upload_key_rules (tenant_id, upload_key_id, rule_key);`,
		`CREATE TABLE upload_records (
			id TEXT PRIMARY KEY,
			tenant_id TEXT NOT NULL,
			provider_id TEXT NOT NULL,
			bucket_id TEXT NOT NULL,
			upload_key_id TEXT NOT NULL,
			rule_id TEXT,
			uploaded_by TEXT,
			original_filename TEXT NOT NULL,
			stored_filename TEXT NOT NULL,
			storage_key TEXT NOT NULL,
			url TEXT NOT NULL,
			mime_type TEXT NOT NULL DEFAULT '',
			size INTEGER NOT NULL DEFAULT 0,
			checksum TEXT NOT NULL DEFAULT '',
			status TEXT NOT NULL DEFAULT 'active',
			meta TEXT NOT NULL DEFAULT '{}',
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME
		);`,
	}
	for _, stmt := range schema {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatalf("create upload test schema failed: %v", err)
		}
	}
	return db
}

func newUploadTestRepository(t *testing.T) (*Repository, *auditRecorderSpy) {
	t.Helper()

	cipher, err := NewSecretCipher(config.UploadConfig{
		SecretMasterKeys:   []string{"local:test-master-key"},
		SecretCurrentKeyID: "local",
		SecretCacheTTL:     60,
	})
	if err != nil {
		t.Fatalf("new secret cipher: %v", err)
	}
	recorder := &auditRecorderSpy{}
	repo := NewRepository(newUploadTestDB(t)).
		WithSecretCipher(cipher).
		WithAuditRecorder(recorder).
		WithResolvedConfigCache(newLocalResolvedConfigCache(time.Minute)).
		WithDefaultUploadKey("media.default")
	return repo, recorder
}

func seedUploadHierarchy(t *testing.T, repo *Repository, tenantID string) (models.StorageProvider, models.StorageBucket, models.UploadKey, models.UploadKeyRule) {
	t.Helper()

	ctx := context.Background()
	provider := models.StorageProvider{
		ID:                 uuid.New(),
		TenantID:           tenantID,
		ProviderKey:        "local-public",
		Name:               "Local Public",
		Driver:             models.UploadProviderDriverLocal,
		BaseURL:            "/uploads",
		AccessKeyEncrypted: "access-key",
		SecretKeyEncrypted: "secret-key",
		IsDefault:          true,
		Status:             models.UploadProviderStatusReady,
	}
	if err := repo.EnsureProvider(ctx, &provider); err != nil {
		t.Fatalf("EnsureProvider() error = %v", err)
	}

	bucket := models.StorageBucket{
		ID:            uuid.New(),
		TenantID:      tenantID,
		ProviderID:    provider.ID,
		BucketKey:     "public-media",
		Name:          "Public Media",
		BucketName:    "public-media",
		BasePath:      "public-media",
		PublicBaseURL: "/uploads/public-media",
		IsPublic:      true,
		Status:        models.UploadProviderStatusReady,
	}
	if err := repo.EnsureBucket(ctx, &bucket); err != nil {
		t.Fatalf("EnsureBucket() error = %v", err)
	}

	uploadKey := models.UploadKey{
		ID:               uuid.New(),
		TenantID:         tenantID,
		BucketID:         bucket.ID,
		Key:              "media.default",
		Name:             "Default Upload",
		PathTemplate:     "{yyyy}/{mm}/{dd}",
		DefaultRuleKey:   "image",
		MaxSizeBytes:     2 * 1024 * 1024,
		AllowedMimeTypes: models.StringList{"image/png", "image/jpeg"},
		Visibility:       "public",
		Status:           models.UploadProviderStatusReady,
	}
	if err := repo.EnsureUploadKey(ctx, &uploadKey); err != nil {
		t.Fatalf("EnsureUploadKey() error = %v", err)
	}

	rule := models.UploadKeyRule{
		ID:               uuid.New(),
		TenantID:         tenantID,
		UploadKeyID:      uploadKey.ID,
		RuleKey:          "image",
		Name:             "Image Rule",
		SubPath:          "images",
		FilenameStrategy: "uuid",
		MaxSizeBytes:     2 * 1024 * 1024,
		AllowedMimeTypes: models.StringList{"image/*"},
		ProcessPipeline:  models.StringList{},
		IsDefault:        true,
		Status:           models.UploadProviderStatusReady,
	}
	if err := repo.EnsureUploadRule(ctx, &rule); err != nil {
		t.Fatalf("EnsureUploadRule() error = %v", err)
	}
	return provider, bucket, uploadKey, rule
}

func TestRepositorySQLiteResolveAndDelete(t *testing.T) {
	repo, recorder := newUploadTestRepository(t)
	provider, bucket, uploadKey, rule := seedUploadHierarchy(t, repo, "tenant-a")

	ctx := context.Background()
	gotProvider, err := repo.GetProviderByKey(ctx, "tenant-a", provider.ProviderKey)
	if err != nil {
		t.Fatalf("GetProviderByKey() error = %v", err)
	}
	if gotProvider.AccessKeyEncrypted != "access-key" || gotProvider.SecretKeyEncrypted != "secret-key" {
		t.Fatalf("provider secrets were not decrypted: %#v", gotProvider)
	}

	cfg, err := repo.GetResolvedConfigByKey(ctx, "tenant-a", uploadKey.Key)
	if err != nil {
		t.Fatalf("GetResolvedConfigByKey() error = %v", err)
	}
	if cfg.Bucket.BucketKey != bucket.BucketKey || cfg.Rule.RuleKey != rule.RuleKey {
		t.Fatalf("resolved config mismatch: bucket=%q rule=%q", cfg.Bucket.BucketKey, cfg.Rule.RuleKey)
	}

	defaultCfg, err := repo.GetDefaultConfig(ctx, "tenant-a")
	if err != nil {
		t.Fatalf("GetDefaultConfig() error = %v", err)
	}
	if defaultCfg.UploadKey.Key != uploadKey.Key {
		t.Fatalf("default upload key = %q, want %q", defaultCfg.UploadKey.Key, uploadKey.Key)
	}
	loadedCfg, err := repo.loadDefaultConfig(ctx, "tenant-a")
	if err != nil {
		t.Fatalf("loadDefaultConfig() error = %v", err)
	}
	if loadedCfg.Provider.ProviderKey != provider.ProviderKey {
		t.Fatalf("loadDefaultConfig() provider_key = %q, want %q", loadedCfg.Provider.ProviderKey, provider.ProviderKey)
	}

	masked, err := repo.GetBucketMasked(ctx, "tenant-a", bucket.ID)
	if err != nil {
		t.Fatalf("GetBucketMasked() error = %v", err)
	}
	if masked.Provider.SecretKeyEncrypted == "secret-key" {
		t.Fatalf("GetBucketMasked() should redact provider secret")
	}

	if err := repo.DeleteUploadRule(ctx, "tenant-a", rule.ID); err != nil {
		t.Fatalf("DeleteUploadRule() error = %v", err)
	}
	if err := repo.DeleteUploadKey(ctx, "tenant-a", uploadKey.ID); err != nil {
		t.Fatalf("DeleteUploadKey() error = %v", err)
	}
	if err := repo.DeleteBucket(ctx, "tenant-a", bucket.ID); err != nil {
		t.Fatalf("DeleteBucket() error = %v", err)
	}

	if len(recorder.events) < 6 {
		t.Fatalf("audit event count = %d, want at least 6", len(recorder.events))
	}
}

func TestRepositorySQLiteRecordCRUD(t *testing.T) {
	repo, _ := newUploadTestRepository(t)
	provider, bucket, uploadKey, rule := seedUploadHierarchy(t, repo, "tenant-a")
	ctx := context.Background()

	record := &models.UploadRecord{
		ID:               uuid.New(),
		TenantID:         "tenant-a",
		ProviderID:       provider.ID,
		BucketID:         bucket.ID,
		UploadKeyID:      uploadKey.ID,
		RuleID:           &rule.ID,
		OriginalFilename: "demo.png",
		StoredFilename:   "stored.png",
		StorageKey:       "public-media/2026/04/16/images/stored.png",
		URL:              "/uploads/public-media/2026/04/16/images/stored.png",
		MimeType:         "image/png",
		Size:             128,
		Checksum:         "abc123",
		Status:           models.UploadRecordStatusActive,
		Meta:             models.MetaJSON{"provider_driver": models.UploadProviderDriverLocal},
	}
	if err := repo.CreateRecord(ctx, record); err != nil {
		t.Fatalf("CreateRecord() error = %v", err)
	}

	items, err := repo.ListRecords(ctx, "tenant-a", 10)
	if err != nil {
		t.Fatalf("ListRecords() error = %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("ListRecords() count = %d, want 1", len(items))
	}

	total, err := repo.CountRecords(ctx, "tenant-a")
	if err != nil {
		t.Fatalf("CountRecords() error = %v", err)
	}
	if total != 1 {
		t.Fatalf("CountRecords() = %d, want 1", total)
	}

	got, err := repo.GetRecord(ctx, "tenant-a", record.ID)
	if err != nil {
		t.Fatalf("GetRecord() error = %v", err)
	}
	if got.StorageKey != record.StorageKey {
		t.Fatalf("GetRecord() storage_key = %q, want %q", got.StorageKey, record.StorageKey)
	}

	if err := repo.DeleteRecord(ctx, "tenant-a", record.ID); err != nil {
		t.Fatalf("DeleteRecord() error = %v", err)
	}
	if _, err := repo.GetRecord(ctx, "tenant-a", record.ID); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("GetRecord() after delete error = %v, want %v", err, gorm.ErrRecordNotFound)
	}
}

func TestServiceSQLiteEnsureSeedUploadListAndDelete(t *testing.T) {
	db := newUploadTestDB(t)
	cfg := &config.Config{
		Redis: config.RedisConfig{},
		Upload: config.UploadConfig{
			LocalRoot:          t.TempDir(),
			PublicBaseURL:      "/uploads",
			DefaultTenantID:    "tenant-a",
			DefaultProviderKey: "local-public",
			DefaultBucketKey:   "public-media",
			DefaultUploadKey:   "media.default",
			DefaultRuleKey:     "image",
			MaxFileSizeBytes:   1024 * 1024,
			SecretMasterKeys:   []string{"local:test-master-key"},
			SecretCurrentKeyID: "local",
		},
	}
	repo := NewRepository(db)
	svc := NewService(repo, cfg, zap.NewNop())

	ctx := context.Background()
	if err := svc.EnsureDefaultSeeds(ctx); err != nil {
		t.Fatalf("EnsureDefaultSeeds() error = %v", err)
	}
	if err := EnsureDefaultSeeds(ctx, repo, zap.NewNop()); err != nil {
		t.Fatalf("EnsureDefaultSeeds() wrapper error = %v", err)
	}

	if _, err := svc.Upload(ctx, "tenant-a", nil, UploadInput{Name: "", File: bytes.NewBuffer(nil)}); !errors.Is(err, ErrInvalidFile) {
		t.Fatalf("Upload() invalid input error = %v, want %v", err, ErrInvalidFile)
	}
	if _, err := svc.Upload(ctx, "tenant-a", nil, UploadInput{
		Name:     "demo.txt",
		File:     bytes.NewBufferString("hello"),
		Size:     5,
		MimeType: "text/plain",
	}); err == nil {
		t.Fatalf("Upload() should reject disallowed mime type")
	}

	userID := uuid.New()
	record, err := svc.Upload(ctx, "tenant-a", &userID, UploadInput{
		Name:     "avatar",
		File:     bytes.NewBufferString("pngdata"),
		Size:     7,
		MimeType: "image/png",
	})
	if err != nil {
		t.Fatalf("Upload() success path error = %v", err)
	}
	if record.URL == "" || filepath.Ext(record.StoredFilename) != ".png" {
		t.Fatalf("Upload() result = %#v", record)
	}
	if _, err := os.Stat(filepath.Join(cfg.Upload.LocalRoot, filepath.FromSlash(record.StorageKey))); err != nil {
		t.Fatalf("uploaded file not found: %v", err)
	}

	items, total, err := svc.List(ctx, "tenant-a", 10)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(items) != 1 || total != 1 {
		t.Fatalf("List() = (%d,%d), want (1,1)", len(items), total)
	}

	if err := svc.Delete(ctx, "tenant-a", uuid.New()); !errors.Is(err, ErrRecordNotFound) {
		t.Fatalf("Delete() missing record error = %v, want %v", err, ErrRecordNotFound)
	}
	if err := svc.Delete(ctx, "tenant-a", record.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(cfg.Upload.LocalRoot, filepath.FromSlash(record.StorageKey))); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("deleted file stat error = %v, want not exist", err)
	}
}

func TestResolvedConfigCacheRedisBroadcastInvalidation(t *testing.T) {
	redisServer, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis.Run() error = %v", err)
	}
	defer redisServer.Close()

	cfg := &config.Config{
		Redis: config.RedisConfig{
			Host: redisServer.Host(),
		},
	}
	port, err := strconv.Atoi(redisServer.Port())
	if err != nil {
		t.Fatalf("strconv.Atoi(redis port) error = %v", err)
	}
	cfg.Redis.Port = port
	cacheA, err := newResolvedConfigCache(cfg, zap.NewNop())
	if err != nil {
		t.Fatalf("newResolvedConfigCache(cacheA) error = %v", err)
	}
	cacheB, err := newResolvedConfigCache(cfg, zap.NewNop())
	if err != nil {
		t.Fatalf("newResolvedConfigCache(cacheB) error = %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cacheA.Start(ctx)
	cacheB.Start(ctx)
	time.Sleep(50 * time.Millisecond)

	cacheKey := uploadKeyCacheKey("tenant-a", "media.default")
	cacheA.Set(ctx, cacheKey, &ResolvedConfig{UploadKey: models.UploadKey{Key: "media.default"}})

	var cached ResolvedConfig
	if !cacheB.Get(ctx, cacheKey, &cached) {
		t.Fatalf("cacheB should read config from redis")
	}

	cacheA.Invalidate(ctx, nil, []string{uploadKeyCachePrefix("tenant-a")})
	deadline := time.Now().Add(time.Second)
	for {
		var after ResolvedConfig
		if !cacheB.Get(ctx, cacheKey, &after) {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("cacheB local cache should be invalidated by broadcast")
		}
		time.Sleep(20 * time.Millisecond)
	}
}

func TestResolvedConfigCacheLogWarnAndMalformedPayload(t *testing.T) {
	cache := &resolvedConfigCache{
		logger:     zap.NewNop(),
		ttl:        time.Minute,
		localItems: make(map[string]localCacheItem),
	}
	cache.logWarn("test")
	cache.handleInvalidationPayload("not-json")
}


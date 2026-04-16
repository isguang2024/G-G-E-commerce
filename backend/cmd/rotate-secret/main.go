package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/gg-ecommerce/backend/internal/config"
	systemmodels "github.com/gg-ecommerce/backend/internal/modules/system/models"
	"github.com/gg-ecommerce/backend/internal/modules/system/upload"
	"github.com/gg-ecommerce/backend/internal/pkg/database"
	loggerpkg "github.com/gg-ecommerce/backend/internal/pkg/logger"
)

type providerSecretRow struct {
	ID                 string
	TenantID           string
	ProviderKey        string
	AccessKeyEncrypted string
	SecretKeyEncrypted string
}

func main() {
	tenantID := flag.String("tenant", "", "仅轮换指定 tenant_id；为空时扫描全部租户")
	dryRun := flag.Bool("dry-run", false, "仅预演，不写回数据库")
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	logger, err := loggerpkg.New(cfg.Log.Level, cfg.Log.Output)
	if err != nil {
		log.Fatalf("初始化日志失败: %v", err)
	}
	defer logger.Sync()

	cipher, err := upload.NewSecretCipher(cfg.Upload)
	if err != nil {
		logger.Fatal("初始化 SecretCipher 失败", zap.Error(err))
	}

	_, err = database.Init(&cfg.DB)
	if err != nil {
		logger.Fatal("初始化数据库失败", zap.Error(err))
	}
	defer database.Close()

	ctx := context.Background()
	rows, err := loadProviders(ctx, strings.TrimSpace(*tenantID))
	if err != nil {
		logger.Fatal("加载 provider 失败", zap.Error(err))
	}

	plan, err := buildRotationPlan(ctx, cipher, rows)
	if err != nil {
		logger.Fatal("生成轮换计划失败", zap.Error(err))
	}

	logger.Info("upload secret rotation plan ready",
		zap.Bool("dry_run", *dryRun),
		zap.Int("providers_total", len(rows)),
		zap.Int("providers_to_rotate", len(plan)),
		zap.String("current_key_id", cipher.CurrentKeyID()),
	)
	for _, item := range plan {
		logger.Info("provider scheduled for rotation",
			zap.String("tenant_id", item.TenantID),
			zap.String("provider_key", item.ProviderKey),
			zap.Bool("rotate_access_key", item.AccessKeyEncrypted != ""),
			zap.Bool("rotate_secret_key", item.SecretKeyEncrypted != ""),
		)
	}
	if *dryRun || len(plan) == 0 {
		return
	}

	if err := database.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return persistRotationPlan(ctx, tx, plan)
	}); err != nil {
		logger.Fatal("写回轮换结果失败", zap.Error(err))
	}

	logger.Info("upload secret rotation finished", zap.Int("providers_rotated", len(plan)))
}

func loadProviders(ctx context.Context, tenantID string) ([]providerSecretRow, error) {
	query := database.DB.WithContext(ctx).
		Model(&systemmodels.StorageProvider{}).
		Where("deleted_at IS NULL")
	if tenantID != "" {
		query = query.Where("tenant_id = ?", tenantID)
	}

	rows := make([]providerSecretRow, 0)
	err := query.
		Select("id", "tenant_id", "provider_key", "access_key_encrypted", "secret_key_encrypted").
		Order("tenant_id ASC, provider_key ASC").
		Scan(&rows).Error
	return rows, err
}

func buildRotationPlan(ctx context.Context, cipher upload.SecretCipher, rows []providerSecretRow) ([]providerSecretRow, error) {
	plan := make([]providerSecretRow, 0, len(rows))
	for _, row := range rows {
		rotatedAccessKey, rotateAccess, err := rotateSecretValue(ctx, cipher, row.AccessKeyEncrypted)
		if err != nil {
			return nil, fmt.Errorf("rotate access key for provider %s: %w", row.ProviderKey, err)
		}
		rotatedSecretKey, rotateSecret, err := rotateSecretValue(ctx, cipher, row.SecretKeyEncrypted)
		if err != nil {
			return nil, fmt.Errorf("rotate secret key for provider %s: %w", row.ProviderKey, err)
		}
		if !rotateAccess && !rotateSecret {
			continue
		}
		row.AccessKeyEncrypted = rotatedAccessKey
		row.SecretKeyEncrypted = rotatedSecretKey
		plan = append(plan, row)
	}
	return plan, nil
}

func persistRotationPlan(ctx context.Context, tx *gorm.DB, plan []providerSecretRow) error {
	now := time.Now()
	for _, row := range plan {
		updates := map[string]any{
			"updated_at": now,
		}
		if row.AccessKeyEncrypted != "" {
			updates["access_key_encrypted"] = row.AccessKeyEncrypted
		}
		if row.SecretKeyEncrypted != "" {
			updates["secret_key_encrypted"] = row.SecretKeyEncrypted
		}
		if err := tx.WithContext(ctx).
			Model(&systemmodels.StorageProvider{}).
			Where("id = ? AND deleted_at IS NULL", row.ID).
			Updates(updates).Error; err != nil {
			return err
		}
	}
	return nil
}

func rotateSecretValue(ctx context.Context, cipher upload.SecretCipher, value string) (string, bool, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", false, nil
	}
	if !shouldRotateSecret(cipher, trimmed) {
		return trimmed, false, nil
	}

	plaintext := trimmed
	if uploadSecretLooksEncrypted(trimmed) {
		decrypted, err := cipher.Decrypt(ctx, trimmed)
		if err != nil {
			return "", false, err
		}
		plaintext = decrypted
	}
	encrypted, err := cipher.Encrypt(ctx, plaintext)
	if err != nil {
		return "", false, err
	}
	return encrypted, true, nil
}

func shouldRotateSecret(cipher upload.SecretCipher, value string) bool {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return false
	}
	if !uploadSecretLooksEncrypted(trimmed) {
		return true
	}
	return uploadSecretKeyID(trimmed) != cipher.CurrentKeyID()
}

func uploadSecretLooksEncrypted(value string) bool {
	return strings.HasPrefix(strings.TrimSpace(value), "gge:v1:")
}

func uploadSecretKeyID(value string) string {
	parts := strings.SplitN(strings.TrimSpace(value), ":", 4)
	if len(parts) != 4 || parts[0] != "gge" || parts[1] != "v1" {
		return ""
	}
	return strings.TrimSpace(parts[2])
}

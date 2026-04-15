// Package main 提供 GGE 多租户 SaaS 沙箱测试夹具。
//
// 职责：为 /system/message、/collaboration-workspace/message、/system/access-trace 等
// 高风险页面的 E2E 深度回归创建**可识别、可回收**的种子数据，避免污染真实消息流。
//
// 约定：所有沙箱数据以 "sandbox." 前缀或 "[SANDBOX] " 前缀区分，
//   - MessageTemplate.TemplateKey = "sandbox.deep-probe.notice"
//   - MessageSender.Name          = "[SANDBOX] Playwright Deep Probe"
//
// 使用：
//
//	go run ./cmd/fixtures-sandbox           # 幂等创建 / 更新
//	go run ./cmd/fixtures-sandbox --cleanup # 软删所有沙箱数据
//	go run ./cmd/fixtures-sandbox --purge   # 硬删（含软删行），仅本地调试使用
//
// 与 dispatch dry_run 配合：沙箱模板+沙箱 sender 仍然触发真实落库；
// 要完全零副作用，请在请求体里加 "dry_run": true（参见 MessageDispatchRequest）。
package main

import (
	"errors"
	"flag"
	"log"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/gg-ecommerce/backend/internal/config"
	systemmodels "github.com/gg-ecommerce/backend/internal/modules/system/models"
	"github.com/gg-ecommerce/backend/internal/pkg/database"
	"github.com/gg-ecommerce/backend/internal/pkg/logger"
)

const (
	sandboxTemplateKey = "sandbox.deep-probe.notice"
	sandboxSenderName  = "[SANDBOX] Playwright Deep Probe"
)

func main() {
	cleanup := flag.Bool("cleanup", false, "soft-delete sandbox fixture rows")
	purge := flag.Bool("purge", false, "hard-delete sandbox fixture rows (dangerous, local-only)")
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	lg, err := logger.New(cfg.Log.Level, cfg.Log.Output)
	if err != nil {
		log.Fatalf("init logger: %v", err)
	}
	defer lg.Sync()

	if _, err := database.Init(&cfg.DB); err != nil {
		lg.Fatal("init db", zap.Error(err))
	}
	defer database.Close()

	db := database.DB

	switch {
	case *purge:
		if err := purgeSandbox(db, lg); err != nil {
			lg.Fatal("purge sandbox", zap.Error(err))
		}
	case *cleanup:
		if err := cleanupSandbox(db, lg); err != nil {
			lg.Fatal("cleanup sandbox", zap.Error(err))
		}
	default:
		if err := ensureSandbox(db, lg); err != nil {
			lg.Fatal("ensure sandbox", zap.Error(err))
		}
	}
}

// ensureSandbox idempotently creates or updates the sandbox template + sender.
func ensureSandbox(db *gorm.DB, lg *zap.Logger) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := ensureSandboxTemplate(tx, lg); err != nil {
			return err
		}
		if err := ensureSandboxSender(tx, lg); err != nil {
			return err
		}
		return nil
	})
}

func ensureSandboxTemplate(tx *gorm.DB, lg *zap.Logger) error {
	var existing systemmodels.MessageTemplate
	err := tx.Where("template_key = ?", sandboxTemplateKey).First(&existing).Error
	switch {
	case err == nil:
		// Refresh content so re-running picks up edits to this file.
		updates := map[string]any{
			"name":             "[SANDBOX] 深测通知模板",
			"description":      "由 cmd/fixtures-sandbox 维护，专用于 Playwright E2E 深度回归。",
			"message_type":     "notice",
			"owner_scope":      "personal",
			"audience_type":    "specified_users",
			"title_template":   "[沙箱] {{title}}",
			"summary_template": "本条仅用于 E2E 深度回归，忽略即可。",
			"content_template": "深测触发时间：{{triggered_at}}。无需人工处理。",
			"action_type":      "none",
			"status":           "normal",
		}
		if err := tx.Model(&existing).Updates(updates).Error; err != nil {
			return err
		}
		lg.Info("sandbox template refreshed",
			zap.String("template_key", sandboxTemplateKey),
			zap.String("id", existing.ID.String()),
		)
		return nil
	case errors.Is(err, gorm.ErrRecordNotFound):
		tpl := systemmodels.MessageTemplate{
			TemplateKey:     sandboxTemplateKey,
			Name:            "[SANDBOX] 深测通知模板",
			Description:     "由 cmd/fixtures-sandbox 维护，专用于 Playwright E2E 深度回归。",
			MessageType:     "notice",
			OwnerScope:      "personal",
			AudienceType:    "specified_users",
			TitleTemplate:   "[沙箱] {{title}}",
			SummaryTemplate: "本条仅用于 E2E 深度回归，忽略即可。",
			ContentTemplate: "深测触发时间：{{triggered_at}}。无需人工处理。",
			ActionType:      "none",
			Status:          "normal",
			Meta: systemmodels.MetaJSON{
				"origin":     "fixtures-sandbox",
				"managed_by": "cmd/fixtures-sandbox",
			},
		}
		if err := tx.Create(&tpl).Error; err != nil {
			return err
		}
		lg.Info("sandbox template created",
			zap.String("template_key", sandboxTemplateKey),
			zap.String("id", tpl.ID.String()),
		)
		return nil
	default:
		return err
	}
}

func ensureSandboxSender(tx *gorm.DB, lg *zap.Logger) error {
	var existing systemmodels.MessageSender
	err := tx.Where("name = ? AND scope_type = ?", sandboxSenderName, "personal").First(&existing).Error
	switch {
	case err == nil:
		updates := map[string]any{
			"description": "由 cmd/fixtures-sandbox 维护，专用于 Playwright E2E 深度回归。",
			"status":      "normal",
			"is_default":  false,
		}
		if err := tx.Model(&existing).Updates(updates).Error; err != nil {
			return err
		}
		lg.Info("sandbox sender refreshed",
			zap.String("name", sandboxSenderName),
			zap.String("id", existing.ID.String()),
		)
		return nil
	case errors.Is(err, gorm.ErrRecordNotFound):
		sender := systemmodels.MessageSender{
			ScopeType:   "personal",
			Name:        sandboxSenderName,
			Description: "由 cmd/fixtures-sandbox 维护，专用于 Playwright E2E 深度回归。",
			IsDefault:   false,
			Status:      "normal",
			Meta: systemmodels.MetaJSON{
				"origin":     "fixtures-sandbox",
				"managed_by": "cmd/fixtures-sandbox",
			},
		}
		if err := tx.Create(&sender).Error; err != nil {
			return err
		}
		lg.Info("sandbox sender created",
			zap.String("name", sandboxSenderName),
			zap.String("id", sender.ID.String()),
		)
		return nil
	default:
		return err
	}
}

func cleanupSandbox(db *gorm.DB, lg *zap.Logger) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("template_key = ?", sandboxTemplateKey).Delete(&systemmodels.MessageTemplate{}).Error; err != nil {
			return err
		}
		if err := tx.Where("name = ? AND scope_type = ?", sandboxSenderName, "personal").Delete(&systemmodels.MessageSender{}).Error; err != nil {
			return err
		}
		lg.Info("sandbox fixtures soft-deleted",
			zap.String("template_key", sandboxTemplateKey),
			zap.String("sender_name", sandboxSenderName),
		)
		return nil
	})
}

func purgeSandbox(db *gorm.DB, lg *zap.Logger) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Unscoped().Where("template_key = ?", sandboxTemplateKey).Delete(&systemmodels.MessageTemplate{}).Error; err != nil {
			return err
		}
		if err := tx.Unscoped().Where("name = ? AND scope_type = ?", sandboxSenderName, "personal").Delete(&systemmodels.MessageSender{}).Error; err != nil {
			return err
		}
		lg.Warn("sandbox fixtures HARD-deleted (purge)",
			zap.String("template_key", sandboxTemplateKey),
			zap.String("sender_name", sandboxSenderName),
		)
		return nil
	})
}

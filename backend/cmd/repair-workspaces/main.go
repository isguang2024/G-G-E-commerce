// repair-workspaces 一次性修复命令：把旧 user / collaboration_workspaces / collaboration_workspace_members
// 数据回填进 V5 canonical 的 workspaces / workspace_members 表。
//
// V5 运行时已不再做读路径自动 backfill，迁移期/历史库使用本命令做一次性补齐。
package main

import (
	"log"

	"go.uber.org/zap"

	"github.com/maben/backend/internal/config"
	"github.com/maben/backend/internal/modules/system/workspace"
	"github.com/maben/backend/internal/pkg/database"
	"github.com/maben/backend/internal/pkg/logger"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	zlog, err := logger.New(cfg.Log.Level, cfg.Log.Output)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer zlog.Sync()

	if _, err := database.Init(&cfg.DB); err != nil {
		zlog.Fatal("Failed to init database", zap.Error(err))
	}

	svc := workspace.NewService(database.DB, zlog)
	zlog.Info("Repairing workspaces (full backfill)...")
	if err := svc.EnsureWorkspaceBackfill(); err != nil {
		zlog.Fatal("Workspace backfill failed", zap.Error(err))
	}
	zlog.Info("Workspace repair completed.")
}


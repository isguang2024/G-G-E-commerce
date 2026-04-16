package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/maben/backend/internal/api/router"
	"github.com/maben/backend/internal/config"
	"github.com/maben/backend/internal/modules/observability/audit"
	"github.com/maben/backend/internal/modules/observability/logpolicy"
	"github.com/maben/backend/internal/modules/observability/telemetry"
	"github.com/maben/backend/internal/pkg/database"
	"github.com/maben/backend/internal/pkg/logger"
)

func main() {
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化日志
	zlog, err := logger.NewWithOptions(logger.Options{
		Level:              cfg.Log.Level,
		Output:             cfg.Log.Output,
		Format:             cfg.Log.Format,
		SamplingInitial:    cfg.Log.Sampling.Initial,
		SamplingThereafter: cfg.Log.Sampling.Thereafter,
		Sampling: &logger.Sampling{
			Initial:    cfg.Log.Sampling.Initial,
			Thereafter: cfg.Log.Sampling.Thereafter,
		},
	})
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer zlog.Sync()
	// 把根 logger 注册给 logger 包，使全局 logger.With(ctx) / logger.FromContext(ctx)
	// 都能派生子 logger，保证 request_id 贯穿整条链路。
	logger.SetBase(zlog)

	zlog.Info("Starting G&G E-commerce Backend Server",
		zap.String("version", "1.0.0"),
		zap.String("env", cfg.Env),
	)

	// 初始化数据库
	db, err := database.Init(&cfg.DB, database.RuntimeOptions{
		Env:      cfg.Env,
		LogLevel: cfg.Log.Level,
	})
	if err != nil {
		zlog.Fatal("Failed to initialize database", zap.Error(err))
	}
	defer database.Close()
	zlog.Info("Database connected successfully")

	policyRepo := logpolicy.NewRepository(db)
	policyEngine := logpolicy.NewEngine(policyRepo, zlog)
	policyCtx, policyCancel := context.WithCancel(context.Background())
	policyEngine.Start(policyCtx)
	defer policyCancel()

	// 初始化审计 recorder（异步 channel + worker）。配置关闭时退化为 Noop。
	auditRecorder := audit.New(db, zlog, audit.Config{
		Enabled:      cfg.Audit.Enabled,
		RedactFields: cfg.Audit.RedactFields,
		QueueSize:    cfg.Audit.QueueSize,
		Workers:      cfg.Audit.Workers,
		BatchSize:    cfg.Audit.BatchSize,
		FlushInterval: resolveServerTimeout(
			cfg.Audit.FlushIntervalSeconds,
			time.Second,
		),
		AsyncMode:    cfg.Audit.AsyncMode,
		DegradedFile: cfg.Audit.DegradedFile,
		PolicyEngine: policyEngine,
	})
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := auditRecorder.Shutdown(shutdownCtx); err != nil {
			zlog.Warn("audit recorder shutdown error", zap.Error(err))
		}
	}()

	// 初始化前端日志 ingester（异步 channel + worker + token bucket 限流）。
	// 配置字段以"每秒条数"暴露，内部换算成 token bucket 的 burst/refill，
	// 便于 ops 同学直接配置而不需要理解令牌桶细节。
	telemetryIngester := telemetry.New(db, zlog, telemetry.Config{
		Enabled:         cfg.Telemetry.IngestEnabled,
		QueueSize:       2048,
		Workers:         2,
		RedactFields:    cfg.Audit.RedactFields,
		PerSessionRate:  float64(cfg.Telemetry.SessionRateLimit),
		PerSessionBurst: float64(cfg.Telemetry.SessionRateLimit) * 3,
		PerIPRate:       float64(cfg.Telemetry.IPRateLimit),
		PerIPBurst:      float64(cfg.Telemetry.IPRateLimit) * 3,
		MaxMessageBytes: cfg.Telemetry.PayloadMaxBytes,
		PolicyEngine:    policyEngine,
	})
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := telemetryIngester.Shutdown(shutdownCtx); err != nil {
			zlog.Warn("telemetry ingester shutdown error", zap.Error(err))
		}
	}()

	// 初始化 Gin 路由
	r := router.SetupRouter(cfg, zlog, db, auditRecorder, telemetryIngester)

	// 创建 HTTP 服务器
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      r,
		ReadTimeout:  resolveServerTimeout(cfg.Server.ReadTimeout, 15*time.Second),
		WriteTimeout: resolveServerTimeout(cfg.Server.WriteTimeout, 30*time.Second),
		IdleTimeout:  resolveServerTimeout(cfg.Server.IdleTimeout, 60*time.Second),
	}

	// 启动服务器（goroutine）
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			zlog.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	zlog.Info("Server started",
		zap.Int("port", cfg.Server.Port),
		zap.String("env", cfg.Env),
	)

	// 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	zlog.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		zlog.Fatal("Server forced to shutdown", zap.Error(err))
	}

	zlog.Info("Server exited")
}

func resolveServerTimeout(seconds int, fallback time.Duration) time.Duration {
	if seconds <= 0 {
		return fallback
	}
	return time.Duration(seconds) * time.Second
}


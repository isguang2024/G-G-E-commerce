package logger

import (
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Options 精细化控制 zap 初始化。Level/Output 必填，其余可选。
type Options struct {
	Level      string
	Output     string
	Format     string // "json"（默认） | "console"
	Sampling   *Sampling
	Production bool // 为 true 时强制 json + 禁 DPanic 直接 panic
}

// Sampling 对应 zap.SamplingConfig。Initial / Thereafter 都 <=0 时表示关闭采样。
type Sampling struct {
	Initial    int
	Thereafter int
}

// New 创建日志实例（保留旧签名，兼容各 CLI）。生产入口请改用 NewWithOptions。
func New(level, output string) (*zap.Logger, error) {
	return NewWithOptions(Options{Level: level, Output: output})
}

// NewWithOptions 是新的推荐入口。支持 format=console、采样、文件输出等。
// 返回的 *zap.Logger 可直接传给 middleware / handler；可同时调用 SetBase(l)
// 把它挂为全局根 logger，之后 logger.With(ctx) 会基于它派生子 logger。
func NewWithOptions(opts Options) (*zap.Logger, error) {
	zapLevel := parseLevel(opts.Level)

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.LowercaseLevelEncoder

	var encoder zapcore.Encoder
	switch strings.ToLower(strings.TrimSpace(opts.Format)) {
	case "console":
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	default:
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	}

	writeSyncer, err := resolveWriter(opts.Output)
	if err != nil {
		return nil, err
	}

	core := zapcore.NewCore(encoder, writeSyncer, zapLevel)

	// 采样：一旦给 Initial/Thereafter，就用 zapcore.NewSamplerWithOptions 包一层。
	if opts.Sampling != nil && (opts.Sampling.Initial > 0 || opts.Sampling.Thereafter > 0) {
		initial := opts.Sampling.Initial
		if initial <= 0 {
			initial = 100
		}
		thereafter := opts.Sampling.Thereafter
		if thereafter <= 0 {
			thereafter = 100
		}
		core = zapcore.NewSamplerWithOptions(core, samplingTick, initial, thereafter)
	}

	return zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel)), nil
}

// parseLevel 解析级别字符串；非法值退回 info。
func parseLevel(level string) zapcore.Level {
	var zl zapcore.Level
	if err := zl.UnmarshalText([]byte(strings.TrimSpace(level))); err != nil {
		return zapcore.InfoLevel
	}
	return zl
}

// resolveWriter 把 "stdout" / 文件路径统一转成 WriteSyncer。
func resolveWriter(output string) (zapcore.WriteSyncer, error) {
	out := strings.TrimSpace(output)
	if out == "" || out == "stdout" {
		return zapcore.AddSync(os.Stdout), nil
	}
	if out == "stderr" {
		return zapcore.AddSync(os.Stderr), nil
	}
	file, err := os.OpenFile(out, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o666)
	if err != nil {
		return nil, err
	}
	return zapcore.AddSync(file), nil
}

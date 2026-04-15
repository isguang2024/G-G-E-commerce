package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

// Config 应用配置结构
type Config struct {
	Env       string          `mapstructure:"env"`
	Server    ServerConfig    `mapstructure:"server"`
	DB        DBConfig        `mapstructure:"db"`
	Redis     RedisConfig     `mapstructure:"redis"`
	ES        ESConfig        `mapstructure:"elasticsearch"`
	MinIO     MinIOConfig     `mapstructure:"minio"`
	JWT       JWTConfig       `mapstructure:"jwt"`
	Log       LogConfig       `mapstructure:"log"`
	Audit     AuditConfig     `mapstructure:"audit"`
	Telemetry TelemetryConfig `mapstructure:"telemetry"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port         int    `mapstructure:"port"`
	Host         string `mapstructure:"host"`
	ReadTimeout  int    `mapstructure:"read_timeout"`
	WriteTimeout int    `mapstructure:"write_timeout"`
	IdleTimeout  int    `mapstructure:"idle_timeout"`
}

// DBConfig 数据库配置
type DBConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
	SSLMode  string `mapstructure:"sslmode"`
	TimeZone string `mapstructure:"timezone"`
}

// RedisConfig Redis 配置
type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

// ESConfig Elasticsearch 配置
type ESConfig struct {
	Addresses []string `mapstructure:"addresses"`
	Username  string   `mapstructure:"username"`
	Password  string   `mapstructure:"password"`
}

// MinIOConfig MinIO 配置
type MinIOConfig struct {
	Endpoint        string `mapstructure:"endpoint"`
	AccessKeyID     string `mapstructure:"access_key_id"`
	SecretAccessKey string `mapstructure:"secret_access_key"`
	UseSSL          bool   `mapstructure:"use_ssl"`
	BucketName      string `mapstructure:"bucket_name"`
}

// JWTConfig JWT 配置
type JWTConfig struct {
	Secret        string `mapstructure:"secret"`
	AccessExpire  int    `mapstructure:"access_expire"`
	RefreshExpire int    `mapstructure:"refresh_expire"`
}

// LogConfig 日志配置
//
// 字段说明：
//   - Level：zap 级别（debug/info/warn/error）；env 覆盖 `GG_LOG_LEVEL`；
//   - Output：stdout 或文件路径（例如 "/var/log/gge/app.log"）；
//   - Format：json（默认）或 console；生产环境强制 json；
//   - Sampling：打爆日志的保护阈值；Initial 为前 N 条按级别全写，Thereafter
//     为每个采样窗口额外放行 1/N 条。不需要采样时两者都设 0。
type LogConfig struct {
	Level    string            `mapstructure:"level"`
	Output   string            `mapstructure:"output"`
	Format   string            `mapstructure:"format"`
	Sampling LogSamplingConfig `mapstructure:"sampling"`
}

// LogSamplingConfig 是 zap.SamplingConfig 的用户态镜像。
type LogSamplingConfig struct {
	Initial    int `mapstructure:"initial"`
	Thereafter int `mapstructure:"thereafter"`
}

// AuditConfig 控制业务审计日志的持久化行为。
//
// 字段说明：
//   - Enabled：总开关；关闭后 recorder 退化为 Noop，不写库；
//   - RedactFields：顶层 redact 白名单之外，额外需要脱敏的字段名；
//   - QueueSize：异步 channel 缓冲；满了按 drop-newest 丢弃并 Warn；
//   - Workers：消费 goroutine 数量；
//   - BatchSize：单次批量落库阈值；
//   - FlushIntervalSeconds：批量最大等待秒数（即使没到 BatchSize 也会刷盘）；
//   - AsyncMode：true=channel+worker；false=同步写（测试/低流量场景）；
//   - DegradedFile：断路器打开时降级写入的 JSONL 文件路径。
type AuditConfig struct {
	Enabled              bool     `mapstructure:"enabled"`
	RedactFields         []string `mapstructure:"redact_fields"`
	QueueSize            int      `mapstructure:"queue_size"`
	Workers              int      `mapstructure:"workers"`
	BatchSize            int      `mapstructure:"batch_size"`
	FlushIntervalSeconds int      `mapstructure:"flush_interval_seconds"`
	AsyncMode            bool     `mapstructure:"async"`
	DegradedFile         string   `mapstructure:"degraded_file"`
}

// TelemetryConfig 控制前端日志上报端点 /telemetry/logs 的 ingest 行为。
//
// 字段说明：
//   - IngestEnabled：总开关；关闭时端点返回 204 但不落库；
//   - MaxBatchSize：单次 POST 最多接收的条数；超额返回 400；
//   - SessionRateLimit：单 session 每秒最多条数；超限静默丢弃 + 返回 429；
//   - IPRateLimit：单 IP 每秒最多条数；同上；
//   - PayloadMaxBytes：单条 telemetry payload JSON 的最大字节数。
type TelemetryConfig struct {
	IngestEnabled    bool `mapstructure:"ingest_enabled"`
	MaxBatchSize     int  `mapstructure:"max_batch_size"`
	SessionRateLimit int  `mapstructure:"session_rate_limit"`
	IPRateLimit      int  `mapstructure:"ip_rate_limit"`
	PayloadMaxBytes  int  `mapstructure:"payload_max_bytes"`
}

// Load 加载配置
func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath("../configs")
	viper.AddConfigPath("../../configs")

	viper.SetEnvPrefix("GG")
	viper.AutomaticEnv()

	setDefaults()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if env := os.Getenv("GG_ENV"); env != "" {
		cfg.Env = env
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func setDefaults() {
	viper.SetDefault("env", "development")
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.read_timeout", 15)
	viper.SetDefault("server.write_timeout", 30)
	viper.SetDefault("server.idle_timeout", 60)
	viper.SetDefault("db.host", "localhost")
	viper.SetDefault("db.port", 5432)
	viper.SetDefault("db.sslmode", "disable")
	viper.SetDefault("db.timezone", "Asia/Shanghai")
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.db", 0)
	viper.SetDefault("log.level", "info")
	viper.SetDefault("log.output", "stdout")
	viper.SetDefault("log.format", "json")
	viper.SetDefault("log.sampling.initial", 100)
	viper.SetDefault("log.sampling.thereafter", 10)

	viper.SetDefault("audit.enabled", true)
	viper.SetDefault("audit.queue_size", 1024)
	viper.SetDefault("audit.workers", 2)
	viper.SetDefault("audit.batch_size", 100)
	viper.SetDefault("audit.flush_interval_seconds", 1)
	viper.SetDefault("audit.async", true)
	viper.SetDefault("audit.degraded_file", "./data/audit_degraded.jsonl")

	viper.SetDefault("telemetry.ingest_enabled", true)
	viper.SetDefault("telemetry.max_batch_size", 100)
	viper.SetDefault("telemetry.session_rate_limit", 60)
	viper.SetDefault("telemetry.ip_rate_limit", 600)
	viper.SetDefault("telemetry.payload_max_bytes", 8192)
}

func (c *Config) Validate() error {
	if c == nil {
		return fmt.Errorf("config is nil")
	}
	if strings.TrimSpace(c.JWT.Secret) == "" {
		return fmt.Errorf("jwt.secret 不能为空")
	}
	if strings.EqualFold(strings.TrimSpace(c.Env), "production") && strings.TrimSpace(c.JWT.Secret) == "your-secret-key-change-in-production" {
		return fmt.Errorf("production 环境禁止使用默认 jwt.secret")
	}
	if c.Server.ReadTimeout < 0 || c.Server.WriteTimeout < 0 || c.Server.IdleTimeout < 0 {
		return fmt.Errorf("server 超时配置不能为负数")
	}
	return nil
}

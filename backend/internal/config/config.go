package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

// Config 应用配置结构
type Config struct {
	Env    string      `mapstructure:"env"`
	Server ServerConfig `mapstructure:"server"`
	DB     DBConfig     `mapstructure:"db"`
	Redis  RedisConfig  `mapstructure:"redis"`
	ES     ESConfig     `mapstructure:"elasticsearch"`
	MinIO  MinIOConfig  `mapstructure:"minio"`
	JWT    JWTConfig    `mapstructure:"jwt"`
	Log    LogConfig    `mapstructure:"log"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port         int    `mapstructure:"port"`
	Host         string `mapstructure:"host"`
	ReadTimeout  int    `mapstructure:"read_timeout"`
	WriteTimeout int    `mapstructure:"write_timeout"`
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
	Secret         string `mapstructure:"secret"`
	AccessExpire  int    `mapstructure:"access_expire"`  // 分钟
	RefreshExpire int    `mapstructure:"refresh_expire"` // 分钟
}

// LogConfig 日志配置
type LogConfig struct {
	Level  string `mapstructure:"level"`
	Output string `mapstructure:"output"`
}

// Load 加载配置
func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath("../configs")
	viper.AddConfigPath("../../configs")

	// 环境变量支持
	viper.SetEnvPrefix("GG")
	viper.AutomaticEnv()

	// 设置默认值
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

	// 从环境变量覆盖
	if env := os.Getenv("GG_ENV"); env != "" {
		cfg.Env = env
	}

	return &cfg, nil
}

func setDefaults() {
	viper.SetDefault("env", "development")
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("db.host", "localhost")
	viper.SetDefault("db.port", 5432)
	viper.SetDefault("db.sslmode", "disable")
	viper.SetDefault("db.timezone", "Asia/Shanghai")
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.db", 0)
	viper.SetDefault("log.level", "info")
	viper.SetDefault("log.output", "stdout")
}

package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	UploadProviderDriverLocal = "local"
	UploadProviderStatusReady = "ready"
	UploadRecordStatusActive  = "active"
	UploadModeAuto            = "auto"
	UploadModeDirect          = "direct"
	UploadModeRelay           = "relay"
	UploadModeInherit         = "inherit"
	VisibilityOverrideInherit = "inherit"
)

type StorageProvider struct {
	ID                 uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID           string         `gorm:"type:varchar(64);not null;default:'default';index" json:"tenant_id"`
	ProviderKey        string         `gorm:"type:varchar(100);not null" json:"provider_key"`
	Name               string         `gorm:"type:varchar(200);not null" json:"name"`
	Driver             string         `gorm:"type:varchar(32);not null" json:"driver"`
	Endpoint           string         `gorm:"type:text;not null;default:''" json:"endpoint"`
	Region             string         `gorm:"type:varchar(100);not null;default:''" json:"region"`
	BaseURL            string         `gorm:"type:text;not null;default:''" json:"base_url"`
	AccessKeyEncrypted string         `gorm:"type:text;not null;default:''" json:"access_key_encrypted"`
	SecretKeyEncrypted string         `gorm:"type:text;not null;default:''" json:"secret_key_encrypted"`
	Extra              MetaJSON       `gorm:"type:jsonb;not null;default:'{}'::jsonb" json:"extra"`
	IsDefault          bool           `gorm:"not null;default:false" json:"is_default"`
	Status             string         `gorm:"type:varchar(20);not null;default:'ready'" json:"status"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
	DeletedAt          gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (StorageProvider) TableName() string   { return "storage_providers" }
func (m StorageProvider) GetStatus() string { return m.Status }

type StorageBucket struct {
	ID            uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID      string         `gorm:"type:varchar(64);not null;default:'default';index" json:"tenant_id"`
	ProviderID    uuid.UUID      `gorm:"type:uuid;not null;index" json:"provider_id"`
	BucketKey     string         `gorm:"type:varchar(100);not null" json:"bucket_key"`
	Name          string         `gorm:"type:varchar(200);not null" json:"name"`
	BucketName    string         `gorm:"type:varchar(200);not null" json:"bucket_name"`
	BasePath      string         `gorm:"type:varchar(500);not null;default:''" json:"base_path"`
	PublicBaseURL string         `gorm:"type:text;not null;default:''" json:"public_base_url"`
	IsPublic      bool           `gorm:"not null;default:true" json:"is_public"`
	Status        string         `gorm:"type:varchar(20);not null;default:'ready'" json:"status"`
	Extra         MetaJSON       `gorm:"type:jsonb;not null;default:'{}'::jsonb" json:"extra"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (StorageBucket) TableName() string   { return "storage_buckets" }
func (m StorageBucket) GetStatus() string { return m.Status }

type UploadKey struct {
	ID                       uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID                 string         `gorm:"type:varchar(64);not null;default:'default';index" json:"tenant_id"`
	BucketID                 uuid.UUID      `gorm:"type:uuid;not null;index" json:"bucket_id"`
	Key                      string         `gorm:"type:varchar(150);not null" json:"key"`
	Name                     string         `gorm:"type:varchar(200);not null" json:"name"`
	PathTemplate             string         `gorm:"type:varchar(500);not null;default:''" json:"path_template"`
	DefaultRuleKey           string         `gorm:"type:varchar(150);not null;default:''" json:"default_rule_key"`
	MaxSizeBytes             int64          `gorm:"not null;default:0" json:"max_size_bytes"`
	AllowedMimeTypes         StringList     `gorm:"type:jsonb;serializer:json;not null" json:"allowed_mime_types"`
	UploadMode               string         `gorm:"type:varchar(20);not null;default:'auto'" json:"upload_mode"`
	IsFrontendVisible        bool           `gorm:"not null;default:false" json:"is_frontend_visible"`
	PermissionKey            string         `gorm:"type:varchar(150);not null;default:''" json:"permission_key"`
	FallbackKey              string         `gorm:"type:varchar(150);not null;default:''" json:"fallback_key"`
	ClientAccept             StringList     `gorm:"type:jsonb;serializer:json;not null" json:"client_accept"`
	DirectSizeThresholdBytes int64          `gorm:"not null;default:0" json:"direct_size_threshold_bytes"`
	ExtraSchema              MetaJSON       `gorm:"type:jsonb;not null;default:'{}'::jsonb" json:"extra_schema"`
	Visibility               string         `gorm:"type:varchar(20);not null;default:'public'" json:"visibility"`
	Status                   string         `gorm:"type:varchar(20);not null;default:'ready'" json:"status"`
	Meta                     MetaJSON       `gorm:"type:jsonb;not null;default:'{}'::jsonb" json:"meta"`
	CreatedAt                time.Time      `json:"created_at"`
	UpdatedAt                time.Time      `json:"updated_at"`
	DeletedAt                gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (UploadKey) TableName() string   { return "upload_keys" }
func (m UploadKey) GetStatus() string { return m.Status }

type UploadKeyRule struct {
	ID                 uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID           string         `gorm:"type:varchar(64);not null;default:'default';index" json:"tenant_id"`
	UploadKeyID        uuid.UUID      `gorm:"type:uuid;not null;index" json:"upload_key_id"`
	RuleKey            string         `gorm:"type:varchar(150);not null" json:"rule_key"`
	Name               string         `gorm:"type:varchar(200);not null" json:"name"`
	SubPath            string         `gorm:"type:varchar(255);not null;default:''" json:"sub_path"`
	FilenameStrategy   string         `gorm:"type:varchar(50);not null;default:'uuid'" json:"filename_strategy"`
	MaxSizeBytes       int64          `gorm:"not null;default:0" json:"max_size_bytes"`
	AllowedMimeTypes   StringList     `gorm:"type:jsonb;serializer:json;not null" json:"allowed_mime_types"`
	ProcessPipeline    StringList     `gorm:"type:jsonb;serializer:json;not null" json:"process_pipeline"`
	ModeOverride       string         `gorm:"type:varchar(20);not null;default:'inherit'" json:"mode_override"`
	VisibilityOverride string         `gorm:"type:varchar(20);not null;default:'inherit'" json:"visibility_override"`
	ClientAccept       StringList     `gorm:"type:jsonb;serializer:json;not null" json:"client_accept"`
	ExtraSchema        MetaJSON       `gorm:"type:jsonb;not null;default:'{}'::jsonb" json:"extra_schema"`
	IsDefault          bool           `gorm:"not null;default:false" json:"is_default"`
	Status             string         `gorm:"type:varchar(20);not null;default:'ready'" json:"status"`
	Meta               MetaJSON       `gorm:"type:jsonb;not null;default:'{}'::jsonb" json:"meta"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
	DeletedAt          gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (UploadKeyRule) TableName() string { return "upload_key_rules" }

type UploadRecord struct {
	ID               uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID         string         `gorm:"type:varchar(64);not null;default:'default';index" json:"tenant_id"`
	ProviderID       uuid.UUID      `gorm:"type:uuid;not null;index" json:"provider_id"`
	BucketID         uuid.UUID      `gorm:"type:uuid;not null;index" json:"bucket_id"`
	UploadKeyID      uuid.UUID      `gorm:"type:uuid;not null;index" json:"upload_key_id"`
	RuleID           *uuid.UUID     `gorm:"type:uuid;index" json:"rule_id"`
	UploadedBy       *uuid.UUID     `gorm:"type:uuid;index" json:"uploaded_by"`
	OriginalFilename string         `gorm:"type:varchar(500);not null" json:"original_filename"`
	StoredFilename   string         `gorm:"type:varchar(500);not null" json:"stored_filename"`
	StorageKey       string         `gorm:"type:varchar(1000);not null" json:"storage_key"`
	URL              string         `gorm:"type:varchar(1000);not null" json:"url"`
	MimeType         string         `gorm:"type:varchar(100);not null;default:''" json:"mime_type"`
	Size             int64          `gorm:"not null;default:0" json:"size"`
	Checksum         string         `gorm:"type:varchar(64);not null;default:'';index" json:"checksum"`
	Status           string         `gorm:"type:varchar(20);not null;default:'active'" json:"status"`
	Meta             MetaJSON       `gorm:"type:jsonb;not null;default:'{}'::jsonb" json:"meta"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (UploadRecord) TableName() string { return "upload_records" }

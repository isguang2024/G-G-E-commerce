package logpolicy

import (
	"time"

	"github.com/google/uuid"
)

const (
	DefaultTenantID = "default"

	PipelineAudit     = "audit"
	PipelineTelemetry = "telemetry"

	MatchFieldAction       = "action"
	MatchFieldOutcome      = "outcome"
	MatchFieldResourceType = "resource_type"
	MatchFieldLevel        = "level"
	MatchFieldEvent        = "event"
	MatchFieldRoute        = "route"

	DecisionAllow  = "allow"
	DecisionDeny   = "deny"
	DecisionSample = "sample"
)

// ComplianceLockedPatterns 是不可被 deny 的审计动作清单（用于 compliance lock）。
var ComplianceLockedPatterns = []string{
	"system.auth.login",
	"system.auth.register",
	"system.auth.logout",
	"system.user.delete",
	"system.role.assign",
	"system.permission.grant",
	"observability.policy.*",
}

type LogPolicy struct {
	ID         uuid.UUID  `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	TenantID   string     `gorm:"column:tenant_id;type:varchar(64);not null;default:'default'" json:"tenant_id"`
	Pipeline   string     `gorm:"column:pipeline;type:varchar(16);not null" json:"pipeline"`
	MatchField string     `gorm:"column:match_field;type:varchar(64);not null" json:"match_field"`
	Pattern    string     `gorm:"column:pattern;type:varchar(256);not null" json:"pattern"`
	Decision   string     `gorm:"column:decision;type:varchar(16);not null" json:"decision"`
	SampleRate *int       `gorm:"column:sample_rate" json:"sample_rate,omitempty"`
	Priority   int        `gorm:"column:priority;not null;default:0" json:"priority"`
	Enabled    bool       `gorm:"column:enabled;not null;default:true" json:"enabled"`
	Note       string     `gorm:"column:note;type:text;not null;default:''" json:"note"`
	CreatedBy  *uuid.UUID `gorm:"column:created_by;type:uuid" json:"created_by,omitempty"`
	CreatedAt  time.Time  `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (LogPolicy) TableName() string { return "log_policies" }

func normalizeTenantID(tenantID string) string {
	if tenantID == "" {
		return DefaultTenantID
	}
	return tenantID
}


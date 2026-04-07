package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// DefaultTenantCode is the built-in tenant every record currently belongs to.
// GGE 5.0 Phase 2a only reserves the dimension; the multi-tenant business
// surface (operator UI, cross-tenant routing, isolation strategies) lands in
// later phases as documented in ch.10 of the architecture doc.
const DefaultTenantCode = "default"

// Tenant is the outermost data isolation boundary. workspaces remain the
// team / permission subject inside a tenant.
type Tenant struct {
	ID        uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Code      string         `gorm:"type:varchar(64);not null;uniqueIndex" json:"code"`
	Name      string         `gorm:"type:varchar(150);not null" json:"name"`
	Status    string         `gorm:"type:varchar(20);not null;default:'active'" json:"status"`
	IsDefault bool           `gorm:"not null;default:false" json:"is_default"`
	Meta      MetaJSON       `gorm:"type:jsonb;default:'{}'::jsonb" json:"meta"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (Tenant) TableName() string { return "tenants" }

// TenantScoped is the embed every multi-tenant aware model carries. Phase 2a
// only embeds it on workspaces and workspace_members; subsequent phases roll
// it onto the rest of the schema together with the per-domain refactor.
type TenantScoped struct {
	TenantID uuid.UUID `gorm:"type:uuid;not null;index" json:"tenant_id"`
}

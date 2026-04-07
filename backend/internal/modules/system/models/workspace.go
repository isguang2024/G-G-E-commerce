package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	WorkspaceTypePersonal      = "personal"
	WorkspaceTypeCollaboration = "collaboration"

	WorkspaceStatusActive = "active"

	WorkspaceMemberOwner  = "owner"
	WorkspaceMemberAdmin  = "admin"
	WorkspaceMemberMember = "member"
	WorkspaceMemberViewer = "viewer"
)

type Workspace struct {
	ID                       uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantScoped
	WorkspaceType            string         `gorm:"type:varchar(20);not null;index" json:"workspace_type"`
	Name                     string         `gorm:"type:varchar(150);not null" json:"name"`
	Code                     string         `gorm:"type:varchar(150);not null;uniqueIndex" json:"code"`
	OwnerUserID              *uuid.UUID     `gorm:"type:uuid;index" json:"owner_user_id"`
	CollaborationWorkspaceID *uuid.UUID     `gorm:"column:collaboration_workspace_id;type:uuid;index" json:"collaboration_workspace_id"`
	Status                   string         `gorm:"type:varchar(20);not null;default:'active'" json:"status"`
	Meta                     MetaJSON       `gorm:"type:jsonb;default:'{}'::jsonb" json:"meta"`
	CreatedAt                time.Time      `json:"created_at"`
	UpdatedAt                time.Time      `json:"updated_at"`
	DeletedAt                gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (Workspace) TableName() string {
	return "workspaces"
}

type WorkspaceMember struct {
	ID                             uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantScoped
	WorkspaceID                    uuid.UUID      `gorm:"type:uuid;not null;index" json:"workspace_id"`
	UserID                         uuid.UUID      `gorm:"type:uuid;not null;index" json:"user_id"`
	MemberType                     string         `gorm:"type:varchar(20);not null;default:'member'" json:"member_type"`
	Status                         string         `gorm:"type:varchar(20);not null;default:'active'" json:"status"`
	CollaborationWorkspaceMemberID *uuid.UUID     `gorm:"column:collaboration_workspace_member_id;type:uuid;index" json:"collaboration_workspace_member_id"`
	CreatedAt                      time.Time      `json:"created_at"`
	UpdatedAt                      time.Time      `json:"updated_at"`
	DeletedAt                      gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (WorkspaceMember) TableName() string {
	return "workspace_members"
}

type WorkspaceRoleBinding struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	WorkspaceID uuid.UUID      `gorm:"type:uuid;not null;index" json:"workspace_id"`
	UserID      uuid.UUID      `gorm:"type:uuid;not null;index" json:"user_id"`
	RoleID      uuid.UUID      `gorm:"type:uuid;not null;index" json:"role_id"`
	Enabled     bool           `gorm:"not null;default:true" json:"enabled"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (WorkspaceRoleBinding) TableName() string {
	return "workspace_role_bindings"
}

type WorkspaceFeaturePackage struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	WorkspaceID uuid.UUID      `gorm:"type:uuid;not null;index" json:"workspace_id"`
	PackageID   uuid.UUID      `gorm:"type:uuid;not null;index" json:"package_id"`
	Enabled     bool           `gorm:"not null;default:true" json:"enabled"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (WorkspaceFeaturePackage) TableName() string {
	return "workspace_feature_packages"
}

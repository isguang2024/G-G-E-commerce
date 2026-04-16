package user

import (
	"github.com/google/uuid"

	"github.com/maben/backend/internal/modules/system/models"
)

type MetaJSON = models.MetaJSON
type User = models.User
type Role = models.Role
type RoleAppScope = models.RoleAppScope
type App = models.App
type AppHostBinding = models.AppHostBinding
type Menu = models.Menu
type MenuDefinition = models.MenuDefinition
type SpaceMenuPlacement = models.SpaceMenuPlacement
type UIPage = models.UIPage
type UserRole = models.UserRole
type PermissionGroup = models.PermissionGroup
type PermissionKey = models.PermissionKey
type FeaturePackage = models.FeaturePackage
type FeaturePackageBundle = models.FeaturePackageBundle
type FeaturePackageKey = models.FeaturePackageKey
type FeaturePackageMenu = models.FeaturePackageMenu
type CollaborationWorkspaceFeaturePackage = models.CollaborationWorkspaceFeaturePackage
type UserFeaturePackage = models.UserFeaturePackage
type RoleFeaturePackage = models.RoleFeaturePackage
type RoleHiddenMenu = models.RoleHiddenMenu
type RoleDisabledAction = models.RoleDisabledAction
type RoleDataPermission = models.RoleDataPermission
type CollaborationWorkspaceBlockedMenu = models.CollaborationWorkspaceBlockedMenu
type CollaborationWorkspaceBlockedAction = models.CollaborationWorkspaceBlockedAction
type UserActionPermission = models.UserActionPermission
type UserHiddenMenu = models.UserHiddenMenu
type APIEndpoint = models.APIEndpoint
type APIEndpointCategory = models.APIEndpointCategory
type APIEndpointPermissionBinding = models.APIEndpointPermissionBinding
type CollaborationWorkspace = models.CollaborationWorkspace
type CollaborationWorkspaceMember = models.CollaborationWorkspaceMember
type Workspace = models.Workspace
type WorkspaceMember = models.WorkspaceMember
type WorkspaceRoleBinding = models.WorkspaceRoleBinding
type WorkspaceFeaturePackage = models.WorkspaceFeaturePackage
type MemberSearchParams = models.MemberSearchParams
type APIKey = models.APIKey
type MediaAsset = models.MediaAsset
type RiskOperationAudit = models.RiskOperationAudit
type FeaturePackageVersion = models.FeaturePackageVersion
type PermissionBatchTemplate = models.PermissionBatchTemplate

type RoleKeyPermission struct {
	RoleID uuid.UUID `json:"role_id"`
	KeyID  uuid.UUID `json:"action_id"`
}


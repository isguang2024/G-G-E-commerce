package user

import (
	"github.com/google/uuid"

	"github.com/gg-ecommerce/backend/internal/modules/system/models"
)

type MetaJSON = models.MetaJSON
type User = models.User
type Role = models.Role
type Menu = models.Menu
type UserRole = models.UserRole
type PermissionAction = models.PermissionAction
type FeaturePackage = models.FeaturePackage
type FeaturePackageBundle = models.FeaturePackageBundle
type FeaturePackageAction = models.FeaturePackageAction
type FeaturePackageMenu = models.FeaturePackageMenu
type TeamFeaturePackage = models.TeamFeaturePackage
type UserFeaturePackage = models.UserFeaturePackage
type RoleFeaturePackage = models.RoleFeaturePackage
type RoleHiddenMenu = models.RoleHiddenMenu
type RoleDisabledAction = models.RoleDisabledAction
type RoleDataPermission = models.RoleDataPermission
type TeamBlockedMenu = models.TeamBlockedMenu
type TeamBlockedAction = models.TeamBlockedAction
type UserActionPermission = models.UserActionPermission
type UserHiddenMenu = models.UserHiddenMenu
type APIEndpoint = models.APIEndpoint
type Tenant = models.Tenant
type TenantMember = models.TenantMember
type MemberSearchParams = models.MemberSearchParams
type APIKey = models.APIKey
type MediaAsset = models.MediaAsset
type MenuBackup = models.MenuBackup

type RoleActionPermission struct {
	RoleID   uuid.UUID `json:"role_id"`
	ActionID uuid.UUID `json:"action_id"`
}

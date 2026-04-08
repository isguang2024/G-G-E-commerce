// subroute_core.go — userSubrouteCore holds the business-logic helpers that
// SubrouteService needs. It mirrors the deps of UserHandler but is entirely
// independent of gin and is not exposed outside this package.
package user

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/gg-ecommerce/backend/internal/modules/system/models"
	appctx "github.com/gg-ecommerce/backend/internal/pkg/appctx"
	"github.com/gg-ecommerce/backend/internal/pkg/authorization"
	"github.com/gg-ecommerce/backend/internal/pkg/collaborationworkspaceboundary"
	"github.com/gg-ecommerce/backend/internal/pkg/permissionkey"
	"github.com/gg-ecommerce/backend/internal/pkg/platformaccess"
	"github.com/gg-ecommerce/backend/internal/pkg/workspacerolebinding"
)

// userSubrouteCore carries the minimal set of dependencies needed to implement
// the SubrouteService methods without touching gin.Context.
type userSubrouteCore struct {
	db          *gorm.DB
	userService UserService
	featurePkgRepo interface {
		GetByIDs(ids []uuid.UUID) ([]FeaturePackage, error)
	}
	keyRepo interface {
		GetByPermissionKey(permissionKey string) (*PermissionKey, error)
	}
	personalWorkspaceAccessService platformaccess.Service
	boundaryService                collaborationworkspaceboundary.Service
	authzService                   interface {
		Authorize(userID uuid.UUID, collaborationWorkspaceID *uuid.UUID, permissionKey string, legacy ...string) (bool, *models.PermissionKey, error)
		AuthorizeInApp(userID uuid.UUID, collaborationWorkspaceID *uuid.UUID, appKey string, permissionKey string, legacy ...string) (bool, *models.PermissionKey, error)
	}
	roleRepo interface {
		GetByIDs(ids []uuid.UUID) ([]Role, error)
	}
	userRoleRepo interface {
		GetEffectiveActiveRoleIDsByUserAndCollaborationWorkspace(userID uuid.UUID, collaborationWorkspaceID *uuid.UUID) ([]uuid.UUID, error)
	}
	collaborationWorkspaceMemberRepo interface {
		GetByUserAndCollaborationWorkspace(userID, collaborationWorkspaceID uuid.UUID) (*CollaborationWorkspaceMember, error)
		GetCollaborationWorkspacesByUserID(userID uuid.UUID) ([]CollaborationWorkspace, error)
	}
	menuRepo interface {
		ListAll() ([]Menu, error)
	}
	refresher interface {
		RefreshPersonalWorkspaceUser(userID uuid.UUID) error
		RefreshCollaborationWorkspace(collaborationWorkspaceID uuid.UUID) error
	}
	logger *zap.Logger
}

// ── snapshot helpers ──────────────────────────────────────────────────────────

func (c *userSubrouteCore) getPersonalWorkspaceSnapshot(userID uuid.UUID, appKey string) (*platformaccess.Snapshot, error) {
	if c.personalWorkspaceAccessService == nil {
		return &platformaccess.Snapshot{
			DirectPackageIDs:   []uuid.UUID{},
			ExpandedPackageIDs: []uuid.UUID{},
			ActionIDs:          []uuid.UUID{},
			ActionSourceMap:    map[uuid.UUID][]uuid.UUID{},
			AvailableMenuIDs:   []uuid.UUID{},
			AvailableMenuMap:   map[uuid.UUID][]uuid.UUID{},
			MenuIDs:            []uuid.UUID{},
			MenuSourceMap:      map[uuid.UUID][]uuid.UUID{},
			HasPackageConfig:   false,
		}, nil
	}
	snapshot, err := c.personalWorkspaceAccessService.GetSnapshot(userID, appctx.NormalizeAppKey(appKey))
	if err != nil {
		return nil, err
	}
	if snapshot == nil {
		return &platformaccess.Snapshot{
			DirectPackageIDs:   []uuid.UUID{},
			ExpandedPackageIDs: []uuid.UUID{},
			ActionIDs:          []uuid.UUID{},
			ActionSourceMap:    map[uuid.UUID][]uuid.UUID{},
			AvailableMenuIDs:   []uuid.UUID{},
			AvailableMenuMap:   map[uuid.UUID][]uuid.UUID{},
			MenuIDs:            []uuid.UUID{},
			MenuSourceMap:      map[uuid.UUID][]uuid.UUID{},
			HasPackageConfig:   false,
		}, nil
	}
	return snapshot, nil
}

func (c *userSubrouteCore) getPersonalWorkspaceSnapshotRecord(userID uuid.UUID, appKey string) (*models.PersonalWorkspaceAccessSnapshot, error) {
	if c.db == nil {
		return nil, nil
	}
	var record models.PersonalWorkspaceAccessSnapshot
	if err := c.db.Where("app_key = ? AND user_id = ?", appctx.NormalizeAppKey(appKey), userID).First(&record).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &record, nil
}

func (c *userSubrouteCore) getCollaborationWorkspaceSnapshotRecord(collaborationWorkspaceID uuid.UUID, appKey string) (*models.CollaborationWorkspaceAccessSnapshot, error) {
	if c.db == nil {
		return nil, nil
	}
	var record models.CollaborationWorkspaceAccessSnapshot
	if err := c.db.Where("app_key = ? AND collaboration_workspace_id = ?", appctx.NormalizeAppKey(appKey), collaborationWorkspaceID).First(&record).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &record, nil
}

func (c *userSubrouteCore) getCollaborationWorkspaceRoleSnapshotRecord(collaborationWorkspaceID, roleID uuid.UUID, appKey string) (*models.CollaborationWorkspaceRoleAccessSnapshot, error) {
	if c.db == nil {
		return nil, nil
	}
	var record models.CollaborationWorkspaceRoleAccessSnapshot
	if err := c.db.Where("app_key = ? AND collaboration_workspace_id = ? AND role_id = ?", appctx.NormalizeAppKey(appKey), collaborationWorkspaceID, roleID).First(&record).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &record, nil
}

// ── permission menu IDs ───────────────────────────────────────────────────────

func (c *userSubrouteCore) getPermissionMenuIDs(userID uuid.UUID, collaborationWorkspaceID *uuid.UUID, appKey string) ([]uuid.UUID, error) {
	userEntity, err := c.userService.Get(userID)
	if err != nil {
		return nil, err
	}
	if collaborationWorkspaceID == nil {
		if userEntity.IsSuperAdmin {
			return c.listEnabledMenuIDs(appKey)
		}
		snapshot, err := c.getPersonalWorkspaceSnapshot(userID, appKey)
		if err != nil {
			return nil, err
		}
		return snapshot.MenuIDs, nil
	}
	return c.getCollaborationWorkspacePermissionMenuIDs(userID, *collaborationWorkspaceID, appKey)
}

func (c *userSubrouteCore) getCollaborationWorkspacePermissionMenuIDs(userID, collaborationWorkspaceID uuid.UUID, appKey string) ([]uuid.UUID, error) {
	if c.userRoleRepo == nil || c.roleRepo == nil || c.boundaryService == nil {
		return c.finalizePermissionMenuIDs(nil, appKey)
	}
	roleIDs, err := c.userRoleRepo.GetEffectiveActiveRoleIDsByUserAndCollaborationWorkspace(userID, &collaborationWorkspaceID)
	if err != nil {
		return nil, err
	}
	if len(roleIDs) == 0 {
		return c.finalizePermissionMenuIDs(nil, appKey)
	}
	roles, err := c.roleRepo.GetByIDs(roleIDs)
	if err != nil {
		return nil, err
	}
	roleMap := make(map[uuid.UUID]Role, len(roles))
	for _, role := range roles {
		roleMap[role.ID] = role
	}
	menuSet := make(map[uuid.UUID]struct{})
	for _, roleID := range roleIDs {
		role, ok := roleMap[roleID]
		if !ok {
			continue
		}
		snapshot, snapshotErr := c.boundaryService.GetRoleSnapshot(collaborationWorkspaceID, roleID, role.CollaborationWorkspaceID == nil, appctx.NormalizeAppKey(appKey))
		if snapshotErr != nil {
			return nil, snapshotErr
		}
		for _, menuID := range snapshot.MenuIDs {
			menuSet[menuID] = struct{}{}
		}
	}
	menuIDs := make([]uuid.UUID, 0, len(menuSet))
	for menuID := range menuSet {
		menuIDs = append(menuIDs, menuID)
	}
	return c.finalizePermissionMenuIDs(menuIDs, appKey)
}

func (c *userSubrouteCore) finalizePermissionMenuIDs(menuIDs []uuid.UUID, appKey string) ([]uuid.UUID, error) {
	allMenus, err := c.menuRepo.ListAll()
	if err != nil {
		return nil, err
	}
	enabledSet := make(map[uuid.UUID]struct{}, len(allMenus))
	publicIDs := make([]uuid.UUID, 0)
	for _, menu := range allMenus {
		if appctx.NormalizeAppKey(menu.AppKey) != appctx.NormalizeAppKey(appKey) {
			continue
		}
		if !isMenuEnabled(menu) {
			continue
		}
		enabledSet[menu.ID] = struct{}{}
		if isPublicMenu(menu) {
			publicIDs = append(publicIDs, menu.ID)
		}
	}
	result := make([]uuid.UUID, 0, len(menuIDs)+len(publicIDs))
	seen := make(map[uuid.UUID]struct{}, len(menuIDs)+len(publicIDs))
	for _, menuID := range mergeUUIDLists(menuIDs, publicIDs) {
		if _, ok := enabledSet[menuID]; !ok {
			continue
		}
		if _, ok := seen[menuID]; ok {
			continue
		}
		seen[menuID] = struct{}{}
		result = append(result, menuID)
	}
	return result, nil
}

func (c *userSubrouteCore) listEnabledMenuIDs(appKey string) ([]uuid.UUID, error) {
	allMenus, err := c.menuRepo.ListAll()
	if err != nil {
		return nil, err
	}
	result := make([]uuid.UUID, 0, len(allMenus))
	for _, menu := range allMenus {
		if appctx.NormalizeAppKey(menu.AppKey) != appctx.NormalizeAppKey(appKey) {
			continue
		}
		if !isMenuEnabled(menu) {
			continue
		}
		result = append(result, menu.ID)
	}
	return result, nil
}

// ── permission diagnosis ──────────────────────────────────────────────────────

func (c *userSubrouteCore) buildPermissionDiagnosis(userID uuid.UUID, collaborationWorkspaceID *uuid.UUID, rawPermissionKey string, appKey string) (gin.H, error) {
	userEntity, err := c.userService.Get(userID)
	if err != nil {
		return nil, err
	}

	permissionKeyValue := permissionkey.Normalize(rawPermissionKey)
	userInfo := gin.H{
		"id":             userEntity.ID.String(),
		"user_name":      userEntity.Username,
		"nick_name":      userEntity.Nickname,
		"status":         userEntity.Status,
		"is_super_admin": userEntity.IsSuperAdmin,
	}

	if collaborationWorkspaceID == nil {
		snapshot, err := c.getPersonalWorkspaceSnapshot(userID, appKey)
		if err != nil {
			return nil, err
		}
		meta, err := c.getPersonalWorkspaceSnapshotRecord(userID, appKey)
		if err != nil {
			return nil, err
		}
		payload := gin.H{
			"user":      userInfo,
			"context":   gin.H{"type": "personal", "binding_workspace_id": "", "current_collaboration_workspace_id": "", "current_collaboration_workspace_name": ""},
			"snapshot":  buildPersonalWorkspaceSnapshotSummary(snapshot, meta),
			"roles":     []gin.H{},
			"diagnosis": nil,
		}
		if permissionKeyValue != "" {
			diagnosis, err := c.buildPersonalWorkspacePermissionDiagnosis(userEntity, snapshot, permissionKeyValue, appKey)
			if err != nil {
				return nil, err
			}
			payload["diagnosis"] = diagnosis
		}
		return payload, nil
	}

	collaborationWorkspaceSnapshot, err := c.boundaryService.GetSnapshot(*collaborationWorkspaceID, appctx.NormalizeAppKey(appKey))
	if err != nil {
		return nil, err
	}
	currentCollaborationWorkspaceID := ""
	if workspace, workspaceErr := workspacerolebinding.GetCollaborationWorkspaceByCollaborationWorkspaceID(c.db, *collaborationWorkspaceID); workspaceErr == nil && workspace != nil {
		currentCollaborationWorkspaceID = workspace.ID.String()
	}
	collaborationWorkspaceMeta, err := c.getCollaborationWorkspaceSnapshotRecord(*collaborationWorkspaceID, appKey)
	if err != nil {
		return nil, err
	}
	roleStates, err := c.buildCollaborationWorkspaceRoleStates(userID, *collaborationWorkspaceID, permissionKeyValue, appKey)
	if err != nil {
		return nil, err
	}
	payload := gin.H{
		"user": userInfo,
		"context": gin.H{
			"type":                                 "collaboration",
			"binding_workspace_id":                 currentCollaborationWorkspaceID,
			"current_collaboration_workspace_id":   collaborationWorkspaceID.String(),
			"current_collaboration_workspace_name": "",
		},
		"snapshot":                         buildCollaborationWorkspaceSnapshotSummary(collaborationWorkspaceSnapshot, collaborationWorkspaceMeta),
		"roles":                            roleStates,
		"collaboration_workspace_member":   c.buildCollaborationWorkspaceMemberMap(userID, *collaborationWorkspaceID),
		"collaboration_workspace_packages": c.buildPackageMapsByIDs(collaborationWorkspaceSnapshot.ExpandedPackageIDs),
		"diagnosis":                        nil,
	}
	if permissionKeyValue != "" {
		diagnosis, err := c.buildCollaborationWorkspacePermissionDiagnosis(userEntity, *collaborationWorkspaceID, collaborationWorkspaceSnapshot, roleStates, permissionKeyValue, appKey)
		if err != nil {
			return nil, err
		}
		payload["diagnosis"] = diagnosis
	}
	return payload, nil
}

func (c *userSubrouteCore) buildPersonalWorkspacePermissionDiagnosis(userEntity *User, snapshot *platformaccess.Snapshot, permissionKeyValue string, appKey string) (gin.H, error) {
	allowed, actionDef, authErr := c.authzService.AuthorizeInApp(userEntity.ID, nil, appctx.NormalizeAppKey(appKey), permissionKeyValue)
	actionDetail, err := c.loadPermissionKeyDetail(permissionKeyValue)
	if err != nil {
		return nil, err
	}

	var actionID uuid.UUID
	sourcePackageIDs := []uuid.UUID{}
	inSnapshot := false
	if actionDef != nil {
		actionID = actionDef.ID
		inSnapshot = containsUUID(snapshot.ActionIDs, actionID)
		sourcePackageIDs = append(sourcePackageIDs, snapshot.ActionSourceMap[actionID]...)
	}
	bypassedBySuperAdmin := userEntity.IsSuperAdmin && authErr == nil && allowed && !inSnapshot
	reasons := buildPermissionDiagnosisReasons(authErr, allowed, "personal", bypassedBySuperAdmin)

	return gin.H{
		"permission_key":          permissionKeyValue,
		"allowed":                 authErr == nil && allowed,
		"reason_text":             strings.Join(reasons, "；"),
		"reasons":                 reasons,
		"matched_in_snapshot":     inSnapshot,
		"bypassed_by_super_admin": bypassedBySuperAdmin,
		"action":                  buildPermissionActionMap(actionDetail, actionDef),
		"source_packages":         c.buildPackageMapsByIDs(sourcePackageIDs),
		"role_results":            []gin.H{},
	}, nil
}

func (c *userSubrouteCore) buildCollaborationWorkspacePermissionDiagnosis(userEntity *User, collaborationWorkspaceID uuid.UUID, collaborationWorkspaceSnapshot *collaborationworkspaceboundary.Snapshot, roleStates []gin.H, permissionKeyValue string, appKey string) (gin.H, error) {
	allowed, actionDef, authErr := c.authzService.AuthorizeInApp(userEntity.ID, &collaborationWorkspaceID, appctx.NormalizeAppKey(appKey), permissionKeyValue)
	actionDetail, err := c.loadPermissionKeyDetail(permissionKeyValue)
	if err != nil {
		return nil, err
	}

	var actionID uuid.UUID
	blockedByCollaborationWorkspace := false
	sourcePackageIDs := []uuid.UUID{}
	if actionDef != nil {
		actionID = actionDef.ID
		blockedByCollaborationWorkspace = containsUUID(collaborationWorkspaceSnapshot.BlockedIDs, actionID)
		for _, roleItem := range roleStates {
			if sourceItems, ok := roleItem["source_package_ids"].([]uuid.UUID); ok {
				sourcePackageIDs = mergeUUIDLists(sourcePackageIDs, sourceItems)
			}
		}
	}
	inSnapshot := actionDef != nil && containsUUID(collaborationWorkspaceSnapshot.EffectiveIDs, actionID)
	bypassedBySuperAdmin := userEntity.IsSuperAdmin && authErr == nil && allowed && !inSnapshot
	reasons := buildPermissionDiagnosisReasons(authErr, allowed, "collaboration", bypassedBySuperAdmin)
	memberStatus, memberMatched, err := c.getCollaborationWorkspaceMemberDiagnosis(userEntity.ID, collaborationWorkspaceID)
	if err != nil {
		return nil, err
	}
	roleMatched, roleDisabled, roleAvailable := summarizeRoleChain(roleStates)
	boundaryConfigured := len(collaborationWorkspaceSnapshot.PackageIDs) > 0 || len(collaborationWorkspaceSnapshot.ExpandedPackageIDs) > 0 || len(collaborationWorkspaceSnapshot.BlockedIDs) > 0
	boundaryState, denialStage, denialReason := buildCollaborationWorkspaceDiagnosisDecision(authErr, allowed, blockedByCollaborationWorkspace, boundaryConfigured, inSnapshot, memberStatus, memberMatched, roleMatched, roleDisabled, roleAvailable, bypassedBySuperAdmin)

	return gin.H{
		"permission_key":                     permissionKeyValue,
		"allowed":                            authErr == nil && allowed,
		"reason_text":                        strings.Join(reasons, "；"),
		"reasons":                            reasons,
		"matched_in_snapshot":                inSnapshot,
		"bypassed_by_super_admin":            bypassedBySuperAdmin,
		"blocked_by_collaboration_workspace": blockedByCollaborationWorkspace,
		"denial_stage":                       denialStage,
		"denial_reason":                      denialReason,
		"member_status":                      memberStatus,
		"member_matched":                     memberMatched,
		"boundary_state":                     boundaryState,
		"boundary_configured":                boundaryConfigured,
		"role_chain_matched":                 roleMatched,
		"role_chain_disabled":                roleDisabled,
		"role_chain_available":               roleAvailable,
		"action":                             buildPermissionActionMap(actionDetail, actionDef),
		"source_packages":                    c.buildPackageMapsByIDs(sourcePackageIDs),
		"role_results":                       roleStates,
	}, nil
}

func (c *userSubrouteCore) buildCollaborationWorkspaceRoleStates(userID, collaborationWorkspaceID uuid.UUID, permissionKeyValue string, appKey string) ([]gin.H, error) {
	if c.userRoleRepo == nil || c.roleRepo == nil || c.boundaryService == nil {
		return []gin.H{}, nil
	}
	roleIDs, err := c.userRoleRepo.GetEffectiveActiveRoleIDsByUserAndCollaborationWorkspace(userID, &collaborationWorkspaceID)
	if err != nil {
		return nil, err
	}
	if len(roleIDs) == 0 {
		return []gin.H{}, nil
	}
	roles, err := c.roleRepo.GetByIDs(roleIDs)
	if err != nil {
		return nil, err
	}

	actionDetail, err := c.loadPermissionKeyDetail(permissionKeyValue)
	if err != nil {
		return nil, err
	}

	roleStates := make([]gin.H, 0, len(roles))
	for _, role := range roles {
		inheritAll := role.CollaborationWorkspaceID == nil
		snapshot, err := c.boundaryService.GetRoleSnapshot(collaborationWorkspaceID, role.ID, inheritAll, appctx.NormalizeAppKey(appKey))
		if err != nil {
			return nil, err
		}
		roleMeta, err := c.getCollaborationWorkspaceRoleSnapshotRecord(collaborationWorkspaceID, role.ID, appKey)
		if err != nil {
			return nil, err
		}

		roleState := gin.H{
			"role_id":                role.ID.String(),
			"role_code":              role.Code,
			"role_name":              role.Name,
			"inherited":              snapshot.Inherited,
			"refreshed_at":           formatTimeValue(timeValue(roleMeta, func(item *models.CollaborationWorkspaceRoleAccessSnapshot) time.Time { return item.RefreshedAt })),
			"available_action_count": len(snapshot.AvailableActionIDs),
			"disabled_action_count":  len(snapshot.DisabledActionIDs),
			"effective_action_count": len(snapshot.ActionIDs),
			"matched":                false,
			"disabled":               false,
			"available":              false,
			"source_packages":        []gin.H{},
			"source_package_ids":     []uuid.UUID{},
		}
		if actionDetail != nil {
			roleState["available"] = containsUUID(snapshot.AvailableActionIDs, actionDetail.ID)
			roleState["disabled"] = containsUUID(snapshot.DisabledActionIDs, actionDetail.ID)
			roleState["matched"] = containsUUID(snapshot.ActionIDs, actionDetail.ID)
			sourceIDs := snapshot.ActionSourceMap[actionDetail.ID]
			roleState["source_package_ids"] = sourceIDs
			roleState["source_packages"] = c.buildPackageMapsByIDs(sourceIDs)
		}
		roleStates = append(roleStates, roleState)
	}
	return roleStates, nil
}

func (c *userSubrouteCore) loadPermissionKeyDetail(permissionKeyValue string) (*PermissionKey, error) {
	if strings.TrimSpace(permissionKeyValue) == "" || c.keyRepo == nil {
		return nil, nil
	}
	item, err := c.keyRepo.GetByPermissionKey(permissionKeyValue)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return item, nil
}

func (c *userSubrouteCore) buildPackageMapsByIDs(ids []uuid.UUID) []gin.H {
	if len(ids) == 0 || c.featurePkgRepo == nil {
		return []gin.H{}
	}
	items, err := c.featurePkgRepo.GetByIDs(ids)
	if err != nil {
		if c.logger != nil {
			c.logger.Warn("Load feature packages for permission diagnosis failed", zap.Error(err))
		}
		return []gin.H{}
	}
	return featurePackageListToMaps(items)
}

func (c *userSubrouteCore) buildCollaborationWorkspaceMemberMap(userID, collaborationWorkspaceID uuid.UUID) gin.H {
	if c.collaborationWorkspaceMemberRepo == nil {
		return nil
	}
	member, err := c.collaborationWorkspaceMemberRepo.GetByUserAndCollaborationWorkspace(userID, collaborationWorkspaceID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return gin.H{
				"matched": false,
				"status":  "missing",
			}
		}
		if c.logger != nil {
			c.logger.Warn("Load collaboration workspace member for permission diagnosis failed", zap.Error(err))
		}
		return nil
	}
	memberType := mapRoleCodeToMemberType(member.RoleCode)
	workspaceID := ""
	workspaceType := ""
	if workspace, err := workspacerolebinding.GetCollaborationWorkspaceByCollaborationWorkspaceID(c.db, collaborationWorkspaceID); err == nil && workspace != nil {
		workspaceID = workspace.ID.String()
		workspaceType = workspace.WorkspaceType
		var workspaceMember models.WorkspaceMember
		if err := c.db.Where("workspace_id = ? AND user_id = ? AND deleted_at IS NULL", workspace.ID, userID).First(&workspaceMember).Error; err == nil {
			if strings.TrimSpace(workspaceMember.MemberType) != "" {
				memberType = workspaceMember.MemberType
			}
		}
	}
	return gin.H{
		"id":                                 member.ID.String(),
		"current_collaboration_workspace_id": member.CollaborationWorkspaceID.String(),
		"user_id":                            member.UserID.String(),
		"role_code":                          member.RoleCode,
		"member_type":                        memberType,
		"binding_workspace_id":               workspaceID,
		"binding_workspace_type":             workspaceType,
		"status":                             member.Status,
		"matched":                            true,
	}
}

func (c *userSubrouteCore) getCollaborationWorkspaceMemberDiagnosis(userID, collaborationWorkspaceID uuid.UUID) (string, bool, error) {
	if c.db == nil {
		return "", false, nil
	}
	var member models.CollaborationWorkspaceMember
	if err := c.db.Where("user_id = ? AND collaboration_workspace_id = ?", userID, collaborationWorkspaceID).First(&member).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return "missing", false, nil
		}
		return "", false, err
	}
	return member.Status, member.Status == "active", nil
}

// ── free-function helpers (moved from handler.go, used by the methods above) ──

func buildUserMenuSourceMaps(sourceMap map[uuid.UUID][]uuid.UUID) []gin.H {
	if len(sourceMap) == 0 {
		return []gin.H{}
	}
	items := make([]gin.H, 0, len(sourceMap))
	for menuID, packageIDs := range sourceMap {
		items = append(items, gin.H{
			"menu_id":     menuID.String(),
			"package_ids": packageIDsToStrings(packageIDs),
		})
	}
	return items
}

func uuidSliceToSet(ids []uuid.UUID) map[uuid.UUID]bool {
	result := make(map[uuid.UUID]bool, len(ids))
	for _, id := range ids {
		result[id] = true
	}
	return result
}

func excludeUUIDs(source []uuid.UUID, selected []uuid.UUID) []uuid.UUID {
	selectedSet := uuidSliceToSet(selected)
	result := make([]uuid.UUID, 0, len(source))
	for _, item := range source {
		if selectedSet[item] {
			continue
		}
		result = append(result, item)
	}
	return result
}

func packageIDsToStrings(ids []uuid.UUID) []string {
	result := make([]string, 0, len(ids))
	for _, id := range ids {
		result = append(result, id.String())
	}
	return result
}

func filterMenusByApp(allMenus []Menu, appKey string) []Menu {
	if len(allMenus) == 0 {
		return []Menu{}
	}
	result := make([]Menu, 0, len(allMenus))
	for _, menu := range allMenus {
		if appctx.NormalizeAppKey(menu.AppKey) != appctx.NormalizeAppKey(appKey) {
			continue
		}
		result = append(result, menu)
	}
	return result
}

func supportsPersonalWorkspaceContext(contextType string) bool {
	return contextType == "" || contextType == "personal" || contextType == "common"
}

func buildMenuTree(allMenus []Menu, allowedIDs map[uuid.UUID]bool) []gin.H {
	parentMap := make(map[uuid.UUID]*uuid.UUID, len(allMenus))
	childrenMap := make(map[uuid.UUID][]Menu, len(allMenus))
	rootMenus := make([]Menu, 0)
	for _, menu := range allMenus {
		parentMap[menu.ID] = menu.ParentID
		if menu.ParentID == nil {
			rootMenus = append(rootMenus, menu)
			continue
		}
		childrenMap[*menu.ParentID] = append(childrenMap[*menu.ParentID], menu)
	}

	allowedMenuIDs := make(map[uuid.UUID]bool, len(allowedIDs))
	for menuID := range allowedIDs {
		allowedMenuIDs[menuID] = true
		parentID := parentMap[menuID]
		for parentID != nil && *parentID != (uuid.UUID{}) {
			allowedMenuIDs[*parentID] = true
			parentID = parentMap[*parentID]
		}
	}

	var build func(parentID *uuid.UUID) []gin.H
	build = func(parentID *uuid.UUID) []gin.H {
		var result []gin.H
		var menus []Menu
		if parentID == nil {
			menus = rootMenus
		} else {
			menus = childrenMap[*parentID]
		}
		for _, menu := range menus {
			if !allowedMenuIDs[menu.ID] {
				continue
			}
			children := build(&menu.ID)
			node := gin.H{
				"id":        menu.ID.String(),
				"name":      menu.Name,
				"title":     menu.Title,
				"path":      menu.Path,
				"component": menu.Component,
				"hidden":    menu.Hidden,
				"sort":      menu.SortOrder,
			}
			if len(children) > 0 {
				node["children"] = children
			}
			result = append(result, node)
		}
		return result
	}

	return build(nil)
}

func featurePackageListToMaps(items []FeaturePackage) []gin.H {
	result := make([]gin.H, 0, len(items))
	for _, item := range items {
		result = append(result, gin.H{
			"id":           item.ID.String(),
			"package_key":  item.PackageKey,
			"package_type": item.PackageType,
			"name":         item.Name,
			"description":  item.Description,
			"context_type": item.ContextType,
			"status":       item.Status,
			"is_builtin":   item.IsBuiltin,
			"sort_order":   item.SortOrder,
		})
	}
	return result
}

func mapRoleCodeToMemberType(roleCode string) string {
	switch strings.ToLower(strings.TrimSpace(roleCode)) {
	case "owner":
		return models.WorkspaceMemberOwner
	case "collaboration_workspace_admin", "admin":
		return models.WorkspaceMemberAdmin
	case "viewer":
		return models.WorkspaceMemberViewer
	default:
		return models.WorkspaceMemberMember
	}
}

func buildPermissionActionMap(detail *PermissionKey, runtime *models.PermissionKey) gin.H {
	target := runtime
	if detail == nil && target == nil {
		return nil
	}
	if target == nil && detail != nil {
		target = &models.PermissionKey{
			ID:                    detail.ID,
			PermissionKey:         detail.PermissionKey,
			AppKey:                detail.AppKey,
			Name:                  detail.Name,
			Description:           detail.Description,
			Status:                detail.Status,
			ContextType:           detail.ContextType,
			FeatureKind:           detail.FeatureKind,
			DataPolicy:            detail.DataPolicy,
			AllowedWorkspaceTypes: detail.AllowedWorkspaceTypes,
			ModuleCode:            detail.ModuleCode,
		}
	}

	result := gin.H{
		"id":                      target.ID.String(),
		"permission_key":          target.PermissionKey,
		"app_key":                 target.AppKey,
		"name":                    target.Name,
		"description":             target.Description,
		"status":                  target.Status,
		"context_type":            target.ContextType,
		"feature_kind":            target.FeatureKind,
		"data_policy":             target.DataPolicy,
		"allowed_workspace_types": target.AllowedWorkspaceTypes,
		"module_code":             target.ModuleCode,
	}
	if detail != nil {
		result["self_status"] = detail.Status
		result["module_group"] = permissionGroupToLiteMap(detail.ModuleGroup)
		result["feature_group"] = permissionGroupToLiteMap(detail.FeatureGroup)
		result["module_group_status"] = groupStatus(detail.ModuleGroup)
		result["feature_group_status"] = groupStatus(detail.FeatureGroup)
	}
	return result
}

func buildPersonalWorkspaceSnapshotSummary(snapshot *platformaccess.Snapshot, meta *models.PersonalWorkspaceAccessSnapshot) gin.H {
	return gin.H{
		"refreshed_at":           formatTimeValue(timeValue(meta, func(item *models.PersonalWorkspaceAccessSnapshot) time.Time { return item.RefreshedAt })),
		"updated_at":             formatTimeValue(timeValue(meta, func(item *models.PersonalWorkspaceAccessSnapshot) time.Time { return item.UpdatedAt })),
		"role_count":             len(snapshot.RoleIDs),
		"direct_package_count":   len(snapshot.DirectPackageIDs),
		"expanded_package_count": len(snapshot.ExpandedPackageIDs),
		"action_count":           len(snapshot.ActionIDs),
		"disabled_action_count":  len(snapshot.DisabledActionIDs),
		"menu_count":             len(snapshot.MenuIDs),
		"has_package_config":     snapshot.HasPackageConfig,
	}
}

func buildCollaborationWorkspaceSnapshotSummary(snapshot *collaborationworkspaceboundary.Snapshot, meta *models.CollaborationWorkspaceAccessSnapshot) gin.H {
	return gin.H{
		"refreshed_at":           formatTimeValue(timeValue(meta, func(item *models.CollaborationWorkspaceAccessSnapshot) time.Time { return item.RefreshedAt })),
		"updated_at":             formatTimeValue(timeValue(meta, func(item *models.CollaborationWorkspaceAccessSnapshot) time.Time { return item.UpdatedAt })),
		"direct_package_count":   len(snapshot.PackageIDs),
		"expanded_package_count": len(snapshot.ExpandedPackageIDs),
		"derived_action_count":   len(snapshot.DerivedIDs),
		"blocked_action_count":   len(snapshot.BlockedIDs),
		"effective_action_count": len(snapshot.EffectiveIDs),
	}
}

func permissionGroupToLiteMap(group *PermissionGroup) gin.H {
	if group == nil {
		return nil
	}
	return gin.H{
		"id":     group.ID.String(),
		"code":   group.Code,
		"name":   group.Name,
		"status": group.Status,
	}
}

func groupStatus(group *PermissionGroup) string {
	if group == nil {
		return ""
	}
	return group.Status
}

func buildPermissionDiagnosisReasons(err error, allowed bool, context string, bypassedBySuperAdmin bool) []string {
	switch {
	case bypassedBySuperAdmin:
		return []string{"当前用户是超级管理员，直接放行，不依赖快照命中"}
	case err == nil && allowed:
		return []string{"权限测试通过"}
	case err == authorization.ErrPermissionKeyMissing:
		return []string{"权限键未注册或未找到"}
	case err == authorization.ErrUserInactive:
		return []string{"用户已停用"}
	case err == authorization.ErrCollaborationWorkspaceMemberNotFound:
		return []string{"当前协作空间下无有效成员或角色"}
	case err == authorization.ErrCollaborationWorkspaceContextRequired:
		return []string{"当前权限需要协作空间上下文"}
	case err == authorization.ErrPermissionDenied:
		if context == "collaboration" {
			return []string{"当前协作空间上下文下未生效此权限"}
		}
		return []string{"当前个人空间下未生效此权限"}
	default:
		return []string{"权限未通过"}
	}
}

func summarizeRoleChain(roleStates []gin.H) (matched bool, disabled bool, available bool) {
	for _, roleItem := range roleStates {
		if value, ok := roleItem["matched"].(bool); ok && value {
			matched = true
		}
		if value, ok := roleItem["disabled"].(bool); ok && value {
			disabled = true
		}
		if value, ok := roleItem["available"].(bool); ok && value {
			available = true
		}
	}
	return matched, disabled, available
}

func buildCollaborationWorkspaceDiagnosisDecision(authErr error, allowed bool, blockedByCollaborationWorkspace bool, boundaryConfigured bool, inSnapshot bool, memberStatus string, memberMatched bool, roleMatched bool, roleDisabled bool, roleAvailable bool, bypassedBySuperAdmin bool) (string, string, string) {
	if bypassedBySuperAdmin {
		return "超级管理员直通", "", ""
	}
	if allowed && authErr == nil {
		switch {
		case blockedByCollaborationWorkspace:
			return "拦截", "协作空间边界校验", "协作空间边界已屏蔽该权限"
		case inSnapshot:
			return "命中", "", ""
		case !boundaryConfigured:
			return "未配置", "", ""
		default:
			return "命中", "", ""
		}
	}

	switch authErr {
	case authorization.ErrUserInactive:
		return currentBoundaryState(boundaryConfigured, blockedByCollaborationWorkspace, inSnapshot), "用户状态校验", "当前用户已停用"
	case authorization.ErrCollaborationWorkspaceMemberNotFound:
		return currentBoundaryState(boundaryConfigured, blockedByCollaborationWorkspace, inSnapshot), "协作空间成员校验", "当前用户不是该协作空间有效成员"
	case authorization.ErrCollaborationWorkspaceMemberInactive:
		return currentBoundaryState(boundaryConfigured, blockedByCollaborationWorkspace, inSnapshot), "协作空间成员校验", "协作空间成员状态不是 active"
	case authorization.ErrCollaborationWorkspaceContextRequired:
		return currentBoundaryState(boundaryConfigured, blockedByCollaborationWorkspace, inSnapshot), "协作空间上下文校验", "当前权限需要协作空间上下文"
	case authorization.ErrPermissionKeyMissing:
		return currentBoundaryState(boundaryConfigured, blockedByCollaborationWorkspace, inSnapshot), "权限键校验", "权限键未注册或未找到"
	case authorization.ErrPermissionDenied:
		switch {
		case blockedByCollaborationWorkspace:
			return "拦截", "协作空间边界校验", "协作空间边界未开通或已屏蔽该权限"
		case !memberMatched || memberStatus == "missing":
			return currentBoundaryState(boundaryConfigured, blockedByCollaborationWorkspace, inSnapshot), "协作空间成员校验", "当前用户不是该协作空间有效成员"
		case memberStatus != "" && memberStatus != "active":
			return currentBoundaryState(boundaryConfigured, blockedByCollaborationWorkspace, inSnapshot), "协作空间成员校验", "协作空间成员状态不是 active"
		case roleMatched:
			return currentBoundaryState(boundaryConfigured, blockedByCollaborationWorkspace, inSnapshot), "角色权限校验", "角色链路命中，但最终权限未通过"
		case roleDisabled:
			return currentBoundaryState(boundaryConfigured, blockedByCollaborationWorkspace, inSnapshot), "角色权限校验", "角色链路存在，但权限被角色层禁用"
		case roleAvailable:
			return currentBoundaryState(boundaryConfigured, blockedByCollaborationWorkspace, inSnapshot), "角色权限校验", "角色链路可用，但最终未生效为可执行权限"
		default:
			return currentBoundaryState(boundaryConfigured, blockedByCollaborationWorkspace, inSnapshot), "角色权限校验", "当前协作空间角色未最终授予该权限"
		}
	default:
		return currentBoundaryState(boundaryConfigured, blockedByCollaborationWorkspace, inSnapshot), "", ""
	}
}

func currentBoundaryState(boundaryConfigured bool, blockedByCollaborationWorkspace bool, inSnapshot bool) string {
	switch {
	case blockedByCollaborationWorkspace:
		return "拦截"
	case inSnapshot:
		return "命中"
	case !boundaryConfigured:
		return "未配置"
	default:
		return "未命中"
	}
}

func containsUUID(items []uuid.UUID, target uuid.UUID) bool {
	if target == uuid.Nil {
		return false
	}
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}

func formatTimeValue(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.Format("2006-01-02 15:04:05")
}

func timeValue[T any](item *T, getter func(*T) time.Time) time.Time {
	if item == nil {
		return time.Time{}
	}
	return getter(item)
}

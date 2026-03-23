package authorization

import (
	"errors"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/gg-ecommerce/backend/internal/api/errcode"
	"github.com/gg-ecommerce/backend/internal/modules/system/models"
	"github.com/gg-ecommerce/backend/internal/pkg/permissionkey"
	"github.com/gg-ecommerce/backend/internal/pkg/platformaccess"
	"github.com/gg-ecommerce/backend/internal/pkg/teamboundary"
)

const tenantContextHeader = "X-Tenant-ID"

var (
	ErrUnauthorized            = errors.New("unauthorized")
	ErrUserInactive            = errors.New("user inactive")
	ErrPermissionActionMissing = errors.New("permission action missing")
	ErrTenantContextRequired   = errors.New("tenant context required")
	ErrTenantMemberNotFound    = errors.New("tenant member not found")
	ErrTenantMemberInactive    = errors.New("tenant member inactive")
	ErrPermissionDenied        = errors.New("permission denied")
)

type Service struct {
	db              *gorm.DB
	logger          *zap.Logger
	boundaryService teamboundary.Service
	platformService platformaccess.Service
}

func NewService(db *gorm.DB, logger *zap.Logger) *Service {
	return &Service{
		db:              db,
		logger:          logger,
		boundaryService: teamboundary.NewService(db),
		platformService: platformaccess.NewService(db),
	}
}

func (s *Service) RequireAction(permissionKey string, legacy ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := userIDFromContext(c)
		if err != nil {
			status, resp := errcode.Response(errcode.ErrUnauthorized)
			c.JSON(status, resp)
			c.Abort()
			return
		}

		tenantID, err := resolveTenantID(c)
		if err != nil {
			status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的团队ID")
			c.JSON(status, resp)
			c.Abort()
			return
		}

		allowed, actionDef, authErr := s.Authorize(userID, tenantID, permissionKey, legacy...)
		if authErr != nil {
			s.respondAuthError(c, authErr, resolvePermissionKey(permissionKey, legacy...))
			return
		}
		if !allowed {
			s.respondAuthError(c, ErrPermissionDenied, resolvePermissionKey(permissionKey, legacy...))
			return
		}

		if actionDef != nil {
			c.Set("permission_action_id", actionDef.ID.String())
			c.Set("permission_action_key", actionDef.PermissionKey)
		}
		c.Next()
	}
}

func (s *Service) Authorize(userID uuid.UUID, tenantID *uuid.UUID, permissionKey string, legacy ...string) (bool, *models.PermissionAction, error) {
	var currentUser models.User
	err := s.db.Where("id = ?", userID).First(&currentUser).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil, ErrUnauthorized
		}
		return false, nil, err
	}
	if currentUser.Status != "active" {
		return false, nil, ErrUserInactive
	}
	var actionDef models.PermissionAction
	err = s.db.Where("permission_actions.permission_key = ?", resolvePermissionKey(permissionKey, legacy...)).First(&actionDef).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil, ErrPermissionActionMissing
		}
		return false, nil, err
	}
	if actionDef.Status != "normal" {
		return false, &actionDef, ErrPermissionDenied
	}
	if currentUser.IsSuperAdmin {
		return true, &actionDef, nil
	}
	if tenantID == nil && s.platformService != nil {
		snapshot, err := s.platformService.GetSnapshot(userID)
		if err != nil {
			return false, &actionDef, err
		}
		if snapshot.HasPackageConfig {
			return containsUUID(snapshot.ActionIDs, actionDef.ID), &actionDef, nil
		}
	}

	if tenantID != nil {
		memberActive, boundaryConfigured, boundaryActionSet, ctxErr := s.resolveTenantActionContext(userID, tenantID)
		if ctxErr != nil {
			return false, &actionDef, ctxErr
		}
		if !memberActive {
			return false, &actionDef, ErrTenantMemberNotFound
		}
		if boundaryConfigured && !boundaryActionSet[actionDef.ID] {
			return false, &actionDef, ErrPermissionDenied
		}
	}

	overrideEffect, hasOverride, err := s.resolveUserOverride(userID, tenantID, actionDef.ID)
	if err != nil {
		return false, &actionDef, err
	}
	if hasOverride {
		return overrideEffect == "allow", &actionDef, nil
	}

	roleIDs, err := s.getEffectiveActiveRoleIDs(userID, tenantID)
	if err != nil {
		return false, &actionDef, err
	}
	var rolePermissions []models.RoleActionPermission
	if len(roleIDs) > 0 {
		err = s.db.Where("role_id IN ? AND action_id = ?", roleIDs, actionDef.ID).Find(&rolePermissions).Error
		if err != nil {
			return false, &actionDef, err
		}
	}
	effect := evaluateEffects(rolePermissions)
	if effect == "allow" {
		return true, &actionDef, nil
	}
	return false, &actionDef, nil
}

func (s *Service) GetUserActionKeys(userID uuid.UUID, tenantID *uuid.UUID) ([]string, error) {
	keys, _, err := s.collectUserActionKeys(userID, tenantID)
	if err != nil {
		return nil, err
	}
	return keys, nil
}

func (s *Service) GetUserActionSnapshot(userID uuid.UUID, tenantID *uuid.UUID) ([]string, error) {
	keys, _, err := s.collectUserActionKeys(userID, tenantID)
	if err != nil {
		return nil, err
	}
	return keys, nil
}

func (s *Service) collectUserActionKeys(userID uuid.UUID, tenantID *uuid.UUID) ([]string, []string, error) {
	var currentUser models.User
	if err := s.db.Where("id = ?", userID).First(&currentUser).Error; err != nil {
		return nil, nil, err
	}
	if currentUser.Status != "active" {
		return []string{}, []string{}, nil
	}

	var actions []models.PermissionAction
	if err := s.db.Where("status = ?", "normal").Order("sort_order ASC, created_at DESC").Find(&actions).Error; err != nil {
		return nil, nil, err
	}
	if currentUser.IsSuperAdmin {
		keys := make([]string, 0, len(actions))
		keySet := make(map[string]struct{}, len(actions))
		for _, action := range actions {
			key := action.ResourceCode + ":" + action.ActionCode
			if _, exists := keySet[key]; !exists {
				keys = append(keys, key)
				keySet[key] = struct{}{}
			}
		}
		return keys, []string{}, nil
	}
	if tenantID == nil && s.platformService != nil {
		snapshot, err := s.platformService.GetSnapshot(userID)
		if err != nil {
			return nil, nil, err
		}
		if snapshot.HasPackageConfig {
			return buildActionKeysFromIDs(actions, snapshot.ActionIDs), []string{}, nil
		}
	}

	memberActive := false
	boundaryConfigured := false
	boundaryActionSet := make(map[uuid.UUID]bool)
	if tenantID != nil {
		var ctxErr error
		memberActive, boundaryConfigured, boundaryActionSet, ctxErr = s.resolveTenantActionContext(userID, tenantID)
		if ctxErr != nil {
			if errors.Is(ctxErr, ErrTenantMemberInactive) {
				return []string{}, []string{}, nil
			}
			return nil, nil, ctxErr
		}
		if !memberActive {
			return []string{}, []string{}, nil
		}
	}

	roleIDs, err := s.getEffectiveActiveRoleIDs(userID, tenantID)
	if err != nil {
		return nil, nil, err
	}
	roleEffectMap, err := s.buildRoleEffectMap(roleIDs)
	if err != nil {
		return nil, nil, err
	}

	globalOverrideMap, tenantOverrideMap, err := s.buildUserOverrideMaps(userID, tenantID)
	if err != nil {
		return nil, nil, err
	}

	keys := make([]string, 0, len(actions))
	keySet := make(map[string]struct{}, len(actions))
	for _, action := range actions {
		if tenantID != nil {
			if boundaryConfigured && !boundaryActionSet[action.ID] {
				continue
			}
			if effect, ok := tenantOverrideMap[action.ID]; ok {
				if effect == "allow" {
					appendActionKeys(&keys, keySet, action)
				}
				continue
			}
			if effect, ok := globalOverrideMap[action.ID]; ok {
				if effect == "allow" {
					appendActionKeys(&keys, keySet, action)
				}
				continue
			}
			if roleEffectMap[action.ID] == "allow" {
				appendActionKeys(&keys, keySet, action)
			}
			continue
		}

		if effect, ok := globalOverrideMap[action.ID]; ok {
			if effect == "allow" {
				appendActionKeys(&keys, keySet, action)
			}
			continue
		}
		if roleEffectMap[action.ID] == "allow" {
			appendActionKeys(&keys, keySet, action)
		}
	}

	return keys, []string{}, nil
}

func appendActionKeys(keys *[]string, keySet map[string]struct{}, action models.PermissionAction) {
	key := strings.TrimSpace(action.PermissionKey)
	if key == "" {
		key = action.ResourceCode + ":" + action.ActionCode
	}
	if _, exists := keySet[key]; !exists {
		*keys = append(*keys, key)
		keySet[key] = struct{}{}
	}
}

func buildActionKeysFromIDs(actions []models.PermissionAction, allowedActionIDs []uuid.UUID) []string {
	allowedSet := make(map[uuid.UUID]struct{}, len(allowedActionIDs))
	for _, actionID := range allowedActionIDs {
		allowedSet[actionID] = struct{}{}
	}
	keys := make([]string, 0, len(allowedSet))
	keySet := make(map[string]struct{}, len(allowedSet))
	for _, action := range actions {
		if _, ok := allowedSet[action.ID]; !ok {
			continue
		}
		appendActionKeys(&keys, keySet, action)
	}
	return keys
}

func containsUUID(ids []uuid.UUID, target uuid.UUID) bool {
	for _, id := range ids {
		if id == target {
			return true
		}
	}
	return false
}

func (s *Service) respondAuthError(c *gin.Context, authErr error, permissionKey string) {
	if authErr != nil {
		s.logger.Warn("Permission denied",
			zap.Error(authErr),
			zap.String("permission_key", permissionKey),
			zap.String("path", c.FullPath()))
	}

	switch {
	case errors.Is(authErr, ErrUnauthorized):
		status, resp := errcode.Response(errcode.ErrUnauthorized)
		c.JSON(status, resp)
	case errors.Is(authErr, ErrTenantContextRequired), errors.Is(authErr, ErrTenantMemberNotFound):
		status, resp := errcode.Response(errcode.ErrNoTeam)
		c.JSON(status, resp)
	case errors.Is(authErr, ErrTenantMemberInactive), errors.Is(authErr, ErrUserInactive):
		status, resp := errcode.ResponseWithMsg(errcode.ErrForbidden, "当前账号状态不可用")
		c.JSON(status, resp)
	case errors.Is(authErr, ErrPermissionActionMissing):
		status, resp := errcode.ResponseWithMsg(errcode.ErrForbidden, "功能权限未注册，禁止访问")
		c.JSON(status, resp)
	case errors.Is(authErr, ErrPermissionDenied):
		status, resp := errcode.ResponseWithMsg(errcode.ErrForbidden, "无权限执行该操作")
		c.JSON(status, resp)
	default:
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "权限校验失败")
		c.JSON(status, resp)
	}
	c.Abort()
}

func resolvePermissionKey(permissionKey string, legacy ...string) string {
	if len(legacy) > 0 {
		resourceCode := strings.TrimSpace(permissionKey)
		actionCode := strings.TrimSpace(legacy[0])
		if resourceCode != "" && actionCode != "" {
			return permissionkey.FromLegacy(resourceCode, actionCode).Key
		}
		if strings.TrimSpace(legacy[0]) != "" {
			actionCode = ""
			if len(legacy) > 1 {
				actionCode = legacy[1]
			}
			return permissionkey.FromLegacy(legacy[0], actionCode).Key
		}
	}
	if strings.Contains(permissionKey, ":") {
		parts := strings.SplitN(permissionKey, ":", 2)
		return permissionkey.FromLegacy(parts[0], parts[1]).Key
	}
	return permissionkey.Normalize(permissionKey)
}

func (s *Service) resolveUserOverride(userID uuid.UUID, tenantID *uuid.UUID, actionID uuid.UUID) (string, bool, error) {
	var records []models.UserActionPermission
	query := s.db.Where("user_id = ? AND action_id = ?", userID, actionID)
	if tenantID == nil {
		query = query.Where("tenant_id IS NULL")
	} else {
		query = query.Where("tenant_id IS NULL OR tenant_id = ?", *tenantID)
	}
	err := query.Find(&records).Error
	if err != nil {
		return "", false, err
	}
	if tenantID != nil {
		if effect, ok := pickOverrideEffect(records, tenantID); ok {
			return effect, true, nil
		}
	}
	if effect, ok := pickOverrideEffect(records, nil); ok {
		return effect, true, nil
	}
	return "", false, nil
}

func pickOverrideEffect(records []models.UserActionPermission, tenantID *uuid.UUID) (string, bool) {
	foundAllow := false
	for _, record := range records {
		if !sameTenant(record.TenantID, tenantID) {
			continue
		}
		if record.Effect == "deny" {
			return "deny", true
		}
		if record.Effect == "allow" {
			foundAllow = true
		}
	}
	if foundAllow {
		return "allow", true
	}
	return "", false
}

func evaluateEffects(records []models.RoleActionPermission) string {
	if len(records) > 0 {
		return "allow"
	}
	return ""
}

func userIDFromContext(c *gin.Context) (uuid.UUID, error) {
	value, ok := c.Get("user_id")
	if !ok {
		return uuid.Nil, ErrUnauthorized
	}
	userIDStr, ok := value.(string)
	if !ok || strings.TrimSpace(userIDStr) == "" {
		return uuid.Nil, ErrUnauthorized
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return uuid.Nil, ErrUnauthorized
	}
	return userID, nil
}

func resolveTenantID(c *gin.Context) (*uuid.UUID, error) {
	candidates := []string{
		strings.TrimSpace(c.Query("tenant_id")),
		strings.TrimSpace(c.GetHeader(tenantContextHeader)),
	}
	for _, candidate := range candidates {
		if candidate == "" {
			continue
		}
		tenantID, err := uuid.Parse(candidate)
		if err != nil {
			return nil, err
		}
		return &tenantID, nil
	}
	return nil, nil
}

func sameTenant(left, right *uuid.UUID) bool {
	if left == nil && right == nil {
		return true
	}
	if left == nil || right == nil {
		return false
	}
	return *left == *right
}

func (s *Service) getEffectiveActiveRoleIDs(userID uuid.UUID, tenantID *uuid.UUID) ([]uuid.UUID, error) {
	var roleIDs []uuid.UUID
	query := s.db.Model(&models.UserRole{}).
		Joins("JOIN roles ON roles.id = user_roles.role_id").
		Where("user_roles.user_id = ?", userID).
		Where("roles.status = ?", "normal")
	if tenantID == nil {
		query = query.Where("user_roles.tenant_id IS NULL")
	} else {
		query = query.Where("user_roles.tenant_id = ?", *tenantID)
	}
	err := query.Distinct("user_roles.role_id").Pluck("user_roles.role_id", &roleIDs).Error
	return roleIDs, err
}

func (s *Service) resolveTenantActionContext(userID uuid.UUID, tenantID *uuid.UUID) (bool, bool, map[uuid.UUID]bool, error) {
	boundaryActionSet := make(map[uuid.UUID]bool)
	if tenantID == nil {
		return false, false, boundaryActionSet, nil
	}

	var member models.TenantMember
	err := s.db.Where("user_id = ? AND tenant_id = ?", userID, *tenantID).First(&member).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, false, boundaryActionSet, nil
		}
		return false, false, nil, err
	}
	if member.Status != "active" {
		return false, false, boundaryActionSet, ErrTenantMemberInactive
	}

	snapshot, err := s.boundaryService.GetSnapshot(*tenantID)
	if err != nil {
		return false, false, nil, err
	}
	if len(snapshot.EffectiveIDs) == 0 {
		return true, false, boundaryActionSet, nil
	}
	for _, actionID := range snapshot.EffectiveIDs {
		boundaryActionSet[actionID] = true
	}
	return true, true, boundaryActionSet, nil
}

func (s *Service) buildRoleEffectMap(roleIDs []uuid.UUID) (map[uuid.UUID]string, error) {
	result := make(map[uuid.UUID]string)
	if len(roleIDs) == 0 {
		return result, nil
	}

	var rolePermissions []models.RoleActionPermission
	if err := s.db.Where("role_id IN ?", roleIDs).Find(&rolePermissions).Error; err != nil {
		return nil, err
	}

	for _, record := range rolePermissions {
		result[record.ActionID] = "allow"
	}
	return result, nil
}

func (s *Service) buildUserOverrideMaps(userID uuid.UUID, tenantID *uuid.UUID) (map[uuid.UUID]string, map[uuid.UUID]string, error) {
	globalOverrides := make(map[uuid.UUID]string)
	tenantOverrides := make(map[uuid.UUID]string)

	query := s.db.Where("user_id = ?", userID)
	if tenantID == nil {
		query = query.Where("tenant_id IS NULL")
	} else {
		query = query.Where("tenant_id IS NULL OR tenant_id = ?", *tenantID)
	}

	var records []models.UserActionPermission
	if err := query.Find(&records).Error; err != nil {
		return nil, nil, err
	}

	for _, record := range records {
		target := globalOverrides
		if record.TenantID != nil {
			target = tenantOverrides
		}
		if target[record.ActionID] == "deny" {
			continue
		}
		if record.Effect == "deny" {
			target[record.ActionID] = "deny"
			continue
		}
		if record.Effect == "allow" {
			target[record.ActionID] = "allow"
		}
	}

	return globalOverrides, tenantOverrides, nil
}

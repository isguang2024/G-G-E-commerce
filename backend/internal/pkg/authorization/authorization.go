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
	db     *gorm.DB
	logger *zap.Logger
}

type actionScopeSelector struct {
	Code string
}

func NewService(db *gorm.DB, logger *zap.Logger) *Service {
	return &Service{db: db, logger: logger}
}

func (s *Service) RequireAction(resourceCode, actionCode string, scopeCodes ...string) gin.HandlerFunc {
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

		allowed, actionDef, authErr := s.Authorize(userID, tenantID, resourceCode, actionCode, scopeCodes...)
		if authErr != nil {
			s.respondAuthError(c, authErr, resourceCode, actionCode)
			return
		}
		if !allowed {
			s.respondAuthError(c, ErrPermissionDenied, resourceCode, actionCode)
			return
		}

		if actionDef != nil {
			c.Set("permission_action_id", actionDef.ID.String())
			c.Set("permission_action_key", actionDef.ResourceCode+":"+actionDef.ActionCode)
		}
		c.Next()
	}
}

func (s *Service) Authorize(userID uuid.UUID, tenantID *uuid.UUID, resourceCode, actionCode string, scopeCodes ...string) (bool, *models.PermissionAction, error) {
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
	query := s.db.Preload("Scope").Joins("JOIN scopes ON permission_actions.scope_id = scopes.id").Where("permission_actions.resource_code = ? AND permission_actions.action_code = ?", resourceCode, actionCode)
	selector := resolveActionScopeSelector(scopeCodes, tenantID)
	if selector.Code != "" {
		if selector.Code == "team" {
			query = query.Where("scopes.code IN ?", []string{"team", "tenant"})
		} else {
			query = query.Where("scopes.code = ?", selector.Code)
		}
	}
	err = query.First(&actionDef).Error
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

	requiresTenant := isTeamScopeCode(actionDef.Scope.Code)
	var member models.TenantMember
	if requiresTenant {
		if tenantID == nil {
			return false, &actionDef, ErrTenantContextRequired
		}
		err = s.db.Where("user_id = ? AND tenant_id = ?", userID, *tenantID).First(&member).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return false, &actionDef, ErrTenantMemberNotFound
			}
			return false, &actionDef, err
		}
		if member.Status != "active" {
			return false, &actionDef, ErrTenantMemberInactive
		}
		if allowed, boundaryErr := s.isActionWithinTenantBoundary(*tenantID, actionDef.ID); boundaryErr != nil {
			return false, &actionDef, boundaryErr
		} else if !allowed {
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

func resolveActionScopeSelector(scopeCodes []string, tenantID *uuid.UUID) actionScopeSelector {
	for _, scopeCode := range scopeCodes {
		trimmed := normalizeActionScopeCode(scopeCode)
		if trimmed != "" {
			return actionScopeSelector{Code: trimmed}
		}
	}
	if tenantID != nil {
		return actionScopeSelector{Code: "team"}
	}
	return actionScopeSelector{Code: "global"}
}

func (s *Service) GetUserActionKeys(userID uuid.UUID, tenantID *uuid.UUID) ([]string, error) {
	keys, _, err := s.collectUserActionKeys(userID, tenantID)
	if err != nil {
		return nil, err
	}
	return keys, nil
}

func (s *Service) GetUserScopedActionKeys(userID uuid.UUID, tenantID *uuid.UUID) ([]string, error) {
	_, scopedKeys, err := s.collectUserActionKeys(userID, tenantID)
	if err != nil {
		return nil, err
	}
	return scopedKeys, nil
}

func (s *Service) GetUserActionSnapshot(userID uuid.UUID, tenantID *uuid.UUID) ([]string, []string, error) {
	return s.collectUserActionKeys(userID, tenantID)
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
	if err := s.db.Preload("Scope").Where("status = ?", "normal").Order("sort_order ASC, created_at DESC").Find(&actions).Error; err != nil {
		return nil, nil, err
	}
	if currentUser.IsSuperAdmin {
		keys := make([]string, 0, len(actions))
		scopedKeys := make([]string, 0, len(actions))
		keySet := make(map[string]struct{}, len(actions))
		scopedKeySet := make(map[string]struct{}, len(actions))
		for _, action := range actions {
			key := action.ResourceCode + ":" + action.ActionCode
			scopedKey := key + "@" + normalizeActionScopeCode(action.Scope.Code)
			if _, exists := keySet[key]; !exists {
				keys = append(keys, key)
				keySet[key] = struct{}{}
			}
			if _, exists := scopedKeySet[scopedKey]; !exists {
				scopedKeys = append(scopedKeys, scopedKey)
				scopedKeySet[scopedKey] = struct{}{}
			}
		}
		return keys, scopedKeys, nil
	}

	memberActive, boundaryAllOpen, boundaryActionSet, err := s.resolveTenantActionContext(userID, tenantID)
	if err != nil {
		return nil, nil, err
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
	scopedKeys := make([]string, 0, len(actions))
	keySet := make(map[string]struct{}, len(actions))
	scopedKeySet := make(map[string]struct{}, len(actions))
	for _, action := range actions {
		requiresTenant := isTeamScopeCode(action.Scope.Code)
		if requiresTenant {
			if tenantID == nil || !memberActive {
				continue
			}
			if !boundaryAllOpen && !boundaryActionSet[action.ID] {
				continue
			}
			if effect, ok := tenantOverrideMap[action.ID]; ok {
				if effect == "allow" {
					appendActionKeys(&keys, &scopedKeys, keySet, scopedKeySet, action)
				}
				continue
			}
			if effect, ok := globalOverrideMap[action.ID]; ok {
				if effect == "allow" {
					appendActionKeys(&keys, &scopedKeys, keySet, scopedKeySet, action)
				}
				continue
			}
			if roleEffectMap[action.ID] == "allow" {
				appendActionKeys(&keys, &scopedKeys, keySet, scopedKeySet, action)
			}
			continue
		}

		if effect, ok := globalOverrideMap[action.ID]; ok {
			if effect == "allow" {
				appendActionKeys(&keys, &scopedKeys, keySet, scopedKeySet, action)
			}
			continue
		}
		if roleEffectMap[action.ID] == "allow" {
			appendActionKeys(&keys, &scopedKeys, keySet, scopedKeySet, action)
		}
	}

	return keys, scopedKeys, nil
}

func appendActionKeys(keys, scopedKeys *[]string, keySet, scopedKeySet map[string]struct{}, action models.PermissionAction) {
	key := action.ResourceCode + ":" + action.ActionCode
	scopedKey := key + "@" + normalizeActionScopeCode(action.Scope.Code)
	if _, exists := keySet[key]; !exists {
		*keys = append(*keys, key)
		keySet[key] = struct{}{}
	}
	if _, exists := scopedKeySet[scopedKey]; !exists {
		*scopedKeys = append(*scopedKeys, scopedKey)
		scopedKeySet[scopedKey] = struct{}{}
	}
}

func normalizeActionScopeCode(scopeCode string) string {
	switch strings.TrimSpace(strings.ToLower(scopeCode)) {
	case "tenant", "team":
		return "team"
	case "":
		return "global"
	default:
		return strings.TrimSpace(strings.ToLower(scopeCode))
	}
}

func isTeamScopeCode(scopeCode string) bool {
	normalized := normalizeActionScopeCode(scopeCode)
	return normalized == "team"
}

func (s *Service) respondAuthError(c *gin.Context, authErr error, resourceCode, actionCode string) {
	if authErr != nil {
		s.logger.Warn("Permission denied",
			zap.Error(authErr),
			zap.String("resource", resourceCode),
			zap.String("action", actionCode),
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

func (s *Service) isActionWithinTenantBoundary(tenantID, actionID uuid.UUID) (bool, error) {
	var count int64
	err := s.db.Model(&models.TenantActionPermission{}).Where("tenant_id = ?", tenantID).Count(&count).Error
	if err != nil {
		return false, err
	}
	if count == 0 {
		return false, nil
	}
	var enabledCount int64
	err = s.db.Model(&models.TenantActionPermission{}).
		Where("tenant_id = ? AND action_id = ? AND enabled = ?", tenantID, actionID, true).
		Count(&enabledCount).Error
	return enabledCount > 0, err
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
		query = query.Where("user_roles.tenant_id IS NULL OR user_roles.tenant_id = ?", *tenantID)
	}
	err := query.Distinct("user_roles.role_id").Pluck("user_roles.role_id", &roleIDs).Error
	return roleIDs, err
}

func (s *Service) resolveTenantActionContext(userID uuid.UUID, tenantID *uuid.UUID) (bool, bool, map[uuid.UUID]bool, error) {
	boundaryActionSet := make(map[uuid.UUID]bool)
	if tenantID == nil {
		return false, true, boundaryActionSet, nil
	}

	var member models.TenantMember
	err := s.db.Where("user_id = ? AND tenant_id = ?", userID, *tenantID).First(&member).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, true, boundaryActionSet, nil
		}
		return false, false, nil, err
	}
	if member.Status != "active" {
		return false, true, boundaryActionSet, nil
	}

	var count int64
	if err := s.db.Model(&models.TenantActionPermission{}).Where("tenant_id = ?", *tenantID).Count(&count).Error; err != nil {
		return false, false, nil, err
	}
	if count == 0 {
		return true, false, boundaryActionSet, nil
	}

	var tenantRecords []models.TenantActionPermission
	if err := s.db.Where("tenant_id = ? AND enabled = ?", *tenantID, true).Find(&tenantRecords).Error; err != nil {
		return false, false, nil, err
	}
	for _, record := range tenantRecords {
		boundaryActionSet[record.ActionID] = true
	}
	return true, false, boundaryActionSet, nil
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

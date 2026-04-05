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
	appctx "github.com/gg-ecommerce/backend/internal/pkg/appctx"
	"github.com/gg-ecommerce/backend/internal/pkg/permissionkey"
	"github.com/gg-ecommerce/backend/internal/pkg/platformaccess"
	"github.com/gg-ecommerce/backend/internal/pkg/teamboundary"
)

const tenantContextHeader = "X-Tenant-ID"

var (
	ErrUnauthorized          = errors.New("unauthorized")
	ErrUserInactive          = errors.New("user inactive")
	ErrPermissionKeyMissing  = errors.New("permission action missing")
	ErrTenantContextRequired = errors.New("tenant context required")
	ErrTenantMemberNotFound  = errors.New("tenant member not found")
	ErrTenantMemberInactive  = errors.New("tenant member inactive")
	ErrPermissionDenied      = errors.New("permission denied")
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
	resolvedKey := resolvePermissionKey(permissionKey, legacy...)
	return s.RequireAnyAction(resolvedKey)
}

func (s *Service) RequireAnyAction(permissionKeys ...string) gin.HandlerFunc {
	resolvedKeys := resolvePermissionKeys(permissionKeys...)
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

		allowed, actionDef, matchedKey, authErr := s.AuthorizeAnyInApp(userID, tenantID, appctx.CurrentAppKey(c), resolvedKeys...)
		targetKey := strings.Join(resolvedKeys, " | ")
		if strings.TrimSpace(matchedKey) != "" {
			targetKey = matchedKey
		}
		if authErr != nil {
			s.respondAuthError(c, authErr, targetKey)
			return
		}
		if !allowed {
			s.respondAuthError(c, ErrPermissionDenied, targetKey)
			return
		}

		if actionDef != nil {
			c.Set("permission_action_id", actionDef.ID.String())
			c.Set("permission_action_key", actionDef.PermissionKey)
		}
		c.Next()
	}
}

func (s *Service) Authorize(userID uuid.UUID, tenantID *uuid.UUID, permissionKey string, legacy ...string) (bool, *models.PermissionKey, error) {
	return s.AuthorizeInApp(userID, tenantID, models.DefaultAppKey, permissionKey, legacy...)
}

func (s *Service) AuthorizeInApp(userID uuid.UUID, tenantID *uuid.UUID, appKey string, permissionKey string, legacy ...string) (bool, *models.PermissionKey, error) {
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
	var actionDef models.PermissionKey
	err = s.db.Where("permission_keys.permission_key = ?", resolvePermissionKey(permissionKey, legacy...)).First(&actionDef).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil, ErrPermissionKeyMissing
		}
		return false, nil, err
	}
	if actionDef.Status != "normal" {
		return false, &actionDef, ErrPermissionDenied
	}
	if currentUser.IsSuperAdmin {
		return true, &actionDef, nil
	}
	if tenantID == nil {
		if s.platformService == nil {
			return false, &actionDef, nil
		}
		snapshot, err := s.platformService.GetSnapshot(userID, appctx.NormalizeAppKey(appKey))
		if err != nil {
			return false, &actionDef, err
		}
		return containsUUID(snapshot.ActionIDs, actionDef.ID), &actionDef, nil
	}

	memberActive, boundaryConfigured, boundaryActionSet, ctxErr := s.resolveTenantActionContext(userID, tenantID, appKey)
	if ctxErr != nil {
		return false, &actionDef, ctxErr
	}
	if !memberActive {
		return false, &actionDef, ErrTenantMemberNotFound
	}
	if boundaryConfigured && !boundaryActionSet[actionDef.ID] {
		return false, &actionDef, ErrPermissionDenied
	}
	roleActionSet, err := s.getTeamRoleSnapshotActionSet(userID, *tenantID, appKey)
	if err != nil {
		return false, &actionDef, err
	}
	return roleActionSet[actionDef.ID], &actionDef, nil
}

func (s *Service) AuthorizeAny(userID uuid.UUID, tenantID *uuid.UUID, permissionKeys ...string) (bool, *models.PermissionKey, string, error) {
	return s.AuthorizeAnyInApp(userID, tenantID, models.DefaultAppKey, permissionKeys...)
}

func (s *Service) AuthorizeAnyInApp(userID uuid.UUID, tenantID *uuid.UUID, appKey string, permissionKeys ...string) (bool, *models.PermissionKey, string, error) {
	resolvedKeys := resolvePermissionKeys(permissionKeys...)
	if len(resolvedKeys) == 0 {
		return false, nil, "", ErrPermissionKeyMissing
	}

	missingCount := 0
	denied := false
	var deniedAction *models.PermissionKey
	for _, key := range resolvedKeys {
		allowed, actionDef, err := s.AuthorizeInApp(userID, tenantID, appKey, key)
		if err == nil {
			if allowed {
				return true, actionDef, key, nil
			}
			denied = true
			if actionDef != nil {
				deniedAction = actionDef
			}
			continue
		}

		switch {
		case errors.Is(err, ErrPermissionKeyMissing):
			missingCount++
		case errors.Is(err, ErrPermissionDenied):
			denied = true
			if actionDef != nil {
				deniedAction = actionDef
			}
		default:
			return false, actionDef, key, err
		}
	}

	if denied {
		return false, deniedAction, "", ErrPermissionDenied
	}
	if missingCount == len(resolvedKeys) {
		return false, nil, "", ErrPermissionKeyMissing
	}
	return false, deniedAction, "", ErrPermissionDenied
}

func (s *Service) GetUserActionKeys(userID uuid.UUID, tenantID *uuid.UUID) ([]string, error) {
	return s.GetUserActionKeysInApp(userID, tenantID, models.DefaultAppKey)
}

func (s *Service) GetUserActionKeysInApp(userID uuid.UUID, tenantID *uuid.UUID, appKey string) ([]string, error) {
	keys, _, err := s.collectUserActionKeysInApp(userID, tenantID, appKey)
	if err != nil {
		return nil, err
	}
	return keys, nil
}

func (s *Service) GetUserActionSnapshot(userID uuid.UUID, tenantID *uuid.UUID) ([]string, error) {
	return s.GetUserActionSnapshotInApp(userID, tenantID, models.DefaultAppKey)
}

func (s *Service) GetUserActionSnapshotInApp(userID uuid.UUID, tenantID *uuid.UUID, appKey string) ([]string, error) {
	keys, _, err := s.collectUserActionKeysInApp(userID, tenantID, appKey)
	if err != nil {
		return nil, err
	}
	return keys, nil
}

func (s *Service) collectUserActionKeys(userID uuid.UUID, tenantID *uuid.UUID) ([]string, []string, error) {
	return s.collectUserActionKeysInApp(userID, tenantID, models.DefaultAppKey)
}

func (s *Service) collectUserActionKeysInApp(userID uuid.UUID, tenantID *uuid.UUID, appKey string) ([]string, []string, error) {
	var currentUser models.User
	if err := s.db.Where("id = ?", userID).First(&currentUser).Error; err != nil {
		return nil, nil, err
	}
	if currentUser.Status != "active" {
		return []string{}, []string{}, nil
	}

	var actions []models.PermissionKey
	if err := s.db.Where("status = ?", "normal").Order("sort_order ASC, created_at DESC").Find(&actions).Error; err != nil {
		return nil, nil, err
	}
	if currentUser.IsSuperAdmin {
		keys := make([]string, 0, len(actions))
		keySet := make(map[string]struct{}, len(actions))
		for _, action := range actions {
			appendActionKeys(&keys, keySet, action)
		}
		return keys, []string{}, nil
	}
	if tenantID == nil {
		if s.platformService == nil {
			return []string{}, []string{}, nil
		}
		snapshot, err := s.platformService.GetSnapshot(userID, appctx.NormalizeAppKey(appKey))
		if err != nil {
			return nil, nil, err
		}
		return buildActionKeysFromIDs(actions, snapshot.ActionIDs), []string{}, nil
	}

	memberActive := false
	var ctxErr error
	memberActive, _, _, ctxErr = s.resolveTenantActionContext(userID, tenantID, appKey)
	if ctxErr != nil {
		if errors.Is(ctxErr, ErrTenantMemberInactive) {
			return []string{}, []string{}, nil
		}
		return nil, nil, ctxErr
	}
	if !memberActive {
		return []string{}, []string{}, nil
	}
	roleActionIDs, roleErr := s.getTeamRoleSnapshotActionIDsInApp(userID, *tenantID, appKey)
	if roleErr != nil {
		return nil, nil, roleErr
	}
	return buildActionKeysFromIDs(actions, roleActionIDs), []string{}, nil
}

func appendActionKeys(keys *[]string, keySet map[string]struct{}, action models.PermissionKey) {
	key := permissionkey.Normalize(action.PermissionKey)
	if _, exists := keySet[key]; !exists {
		*keys = append(*keys, key)
		keySet[key] = struct{}{}
	}
}

func buildActionKeysFromIDs(actions []models.PermissionKey, allowedActionIDs []uuid.UUID) []string {
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

func (s *Service) getTeamRoleSnapshotActionIDs(userID, teamID uuid.UUID) ([]uuid.UUID, error) {
	return s.getTeamRoleSnapshotActionIDsInApp(userID, teamID, models.DefaultAppKey)
}

func (s *Service) getTeamRoleSnapshotActionIDsInApp(userID, teamID uuid.UUID, appKey string) ([]uuid.UUID, error) {
	actionSet, err := s.getTeamRoleSnapshotActionSet(userID, teamID, appKey)
	if err != nil {
		return nil, err
	}
	actionIDs := make([]uuid.UUID, 0, len(actionSet))
	for actionID := range actionSet {
		actionIDs = append(actionIDs, actionID)
	}
	return actionIDs, nil
}

func (s *Service) getTeamRoleSnapshotActionSet(userID, teamID uuid.UUID, appKey string) (map[uuid.UUID]bool, error) {
	roleIDs, err := s.getEffectiveActiveRoleIDs(userID, &teamID)
	if err != nil {
		return nil, err
	}
	result := make(map[uuid.UUID]bool)
	if len(roleIDs) == 0 || s.boundaryService == nil {
		return result, nil
	}
	var roles []models.Role
	if err := s.db.Where("id IN ?", roleIDs).Find(&roles).Error; err != nil {
		return nil, err
	}
	roleMap := make(map[uuid.UUID]models.Role, len(roles))
	for _, role := range roles {
		roleMap[role.ID] = role
	}
	for _, roleID := range roleIDs {
		role, ok := roleMap[roleID]
		if !ok {
			continue
		}
		snapshot, err := s.boundaryService.GetRoleSnapshot(teamID, roleID, role.TenantID == nil, appctx.NormalizeAppKey(appKey))
		if err != nil {
			return nil, err
		}
		for _, actionID := range snapshot.ActionIDs {
			result[actionID] = true
		}
	}
	return result, nil
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
	case errors.Is(authErr, ErrPermissionKeyMissing):
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

func resolvePermissionKeys(permissionKeys ...string) []string {
	result := make([]string, 0, len(permissionKeys))
	seen := make(map[string]struct{}, len(permissionKeys))
	for _, item := range permissionKeys {
		key := resolvePermissionKey(item)
		if key == "" {
			continue
		}
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, key)
	}
	return result
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

func (s *Service) resolveTenantActionContext(userID uuid.UUID, tenantID *uuid.UUID, appKey string) (bool, bool, map[uuid.UUID]bool, error) {
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

	snapshot, err := s.boundaryService.GetSnapshot(*tenantID, appctx.NormalizeAppKey(appKey))
	if err != nil {
		return false, false, nil, err
	}
	configured := len(snapshot.PackageIDs) > 0 || len(snapshot.ExpandedPackageIDs) > 0 || len(snapshot.BlockedIDs) > 0
	if !configured && len(snapshot.EffectiveIDs) == 0 {
		return true, false, boundaryActionSet, nil
	}
	if len(snapshot.EffectiveIDs) == 0 {
		return true, true, boundaryActionSet, nil
	}
	for _, actionID := range snapshot.EffectiveIDs {
		boundaryActionSet[actionID] = true
	}
	return true, true, boundaryActionSet, nil
}

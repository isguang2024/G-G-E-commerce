package role

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/gg-ecommerce/backend/internal/api/dto"
	"github.com/gg-ecommerce/backend/internal/api/errcode"
	"github.com/gg-ecommerce/backend/internal/modules/system/user"
)

type RoleHandler struct {
	roleService RoleService
	userRepo    user.UserRepository
	logger      *zap.Logger
}

func NewRoleHandler(roleService RoleService, userRepo user.UserRepository, logger *zap.Logger) *RoleHandler {
	return &RoleHandler{
		roleService: roleService,
		userRepo:    userRepo,
		logger:      logger,
	}
}

func (h *RoleHandler) List(c *gin.Context) {
	var req dto.RoleListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		status, resp := errcode.Response(errcode.ErrParamInvalid)
		c.JSON(status, resp)
		return
	}
	if req.Current <= 0 {
		req.Current = 1
	}
	if req.Size <= 0 {
		req.Size = 20
	}
	list, total, err := h.roleService.List(&req)
	if err != nil {
		h.logger.Error("Role list failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取角色列表失败")
		c.JSON(status, resp)
		return
	}
	records := make([]gin.H, 0, len(list))
	for _, r := range list {
		canEditPermission := true
		scopeIDs, scopes := buildRoleScopePayload(r)
		records = append(records, gin.H{
			"roleId":            r.ID.String(),
			"roleName":          r.Name,
			"roleCode":          r.Code,
			"description":       r.Description,
			"status":            r.Status,
			"sortOrder":         r.SortOrder,
			"priority":          r.Priority,
			"createTime":        r.CreatedAt.Format("2006-01-02 15:04:05"),
			"scopeIds":          scopeIDs,
			"scopes":            scopes,
			"canEditPermission": canEditPermission,
		})
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"records": records,
		"total":   total,
		"current": req.Current,
		"size":    req.Size,
	}))
}

func (h *RoleHandler) Get(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的角色ID")
		c.JSON(status, resp)
		return
	}
	role, err := h.roleService.Get(id)
	if err != nil {
		if err == ErrRoleNotFound {
			status, resp := errcode.Response(errcode.ErrRoleNotFound)
			c.JSON(status, resp)
			return
		}
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取角色失败")
		c.JSON(status, resp)
		return
	}
	scopeIDs, scopes := buildRoleScopePayload(*role)
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"roleId":      role.ID.String(),
		"roleName":    role.Name,
		"roleCode":    role.Code,
		"description": role.Description,
		"status":      role.Status,
		"sortOrder":   role.SortOrder,
		"priority":    role.Priority,
		"createTime":  role.CreatedAt.Format("2006-01-02 15:04:05"),
		"scopeIds":    scopeIDs,
		"scopes":      scopes,
	}))
}

func (h *RoleHandler) Create(c *gin.Context) {
	var req dto.RoleCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		status, resp := errcode.Response(errcode.ErrParamInvalid)
		c.JSON(status, resp)
		return
	}
	role, err := h.roleService.Create(&req)
	if err != nil {
		if err == ErrRoleCodeExists {
			status, resp := errcode.Response(errcode.ErrRoleCodeExists)
			c.JSON(status, resp)
			return
		}
		h.logger.Error("Create role failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "创建角色失败")
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{"roleId": role.ID.String()}))
}

func (h *RoleHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的角色ID")
		c.JSON(status, resp)
		return
	}
	var req dto.RoleUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		status, resp := errcode.Response(errcode.ErrParamInvalid)
		c.JSON(status, resp)
		return
	}
	h.logger.Info("更新角色请求", zap.String("roleId", idStr), zap.Any("scopeIds", req.ScopeIDs), zap.Any("req", req))
	if err := h.roleService.Update(id, &req); err != nil {
		if err == ErrRoleNotFound {
			status, resp := errcode.Response(errcode.ErrRoleNotFound)
			c.JSON(status, resp)
			return
		}
		h.logger.Error("Update role failed", zap.String("roleId", idStr), zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "更新角色失败: "+err.Error())
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(nil))
}

func (h *RoleHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的角色ID")
		c.JSON(status, resp)
		return
	}
	if err := h.roleService.Delete(id); err != nil {
		if err == ErrRoleNotFound {
			status, resp := errcode.Response(errcode.ErrRoleNotFound)
			c.JSON(status, resp)
			return
		}
		if err == ErrSystemRoleCannotDelete {
			status, resp := errcode.Response(errcode.ErrSystemRoleProtected)
			c.JSON(status, resp)
			return
		}
		h.logger.Error("Delete role failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "删除角色失败")
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(nil))
}

func (h *RoleHandler) GetRoleMenus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的角色ID")
		c.JSON(status, resp)
		return
	}
	menuIDs, err := h.roleService.GetRoleMenuIDs(id)
	if err != nil {
		if err == ErrRoleNotFound {
			status, resp := errcode.Response(errcode.ErrRoleNotFound)
			c.JSON(status, resp)
			return
		}
		h.logger.Error("Get role menus failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取角色菜单失败")
		c.JSON(status, resp)
		return
	}
	ids := make([]string, 0, len(menuIDs))
	for _, u := range menuIDs {
		ids = append(ids, u.String())
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{"menu_ids": ids}))
}

func (h *RoleHandler) isSuperAdmin(c *gin.Context) bool {
	userIDStr, ok := c.Get("user_id")
	if !ok {
		return false
	}
	userIDStrValue, ok := userIDStr.(string)
	if !ok || userIDStrValue == "" {
		return false
	}
	userID, err := uuid.Parse(userIDStrValue)
	if err != nil {
		return false
	}
	user, err := h.userRepo.GetByID(userID)
	if err != nil || user == nil {
		return false
	}
	return user.IsSuperAdmin
}

func (h *RoleHandler) SetRoleMenus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的角色ID")
		c.JSON(status, resp)
		return
	}
	var req dto.RoleMenusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		status, resp := errcode.Response(errcode.ErrParamInvalid)
		c.JSON(status, resp)
		return
	}
	menuIDs := make([]uuid.UUID, 0, len(req.MenuIDs))
	for _, s := range req.MenuIDs {
		if s == "" {
			continue
		}
		u, err := uuid.Parse(s)
		if err != nil {
			status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的菜单ID: "+s)
			c.JSON(status, resp)
			return
		}
		menuIDs = append(menuIDs, u)
	}
	h.logger.Info("Setting role menus",
		zap.String("roleId", id.String()),
		zap.Int("menuCount", len(menuIDs)))

	if err := h.roleService.SetRoleMenus(id, menuIDs); err != nil {
		if err == ErrRoleNotFound {
			status, resp := errcode.Response(errcode.ErrRoleNotFound)
			c.JSON(status, resp)
			return
		}
		h.logger.Error("Set role menus failed",
			zap.String("roleId", id.String()),
			zap.Error(err),
			zap.Int("menuCount", len(menuIDs)))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, fmt.Sprintf("保存角色菜单失败: %v", err))
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(nil))
}

func (h *RoleHandler) GetRoleActions(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的角色ID")
		c.JSON(status, resp)
		return
	}
	records, err := h.roleService.GetRoleActions(id)
	if err != nil {
		if err == ErrRoleNotFound {
			status, resp := errcode.Response(errcode.ErrRoleNotFound)
			c.JSON(status, resp)
			return
		}
		h.logger.Error("Get role actions failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取角色功能权限失败")
		c.JSON(status, resp)
		return
	}
	items := make([]gin.H, 0, len(records))
	for _, record := range records {
		items = append(items, gin.H{
			"action_id": record.ActionID.String(),
			"effect":    record.Effect,
		})
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{"actions": items}))
}

func (h *RoleHandler) SetRoleActions(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的角色ID")
		c.JSON(status, resp)
		return
	}
	var req dto.RoleActionPermissionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		status, resp := errcode.Response(errcode.ErrParamInvalid)
		c.JSON(status, resp)
		return
	}
	actions := make([]user.RoleActionPermission, 0, len(req.Actions))
	for _, item := range req.Actions {
		actionID, parseErr := uuid.Parse(item.ActionID)
		if parseErr != nil {
			status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的功能权限ID")
			c.JSON(status, resp)
			return
		}
		actions = append(actions, user.RoleActionPermission{
			RoleID:   id,
			ActionID: actionID,
			Effect:   item.Effect,
		})
	}
	if err := h.roleService.SetRoleActions(id, actions); err != nil {
		if err == ErrRoleNotFound {
			status, resp := errcode.Response(errcode.ErrRoleNotFound)
			c.JSON(status, resp)
			return
		}
		h.logger.Error("Set role actions failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "保存角色功能权限失败")
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(nil))
}

func (h *RoleHandler) GetRoleDataPermissions(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的角色ID")
		c.JSON(status, resp)
		return
	}
	records, resourceCodes, scopeOptions, err := h.roleService.GetRoleDataPermissions(id)
	if err != nil {
		if err == ErrRoleNotFound {
			status, resp := errcode.Response(errcode.ErrRoleNotFound)
			c.JSON(status, resp)
			return
		}
		h.logger.Error("Get role data permissions failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取角色数据权限失败")
		c.JSON(status, resp)
		return
	}

	permissions := make([]gin.H, 0, len(records))
	for _, record := range records {
		permissions = append(permissions, gin.H{
			"resource_code": record.ResourceCode,
			"scope_code":    record.ScopeCode,
		})
	}

	resources := make([]gin.H, 0, len(resourceCodes))
	for _, resourceCode := range resourceCodes {
		resources = append(resources, gin.H{
			"resource_code": resourceCode,
			"resource_name": formatRoleDataResourceName(resourceCode),
		})
	}

	scopes := make([]gin.H, 0, len(scopeOptions))
	for _, option := range scopeOptions {
		scopes = append(scopes, gin.H{
			"scope_code": option.Code,
			"scope_name": option.Name,
		})
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"permissions":      permissions,
		"resources":        resources,
		"available_scopes": scopes,
	}))
}

func (h *RoleHandler) SetRoleDataPermissions(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的角色ID")
		c.JSON(status, resp)
		return
	}
	var req dto.RoleDataPermissionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		status, resp := errcode.Response(errcode.ErrParamInvalid)
		c.JSON(status, resp)
		return
	}

	permissions := make([]user.RoleDataPermission, 0, len(req.Permissions))
	for _, item := range req.Permissions {
		permissions = append(permissions, user.RoleDataPermission{
			RoleID:       id,
			ResourceCode: item.ResourceCode,
			ScopeCode:    item.ScopeCode,
		})
	}
	if err := h.roleService.SetRoleDataPermissions(id, permissions); err != nil {
		if err == ErrRoleNotFound {
			status, resp := errcode.Response(errcode.ErrRoleNotFound)
			c.JSON(status, resp)
			return
		}
		h.logger.Error("Set role data permissions failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "保存角色数据权限失败: "+err.Error())
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(nil))
}

func buildRoleScopePayload(role user.Role) ([]string, []gin.H) {
	scopeMap := make(map[string]struct{})
	scopeIDs := make([]string, 0, len(role.Scopes))
	scopes := make([]gin.H, 0, len(role.Scopes))

	for _, scope := range role.Scopes {
		if scope.ID == (uuid.UUID{}) && scope.Code == "" {
			continue
		}
		key := scope.ID.String()
		if key == "" {
			key = scope.Code
		}
		if _, ok := scopeMap[key]; ok {
			continue
		}
		scopeMap[key] = struct{}{}
		if scope.ID != (uuid.UUID{}) {
			scopeIDs = append(scopeIDs, scope.ID.String())
		}
		scopes = append(scopes, gin.H{
			"scopeId":   scope.ID.String(),
			"scopeCode": scope.Code,
			"scopeName": scope.Name,
		})
	}

	return scopeIDs, scopes
}

func formatRoleDataResourceName(resourceCode string) string {
	names := map[string]string{
		"user":                "用户",
		"role":                "角色",
		"scope":               "作用域",
		"menu":                "菜单",
		"menu_backup":         "菜单备份",
		"permission_action":   "功能权限",
		"tenant":              "团队",
		"tenant_member_admin": "团队成员（系统）",
		"team":                "当前团队",
		"team_member":         "当前团队成员",
		"api_endpoint":        "API 注册表",
		"system":              "系统",
	}
	if name, ok := names[resourceCode]; ok {
		return name
	}
	return resourceCode
}

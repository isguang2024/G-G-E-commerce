package role

import (
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
		records = append(records, gin.H{
			"roleId":            r.ID.String(),
			"roleName":          r.Name,
			"roleCode":          r.Code,
			"description":       r.Description,
			"status":            r.Status,
			"sortOrder":         r.SortOrder,
			"priority":          r.Priority,
			"createTime":        r.CreatedAt.Format("2006-01-02 15:04:05"),
			"canEditPermission": true,
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
	id, err := uuid.Parse(c.Param("id"))
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
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"roleId":      role.ID.String(),
		"roleName":    role.Name,
		"roleCode":    role.Code,
		"description": role.Description,
		"status":      role.Status,
		"sortOrder":   role.SortOrder,
		"priority":    role.Priority,
		"createTime":  role.CreatedAt.Format("2006-01-02 15:04:05"),
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
	id, err := uuid.Parse(c.Param("id"))
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
	if err := h.roleService.Update(id, &req); err != nil {
		if err == ErrRoleNotFound {
			status, resp := errcode.Response(errcode.ErrRoleNotFound)
			c.JSON(status, resp)
			return
		}
		if err == ErrTenantRoleManagedByTeam {
			status, resp := errcode.ResponseWithMsg(errcode.ErrForbidden, "团队自定义角色需要在团队上下文中维护")
			c.JSON(status, resp)
			return
		}
		h.logger.Error("Update role failed", zap.String("roleId", id.String()), zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "更新角色失败: "+err.Error())
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(nil))
}

func (h *RoleHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
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
		if err == ErrTenantRoleManagedByTeam {
			status, resp := errcode.ResponseWithMsg(errcode.ErrForbidden, "团队自定义角色需要在团队上下文中维护")
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
	id, err := uuid.Parse(c.Param("id"))
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

func (h *RoleHandler) SetRoleMenus(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
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
	for _, item := range req.MenuIDs {
		menuID, parseErr := uuid.Parse(item)
		if parseErr != nil {
			status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的菜单ID")
			c.JSON(status, resp)
			return
		}
		menuIDs = append(menuIDs, menuID)
	}
	if err := h.roleService.SetRoleMenus(id, menuIDs); err != nil {
		if err == ErrRoleNotFound {
			status, resp := errcode.Response(errcode.ErrRoleNotFound)
			c.JSON(status, resp)
			return
		}
		if err == ErrTenantRoleManagedByTeam {
			status, resp := errcode.ResponseWithMsg(errcode.ErrForbidden, "团队自定义角色需要在团队上下文中维护")
			c.JSON(status, resp)
			return
		}
		h.logger.Error("Set role menus failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "保存角色菜单失败")
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(nil))
}

func (h *RoleHandler) GetRoleActions(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
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
	actionIDs := make([]string, 0, len(records))
	for _, record := range records {
		actionIDs = append(actionIDs, record.ActionID.String())
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{"action_ids": actionIDs}))
}

func (h *RoleHandler) SetRoleActions(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
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
	actions := make([]user.RoleActionPermission, 0, len(req.ActionIDs))
	for _, item := range req.ActionIDs {
		actionID, parseErr := uuid.Parse(item)
		if parseErr != nil {
			status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的功能权限ID")
			c.JSON(status, resp)
			return
		}
		actions = append(actions, user.RoleActionPermission{
			RoleID:   id,
			ActionID: actionID,
		})
	}
	if err := h.roleService.SetRoleActions(id, actions); err != nil {
		if err == ErrRoleNotFound {
			status, resp := errcode.Response(errcode.ErrRoleNotFound)
			c.JSON(status, resp)
			return
		}
		if err == ErrTenantRoleManagedByTeam {
			status, resp := errcode.ResponseWithMsg(errcode.ErrForbidden, "团队自定义角色需要在团队上下文中维护")
			c.JSON(status, resp)
			return
		}
		if err == ErrTeamRoleActionReadonly {
			status, resp := errcode.ResponseWithMsg(errcode.ErrForbidden, "团队角色功能权限由团队能力边界控制，不支持在系统角色页直接修改")
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
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的角色ID")
		c.JSON(status, resp)
		return
	}
	records, resourceCodes, dataScopeOptions, err := h.roleService.GetRoleDataPermissions(id)
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
			"data_scope":    record.DataScope,
		})
	}

	resources := make([]gin.H, 0, len(resourceCodes))
	for _, resourceCode := range resourceCodes {
		resources = append(resources, gin.H{
			"resource_code": resourceCode,
			"resource_name": formatRoleDataResourceName(resourceCode),
		})
	}

	dataScopes := make([]gin.H, 0, len(dataScopeOptions))
	for _, option := range dataScopeOptions {
		dataScopes = append(dataScopes, gin.H{
			"data_scope": option.Code,
			"label":      option.Name,
		})
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"permissions":           permissions,
		"resources":             resources,
		"available_data_scopes": dataScopes,
	}))
}

func (h *RoleHandler) SetRoleDataPermissions(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
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
			DataScope:    item.DataScope,
		})
	}
	if err := h.roleService.SetRoleDataPermissions(id, permissions); err != nil {
		if err == ErrRoleNotFound {
			status, resp := errcode.Response(errcode.ErrRoleNotFound)
			c.JSON(status, resp)
			return
		}
		if err == ErrTenantRoleManagedByTeam {
			status, resp := errcode.ResponseWithMsg(errcode.ErrForbidden, "团队自定义角色需要在团队上下文中维护")
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

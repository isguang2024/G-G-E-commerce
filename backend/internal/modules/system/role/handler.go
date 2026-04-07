package role

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/gg-ecommerce/backend/internal/api/dto"
	"github.com/gg-ecommerce/backend/internal/api/errcode"
	"github.com/gg-ecommerce/backend/internal/modules/system/user"
	"github.com/gg-ecommerce/backend/internal/pkg/appctx"
	"github.com/gg-ecommerce/backend/internal/pkg/permissionkey"
)

type RoleHandler struct {
	roleService RoleService
	userRepo    user.UserRepository
	keyRepo     user.PermissionKeyRepository
	logger      *zap.Logger
}

func NewRoleHandler(roleService RoleService, userRepo user.UserRepository, keyRepo user.PermissionKeyRepository, logger *zap.Logger) *RoleHandler {
	return &RoleHandler{
		roleService: roleService,
		userRepo:    userRepo,
		keyRepo:     keyRepo,
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
			"appKeys":           r.AppKeys,
			"isGlobal":          len(r.AppKeys) == 0,
			"status":            r.Status,
			"sortOrder":         r.SortOrder,
			"priority":          r.Priority,
			"customParams":      r.CustomParams,
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

func (h *RoleHandler) ListOptions(c *gin.Context) {
	list, err := h.roleService.ListOptions()
	if err != nil {
		h.logger.Error("Role options failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取角色候选失败")
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
			"appKeys":           r.AppKeys,
			"isGlobal":          len(r.AppKeys) == 0,
			"status":            r.Status,
			"sortOrder":         r.SortOrder,
			"priority":          r.Priority,
			"customParams":      r.CustomParams,
			"createTime":        r.CreatedAt.Format("2006-01-02 15:04:05"),
			"canEditPermission": true,
		})
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"records": records,
		"total":   len(records),
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
		"roleId":       role.ID.String(),
		"roleName":     role.Name,
		"roleCode":     role.Code,
		"description":  role.Description,
		"appKeys":      role.AppKeys,
		"isGlobal":     len(role.AppKeys) == 0,
		"status":       role.Status,
		"sortOrder":    role.SortOrder,
		"priority":     role.Priority,
		"customParams": role.CustomParams,
		"createTime":   role.CreatedAt.Format("2006-01-02 15:04:05"),
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
		if err == ErrCollaborationWorkspaceRoleManaged {
			status, resp := errcode.ResponseWithMsg(errcode.ErrForbidden, "协作空间自定义角色需要在协作空间上下文中维护")
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
		if err == ErrCollaborationWorkspaceRoleManaged {
			status, resp := errcode.ResponseWithMsg(errcode.ErrForbidden, "协作空间自定义角色需要在协作空间上下文中维护")
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

func (h *RoleHandler) GetRolePackages(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的角色ID")
		c.JSON(status, resp)
		return
	}
	appKey, err := appctx.RequireRequestAppKey(c)
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrParamInvalid, "app_key 为必填项")
		c.JSON(status, resp)
		return
	}
	packageIDs, packages, err := h.roleService.GetRolePackages(id, appKey)
	if err != nil {
		if err == ErrRoleNotFound {
			status, resp := errcode.Response(errcode.ErrRoleNotFound)
			c.JSON(status, resp)
			return
		}
		if err == ErrCollaborationWorkspaceRoleManaged {
			status, resp := errcode.ResponseWithMsg(errcode.ErrForbidden, "协作空间自定义角色需要在协作空间上下文中维护")
			c.JSON(status, resp)
			return
		}
		if err == ErrRoleAppScopeMismatch {
			status, resp := errcode.ResponseWithMsg(errcode.ErrForbidden, "当前 App 不在角色生效范围内")
			c.JSON(status, resp)
			return
		}
		h.logger.Error("Get role packages failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取角色功能包失败")
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"package_ids": packageIDsToStrings(packageIDs),
		"packages":    featurePackageListToMaps(packages),
	}))
}

func (h *RoleHandler) SetRolePackages(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的角色ID")
		c.JSON(status, resp)
		return
	}
	var req dto.RoleFeaturePackagesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		status, resp := errcode.Response(errcode.ErrParamInvalid)
		c.JSON(status, resp)
		return
	}
	appKey, err := appctx.RequireRequestAppKey(c)
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrParamInvalid, "app_key 为必填项")
		c.JSON(status, resp)
		return
	}
	packageIDs := make([]uuid.UUID, 0, len(req.PackageIDs))
	for _, item := range req.PackageIDs {
		packageID, parseErr := uuid.Parse(item)
		if parseErr != nil {
			status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的功能包ID")
			c.JSON(status, resp)
			return
		}
		packageIDs = append(packageIDs, packageID)
	}
	var grantedBy *uuid.UUID
	if rawUserID, ok := c.Get("user_id"); ok {
		if userIDStr, ok := rawUserID.(string); ok {
			if userID, parseErr := uuid.Parse(userIDStr); parseErr == nil {
				grantedBy = &userID
			}
		}
	}
	if err := h.roleService.SetRolePackages(id, packageIDs, grantedBy, appKey); err != nil {
		if err == ErrRoleNotFound {
			status, resp := errcode.Response(errcode.ErrRoleNotFound)
			c.JSON(status, resp)
			return
		}
		if err == ErrCollaborationWorkspaceRoleManaged {
			status, resp := errcode.ResponseWithMsg(errcode.ErrForbidden, "协作空间自定义角色需要在协作空间上下文中维护")
			c.JSON(status, resp)
			return
		}
		if err == ErrRoleAppScopeMismatch {
			status, resp := errcode.ResponseWithMsg(errcode.ErrForbidden, "当前 App 不在角色生效范围内")
			c.JSON(status, resp)
			return
		}
		h.logger.Error("Set role packages failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "保存角色功能包失败")
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
	appKey, err := appctx.RequireRequestAppKey(c)
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrParamInvalid, "app_key 为必填项")
		c.JSON(status, resp)
		return
	}
	boundary, err := h.roleService.GetRoleMenuBoundary(id, appKey)
	if err != nil {
		if err == ErrRoleNotFound {
			status, resp := errcode.Response(errcode.ErrRoleNotFound)
			c.JSON(status, resp)
			return
		}
		if err == ErrCollaborationWorkspaceRoleManaged {
			status, resp := errcode.ResponseWithMsg(errcode.ErrForbidden, "协作空间自定义角色需要在协作空间上下文中维护")
			c.JSON(status, resp)
			return
		}
		if err == ErrRoleAppScopeMismatch {
			status, resp := errcode.ResponseWithMsg(errcode.ErrForbidden, "当前 App 不在角色生效范围内")
			c.JSON(status, resp)
			return
		}
		h.logger.Error("Get role menus failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取角色菜单失败")
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"menu_ids":             packageIDsToStrings(boundary.EffectiveMenuIDs),
		"available_menu_ids":   packageIDsToStrings(boundary.AvailableMenuIDs),
		"hidden_menu_ids":      packageIDsToStrings(boundary.HiddenMenuIDs),
		"expanded_package_ids": packageIDsToStrings(boundary.ExpandedPackageIDs),
		"derived_sources":      buildMenuSourceMaps(boundary.MenuSourceMap),
	}))
}

func packageIDsToStrings(ids []uuid.UUID) []string {
	result := make([]string, 0, len(ids))
	for _, id := range ids {
		result = append(result, id.String())
	}
	return result
}

func featurePackageListToMaps(items []user.FeaturePackage) []gin.H {
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

func buildKeySourceMaps(sourceMap map[uuid.UUID][]uuid.UUID) []gin.H {
	if len(sourceMap) == 0 {
		return []gin.H{}
	}
	result := make([]gin.H, 0, len(sourceMap))
	for keyID, packageIDs := range sourceMap {
		result = append(result, gin.H{
			"action_id":   keyID.String(),
			"package_ids": packageIDsToStrings(packageIDs),
		})
	}
	return result
}

func buildMenuSourceMaps(sourceMap map[uuid.UUID][]uuid.UUID) []gin.H {
	if len(sourceMap) == 0 {
		return []gin.H{}
	}
	result := make([]gin.H, 0, len(sourceMap))
	for menuID, packageIDs := range sourceMap {
		result = append(result, gin.H{
			"menu_id":     menuID.String(),
			"package_ids": packageIDsToStrings(packageIDs),
		})
	}
	return result
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
	appKey, err := appctx.RequireRequestAppKey(c)
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrParamInvalid, "app_key 为必填项")
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
	if err := h.roleService.SetRoleMenus(id, menuIDs, appKey); err != nil {
		if err == ErrRoleNotFound {
			status, resp := errcode.Response(errcode.ErrRoleNotFound)
			c.JSON(status, resp)
			return
		}
		if err == ErrCollaborationWorkspaceRoleManaged {
			status, resp := errcode.ResponseWithMsg(errcode.ErrForbidden, "协作空间自定义角色需要在协作空间上下文中维护")
			c.JSON(status, resp)
			return
		}
		if err == ErrRoleAppScopeMismatch {
			status, resp := errcode.ResponseWithMsg(errcode.ErrForbidden, "当前 App 不在角色生效范围内")
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

func (h *RoleHandler) GetRoleKeys(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的角色ID")
		c.JSON(status, resp)
		return
	}
	appKey, err := appctx.RequireRequestAppKey(c)
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrParamInvalid, "app_key 为必填项")
		c.JSON(status, resp)
		return
	}
	boundary, err := h.roleService.GetRoleKeyBoundary(id, appKey)
	if err != nil {
		if err == ErrRoleNotFound {
			status, resp := errcode.Response(errcode.ErrRoleNotFound)
			c.JSON(status, resp)
			return
		}
		if err == ErrCollaborationWorkspaceRoleManaged {
			status, resp := errcode.ResponseWithMsg(errcode.ErrForbidden, "协作空间自定义角色需要在协作空间上下文中维护")
			c.JSON(status, resp)
			return
		}
		if err == ErrRoleAppScopeMismatch {
			status, resp := errcode.ResponseWithMsg(errcode.ErrForbidden, "当前 App 不在角色生效范围内")
			c.JSON(status, resp)
			return
		}
		h.logger.Error("Get role permission keys failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取角色功能权限失败")
		c.JSON(status, resp)
		return
	}

	actionIDs := boundary.AvailableKeyIDs
	if len(actionIDs) == 0 {
		// 兜底：历史快照延迟或脏数据时，至少回显已生效动作，避免前端列表空白。
		actionIDs = boundary.EffectiveKeyIDs
	}
	actions, err := h.keyRepo.GetByIDs(actionIDs)
	if err != nil {
		h.logger.Error("Load role permission action details failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取角色功能权限详情失败")
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"action_ids":           packageIDsToStrings(boundary.EffectiveKeyIDs),
		"available_action_ids": packageIDsToStrings(boundary.AvailableKeyIDs),
		"actions":              permissionActionListToMaps(actions),
		"disabled_action_ids":  packageIDsToStrings(boundary.DisabledKeyIDs),
		"expanded_package_ids": packageIDsToStrings(boundary.ExpandedPackageIDs),
		"derived_sources":      buildKeySourceMaps(boundary.KeySourceMap),
	}))
}

func (h *RoleHandler) SetRoleKeys(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的角色ID")
		c.JSON(status, resp)
		return
	}
	var req dto.RoleKeyPermissionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		status, resp := errcode.Response(errcode.ErrParamInvalid)
		c.JSON(status, resp)
		return
	}
	appKey, err := appctx.RequireRequestAppKey(c)
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrParamInvalid, "app_key 为必填项")
		c.JSON(status, resp)
		return
	}
	keys := make([]user.RoleKeyPermission, 0, len(req.KeyIDs))
	for _, item := range req.KeyIDs {
		keyID, parseErr := uuid.Parse(item)
		if parseErr != nil {
			status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的功能权限ID")
			c.JSON(status, resp)
			return
		}
		keys = append(keys, user.RoleKeyPermission{
			RoleID: id,
			KeyID:  keyID,
		})
	}
	if err := h.roleService.SetRoleKeys(id, keys, appKey); err != nil {
		if err == ErrRoleNotFound {
			status, resp := errcode.Response(errcode.ErrRoleNotFound)
			c.JSON(status, resp)
			return
		}
		if err == ErrCollaborationWorkspaceRoleManaged {
			status, resp := errcode.ResponseWithMsg(errcode.ErrForbidden, "协作空间自定义角色需要在协作空间上下文中维护")
			c.JSON(status, resp)
			return
		}
		if err == ErrRoleAppScopeMismatch {
			status, resp := errcode.ResponseWithMsg(errcode.ErrForbidden, "当前 App 不在角色生效范围内")
			c.JSON(status, resp)
			return
		}
		if err == ErrCollaborationWorkspaceRoleKeyReadonly {
			status, resp := errcode.ResponseWithMsg(errcode.ErrForbidden, "协作空间角色功能权限由协作空间能力边界控制，不支持在系统角色页直接修改")
			c.JSON(status, resp)
			return
		}
		h.logger.Error("Set role permission keys failed", zap.Error(err))
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
		if err == ErrCollaborationWorkspaceRoleManaged {
			status, resp := errcode.ResponseWithMsg(errcode.ErrForbidden, "协作空间自定义角色需要在协作空间上下文中维护")
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
		"user":                                 "用户",
		"role":                                 "角色",
		"scope":                                "作用域",
		"menu":                                 "菜单",
		"menu_backup":                          "菜单备份",
		"permission_key":                       "功能权限",
		"collaboration_workspace":              "协作空间",
		"collaboration_workspace_member_admin": "协作空间成员（系统）",
		"collaboration":                        "当前协作空间",
		"collaboration_workspace_member":       "当前协作空间成员",
		"api_endpoint":                         "API 注册表",
		"system":                               "系统",
	}
	if name, ok := names[resourceCode]; ok {
		return name
	}
	return resourceCode
}

func permissionActionListToMaps(items []user.PermissionKey) []gin.H {
	result := make([]gin.H, 0, len(items))
	for _, item := range items {
		parsedKey := permissionkey.FromKey(item.PermissionKey)
		row := gin.H{
			"id":             item.ID.String(),
			"permissionKey":  item.PermissionKey,
			"resourceCode":   parsedKey.ResourceCode,
			"actionCode":     parsedKey.ActionCode,
			"moduleCode":     item.ModuleCode,
			"contextType":    item.ContextType,
			"featureKind":    item.FeatureKind,
			"name":           item.Name,
			"description":    item.Description,
			"status":         item.Status,
			"sortOrder":      item.SortOrder,
			"isBuiltin":      item.IsBuiltin,
			"moduleGroupId":  "",
			"featureGroupId": "",
		}
		if item.ModuleGroupID != nil {
			row["moduleGroupId"] = item.ModuleGroupID.String()
		}
		if item.FeatureGroupID != nil {
			row["featureGroupId"] = item.FeatureGroupID.String()
		}
		if item.ModuleGroup != nil {
			row["moduleGroup"] = gin.H{
				"id":        item.ModuleGroup.ID.String(),
				"groupType": item.ModuleGroup.GroupType,
				"code":      item.ModuleGroup.Code,
				"name":      item.ModuleGroup.Name,
				"nameEn":    item.ModuleGroup.NameEn,
				"status":    item.ModuleGroup.Status,
				"sortOrder": item.ModuleGroup.SortOrder,
				"isBuiltin": item.ModuleGroup.IsBuiltin,
			}
		}
		if item.FeatureGroup != nil {
			row["featureGroup"] = gin.H{
				"id":        item.FeatureGroup.ID.String(),
				"groupType": item.FeatureGroup.GroupType,
				"code":      item.FeatureGroup.Code,
				"name":      item.FeatureGroup.Name,
				"nameEn":    item.FeatureGroup.NameEn,
				"status":    item.FeatureGroup.Status,
				"sortOrder": item.FeatureGroup.SortOrder,
				"isBuiltin": item.FeatureGroup.IsBuiltin,
			}
		}
		result = append(result, row)
	}
	return result
}

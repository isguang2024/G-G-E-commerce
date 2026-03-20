package user

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/gg-ecommerce/backend/internal/api/dto"
	"github.com/gg-ecommerce/backend/internal/api/errcode"
)

const tenantContextHeader = "X-Tenant-ID"

type UserHandler struct {
	userService       UserService
	permissionService interface {
		GetUserMenuIDs(userID uuid.UUID, tenantID *uuid.UUID) ([]uuid.UUID, error)
	}
	actionRepo interface {
		GetByIDs(ids []uuid.UUID) ([]PermissionAction, error)
	}
	userActionRepo interface {
		GetByUserAndTenant(userID uuid.UUID, tenantID *uuid.UUID) ([]UserActionPermission, error)
		ReplaceUserActions(userID uuid.UUID, tenantID *uuid.UUID, actions []UserActionPermission) error
	}
	menuRepo interface {
		ListAll() ([]Menu, error)
	}
	logger *zap.Logger
}

func NewUserHandler(userService UserService, permissionService interface {
	GetUserMenuIDs(userID uuid.UUID, tenantID *uuid.UUID) ([]uuid.UUID, error)
}, actionRepo interface {
	GetByIDs(ids []uuid.UUID) ([]PermissionAction, error)
}, userActionRepo interface {
	GetByUserAndTenant(userID uuid.UUID, tenantID *uuid.UUID) ([]UserActionPermission, error)
	ReplaceUserActions(userID uuid.UUID, tenantID *uuid.UUID, actions []UserActionPermission) error
}, menuRepo interface {
	ListAll() ([]Menu, error)
}, logger *zap.Logger) *UserHandler {
	return &UserHandler{
		userService:       userService,
		permissionService: permissionService,
		actionRepo:        actionRepo,
		userActionRepo:    userActionRepo,
		menuRepo:          menuRepo,
		logger:            logger,
	}
}

func (h *UserHandler) List(c *gin.Context) {
	var req dto.UserListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		status, resp := errcode.Response(errcode.ErrParamInvalid)
		c.JSON(status, resp)
		return
	}
	if req.ID == "" {
		req.ID = c.Query("id")
	}
	if req.UserPhone == "" {
		req.UserPhone = c.Query("userPhone")
	}
	if req.UserEmail == "" {
		req.UserEmail = c.Query("userEmail")
	}
	if req.UserName == "" {
		req.UserName = c.Query("userName")
	}
	if req.Status == "" {
		req.Status = c.Query("status")
	}
	if req.RoleID == "" {
		req.RoleID = c.Query("roleId")
	}
	if req.RegisterSource == "" {
		req.RegisterSource = c.Query("registerSource")
	}
	if req.InvitedBy == "" {
		req.InvitedBy = c.Query("invitedBy")
	}
	if req.Current <= 0 {
		req.Current = 1
	}
	if req.Size <= 0 {
		req.Size = 20
	}
	list, total, err := h.userService.List(&req)
	if err != nil {
		h.logger.Error("User list failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取用户列表失败")
		c.JSON(status, resp)
		return
	}

	invitedByIDs := make([]uuid.UUID, 0, len(list))
	for _, u := range list {
		if u.InvitedBy != nil {
			invitedByIDs = append(invitedByIDs, *u.InvitedBy)
		}
	}

	inviterMap := make(map[string]gin.H)
	if len(invitedByIDs) > 0 {
		inviters, err := h.userService.GetByIDs(invitedByIDs)
		if err == nil {
			for _, inviter := range inviters {
				inviterName := inviter.Nickname
				if inviterName == "" {
					inviterName = inviter.Username
				}
				inviterMap[inviter.ID.String()] = gin.H{
					"id":       inviter.ID.String(),
					"nickName": inviter.Nickname,
					"userName": inviter.Username,
					"name":     inviterName,
				}
			}
		}
	}

	records := make([]gin.H, 0, len(list))
	for _, u := range list {
		roles := make([]gin.H, 0)
		for _, r := range u.Roles {
			roles = append(roles, gin.H{"id": r.ID.String(), "code": r.Code, "name": r.Name})
		}

		var inviterName string
		if u.InvitedBy != nil {
			if inviter, ok := inviterMap[u.InvitedBy.String()]; ok {
				inviterName = inviter["name"].(string)
			} else {
				inviterName = "未知邀请人"
			}
		}

		records = append(records, gin.H{
			"id":             u.ID.String(),
			"userName":       u.Username,
			"userEmail":      u.Email,
			"nickName":       u.Nickname,
			"userPhone":      u.Phone,
			"systemRemark":   u.SystemRemark,
			"lastLoginTime":  formatNullableTime(u.LastLoginAt),
			"lastLoginIP":    u.LastLoginIP,
			"status":         u.Status,
			"avatar":         u.AvatarURL,
			"createTime":     u.CreatedAt.Format("2006-01-02 15:04:05"),
			"updateTime":     u.UpdatedAt.Format("2006-01-02 15:04:05"),
			"userRoles":      roleCodes(u.Roles),
			"roleDetails":    roleInfos(u.Roles),
			"registerSource": u.RegisterSource,
			"invitedBy":      nullUUIDToString(u.InvitedBy),
			"invitedByName":  inviterName,
		})
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"records": records,
		"total":   total,
		"current": req.Current,
		"size":    req.Size,
	}))
}

func (h *UserHandler) Get(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的用户ID")
		c.JSON(status, resp)
		return
	}
	user, err := h.userService.Get(id)
	if err != nil {
		if err == ErrUserNotFound {
			status, resp := errcode.Response(errcode.ErrUserNotFound)
			c.JSON(status, resp)
			return
		}
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取用户失败")
		c.JSON(status, resp)
		return
	}
	roles := make([]gin.H, 0)
	for _, r := range user.Roles {
		roles = append(roles, gin.H{"id": r.ID.String(), "code": r.Code, "name": r.Name})
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"id":            user.ID.String(),
		"userName":      user.Username,
		"userEmail":     user.Email,
		"nickName":      user.Nickname,
		"userPhone":     user.Phone,
		"systemRemark":  user.SystemRemark,
		"lastLoginTime": formatNullableTime(user.LastLoginAt),
		"lastLoginIP":   user.LastLoginIP,
		"status":        user.Status,
		"avatar":        user.AvatarURL,
		"createTime":    user.CreatedAt.Format("2006-01-02 15:04:05"),
		"updateTime":    user.UpdatedAt.Format("2006-01-02 15:04:05"),
		"roles":         roles,
		"userRoles":     roleCodes(user.Roles),
		"roleDetails":   roleInfos(user.Roles),
	}))
}

func (h *UserHandler) Create(c *gin.Context) {
	var req dto.UserCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		status, resp := errcode.Response(errcode.ErrParamInvalid)
		c.JSON(status, resp)
		return
	}
	user, err := h.userService.Create(&req)
	if err != nil {
		if err == ErrUserExists {
			status, resp := errcode.Response(errcode.ErrUsernameExists)
			c.JSON(status, resp)
			return
		}
		h.logger.Error("Create user failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "创建用户失败")
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{"id": user.ID.String()}))
}

func (h *UserHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的用户ID")
		c.JSON(status, resp)
		return
	}
	var req dto.UserUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		status, resp := errcode.Response(errcode.ErrParamInvalid)
		c.JSON(status, resp)
		return
	}
	if err := h.userService.Update(id, &req); err != nil {
		if err == ErrUserNotFound {
			status, resp := errcode.Response(errcode.ErrUserNotFound)
			c.JSON(status, resp)
			return
		}
		h.logger.Error("Update user failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "更新用户失败")
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(nil))
}

func (h *UserHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的用户ID")
		c.JSON(status, resp)
		return
	}
	if err := h.userService.Delete(id); err != nil {
		if err == ErrUserNotFound {
			status, resp := errcode.Response(errcode.ErrUserNotFound)
			c.JSON(status, resp)
			return
		}
		h.logger.Error("Delete user failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "删除用户失败")
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(nil))
}

func (h *UserHandler) AssignRoles(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的用户ID")
		c.JSON(status, resp)
		return
	}
	var req dto.UserAssignRolesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		status, resp := errcode.Response(errcode.ErrParamInvalid)
		c.JSON(status, resp)
		return
	}
	if err := h.userService.AssignRoles(id, req.RoleIDs); err != nil {
		if err == ErrUserNotFound {
			status, resp := errcode.Response(errcode.ErrUserNotFound)
			c.JSON(status, resp)
			return
		}
		h.logger.Error("Assign roles failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "分配角色失败")
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(nil))
}

func (h *UserHandler) GetActions(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的用户ID")
		c.JSON(status, resp)
		return
	}
	if _, err := h.userService.Get(id); err != nil {
		if err == ErrUserNotFound {
			status, resp := errcode.Response(errcode.ErrUserNotFound)
			c.JSON(status, resp)
			return
		}
		h.logger.Error("Get user before actions failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取用户失败")
		c.JSON(status, resp)
		return
	}

	records, err := h.userActionRepo.GetByUserAndTenant(id, nil)
	if err != nil {
		h.logger.Error("Get user action permissions failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取用户功能权限失败")
		c.JSON(status, resp)
		return
	}

	actionIDs := make([]uuid.UUID, 0, len(records))
	for _, record := range records {
		actionIDs = append(actionIDs, record.ActionID)
	}
	actionMap := make(map[uuid.UUID]PermissionAction)
	if len(actionIDs) > 0 {
		actions, actionErr := h.actionRepo.GetByIDs(actionIDs)
		if actionErr != nil {
			h.logger.Error("Get user actions metadata failed", zap.Error(actionErr))
			status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取功能权限详情失败")
			c.JSON(status, resp)
			return
		}
		for _, action := range actions {
			actionMap[action.ID] = action
		}
	}

	items := make([]gin.H, 0, len(records))
	for _, record := range records {
		item := gin.H{
			"action_id": record.ActionID.String(),
			"effect":    record.Effect,
		}
		if action, ok := actionMap[record.ActionID]; ok {
			item["action"] = gin.H{
				"id":                      action.ID.String(),
				"resource_code":           action.ResourceCode,
				"action_code":             action.ActionCode,
				"name":                    action.Name,
				"description":             action.Description,
				"scope_id":                action.ScopeID.String(),
				"scope_code":              action.Scope.Code,
				"scope_name":              action.Scope.Name,
				"data_permission_code":    action.Scope.DataPermissionCode,
				"data_permission_name":    action.Scope.DataPermissionName,
				"status":                  action.Status,
				"sort_order":              action.SortOrder,
			}
		}
		items = append(items, item)
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{"actions": items}))
}

func (h *UserHandler) SetActions(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的用户ID")
		c.JSON(status, resp)
		return
	}
	if _, err := h.userService.Get(id); err != nil {
		if err == ErrUserNotFound {
			status, resp := errcode.Response(errcode.ErrUserNotFound)
			c.JSON(status, resp)
			return
		}
		h.logger.Error("Get user before setting actions failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取用户失败")
		c.JSON(status, resp)
		return
	}

	var req dto.UserActionPermissionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		status, resp := errcode.Response(errcode.ErrParamInvalid)
		c.JSON(status, resp)
		return
	}

	actionIDs := make([]uuid.UUID, 0, len(req.Actions))
	records := make([]UserActionPermission, 0, len(req.Actions))
	for _, item := range req.Actions {
		actionID, parseErr := uuid.Parse(item.ActionID)
		if parseErr != nil {
			status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的功能权限ID")
			c.JSON(status, resp)
			return
		}
		actionIDs = append(actionIDs, actionID)
		records = append(records, UserActionPermission{
			UserID:   id,
			ActionID: actionID,
			TenantID: nil,
			Effect:   item.Effect,
		})
	}

	actions, err := h.actionRepo.GetByIDs(actionIDs)
	if err != nil {
		h.logger.Error("Get permission actions failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取功能权限失败")
		c.JSON(status, resp)
		return
	}
	if len(actions) != len(actionIDs) {
		status, resp := errcode.ResponseWithMsg(errcode.ErrParamInvalid, "包含不存在的功能权限")
		c.JSON(status, resp)
		return
	}
	for _, action := range actions {
		if action.Scope.Code != "global" {
			status, resp := errcode.ResponseWithMsg(errcode.ErrParamInvalid, "单用户权限只能配置平台级功能权限")
			c.JSON(status, resp)
			return
		}
	}

	if err := h.userActionRepo.ReplaceUserActions(id, nil, records); err != nil {
		h.logger.Error("Set user action permissions failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "保存用户功能权限失败")
		c.JSON(status, resp)
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(nil))
}

func (h *UserHandler) GetPermissions(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的用户ID")
		c.JSON(status, resp)
		return
	}

	var tenantID *uuid.UUID
	tenantIDStr := strings.TrimSpace(c.Query("tenant_id"))
	if tenantIDStr == "" {
		tenantIDStr = strings.TrimSpace(c.GetHeader(tenantContextHeader))
	}
	if tenantIDStr != "" {
		if parsed, parseErr := uuid.Parse(tenantIDStr); parseErr == nil {
			tenantID = &parsed
		}
	}

	menuIDs, err := h.permissionService.GetUserMenuIDs(id, tenantID)
	if err != nil {
		h.logger.Error("Get user permissions failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取用户权限失败")
		c.JSON(status, resp)
		return
	}

	allMenus, err := h.menuRepo.ListAll()
	if err != nil {
		h.logger.Error("Get menus failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取菜单列表失败")
		c.JSON(status, resp)
		return
	}

	menuIDSet := make(map[uuid.UUID]bool)
	for _, mid := range menuIDs {
		menuIDSet[mid] = true
	}

	menuTree := buildMenuTree(allMenus, menuIDSet)

	c.JSON(http.StatusOK, dto.SuccessResponse(menuTree))
}

func roleCodes(roles []Role) []string {
	codes := make([]string, 0, len(roles))
	for _, r := range roles {
		codes = append(codes, r.Code)
	}
	return codes
}

func roleInfos(roles []Role) []gin.H {
	infos := make([]gin.H, 0, len(roles))
	for _, r := range roles {
		infos = append(infos, gin.H{"code": r.Code, "name": r.Name})
	}
	return infos
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

func formatNullableTime(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format("2006-01-02 15:04:05")
}

func nullUUIDToString(id *uuid.UUID) string {
	if id == nil {
		return ""
	}
	return id.String()
}

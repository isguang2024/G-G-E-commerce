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
	"github.com/gg-ecommerce/backend/internal/pkg/platformaccess"
	"github.com/gg-ecommerce/backend/internal/pkg/teamboundary"
)

const tenantContextHeader = "X-Tenant-ID"

type UserHandler struct {
	userService    UserService
	featurePkgRepo interface {
		GetByIDs(ids []uuid.UUID) ([]FeaturePackage, error)
	}
	platformService platformaccess.Service
	boundaryService teamboundary.Service
	roleRepo        interface {
		GetByIDs(ids []uuid.UUID) ([]Role, error)
	}
	userRoleRepo interface {
		GetEffectiveActiveRoleIDsByUserAndTenant(userID uuid.UUID, tenantID *uuid.UUID) ([]uuid.UUID, error)
	}
	userPackageRepo interface {
		GetPackageIDsByUserID(userID uuid.UUID) ([]uuid.UUID, error)
		ReplaceUserPackages(userID uuid.UUID, packageIDs []uuid.UUID, grantedBy *uuid.UUID) error
	}
	userHiddenMenuRepo interface {
		GetMenuIDsByUserID(userID uuid.UUID) ([]uuid.UUID, error)
		ReplaceUserHiddenMenus(userID uuid.UUID, menuIDs []uuid.UUID) error
	}
	menuRepo interface {
		ListAll() ([]Menu, error)
	}
	refresher interface {
		RefreshPlatformUser(userID uuid.UUID) error
	}
	logger *zap.Logger
}

func NewUserHandler(userService UserService, featurePkgRepo interface {
	GetByIDs(ids []uuid.UUID) ([]FeaturePackage, error)
}, platformService platformaccess.Service, boundaryService teamboundary.Service, roleRepo interface {
	GetByIDs(ids []uuid.UUID) ([]Role, error)
}, userRoleRepo interface {
	GetEffectiveActiveRoleIDsByUserAndTenant(userID uuid.UUID, tenantID *uuid.UUID) ([]uuid.UUID, error)
}, userPackageRepo interface {
	GetPackageIDsByUserID(userID uuid.UUID) ([]uuid.UUID, error)
	ReplaceUserPackages(userID uuid.UUID, packageIDs []uuid.UUID, grantedBy *uuid.UUID) error
}, userHiddenMenuRepo interface {
	GetMenuIDsByUserID(userID uuid.UUID) ([]uuid.UUID, error)
	ReplaceUserHiddenMenus(userID uuid.UUID, menuIDs []uuid.UUID) error
}, menuRepo interface {
	ListAll() ([]Menu, error)
}, refresher interface {
	RefreshPlatformUser(userID uuid.UUID) error
}, logger *zap.Logger) *UserHandler {
	return &UserHandler{
		userService:        userService,
		featurePkgRepo:     featurePkgRepo,
		platformService:    platformService,
		boundaryService:    boundaryService,
		roleRepo:           roleRepo,
		userRoleRepo:       userRoleRepo,
		userPackageRepo:    userPackageRepo,
		userHiddenMenuRepo: userHiddenMenuRepo,
		menuRepo:           menuRepo,
		refresher:          refresher,
		logger:             logger,
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
	if h.refresher != nil {
		if err := h.refresher.RefreshPlatformUser(id); err != nil {
			h.logger.Error("Refresh platform user after assigning roles failed", zap.Error(err))
			status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "刷新用户权限快照失败")
			c.JSON(status, resp)
			return
		}
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(nil))
}

func (h *UserHandler) GetMenus(c *gin.Context) {
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
		h.logger.Error("Get user before menus failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取用户失败")
		c.JSON(status, resp)
		return
	}
	snapshot, err := h.getPlatformSnapshot(id)
	if err != nil {
		h.logger.Error("Get user platform snapshot for menus failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取用户功能包范围失败")
		c.JSON(status, resp)
		return
	}

	menuIDs := snapshot.MenuIDs
	if snapshot.MenuIDs == nil {
		menuIDs = []uuid.UUID{}
	}
	availableMenuIDs := snapshot.AvailableMenuIDs
	if snapshot.AvailableMenuIDs == nil {
		availableMenuIDs = []uuid.UUID{}
	}
	hiddenMenuIDs := snapshot.HiddenMenuIDs
	if hiddenMenuIDs == nil {
		hiddenMenuIDs = []uuid.UUID{}
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"menu_ids":             packageIDsToStrings(menuIDs),
		"available_menu_ids":   packageIDsToStrings(availableMenuIDs),
		"hidden_menu_ids":      packageIDsToStrings(hiddenMenuIDs),
		"expanded_package_ids": packageIDsToStrings(snapshot.ExpandedPackageIDs),
		"derived_sources":      buildUserMenuSourceMaps(snapshot.AvailableMenuMap),
		"has_package_config":   snapshot.HasPackageConfig,
	}))
}

func (h *UserHandler) SetMenus(c *gin.Context) {
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
		h.logger.Error("Get user before setting menus failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取用户失败")
		c.JSON(status, resp)
		return
	}
	snapshot, err := h.getPlatformSnapshot(id)
	if err != nil {
		h.logger.Error("Get user platform snapshot for setting menus failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取用户功能包范围失败")
		c.JSON(status, resp)
		return
	}
	if !snapshot.HasPackageConfig {
		status, resp := errcode.ResponseWithMsg(errcode.ErrForbidden, "当前用户尚未绑定功能包，不能配置菜单裁剪")
		c.JSON(status, resp)
		return
	}

	var req dto.TenantMenuPermissionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		status, resp := errcode.Response(errcode.ErrParamInvalid)
		c.JSON(status, resp)
		return
	}

	availableMenuSet := uuidSliceToSet(snapshot.AvailableMenuIDs)
	menuIDs := make([]uuid.UUID, 0, len(req.MenuIDs))
	for _, item := range req.MenuIDs {
		menuID, parseErr := uuid.Parse(item)
		if parseErr != nil {
			status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的菜单ID")
			c.JSON(status, resp)
			return
		}
		if !availableMenuSet[menuID] {
			status, resp := errcode.ResponseWithMsg(errcode.ErrForbidden, "存在超出当前用户已生效功能包范围的菜单")
			c.JSON(status, resp)
			return
		}
		menuIDs = append(menuIDs, menuID)
	}
	blockedMenuIDs := excludeUUIDs(snapshot.AvailableMenuIDs, menuIDs)
	if err := h.userHiddenMenuRepo.ReplaceUserHiddenMenus(id, blockedMenuIDs); err != nil {
		h.logger.Error("Set user hidden menus failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "保存用户菜单裁剪失败")
		c.JSON(status, resp)
		return
	}
	if h.refresher != nil {
		if err := h.refresher.RefreshPlatformUser(id); err != nil {
			h.logger.Error("Refresh platform user after setting menus failed", zap.Error(err))
			status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "刷新用户权限快照失败")
			c.JSON(status, resp)
			return
		}
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(nil))
}

func (h *UserHandler) getPlatformSnapshot(userID uuid.UUID) (*platformaccess.Snapshot, error) {
	if h.platformService == nil {
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
	snapshot, err := h.platformService.GetSnapshot(userID)
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

	menuIDs, err := h.getPermissionMenuIDs(id, tenantID)
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

func (h *UserHandler) getPermissionMenuIDs(userID uuid.UUID, tenantID *uuid.UUID) ([]uuid.UUID, error) {
	userEntity, err := h.userService.Get(userID)
	if err != nil {
		return nil, err
	}
	if tenantID == nil {
		if userEntity.IsSuperAdmin {
			return h.listEnabledMenuIDs()
		}
		snapshot, err := h.getPlatformSnapshot(userID)
		if err != nil {
			return nil, err
		}
		return snapshot.MenuIDs, nil
	}
	return h.getTeamPermissionMenuIDs(userID, *tenantID)
}

func (h *UserHandler) getTeamPermissionMenuIDs(userID, teamID uuid.UUID) ([]uuid.UUID, error) {
	if h.userRoleRepo == nil || h.roleRepo == nil || h.boundaryService == nil {
		return h.finalizePermissionMenuIDs(nil)
	}
	roleIDs, err := h.userRoleRepo.GetEffectiveActiveRoleIDsByUserAndTenant(userID, &teamID)
	if err != nil {
		return nil, err
	}
	if len(roleIDs) == 0 {
		return h.finalizePermissionMenuIDs(nil)
	}
	roles, err := h.roleRepo.GetByIDs(roleIDs)
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
		snapshot, snapshotErr := h.boundaryService.GetRoleSnapshot(teamID, roleID, role.TenantID == nil)
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
	return h.finalizePermissionMenuIDs(menuIDs)
}

func (h *UserHandler) finalizePermissionMenuIDs(menuIDs []uuid.UUID) ([]uuid.UUID, error) {
	allMenus, err := h.menuRepo.ListAll()
	if err != nil {
		return nil, err
	}
	enabledSet := make(map[uuid.UUID]struct{}, len(allMenus))
	publicIDs := make([]uuid.UUID, 0)
	for _, menu := range allMenus {
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

func (h *UserHandler) listEnabledMenuIDs() ([]uuid.UUID, error) {
	allMenus, err := h.menuRepo.ListAll()
	if err != nil {
		return nil, err
	}
	result := make([]uuid.UUID, 0, len(allMenus))
	for _, menu := range allMenus {
		if !isMenuEnabled(menu) {
			continue
		}
		result = append(result, menu.ID)
	}
	return result, nil
}

func intersectUUIDSlices(left []uuid.UUID, right []uuid.UUID) []uuid.UUID {
	if len(left) == 0 || len(right) == 0 {
		return []uuid.UUID{}
	}
	rightSet := make(map[uuid.UUID]struct{}, len(right))
	for _, id := range right {
		rightSet[id] = struct{}{}
	}
	result := make([]uuid.UUID, 0, len(left))
	seen := make(map[uuid.UUID]struct{}, len(left))
	for _, id := range left {
		if _, ok := rightSet[id]; !ok {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	return result
}

func (h *UserHandler) GetPackages(c *gin.Context) {
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
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取用户失败")
		c.JSON(status, resp)
		return
	}
	packageIDs, err := h.userPackageRepo.GetPackageIDsByUserID(id)
	if err != nil {
		h.logger.Error("Get user packages failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取用户功能包失败")
		c.JSON(status, resp)
		return
	}
	packages, err := h.featurePkgRepo.GetByIDs(packageIDs)
	if err != nil {
		h.logger.Error("Get user package details failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取功能包详情失败")
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"package_ids": packageIDsToStrings(packageIDs),
		"packages":    featurePackageListToMaps(packages),
	}))
}

func (h *UserHandler) SetPackages(c *gin.Context) {
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
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取用户失败")
		c.JSON(status, resp)
		return
	}

	var req dto.RoleFeaturePackagesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		status, resp := errcode.Response(errcode.ErrParamInvalid)
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
	if len(packageIDs) > 0 {
		packages, err := h.featurePkgRepo.GetByIDs(packageIDs)
		if err != nil {
			h.logger.Error("Get user package details failed", zap.Error(err))
			status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取功能包失败")
			c.JSON(status, resp)
			return
		}
		if len(packages) != len(packageIDs) {
			status, resp := errcode.ResponseWithMsg(errcode.ErrParamInvalid, "包含不存在的功能包")
			c.JSON(status, resp)
			return
		}
		for _, item := range packages {
			if !supportsPlatformContext(item.ContextType) {
				status, resp := errcode.ResponseWithMsg(errcode.ErrForbidden, "仅支持绑定平台功能包")
				c.JSON(status, resp)
				return
			}
		}
	}
	var grantedBy *uuid.UUID
	if rawUserID, ok := c.Get("user_id"); ok {
		if userIDStr, ok := rawUserID.(string); ok {
			if userID, parseErr := uuid.Parse(userIDStr); parseErr == nil {
				grantedBy = &userID
			}
		}
	}
	if err := h.userPackageRepo.ReplaceUserPackages(id, packageIDs, grantedBy); err != nil {
		h.logger.Error("Set user packages failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "保存用户功能包失败")
		c.JSON(status, resp)
		return
	}
	if h.refresher != nil {
		if err := h.refresher.RefreshPlatformUser(id); err != nil {
			h.logger.Error("Refresh platform user after setting packages failed", zap.Error(err))
			status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "刷新用户权限快照失败")
			c.JSON(status, resp)
			return
		}
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(nil))
}

func roleCodes(roles []Role) []string {
	codes := make([]string, 0, len(roles))
	for _, r := range roles {
		codes = append(codes, r.Code)
	}
	return codes
}

func packageIDsToStrings(ids []uuid.UUID) []string {
	result := make([]string, 0, len(ids))
	for _, id := range ids {
		result = append(result, id.String())
	}
	return result
}

func actionIDsToStrings(ids []uuid.UUID) []string {
	return packageIDsToStrings(ids)
}

func actionListToMaps(items []PermissionKey) []gin.H {
	result := make([]gin.H, 0, len(items))
	for _, item := range items {
		result = append(result, gin.H{
			"id":             item.ID.String(),
			"permission_key": item.PermissionKey,
			"name":           item.Name,
			"description":    item.Description,
			"module_code":    item.ModuleCode,
			"context_type":   item.ContextType,
			"feature_kind":   item.FeatureKind,
			"status":         item.Status,
			"sort_order":     item.SortOrder,
		})
	}
	return result
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

func supportsPlatformContext(contextType string) bool {
	return contextType == "" || contextType == "platform" || contextType == "common"
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

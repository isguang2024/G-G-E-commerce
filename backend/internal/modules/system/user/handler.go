package user

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/gg-ecommerce/backend/internal/api/dto"
	"github.com/gg-ecommerce/backend/internal/api/errcode"
	"github.com/gg-ecommerce/backend/internal/modules/system/models"
	appctx "github.com/gg-ecommerce/backend/internal/pkg/appctx"
	"github.com/gg-ecommerce/backend/internal/pkg/appscope"
	"github.com/gg-ecommerce/backend/internal/pkg/authorization"
	"github.com/gg-ecommerce/backend/internal/pkg/collaborationworkspaceboundary"
	"github.com/gg-ecommerce/backend/internal/pkg/permissionkey"
	"github.com/gg-ecommerce/backend/internal/pkg/platformaccess"
	"github.com/gg-ecommerce/backend/internal/pkg/workspacerolebinding"
)

const collaborationWorkspaceContextHeader = "X-Collaboration-Workspace-Id"

type UserHandler struct {
	db             *gorm.DB
	userService    UserService
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
		RefreshPersonalWorkspaceUser(userID uuid.UUID) error
		RefreshCollaborationWorkspace(collaborationWorkspaceID uuid.UUID) error
	}
	logger *zap.Logger
}

func NewUserHandler(db *gorm.DB, userService UserService, featurePkgRepo interface {
	GetByIDs(ids []uuid.UUID) ([]FeaturePackage, error)
}, keyRepo interface {
	GetByPermissionKey(permissionKey string) (*PermissionKey, error)
}, personalWorkspaceAccessService platformaccess.Service, boundaryService collaborationworkspaceboundary.Service, roleRepo interface {
	GetByIDs(ids []uuid.UUID) ([]Role, error)
}, authzService interface {
	Authorize(userID uuid.UUID, collaborationWorkspaceID *uuid.UUID, permissionKey string, legacy ...string) (bool, *models.PermissionKey, error)
	AuthorizeInApp(userID uuid.UUID, collaborationWorkspaceID *uuid.UUID, appKey string, permissionKey string, legacy ...string) (bool, *models.PermissionKey, error)
}, userRoleRepo interface {
	GetEffectiveActiveRoleIDsByUserAndCollaborationWorkspace(userID uuid.UUID, collaborationWorkspaceID *uuid.UUID) ([]uuid.UUID, error)
}, collaborationWorkspaceMemberRepo interface {
	GetByUserAndCollaborationWorkspace(userID, collaborationWorkspaceID uuid.UUID) (*CollaborationWorkspaceMember, error)
	GetCollaborationWorkspacesByUserID(userID uuid.UUID) ([]CollaborationWorkspace, error)
}, userPackageRepo interface {
	GetPackageIDsByUserID(userID uuid.UUID) ([]uuid.UUID, error)
	ReplaceUserPackages(userID uuid.UUID, packageIDs []uuid.UUID, grantedBy *uuid.UUID) error
}, userHiddenMenuRepo interface {
	GetMenuIDsByUserID(userID uuid.UUID) ([]uuid.UUID, error)
	ReplaceUserHiddenMenus(userID uuid.UUID, menuIDs []uuid.UUID) error
}, menuRepo interface {
	ListAll() ([]Menu, error)
}, refresher interface {
	RefreshPersonalWorkspaceUser(userID uuid.UUID) error
	RefreshCollaborationWorkspace(collaborationWorkspaceID uuid.UUID) error
}, logger *zap.Logger) *UserHandler {
	return &UserHandler{
		db:                               db,
		userService:                      userService,
		featurePkgRepo:                   featurePkgRepo,
		keyRepo:                          keyRepo,
		personalWorkspaceAccessService:   personalWorkspaceAccessService,
		boundaryService:                  boundaryService,
		authzService:                     authzService,
		roleRepo:                         roleRepo,
		userRoleRepo:                     userRoleRepo,
		collaborationWorkspaceMemberRepo: collaborationWorkspaceMemberRepo,
		userPackageRepo:                  userPackageRepo,
		userHiddenMenuRepo:               userHiddenMenuRepo,
		menuRepo:                         menuRepo,
		refresher:                        refresher,
		logger:                           logger,
	}
}

func (h *UserHandler) GetCollaborationWorkspaces(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
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
	if h.collaborationWorkspaceMemberRepo == nil {
		c.JSON(http.StatusOK, dto.SuccessResponse([]gin.H{}))
		return
	}
	collaboration_workspaces, err := h.collaborationWorkspaceMemberRepo.GetCollaborationWorkspacesByUserID(id)
	if err != nil {
		h.logger.Error("Get user collaboration workspaces failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取用户所在协作空间失败")
		c.JSON(status, resp)
		return
	}
	result := make([]gin.H, 0, len(collaboration_workspaces))
	for _, item := range collaboration_workspaces {
		result = append(result, gin.H{
			"id":     item.ID.String(),
			"name":   item.Name,
			"status": item.Status,
		})
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(result))
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
		if err == ErrEmailExists {
			status, resp := errcode.Response(errcode.ErrEmailExists)
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
		if err := h.refresher.RefreshPersonalWorkspaceUser(id); err != nil {
			h.logger.Error("Refresh personal workspace user after assigning roles failed", zap.Error(err))
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
	appKey, err := appctx.RequireRequestAppKey(c)
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrParamInvalid, "app_key 为必填项")
		c.JSON(status, resp)
		return
	}
		snapshot, err := h.getPersonalWorkspaceSnapshot(id, appKey)
	if err != nil {
		h.logger.Error("Get user personal workspace snapshot for menus failed", zap.Error(err))
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
	appKey, err := appctx.RequireRequestAppKey(c)
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrParamInvalid, "app_key 为必填项")
		c.JSON(status, resp)
		return
	}
		snapshot, err := h.getPersonalWorkspaceSnapshot(id, appKey)
	if err != nil {
		h.logger.Error("Get user personal workspace snapshot for setting menus failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取用户功能包范围失败")
		c.JSON(status, resp)
		return
	}
	if !snapshot.HasPackageConfig {
		status, resp := errcode.ResponseWithMsg(errcode.ErrForbidden, "当前用户尚未绑定功能包，不能配置菜单裁剪")
		c.JSON(status, resp)
		return
	}

	var req dto.CollaborationWorkspaceMenuPermissionsRequest
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
	if err := appscope.ReplaceUserHiddenMenusInApp(h.db, id, appKey, blockedMenuIDs); err != nil {
		h.logger.Error("Set user hidden menus failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "保存用户菜单裁剪失败")
		c.JSON(status, resp)
		return
	}
	if h.refresher != nil {
		if err := h.refresher.RefreshPersonalWorkspaceUser(id); err != nil {
			h.logger.Error("Refresh personal workspace user after setting menus failed", zap.Error(err))
			status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "刷新用户权限快照失败")
			c.JSON(status, resp)
			return
		}
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(nil))
}

func (h *UserHandler) getPersonalWorkspaceSnapshot(userID uuid.UUID, appKey string) (*platformaccess.Snapshot, error) {
	if h.personalWorkspaceAccessService == nil {
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
	snapshot, err := h.personalWorkspaceAccessService.GetSnapshot(userID, appctx.NormalizeAppKey(appKey))
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

	var collaborationWorkspaceID *uuid.UUID
	collaborationWorkspaceIDStr := strings.TrimSpace(c.Query("collaboration_workspace_id"))
	if collaborationWorkspaceIDStr == "" {
		collaborationWorkspaceIDStr = strings.TrimSpace(c.GetHeader(collaborationWorkspaceContextHeader))
	}
	if collaborationWorkspaceIDStr != "" {
		if parsed, parseErr := uuid.Parse(collaborationWorkspaceIDStr); parseErr == nil {
			collaborationWorkspaceID = &parsed
		}
	}

	currentAppKey := appctx.CurrentAppKey(c)
	menuIDs, err := h.getPermissionMenuIDs(id, collaborationWorkspaceID, currentAppKey)
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

	menuTree := buildMenuTree(filterMenusByApp(allMenus, currentAppKey), menuIDSet)

	c.JSON(http.StatusOK, dto.SuccessResponse(menuTree))
}

func (h *UserHandler) GetPermissionDiagnosis(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的用户ID")
		c.JSON(status, resp)
		return
	}

	var req dto.UserPermissionDiagnosisRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		status, resp := errcode.Response(errcode.ErrParamInvalid)
		c.JSON(status, resp)
		return
	}

	collaborationWorkspaceID, parseErr := parseCollaborationWorkspaceContextID(req.CollaborationWorkspaceID, c.GetHeader(collaborationWorkspaceContextHeader))
	if parseErr != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的协作空间ID")
		c.JSON(status, resp)
		return
	}

	payload, err := h.buildPermissionDiagnosis(id, collaborationWorkspaceID, req.PermissionKey, appctx.CurrentAppKey(c))
	if err != nil {
		if err == ErrUserNotFound {
			status, resp := errcode.Response(errcode.ErrUserNotFound)
			c.JSON(status, resp)
			return
		}
		h.logger.Error("Get user permission diagnosis failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取用户权限诊断失败")
		c.JSON(status, resp)
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(payload))
}

func (h *UserHandler) RefreshPermissionSnapshot(c *gin.Context) {
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
		h.logger.Error("Get user before refreshing permission snapshot failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取用户失败")
		c.JSON(status, resp)
		return
	}

	var req dto.UserPermissionRefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil && err.Error() != "EOF" {
		status, resp := errcode.Response(errcode.ErrParamInvalid)
		c.JSON(status, resp)
		return
	}

	collaborationWorkspaceID, parseErr := parseCollaborationWorkspaceContextID(req.CollaborationWorkspaceID, c.GetHeader(collaborationWorkspaceContextHeader))
	if parseErr != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的协作空间ID")
		c.JSON(status, resp)
		return
	}

	if collaborationWorkspaceID == nil {
		if h.refresher != nil {
			if err := h.refresher.RefreshPersonalWorkspaceUser(id); err != nil {
				h.logger.Error("Refresh platform permission snapshot failed", zap.Error(err))
				status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "刷新个人空间权限快照失败")
				c.JSON(status, resp)
				return
			}
		}
	} else if h.refresher != nil {
		if err := h.refresher.RefreshCollaborationWorkspace(*collaborationWorkspaceID); err != nil {
			h.logger.Error("Refresh collaboration workspace permission snapshot failed", zap.Error(err))
			status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "刷新协作空间权限快照失败")
			c.JSON(status, resp)
			return
		}
	}

	payload, err := h.buildPermissionDiagnosis(id, collaborationWorkspaceID, "", appctx.CurrentAppKey(c))
	if err != nil {
		h.logger.Error("Build permission diagnosis after refresh failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "刷新权限快照后读取失败")
		c.JSON(status, resp)
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(payload))
}

func (h *UserHandler) buildPermissionDiagnosis(userID uuid.UUID, collaborationWorkspaceID *uuid.UUID, rawPermissionKey string, appKey string) (gin.H, error) {
	userEntity, err := h.userService.Get(userID)
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
	snapshot, err := h.getPersonalWorkspaceSnapshot(userID, appKey)
		if err != nil {
			return nil, err
		}
		meta, err := h.getPersonalWorkspaceSnapshotRecord(userID, appKey)
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
			diagnosis, err := h.buildPersonalWorkspacePermissionDiagnosis(userEntity, snapshot, permissionKeyValue, appKey)
			if err != nil {
				return nil, err
			}
			payload["diagnosis"] = diagnosis
		}
		return payload, nil
	}

	collaborationWorkspaceSnapshot, err := h.boundaryService.GetSnapshot(*collaborationWorkspaceID, appctx.NormalizeAppKey(appKey))
	if err != nil {
		return nil, err
	}
	currentCollaborationWorkspaceID := ""
	if workspace, workspaceErr := workspacerolebinding.GetCollaborationWorkspaceByCollaborationWorkspaceID(h.db, *collaborationWorkspaceID); workspaceErr == nil && workspace != nil {
		currentCollaborationWorkspaceID = workspace.ID.String()
	}
	collaborationWorkspaceMeta, err := h.getCollaborationWorkspaceSnapshotRecord(*collaborationWorkspaceID, appKey)
	if err != nil {
		return nil, err
	}
	roleStates, err := h.buildCollaborationWorkspaceRoleStates(userID, *collaborationWorkspaceID, permissionKeyValue, appKey)
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
		"collaboration_workspace_member":   h.buildCollaborationWorkspaceMemberMap(userID, *collaborationWorkspaceID),
		"collaboration_workspace_packages": h.buildPackageMapsByIDs(collaborationWorkspaceSnapshot.ExpandedPackageIDs),
		"diagnosis":                        nil,
	}
	if permissionKeyValue != "" {
		diagnosis, err := h.buildCollaborationWorkspacePermissionDiagnosis(userEntity, *collaborationWorkspaceID, collaborationWorkspaceSnapshot, roleStates, permissionKeyValue, appKey)
		if err != nil {
			return nil, err
		}
		payload["diagnosis"] = diagnosis
	}
	return payload, nil
}

func (h *UserHandler) buildPersonalWorkspacePermissionDiagnosis(userEntity *User, snapshot *platformaccess.Snapshot, permissionKeyValue string, appKey string) (gin.H, error) {
	allowed, actionDef, authErr := h.authzService.AuthorizeInApp(userEntity.ID, nil, appctx.NormalizeAppKey(appKey), permissionKeyValue)
	actionDetail, err := h.loadPermissionKeyDetail(permissionKeyValue)
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
		"source_packages":         h.buildPackageMapsByIDs(sourcePackageIDs),
		"role_results":            []gin.H{},
	}, nil
}

func (h *UserHandler) buildCollaborationWorkspacePermissionDiagnosis(userEntity *User, collaborationWorkspaceID uuid.UUID, collaborationWorkspaceSnapshot *collaborationworkspaceboundary.Snapshot, roleStates []gin.H, permissionKeyValue string, appKey string) (gin.H, error) {
	allowed, actionDef, authErr := h.authzService.AuthorizeInApp(userEntity.ID, &collaborationWorkspaceID, appctx.NormalizeAppKey(appKey), permissionKeyValue)
	actionDetail, err := h.loadPermissionKeyDetail(permissionKeyValue)
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
	memberStatus, memberMatched, err := h.getCollaborationWorkspaceMemberDiagnosis(userEntity.ID, collaborationWorkspaceID)
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
		"source_packages":                    h.buildPackageMapsByIDs(sourcePackageIDs),
		"role_results":                       roleStates,
	}, nil
}

func (h *UserHandler) buildCollaborationWorkspaceRoleStates(userID, collaborationWorkspaceID uuid.UUID, permissionKeyValue string, appKey string) ([]gin.H, error) {
	if h.userRoleRepo == nil || h.roleRepo == nil || h.boundaryService == nil {
		return []gin.H{}, nil
	}
	roleIDs, err := h.userRoleRepo.GetEffectiveActiveRoleIDsByUserAndCollaborationWorkspace(userID, &collaborationWorkspaceID)
	if err != nil {
		return nil, err
	}
	if len(roleIDs) == 0 {
		return []gin.H{}, nil
	}
	roles, err := h.roleRepo.GetByIDs(roleIDs)
	if err != nil {
		return nil, err
	}

	actionDetail, err := h.loadPermissionKeyDetail(permissionKeyValue)
	if err != nil {
		return nil, err
	}

	roleStates := make([]gin.H, 0, len(roles))
	for _, role := range roles {
		inheritAll := role.CollaborationWorkspaceID == nil
		snapshot, err := h.boundaryService.GetRoleSnapshot(collaborationWorkspaceID, role.ID, inheritAll, appctx.NormalizeAppKey(appKey))
		if err != nil {
			return nil, err
		}
		roleMeta, err := h.getCollaborationWorkspaceRoleSnapshotRecord(collaborationWorkspaceID, role.ID, appKey)
		if err != nil {
			return nil, err
		}

		roleState := gin.H{
			"role_id":                role.ID.String(),
			"role_code":              role.Code,
			"role_name":              role.Name,
			"inherited":              snapshot.Inherited,
			"refreshed_at":           formatTimeValue(roleMeta.RefreshedAt),
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
			roleState["source_packages"] = h.buildPackageMapsByIDs(sourceIDs)
		}
		roleStates = append(roleStates, roleState)
	}
	return roleStates, nil
}

func (h *UserHandler) loadPermissionKeyDetail(permissionKeyValue string) (*PermissionKey, error) {
	if strings.TrimSpace(permissionKeyValue) == "" || h.keyRepo == nil {
		return nil, nil
	}
	item, err := h.keyRepo.GetByPermissionKey(permissionKeyValue)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return item, nil
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

func (h *UserHandler) buildPackageMapsByIDs(ids []uuid.UUID) []gin.H {
	if len(ids) == 0 || h.featurePkgRepo == nil {
		return []gin.H{}
	}
	items, err := h.featurePkgRepo.GetByIDs(ids)
	if err != nil {
		h.logger.Warn("Load feature packages for permission diagnosis failed", zap.Error(err))
		return []gin.H{}
	}
	return featurePackageListToMaps(items)
}

func (h *UserHandler) buildCollaborationWorkspaceMemberMap(userID, collaborationWorkspaceID uuid.UUID) gin.H {
	if h.collaborationWorkspaceMemberRepo == nil {
		return nil
	}
	member, err := h.collaborationWorkspaceMemberRepo.GetByUserAndCollaborationWorkspace(userID, collaborationWorkspaceID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return gin.H{
				"matched": false,
				"status":  "missing",
			}
		}
		h.logger.Warn("Load collaboration workspace member for permission diagnosis failed", zap.Error(err))
		return nil
	}
	memberType := mapRoleCodeToMemberType(member.RoleCode)
	workspaceID := ""
	workspaceType := ""
	if workspace, err := workspacerolebinding.GetCollaborationWorkspaceByCollaborationWorkspaceID(h.db, collaborationWorkspaceID); err == nil && workspace != nil {
		workspaceID = workspace.ID.String()
		workspaceType = workspace.WorkspaceType
		var workspaceMember models.WorkspaceMember
		if err := h.db.Where("workspace_id = ? AND user_id = ? AND deleted_at IS NULL", workspace.ID, userID).First(&workspaceMember).Error; err == nil {
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

func (h *UserHandler) getPersonalWorkspaceSnapshotRecord(userID uuid.UUID, appKey string) (*models.PersonalWorkspaceAccessSnapshot, error) {
	if h.db == nil {
		return nil, nil
	}
	var record models.PersonalWorkspaceAccessSnapshot
	if err := h.db.Where("app_key = ? AND user_id = ?", appctx.NormalizeAppKey(appKey), userID).First(&record).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &record, nil
}

func (h *UserHandler) getCollaborationWorkspaceSnapshotRecord(collaborationWorkspaceID uuid.UUID, appKey string) (*models.CollaborationWorkspaceAccessSnapshot, error) {
	if h.db == nil {
		return nil, nil
	}
	var record models.CollaborationWorkspaceAccessSnapshot
	if err := h.db.Where("app_key = ? AND collaboration_workspace_id = ?", appctx.NormalizeAppKey(appKey), collaborationWorkspaceID).First(&record).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &record, nil
}

func (h *UserHandler) getCollaborationWorkspaceRoleSnapshotRecord(collaborationWorkspaceID, roleID uuid.UUID, appKey string) (*models.CollaborationWorkspaceRoleAccessSnapshot, error) {
	if h.db == nil {
		return nil, nil
	}
	var record models.CollaborationWorkspaceRoleAccessSnapshot
	if err := h.db.Where("app_key = ? AND collaboration_workspace_id = ? AND role_id = ?", appctx.NormalizeAppKey(appKey), collaborationWorkspaceID, roleID).First(&record).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &record, nil
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

func (h *UserHandler) getCollaborationWorkspaceMemberDiagnosis(userID, collaborationWorkspaceID uuid.UUID) (string, bool, error) {
	if h.db == nil {
		return "", false, nil
	}
	var member models.CollaborationWorkspaceMember
	if err := h.db.Where("user_id = ? AND collaboration_workspace_id = ?", userID, collaborationWorkspaceID).First(&member).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return "missing", false, nil
		}
		return "", false, err
	}
	return member.Status, member.Status == "active", nil
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

func parseCollaborationWorkspaceContextID(values ...string) (*uuid.UUID, error) {
	for _, value := range values {
		target := strings.TrimSpace(value)
		if target == "" {
			continue
		}
		parsed, err := uuid.Parse(target)
		if err != nil {
			return nil, err
		}
		return &parsed, nil
	}
	return nil, nil
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

func (h *UserHandler) getPermissionMenuIDs(userID uuid.UUID, collaborationWorkspaceID *uuid.UUID, appKey string) ([]uuid.UUID, error) {
	userEntity, err := h.userService.Get(userID)
	if err != nil {
		return nil, err
	}
	if collaborationWorkspaceID == nil {
		if userEntity.IsSuperAdmin {
			return h.listEnabledMenuIDs(appKey)
		}
	snapshot, err := h.getPersonalWorkspaceSnapshot(userID, appKey)
		if err != nil {
			return nil, err
		}
		return snapshot.MenuIDs, nil
	}
	return h.getCollaborationWorkspacePermissionMenuIDs(userID, *collaborationWorkspaceID, appKey)
}

func (h *UserHandler) getCollaborationWorkspacePermissionMenuIDs(userID, collaborationWorkspaceID uuid.UUID, appKey string) ([]uuid.UUID, error) {
	if h.userRoleRepo == nil || h.roleRepo == nil || h.boundaryService == nil {
		return h.finalizePermissionMenuIDs(nil, appKey)
	}
	roleIDs, err := h.userRoleRepo.GetEffectiveActiveRoleIDsByUserAndCollaborationWorkspace(userID, &collaborationWorkspaceID)
	if err != nil {
		return nil, err
	}
	if len(roleIDs) == 0 {
		return h.finalizePermissionMenuIDs(nil, appKey)
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
		snapshot, snapshotErr := h.boundaryService.GetRoleSnapshot(collaborationWorkspaceID, roleID, role.CollaborationWorkspaceID == nil, appctx.NormalizeAppKey(appKey))
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
	return h.finalizePermissionMenuIDs(menuIDs, appKey)
}

func (h *UserHandler) finalizePermissionMenuIDs(menuIDs []uuid.UUID, appKey string) ([]uuid.UUID, error) {
	allMenus, err := h.menuRepo.ListAll()
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

func (h *UserHandler) listEnabledMenuIDs(appKey string) ([]uuid.UUID, error) {
	allMenus, err := h.menuRepo.ListAll()
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
	appKey, err := appctx.RequireRequestAppKey(c)
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrParamInvalid, "app_key 为必填项")
		c.JSON(status, resp)
		return
	}
	packageIDs, err := appscope.PackageIDsByUser(h.db, id, appKey)
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
	bindingWorkspaceID := ""
	bindingWorkspaceType := models.WorkspaceTypePersonal
	if workspace, workspaceErr := workspacerolebinding.GetPersonalWorkspaceByUserID(h.db, id); workspaceErr == nil && workspace != nil {
		bindingWorkspaceID = workspace.ID.String()
		bindingWorkspaceType = workspace.WorkspaceType
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"package_ids":             packageIDsToStrings(packageIDs),
		"packages":                featurePackageListToMaps(packages),
		"binding_workspace_id":    bindingWorkspaceID,
		"binding_workspace_type":  bindingWorkspaceType,
		"binding_workspace_label": "personal_workspace",
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
			if appctx.NormalizeAppKey(item.AppKey) != appctx.NormalizeAppKey(appKey) {
				status, resp := errcode.ResponseWithMsg(errcode.ErrForbidden, "仅支持绑定当前应用内的功能包")
				c.JSON(status, resp)
				return
			}
			if !supportsPersonalWorkspaceContext(item.ContextType) {
				status, resp := errcode.ResponseWithMsg(errcode.ErrForbidden, "仅支持绑定个人空间功能包")
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
	if err := appscope.ReplaceUserPackagesInApp(h.db, id, appKey, packageIDs, grantedBy); err != nil {
		h.logger.Error("Set user packages failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "保存用户功能包失败")
		c.JSON(status, resp)
		return
	}
	if h.refresher != nil {
		if err := h.refresher.RefreshPersonalWorkspaceUser(id); err != nil {
			h.logger.Error("Refresh personal workspace user after setting packages failed", zap.Error(err))
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

func supportsPersonalWorkspaceContext(contextType string) bool {
	return contextType == "" || contextType == "personal" || contextType == "common"
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

package collaborationworkspace

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/gg-ecommerce/backend/internal/api/dto"
	"github.com/gg-ecommerce/backend/internal/api/errcode"
	"github.com/gg-ecommerce/backend/internal/modules/system/models"
	"github.com/gg-ecommerce/backend/internal/modules/system/user"
	workspacepkg "github.com/gg-ecommerce/backend/internal/modules/system/workspace"
	appctx "github.com/gg-ecommerce/backend/internal/pkg/appctx"
	"github.com/gg-ecommerce/backend/internal/pkg/appscope"
	"github.com/gg-ecommerce/backend/internal/pkg/authorization"
	"github.com/gg-ecommerce/backend/internal/pkg/collaborationworkspaceboundary"
	"github.com/gg-ecommerce/backend/internal/pkg/database"
	"github.com/gg-ecommerce/backend/internal/pkg/workspacerolebinding"
)

const collaborationWorkspaceContextHeader = "X-Collaboration-Workspace-Id"

var ErrCurrentCollaborationWorkspaceRoleForbidden = errors.New("collaboration workspace role forbidden")

type CollaborationWorkspaceHandler struct {
	collaborationWorkspaceService            CollaborationWorkspaceService
	collaborationWorkspaceMemberRepo         user.CollaborationWorkspaceMemberRepository
	userRepo                                 user.UserRepository
	roleRepo                                 user.RoleRepository
	roleHiddenMenuRepo                       user.RoleHiddenMenuRepository
	roleDisabledActionRepo                   user.RoleDisabledActionRepository
	userRoleRepo                             user.UserRoleRepository
	actionRepo                               user.PermissionKeyRepository
	blockedMenuRepo                          user.CollaborationWorkspaceBlockedMenuRepository
	blockedActionRepo                        user.CollaborationWorkspaceBlockedActionRepository
	collaborationWorkspaceFeaturePackageRepo user.CollaborationWorkspaceFeaturePackageRepository
	rolePackageRepo                          user.RoleFeaturePackageRepository
	featurePkgRepo                           user.FeaturePackageRepository
	packageActionRepo                        user.FeaturePackageKeyRepository
	packageMenuRepo                          user.FeaturePackageMenuRepository
	boundaryService                          collaborationworkspaceboundary.Service
	refresher                                interface {
		RefreshCollaborationWorkspace(collaborationWorkspaceID uuid.UUID) error
	}
	workspaceService workspacepkg.Service
	authz            *authorization.Service
	logger           *zap.Logger
}

func NewCollaborationWorkspaceHandler(collaborationWorkspaceService CollaborationWorkspaceService, collaborationWorkspaceMemberRepo user.CollaborationWorkspaceMemberRepository, userRepo user.UserRepository, roleRepo user.RoleRepository, roleHiddenMenuRepo user.RoleHiddenMenuRepository, roleDisabledActionRepo user.RoleDisabledActionRepository, userRoleRepo user.UserRoleRepository, actionRepo user.PermissionKeyRepository, blockedMenuRepo user.CollaborationWorkspaceBlockedMenuRepository, blockedActionRepo user.CollaborationWorkspaceBlockedActionRepository, collaborationWorkspaceFeaturePackageRepo user.CollaborationWorkspaceFeaturePackageRepository, rolePackageRepo user.RoleFeaturePackageRepository, featurePkgRepo user.FeaturePackageRepository, packageActionRepo user.FeaturePackageKeyRepository, packageMenuRepo user.FeaturePackageMenuRepository, boundaryService collaborationworkspaceboundary.Service, refresher interface {
	RefreshCollaborationWorkspace(collaborationWorkspaceID uuid.UUID) error
}, workspaceService workspacepkg.Service, authz *authorization.Service, logger *zap.Logger) *CollaborationWorkspaceHandler {
	return &CollaborationWorkspaceHandler{
		collaborationWorkspaceService:            collaborationWorkspaceService,
		collaborationWorkspaceMemberRepo:         collaborationWorkspaceMemberRepo,
		userRepo:                                 userRepo,
		roleRepo:                                 roleRepo,
		roleHiddenMenuRepo:                       roleHiddenMenuRepo,
		roleDisabledActionRepo:                   roleDisabledActionRepo,
		userRoleRepo:                             userRoleRepo,
		actionRepo:                               actionRepo,
		blockedMenuRepo:                          blockedMenuRepo,
		blockedActionRepo:                        blockedActionRepo,
		collaborationWorkspaceFeaturePackageRepo: collaborationWorkspaceFeaturePackageRepo,
		rolePackageRepo:                          rolePackageRepo,
		featurePkgRepo:                           featurePkgRepo,
		packageActionRepo:                        packageActionRepo,
		packageMenuRepo:                          packageMenuRepo,
		boundaryService:                          boundaryService,
		refresher:                                refresher,
		workspaceService:                         workspaceService,
		authz:                                    authz,
		logger:                                   logger,
	}
}

func (h *CollaborationWorkspaceHandler) List(c *gin.Context) {
	var req dto.CollaborationWorkspaceListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		status, resp := errcode.Response(errcode.ErrParamInvalid)
		c.JSON(status, resp)
		return
	}
	list, total, err := h.collaborationWorkspaceService.List(&req)
	if err != nil {
		h.logger.Error("Collaboration workspace list failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取协作空间列表失败")
		c.JSON(status, resp)
		return
	}
	records := make([]gin.H, 0, len(list))
	for _, t := range list {
		adminUsers, _ := h.collaborationWorkspaceMemberRepo.GetAdminUsersByCollaborationWorkspaceID(t.ID)
		records = append(records, h.collaborationWorkspaceToMap(&t, adminUsers))
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"records": records,
		"total":   total,
		"current": req.Current,
		"size":    req.Size,
	}))
}

func (h *CollaborationWorkspaceHandler) ListOptions(c *gin.Context) {
	var req dto.CollaborationWorkspaceListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		status, resp := errcode.Response(errcode.ErrParamInvalid)
		c.JSON(status, resp)
		return
	}
	list, err := h.collaborationWorkspaceService.ListOptions(&req)
	if err != nil {
		h.logger.Error("Collaboration workspace options failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取协作空间候选失败")
		c.JSON(status, resp)
		return
	}
	records := make([]gin.H, 0, len(list))
	for _, t := range list {
		collaborationWorkspace := t
		records = append(records, h.collaborationWorkspaceToMap(&collaborationWorkspace, nil))
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"records": records,
		"total":   len(records),
	}))
}

func (h *CollaborationWorkspaceHandler) Get(c *gin.Context) {
	idStr := c.Param("id")
	if normalized := strings.TrimSpace(strings.ToLower(idStr)); normalized == "current" {
		h.GetCurrentCollaborationWorkspace(c)
		return
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的协作空间ID")
		c.JSON(status, resp)
		return
	}
	if err := h.requireTargetCollaborationWorkspace(c, id); err != nil {
		return
	}
	t, err := h.collaborationWorkspaceService.Get(id)
	if err != nil {
		if err == ErrCollaborationWorkspaceNotFound {
			status, resp := errcode.Response(errcode.ErrCollaborationWorkspaceNotFound)
			c.JSON(status, resp)
			return
		}
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取协作空间失败")
		c.JSON(status, resp)
		return
	}
	adminUsers, _ := h.collaborationWorkspaceMemberRepo.GetAdminUsersByCollaborationWorkspaceID(id)
	c.JSON(http.StatusOK, dto.SuccessResponse(h.collaborationWorkspaceToMap(t, adminUsers)))
}

func (h *CollaborationWorkspaceHandler) Create(c *gin.Context) {
	var req dto.CollaborationWorkspaceCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		status, resp := errcode.Response(errcode.ErrParamInvalid)
		c.JSON(status, resp)
		return
	}
	var ownerID *uuid.UUID
	if uid, ok := c.Get("user_id"); ok {
		if s, ok := uid.(string); ok {
			if parsed, err := uuid.Parse(s); err == nil {
				ownerID = &parsed
			}
		}
	}
	t, err := h.collaborationWorkspaceService.Create(&req, ownerID)
	if err != nil {
		h.logger.Error("Collaboration workspace create failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, err.Error())
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{"id": t.ID.String()}))
}

func (h *CollaborationWorkspaceHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的协作空间ID")
		c.JSON(status, resp)
		return
	}
	if err := h.requireTargetCollaborationWorkspace(c, id); err != nil {
		return
	}
	var req dto.CollaborationWorkspaceUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		status, resp := errcode.Response(errcode.ErrParamInvalid)
		c.JSON(status, resp)
		return
	}
	if err := h.collaborationWorkspaceService.Update(id, &req); err != nil {
		if err == ErrCollaborationWorkspaceNotFound {
			status, resp := errcode.Response(errcode.ErrCollaborationWorkspaceNotFound)
			c.JSON(status, resp)
			return
		}
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "更新协作空间失败")
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(nil))
}

func (h *CollaborationWorkspaceHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的协作空间ID")
		c.JSON(status, resp)
		return
	}
	if err := h.requireTargetCollaborationWorkspace(c, id); err != nil {
		return
	}
	if err := h.collaborationWorkspaceService.Delete(id); err != nil {
		if err == ErrCollaborationWorkspaceNotFound {
			status, resp := errcode.Response(errcode.ErrCollaborationWorkspaceNotFound)
			c.JSON(status, resp)
			return
		}
		h.logger.Error("Collaboration workspace delete failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "删除协作空间失败")
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(nil))
}

func (h *CollaborationWorkspaceHandler) ListMembers(c *gin.Context) {
	collaborationWorkspaceIDStr := c.Param("id")
	collaborationWorkspaceID, err := uuid.Parse(collaborationWorkspaceIDStr)
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的协作空间ID")
		c.JSON(status, resp)
		return
	}
	if err := h.requireTargetCollaborationWorkspace(c, collaborationWorkspaceID); err != nil {
		return
	}
	searchParams := &user.MemberSearchParams{}
	if v := c.Query("user_id"); v != "" {
		searchParams.UserID = v
	}
	if v := c.Query("user_name"); v != "" {
		searchParams.UserName = v
	}
	if v := c.Query("role_code"); v != "" {
		searchParams.RoleCode = v
	}
	members, err := h.collaborationWorkspaceService.ListMembers(collaborationWorkspaceID, searchParams)
	if err != nil {
		h.logger.Error("List members failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取成员列表失败")
		c.JSON(status, resp)
		return
	}
	records := make([]gin.H, 0, len(members))
	for _, m := range members {
		userInfo, _ := h.userRepo.GetByID(m.UserID)
		records = append(records, h.memberToMap(&m, userInfo))
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(records))
}

func (h *CollaborationWorkspaceHandler) AddMember(c *gin.Context) {
	collaborationWorkspaceIDStr := c.Param("id")
	collaborationWorkspaceID, err := uuid.Parse(collaborationWorkspaceIDStr)
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的协作空间ID")
		c.JSON(status, resp)
		return
	}
	if err := h.requireTargetCollaborationWorkspace(c, collaborationWorkspaceID); err != nil {
		return
	}
	var req dto.CollaborationWorkspaceAddMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		status, resp := errcode.Response(errcode.ErrParamInvalid)
		c.JSON(status, resp)
		return
	}
	var invitedBy *uuid.UUID
	if inviterID, ok := c.Get("user_id"); ok {
		if s, ok := inviterID.(string); ok {
			if parsed, err := uuid.Parse(s); err == nil {
				invitedBy = &parsed
			}
		}
	}
	if err := h.collaborationWorkspaceService.AddMember(collaborationWorkspaceID, &req, invitedBy); err != nil {
		if err == ErrCollaborationWorkspaceMemberExists {
			status, resp := errcode.Response(errcode.ErrCollaborationWorkspaceMemberExists)
			c.JSON(status, resp)
			return
		}
		h.logger.Error("Add member failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, err.Error())
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(nil))
}

func (h *CollaborationWorkspaceHandler) RemoveMember(c *gin.Context) {
	collaborationWorkspaceIDStr := c.Param("id")
	collaborationWorkspaceID, err := uuid.Parse(collaborationWorkspaceIDStr)
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的协作空间ID")
		c.JSON(status, resp)
		return
	}
	if err := h.requireTargetCollaborationWorkspace(c, collaborationWorkspaceID); err != nil {
		return
	}
	userIDStr := c.Param("userId")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的用户ID")
		c.JSON(status, resp)
		return
	}
	if err := h.collaborationWorkspaceService.RemoveMember(collaborationWorkspaceID, userID); err != nil {
		if err == ErrCollaborationWorkspaceMemberNotFound {
			status, resp := errcode.Response(errcode.ErrCollaborationWorkspaceMemberNotFound)
			c.JSON(status, resp)
			return
		}
		h.logger.Error("Remove member failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "移除成员失败")
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(nil))
}

func (h *CollaborationWorkspaceHandler) UpdateMemberRole(c *gin.Context) {
	collaborationWorkspaceIDStr := c.Param("id")
	collaborationWorkspaceID, err := uuid.Parse(collaborationWorkspaceIDStr)
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的协作空间ID")
		c.JSON(status, resp)
		return
	}
	if err := h.requireTargetCollaborationWorkspace(c, collaborationWorkspaceID); err != nil {
		return
	}
	userIDStr := c.Param("userId")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的用户ID")
		c.JSON(status, resp)
		return
	}
	var req dto.CollaborationWorkspaceUpdateMemberRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		status, resp := errcode.Response(errcode.ErrParamInvalid)
		c.JSON(status, resp)
		return
	}
	roleCode := strings.TrimSpace(req.RoleCode)
	if err := h.collaborationWorkspaceService.UpdateMemberRole(collaborationWorkspaceID, userID, roleCode); err != nil {
		if err == ErrCollaborationWorkspaceMemberNotFound {
			status, resp := errcode.Response(errcode.ErrCollaborationWorkspaceMemberNotFound)
			c.JSON(status, resp)
			return
		}
		h.logger.Error("Update member role failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, err.Error())
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(nil))
}

func (h *CollaborationWorkspaceHandler) GetCurrentCollaborationWorkspace(c *gin.Context) {
	member, err := h.resolveCollaborationWorkspaceMember(c)
	if err != nil {
		if err == gorm.ErrRecordNotFound || err == ErrCollaborationWorkspaceMemberNotFound {
			status, resp := errcode.Response(errcode.ErrNoCurrentCollaborationWorkspace)
			c.JSON(status, resp)
			return
		}
		h.logger.Error("Get current collaboration workspace failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取我的协作空间失败")
		c.JSON(status, resp)
		return
	}

	collaborationWorkspace, err := h.collaborationWorkspaceService.Get(member.CollaborationWorkspaceID)
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取协作空间信息失败")
		c.JSON(status, resp)
		return
	}
	adminUsers, _ := h.collaborationWorkspaceMemberRepo.GetAdminUsersByCollaborationWorkspaceID(collaborationWorkspace.ID)
	c.JSON(http.StatusOK, dto.SuccessResponse(h.collaborationWorkspaceToMap(collaborationWorkspace, adminUsers)))
}

func (h *CollaborationWorkspaceHandler) ListMyMembers(c *gin.Context) {
	member, err := h.resolveCollaborationWorkspaceMember(c)
	if err != nil {
		if err == gorm.ErrRecordNotFound || err == ErrCollaborationWorkspaceMemberNotFound {
			c.JSON(http.StatusOK, dto.SuccessResponse([]gin.H{}))
			return
		}
		h.logger.Error("Get current collaboration workspace member failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取我的协作空间成员失败")
		c.JSON(status, resp)
		return
	}

	searchParams := &user.MemberSearchParams{}
	if v := c.Query("user_id"); v != "" {
		searchParams.UserID = v
	}
	if v := c.Query("user_name"); v != "" {
		searchParams.UserName = v
	}
	if v := c.Query("role_code"); v != "" {
		searchParams.RoleCode = v
	}
	members, err := h.collaborationWorkspaceService.ListMembers(member.CollaborationWorkspaceID, searchParams)
	if err != nil {
		h.logger.Error("List my members failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取协作空间成员列表失败")
		c.JSON(status, resp)
		return
	}

	records := make([]gin.H, 0, len(members))
	for _, m := range members {
		userInfo, _ := h.userRepo.GetByID(m.UserID)
		records = append(records, h.memberToMap(&m, userInfo))
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(records))
}

func (h *CollaborationWorkspaceHandler) AddMyMember(c *gin.Context) {
	uid, err := h.mustUserID(c)
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的用户ID")
		c.JSON(status, resp)
		return
	}

	member, err := h.resolveCollaborationWorkspaceMember(c)
	if err != nil {
		if err == gorm.ErrRecordNotFound || err == ErrCollaborationWorkspaceMemberNotFound {
			status, resp := errcode.Response(errcode.ErrNoCurrentCollaborationWorkspace)
			c.JSON(status, resp)
			return
		}
		h.logger.Error("Get current collaboration workspace member failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取我的协作空间失败")
		c.JSON(status, resp)
		return
	}

	var req dto.CollaborationWorkspaceAddMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		status, resp := errcode.Response(errcode.ErrParamInvalid)
		c.JSON(status, resp)
		return
	}
	invitedBy := &uid
	if err := h.collaborationWorkspaceService.AddMember(member.CollaborationWorkspaceID, &req, invitedBy); err != nil {
		if err == ErrCollaborationWorkspaceMemberExists {
			status, resp := errcode.Response(errcode.ErrCollaborationWorkspaceMemberExists)
			c.JSON(status, resp)
			return
		}
		h.logger.Error("Add my member failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, err.Error())
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(nil))
}

func (h *CollaborationWorkspaceHandler) RemoveMyMember(c *gin.Context) {
	member, err := h.resolveCollaborationWorkspaceMember(c)
	if err != nil {
		if err == gorm.ErrRecordNotFound || err == ErrCollaborationWorkspaceMemberNotFound {
			status, resp := errcode.Response(errcode.ErrNoCurrentCollaborationWorkspace)
			c.JSON(status, resp)
			return
		}
		h.logger.Error("Get current collaboration workspace member failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取我的协作空间失败")
		c.JSON(status, resp)
		return
	}

	userIDStr := c.Param("userId")
	targetUserID, err := uuid.Parse(userIDStr)
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的用户ID")
		c.JSON(status, resp)
		return
	}

	if err := h.collaborationWorkspaceService.RemoveMember(member.CollaborationWorkspaceID, targetUserID); err != nil {
		if err == ErrCollaborationWorkspaceMemberNotFound {
			status, resp := errcode.Response(errcode.ErrCollaborationWorkspaceMemberNotFound)
			c.JSON(status, resp)
			return
		}
		h.logger.Error("Remove my member failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "移除成员失败")
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(nil))
}

func (h *CollaborationWorkspaceHandler) UpdateMyMemberRole(c *gin.Context) {
	member, err := h.resolveCollaborationWorkspaceMember(c)
	if err != nil {
		if err == gorm.ErrRecordNotFound || err == ErrCollaborationWorkspaceMemberNotFound {
			status, resp := errcode.Response(errcode.ErrNoCurrentCollaborationWorkspace)
			c.JSON(status, resp)
			return
		}
		h.logger.Error("Get current collaboration workspace member failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取我的协作空间失败")
		c.JSON(status, resp)
		return
	}

	targetUserIDStr := c.Param("userId")
	targetUserID, err := uuid.Parse(targetUserIDStr)
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的用户ID")
		c.JSON(status, resp)
		return
	}

	var req dto.CollaborationWorkspaceUpdateMemberRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		status, resp := errcode.Response(errcode.ErrParamInvalid)
		c.JSON(status, resp)
		return
	}
	roleCode := strings.TrimSpace(req.RoleCode)
	if err := h.collaborationWorkspaceService.UpdateMemberRole(member.CollaborationWorkspaceID, targetUserID, roleCode); err != nil {
		if err == ErrCollaborationWorkspaceMemberNotFound {
			status, resp := errcode.Response(errcode.ErrCollaborationWorkspaceMemberNotFound)
			c.JSON(status, resp)
			return
		}
		h.logger.Error("Update my member role failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, err.Error())
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(nil))
}

func (h *CollaborationWorkspaceHandler) GetMyCollaborationWorkspaceMemberRoles(c *gin.Context) {
	member, err := h.resolveCollaborationWorkspaceMember(c)
	if err != nil {
		if err == gorm.ErrRecordNotFound || err == ErrCollaborationWorkspaceMemberNotFound {
			c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
				"role_ids": []string{},
			}))
			return
		}
		h.logger.Error("Get current collaboration workspace member failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取我的协作空间失败")
		c.JSON(status, resp)
		return
	}

	targetUserIDStr := c.Param("userId")
	targetUserID, err := uuid.Parse(targetUserIDStr)
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的用户ID")
		c.JSON(status, resp)
		return
	}
	targetMember, err := h.collaborationWorkspaceMemberRepo.GetByUserAndCollaborationWorkspace(targetUserID, member.CollaborationWorkspaceID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			status, resp := errcode.Response(errcode.ErrCollaborationWorkspaceMemberNotFound)
			c.JSON(status, resp)
			return
		}
		h.logger.Error("Get collaboration workspace member failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取成员失败")
		c.JSON(status, resp)
		return
	}

	roleIDs, err := h.getWorkspaceAwareTeamRoleIDs(targetUserID, member.CollaborationWorkspaceID)
	if err != nil {
		h.logger.Error("Get user role ids failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取用户角色失败")
		c.JSON(status, resp)
		return
	}

	roles, err := h.roleRepo.GetByIDs(roleIDs)
	if err != nil {
		h.logger.Error("Get roles failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取角色失败")
		c.JSON(status, resp)
		return
	}

	roleList := make([]gin.H, 0, len(roles))
	for _, r := range roles {
		roleList = append(roleList, gin.H{"id": r.ID.String(), "code": r.Code, "name": r.Name})
	}
	roleIDsStr := make([]string, 0, len(roles))
	for _, r := range roles {
		roleIDsStr = append(roleIDsStr, r.ID.String())
	}
	bindingWorkspaceID, bindingWorkspaceType, memberType := h.resolveCollaborationWorkspaceRoleBindingMeta(member.CollaborationWorkspaceID, targetUserID, targetMember)
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"role_ids":               roleIDsStr,
		"roles":                  roleList,
		"binding_workspace_id":   bindingWorkspaceID,
		"binding_workspace_type": bindingWorkspaceType,
		"member_type":            memberType,
	}))
}

func (h *CollaborationWorkspaceHandler) SetMyCollaborationWorkspaceMemberRoles(c *gin.Context) {
	member, err := h.resolveCollaborationWorkspaceMember(c)
	if err != nil {
		if err == gorm.ErrRecordNotFound || err == ErrCollaborationWorkspaceMemberNotFound {
			status, resp := errcode.Response(errcode.ErrNoCurrentCollaborationWorkspace)
			c.JSON(status, resp)
			return
		}
		h.logger.Error("Get current collaboration workspace member failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取我的协作空间失败")
		c.JSON(status, resp)
		return
	}

	targetUserIDStr := c.Param("userId")
	targetUserID, err := uuid.Parse(targetUserIDStr)
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的用户ID")
		c.JSON(status, resp)
		return
	}

	var req dto.CollaborationWorkspaceSetMemberRolesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		status, resp := errcode.Response(errcode.ErrParamInvalid)
		c.JSON(status, resp)
		return
	}

	roleIDs := make([]uuid.UUID, 0, len(req.RoleIDs))
	for _, rid := range req.RoleIDs {
		if parsed, err := uuid.Parse(rid); err == nil {
			roleIDs = append(roleIDs, parsed)
		}
	}

	memberRecord, err := h.collaborationWorkspaceMemberRepo.GetByUserAndCollaborationWorkspace(targetUserID, member.CollaborationWorkspaceID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			status, resp := errcode.Response(errcode.ErrCollaborationWorkspaceMemberNotFound)
			c.JSON(status, resp)
			return
		}
		h.logger.Error("Get collaboration workspace member failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取成员失败")
		c.JSON(status, resp)
		return
	}

	allRoles, err := h.roleRepo.ListCollaborationWorkspaceRoles(member.CollaborationWorkspaceID)
	if err != nil {
		h.logger.Error("Get collaboration workspace roles failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取协作空间角色失败")
		c.JSON(status, resp)
		return
	}

	allowedTeamRoleIDs := make(map[uuid.UUID]user.Role)
	protectedRoleID := uuid.Nil
	for _, role := range allRoles {
		allowedTeamRoleIDs[role.ID] = role
		if role.Code == memberRecord.RoleCode {
			protectedRoleID = role.ID
		}
	}

	filteredRoleIDs := make([]uuid.UUID, 0, len(roleIDs)+1)
	seenRoleIDs := make(map[uuid.UUID]struct{}, len(roleIDs)+1)
	for _, roleID := range roleIDs {
		if _, ok := allowedTeamRoleIDs[roleID]; !ok {
			continue
		}
		if _, exists := seenRoleIDs[roleID]; exists {
			continue
		}
		seenRoleIDs[roleID] = struct{}{}
		filteredRoleIDs = append(filteredRoleIDs, roleID)
	}
	if protectedRoleID != uuid.Nil {
		if _, exists := seenRoleIDs[protectedRoleID]; !exists {
			filteredRoleIDs = append(filteredRoleIDs, protectedRoleID)
		}
	}

	if err := h.userRoleRepo.SetUserRoles(targetUserID, filteredRoleIDs, &member.CollaborationWorkspaceID); err != nil {
		h.logger.Error("Set user roles failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "设置角色失败")
		c.JSON(status, resp)
		return
	}
	if err := h.syncWorkspaceRoleBindings(member.CollaborationWorkspaceID, targetUserID, filteredRoleIDs); err != nil {
		h.logger.Error("Sync workspace role bindings failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "同步工作空间角色失败")
		c.JSON(status, resp)
		return
	}
	if h.refresher != nil {
		if err := h.refresher.RefreshCollaborationWorkspace(member.CollaborationWorkspaceID); err != nil {
			h.logger.Error("Refresh collaboration workspace after setting member roles failed", zap.Error(err))
			status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "刷新协作空间权限快照失败")
			c.JSON(status, resp)
			return
		}
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(nil))
}

func (h *CollaborationWorkspaceHandler) getWorkspaceAwareTeamRoleIDs(userID, collaborationWorkspaceID uuid.UUID) ([]uuid.UUID, error) {
	roleIDs, err := workspacerolebinding.ListCollaborationWorkspaceRoleIDsByUser(database.DB, collaborationWorkspaceID, userID, false)
	if err != nil {
		return nil, err
	}
	if len(roleIDs) > 0 {
		return roleIDs, nil
	}
	return h.userRoleRepo.GetRoleIDsByUserAndCollaborationWorkspace(userID, &collaborationWorkspaceID, h.collaborationWorkspaceMemberRepo)
}

func (h *CollaborationWorkspaceHandler) syncWorkspaceRoleBindings(collaborationWorkspaceID, userID uuid.UUID, roleIDs []uuid.UUID) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		return workspacerolebinding.ReplaceCollaborationWorkspaceRoleBindings(tx, collaborationWorkspaceID, userID, roleIDs)
	})
}

func (h *CollaborationWorkspaceHandler) ListCurrentCollaborationWorkspaceRoles(c *gin.Context) {
	member, err := h.resolveCollaborationWorkspaceMember(c)
	if err != nil {
		if err == gorm.ErrRecordNotFound || err == ErrCollaborationWorkspaceMemberNotFound {
			c.JSON(http.StatusOK, dto.SuccessResponse([]gin.H{}))
			return
		}
		h.logger.Error("Get current collaboration workspace member failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取我的协作空间失败")
		c.JSON(status, resp)
		return
	}

	_ = member.CollaborationWorkspaceID
	allRoles, err := h.roleRepo.ListCollaborationWorkspaceRoles(member.CollaborationWorkspaceID)
	if err != nil {
		h.logger.Error("List roles failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取角色列表失败")
		c.JSON(status, resp)
		return
	}

	roleList := make([]gin.H, 0, len(allRoles))
	for _, r := range allRoles {
		roleList = append(roleList, gin.H{
			"id":                         r.ID.String(),
			"code":                       r.Code,
			"name":                       r.Name,
			"description":                r.Description,
			"status":                     r.Status,
			"is_system":                  r.IsSystem,
			"collaboration_workspace_id": uuidPtrToString(r.CollaborationWorkspaceID),
			"is_global":                  r.CollaborationWorkspaceID == nil,
			"create_time":                r.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(roleList))
}

func (h *CollaborationWorkspaceHandler) ListCollaborationWorkspaceRoles(c *gin.Context) {
	collaborationWorkspaceIDStr := c.Param("id")
	collaborationWorkspaceID, err := uuid.Parse(collaborationWorkspaceIDStr)
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的协作空间ID")
		c.JSON(status, resp)
		return
	}
	if err := h.requireTargetCollaborationWorkspace(c, collaborationWorkspaceID); err != nil {
		return
	}

	if _, err := h.collaborationWorkspaceService.Get(collaborationWorkspaceID); err != nil {
		if err == ErrCollaborationWorkspaceNotFound {
			status, resp := errcode.Response(errcode.ErrCollaborationWorkspaceNotFound)
			c.JSON(status, resp)
			return
		}
		h.logger.Error("Get collaboration workspace failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取协作空间失败")
		c.JSON(status, resp)
		return
	}

	allRoles, err := h.roleRepo.ListCollaborationWorkspaceRoles(collaborationWorkspaceID)
	if err != nil {
		h.logger.Error("List collaboration workspace roles failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取协作空间角色失败")
		c.JSON(status, resp)
		return
	}

	roleList := make([]gin.H, 0, len(allRoles))
	for _, r := range allRoles {
		roleList = append(roleList, gin.H{
			"id":                         r.ID.String(),
			"code":                       r.Code,
			"name":                       r.Name,
			"description":                r.Description,
			"status":                     r.Status,
			"is_system":                  r.IsSystem,
			"collaboration_workspace_id": uuidPtrToString(r.CollaborationWorkspaceID),
			"is_global":                  r.CollaborationWorkspaceID == nil,
			"create_time":                r.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(roleList))
}

func (h *CollaborationWorkspaceHandler) CreateCurrentCollaborationWorkspaceRole(c *gin.Context) {
	member, err := h.resolveCollaborationWorkspaceMember(c)
	if err != nil {
		if err == gorm.ErrRecordNotFound || err == ErrCollaborationWorkspaceMemberNotFound {
			status, resp := errcode.Response(errcode.ErrNoCurrentCollaborationWorkspace)
			c.JSON(status, resp)
			return
		}
		h.logger.Error("Get current collaboration workspace member failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取我的协作空间失败")
		c.JSON(status, resp)
		return
	}

	var req dto.RoleCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		status, resp := errcode.Response(errcode.ErrParamInvalid)
		c.JSON(status, resp)
		return
	}

	code := strings.TrimSpace(req.Code)
	if code == "" || strings.TrimSpace(req.Name) == "" {
		status, resp := errcode.ResponseWithMsg(errcode.ErrParamInvalid, "角色编码和角色名称不能为空")
		c.JSON(status, resp)
		return
	}
	existingRoles, err := h.roleRepo.FindByCode(code)
	if err != nil {
		h.logger.Error("Find collaboration workspace role code failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "校验角色编码失败")
		c.JSON(status, resp)
		return
	}
	for _, existing := range existingRoles {
		if existing.CollaborationWorkspaceID != nil && *existing.CollaborationWorkspaceID == member.CollaborationWorkspaceID {
			status, resp := errcode.Response(errcode.ErrRoleCodeExists)
			c.JSON(status, resp)
			return
		}
	}

	role := &user.Role{
		CollaborationWorkspaceID: &member.CollaborationWorkspaceID,
		Code:                     code,
		Name:                     strings.TrimSpace(req.Name),
		Description:              strings.TrimSpace(req.Description),
		SortOrder:                req.SortOrder,
		Priority:                 req.Priority,
		Status:                   normalizeRoleStatus(req.Status),
	}
	if err := h.roleRepo.Create(role); err != nil {
		h.logger.Error("Create collaboration workspace role failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "创建协作空间角色失败")
		c.JSON(status, resp)
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{"roleId": role.ID.String()}))
}

func (h *CollaborationWorkspaceHandler) UpdateCurrentCollaborationWorkspaceRole(c *gin.Context) {
	member, role, err := h.resolveEditableCurrentCollaborationWorkspaceRole(c)
	if err != nil {
		h.respondCurrentCollaborationWorkspaceRoleError(c, err)
		return
	}

	var req dto.RoleUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		status, resp := errcode.Response(errcode.ErrParamInvalid)
		c.JSON(status, resp)
		return
	}

	updates := map[string]interface{}{
		"name":        strings.TrimSpace(defaultString(req.Name, role.Name)),
		"description": strings.TrimSpace(defaultString(req.Description, role.Description)),
		"sort_order":  req.SortOrder,
		"priority":    req.Priority,
		"status":      normalizeRoleStatus(defaultString(req.Status, role.Status)),
	}
	if code := strings.TrimSpace(req.Code); code != "" && code != role.Code {
		existingRoles, findErr := h.roleRepo.FindByCode(code)
		if findErr != nil {
			h.logger.Error("Find collaboration workspace role code failed", zap.Error(findErr))
			status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "校验角色编码失败")
			c.JSON(status, resp)
			return
		}
		for _, existing := range existingRoles {
			if existing.ID == role.ID {
				continue
			}
			if existing.CollaborationWorkspaceID != nil && *existing.CollaborationWorkspaceID == member.CollaborationWorkspaceID {
				status, resp := errcode.Response(errcode.ErrRoleCodeExists)
				c.JSON(status, resp)
				return
			}
		}
		updates["code"] = code
	}

	if err := h.roleRepo.UpdateWithMap(role.ID, updates); err != nil {
		h.logger.Error("Update collaboration workspace role failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "更新协作空间角色失败")
		c.JSON(status, resp)
		return
	}
	if h.refresher != nil {
		if err := h.refresher.RefreshCollaborationWorkspace(member.CollaborationWorkspaceID); err != nil {
			h.logger.Error("Refresh collaboration workspace after updating role failed", zap.Error(err))
			status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "刷新协作空间权限快照失败")
			c.JSON(status, resp)
			return
		}
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(nil))
}

func (h *CollaborationWorkspaceHandler) DeleteCurrentCollaborationWorkspaceRole(c *gin.Context) {
	member, role, err := h.resolveEditableCurrentCollaborationWorkspaceRole(c)
	if err != nil {
		h.respondCurrentCollaborationWorkspaceRoleError(c, err)
		return
	}

	if err := database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("role_id = ? AND collaboration_workspace_id = ?", role.ID, member.CollaborationWorkspaceID).Delete(&user.UserRole{}).Error; err != nil {
			return err
		}
		if err := tx.Where("role_id = ?", role.ID).Delete(&user.RoleFeaturePackage{}).Error; err != nil {
			return err
		}
		if err := tx.Where("role_id = ?", role.ID).Delete(&user.RoleHiddenMenu{}).Error; err != nil {
			return err
		}
		if err := tx.Where("role_id = ?", role.ID).Delete(&user.RoleDisabledAction{}).Error; err != nil {
			return err
		}
		if err := tx.Where("role_id = ?", role.ID).Delete(&user.RoleDataPermission{}).Error; err != nil {
			return err
		}
		if err := tx.Where("collaboration_workspace_id = ? AND role_id = ?", member.CollaborationWorkspaceID, role.ID).Delete(&models.CollaborationWorkspaceRoleAccessSnapshot{}).Error; err != nil {
			return err
		}
		return tx.Delete(&user.Role{}, role.ID).Error
	}); err != nil {
		h.logger.Error("Delete collaboration workspace role failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "删除协作空间角色失败")
		c.JSON(status, resp)
		return
	}
	if h.refresher != nil {
		if err := h.refresher.RefreshCollaborationWorkspace(member.CollaborationWorkspaceID); err != nil {
			h.logger.Error("Refresh collaboration workspace after deleting role failed", zap.Error(err))
			status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "刷新协作空间权限快照失败")
			c.JSON(status, resp)
			return
		}
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(nil))
}

func (h *CollaborationWorkspaceHandler) GetCurrentCollaborationWorkspaceRolePackages(c *gin.Context) {
	member, role, err := h.resolveCurrentCollaborationWorkspaceRole(c)
	if err != nil {
		h.respondCurrentCollaborationWorkspaceRoleError(c, err)
		return
	}

	appKey, err := appctx.RequireRequestAppKey(c)
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrParamInvalid, "app_key 为必填项")
		c.JSON(status, resp)
		return
	}
	snapshot, err := h.getRoleSnapshot(member.CollaborationWorkspaceID, role, appKey)
	if err != nil {
		h.logger.Error("Get collaboration workspace role packages failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取协作空间角色功能包失败")
		c.JSON(status, resp)
		return
	}
	packages, err := h.featurePkgRepo.GetByIDs(snapshot.PackageIDs)
	if err != nil {
		h.logger.Error("Get role feature packages failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取功能包详情失败")
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"package_ids": actionIDsToStrings(snapshot.PackageIDs),
		"packages":    featurePackageListToMaps(packages),
		"inherited":   snapshot.Inherited,
	}))
}

func (h *CollaborationWorkspaceHandler) SetCurrentCollaborationWorkspaceRolePackages(c *gin.Context) {
	member, role, err := h.resolveEditableCurrentCollaborationWorkspaceRole(c)
	if err != nil {
		h.respondCurrentCollaborationWorkspaceRoleError(c, err)
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

	packageIDs, parseErr := parseUUIDSlice(req.PackageIDs)
	if parseErr != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的功能包ID")
		c.JSON(status, resp)
		return
	}

	collaborationWorkspaceFeaturePackageIDs, err := h.collaborationWorkspaceFeaturePackageRepo.GetPackageIDsByCollaborationWorkspaceID(member.CollaborationWorkspaceID)
	if err != nil {
		h.logger.Error("Get collaboration workspace feature packages failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取协作空间功能包失败")
		c.JSON(status, resp)
		return
	}
	allowedPackageSet := make(map[uuid.UUID]struct{}, len(collaborationWorkspaceFeaturePackageIDs))
	for _, packageID := range collaborationWorkspaceFeaturePackageIDs {
		allowedPackageSet[packageID] = struct{}{}
	}
	if len(packageIDs) > 0 {
		packages, getErr := h.featurePkgRepo.GetByIDs(packageIDs)
		if getErr != nil {
			h.logger.Error("Get role package detail failed", zap.Error(getErr))
			status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取功能包失败")
			c.JSON(status, resp)
			return
		}
		if len(packages) != len(packageIDs) {
			status, resp := errcode.ResponseWithMsg(errcode.ErrNotFound, "存在无效的功能包")
			c.JSON(status, resp)
			return
		}
		for _, item := range packages {
			if appctx.NormalizeAppKey(item.AppKey) != appctx.NormalizeAppKey(appKey) {
				status, resp := errcode.ResponseWithMsg(errcode.ErrForbidden, "仅支持绑定当前应用内的功能包")
				c.JSON(status, resp)
				return
			}
			if item.ContextType != "" && item.ContextType != "collaboration" && item.ContextType != "common" {
				status, resp := errcode.ResponseWithMsg(errcode.ErrForbidden, "仅支持绑定协作空间上下文可用的功能包")
				c.JSON(status, resp)
				return
			}
		}
	}
	for _, packageID := range packageIDs {
		if _, ok := allowedPackageSet[packageID]; !ok {
			status, resp := errcode.ResponseWithMsg(errcode.ErrForbidden, "存在未向当前协作空间开通的功能包")
			c.JSON(status, resp)
			return
		}
	}

	userID, _ := h.mustUserID(c)
	if err := appscope.ReplaceRolePackagesInApp(database.DB, role.ID, appKey, packageIDs, &userID); err != nil {
		h.logger.Error("Set collaboration workspace role packages failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "保存协作空间角色功能包失败")
		c.JSON(status, resp)
		return
	}
	if h.refresher != nil {
		if err := h.refresher.RefreshCollaborationWorkspace(member.CollaborationWorkspaceID); err != nil {
			h.logger.Error("Refresh collaboration workspace after setting role packages failed", zap.Error(err))
			status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "刷新协作空间权限快照失败")
			c.JSON(status, resp)
			return
		}
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(nil))
}

func (h *CollaborationWorkspaceHandler) GetCurrentCollaborationWorkspaceRoleMenus(c *gin.Context) {
	member, role, err := h.resolveCurrentCollaborationWorkspaceRole(c)
	if err != nil {
		h.respondCurrentCollaborationWorkspaceRoleError(c, err)
		return
	}

	appKey, err := appctx.RequireRequestAppKey(c)
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrParamInvalid, "app_key 为必填项")
		c.JSON(status, resp)
		return
	}
	snapshot, err := h.getRoleSnapshot(member.CollaborationWorkspaceID, role, appKey)
	if err != nil {
		h.logger.Error("Get role menu boundary failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取协作空间角色菜单范围失败")
		c.JSON(status, resp)
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"menu_ids":             actionIDsToStrings(snapshot.MenuIDs),
		"available_menu_ids":   actionIDsToStrings(snapshot.AvailableMenuIDs),
		"hidden_menu_ids":      actionIDsToStrings(snapshot.HiddenMenuIDs),
		"expanded_package_ids": actionIDsToStrings(snapshot.ExpandedPackageIDs),
		"derived_sources":      buildMenuSourceMaps(snapshot.MenuSourceMap),
	}))
}

func (h *CollaborationWorkspaceHandler) SetCurrentCollaborationWorkspaceRoleMenus(c *gin.Context) {
	member, role, err := h.resolveEditableCurrentCollaborationWorkspaceRole(c)
	if err != nil {
		h.respondCurrentCollaborationWorkspaceRoleError(c, err)
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

	menuIDs, parseErr := parseUUIDSlice(req.MenuIDs)
	if parseErr != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的菜单ID")
		c.JSON(status, resp)
		return
	}

	snapshot, err := h.getRoleSnapshot(member.CollaborationWorkspaceID, role, appKey)
	if err != nil {
		h.logger.Error("Get role menu boundary failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取协作空间角色菜单范围失败")
		c.JSON(status, resp)
		return
	}
	enabledMenuSet := uuidSliceToSet(snapshot.AvailableMenuIDs)
	for _, menuID := range menuIDs {
		if !enabledMenuSet[menuID] {
			status, resp := errcode.ResponseWithMsg(errcode.ErrForbidden, "存在超出当前角色已绑定功能包范围的菜单")
			c.JSON(status, resp)
			return
		}
	}

	hiddenMenuIDs := excludeActionIDs(snapshot.AvailableMenuIDs, menuIDs)
	if err := appscope.ReplaceRoleHiddenMenusInApp(database.DB, role.ID, appKey, hiddenMenuIDs); err != nil {
		h.logger.Error("Set collaboration workspace role hidden menus failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "保存协作空间角色菜单权限失败")
		c.JSON(status, resp)
		return
	}
	if h.refresher != nil {
		if err := h.refresher.RefreshCollaborationWorkspace(member.CollaborationWorkspaceID); err != nil {
			h.logger.Error("Refresh collaboration workspace after setting role menus failed", zap.Error(err))
			status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "刷新协作空间权限快照失败")
			c.JSON(status, resp)
			return
		}
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(nil))
}

func (h *CollaborationWorkspaceHandler) GetCurrentCollaborationWorkspaceRoleActions(c *gin.Context) {
	member, role, err := h.resolveCurrentCollaborationWorkspaceRole(c)
	if err != nil {
		h.respondCurrentCollaborationWorkspaceRoleError(c, err)
		return
	}

	appKey, err := appctx.RequireRequestAppKey(c)
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrParamInvalid, "app_key 为必填项")
		c.JSON(status, resp)
		return
	}
	snapshot, err := h.getRoleSnapshot(member.CollaborationWorkspaceID, role, appKey)
	if err != nil {
		h.logger.Error("Get role action boundary failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取协作空间角色功能范围失败")
		c.JSON(status, resp)
		return
	}
	actions, err := h.actionRepo.GetByIDs(snapshot.AvailableActionIDs)
	if err != nil {
		h.logger.Error("Get role boundary actions failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取协作空间角色功能范围失败")
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"action_ids":           actionIDsToStrings(snapshot.ActionIDs),
		"available_action_ids": actionIDsToStrings(snapshot.AvailableActionIDs),
		"disabled_action_ids":  actionIDsToStrings(snapshot.DisabledActionIDs),
		"actions":              actionListToMaps(actions),
		"expanded_package_ids": actionIDsToStrings(snapshot.ExpandedPackageIDs),
		"derived_sources":      buildDerivedSourceMaps(snapshot.ActionSourceMap),
	}))
}

func (h *CollaborationWorkspaceHandler) SetCurrentCollaborationWorkspaceRoleActions(c *gin.Context) {
	member, role, err := h.resolveEditableCurrentCollaborationWorkspaceRole(c)
	if err != nil {
		h.respondCurrentCollaborationWorkspaceRoleError(c, err)
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

	actionIDs, parseErr := parseUUIDSlice(req.KeyIDs)
	if parseErr != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的功能权限ID")
		c.JSON(status, resp)
		return
	}
	if len(actionIDs) > 0 {
		actionList, actionErr := h.actionRepo.GetByIDs(actionIDs)
		if actionErr != nil {
			h.logger.Error("Get role action detail failed", zap.Error(actionErr))
			status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取功能权限失败")
			c.JSON(status, resp)
			return
		}
		if len(actionList) != len(actionIDs) {
			status, resp := errcode.ResponseWithMsg(errcode.ErrNotFound, "存在无效的功能权限")
			c.JSON(status, resp)
			return
		}
	}

	snapshot, err := h.getRoleSnapshot(member.CollaborationWorkspaceID, role, appKey)
	if err != nil {
		h.logger.Error("Get role action boundary failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取协作空间角色功能范围失败")
		c.JSON(status, resp)
		return
	}
	enabledActionSet := uuidSliceToSet(snapshot.AvailableActionIDs)
	for _, actionID := range actionIDs {
		if !enabledActionSet[actionID] {
			status, resp := errcode.ResponseWithMsg(errcode.ErrForbidden, "存在超出当前角色已绑定功能包范围的功能权限")
			c.JSON(status, resp)
			return
		}
	}

	disabledActionIDs := excludeActionIDs(snapshot.AvailableActionIDs, actionIDs)
	if err := appscope.ReplaceRoleDisabledActionsInScope(database.DB, role.ID, snapshot.AvailableActionIDs, disabledActionIDs); err != nil {
		h.logger.Error("Set collaboration workspace role disabled actions failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "保存协作空间角色功能权限失败")
		c.JSON(status, resp)
		return
	}
	if h.refresher != nil {
		if err := h.refresher.RefreshCollaborationWorkspace(member.CollaborationWorkspaceID); err != nil {
			h.logger.Error("Refresh collaboration workspace after setting role actions failed", zap.Error(err))
			status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "刷新协作空间权限快照失败")
			c.JSON(status, resp)
			return
		}
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(nil))
}

func (h *CollaborationWorkspaceHandler) GetCollaborationWorkspaceActions(c *gin.Context) {
	collaborationWorkspaceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的协作空间ID")
		c.JSON(status, resp)
		return
	}
	if err := h.requireTargetCollaborationWorkspace(c, collaborationWorkspaceID); err != nil {
		return
	}
	appKey, err := appctx.RequireRequestAppKey(c)
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrParamInvalid, "app_key 为必填项")
		c.JSON(status, resp)
		return
	}
	snapshot, err := h.boundaryService.GetSnapshot(collaborationWorkspaceID, appKey)
	if err != nil {
		h.logger.Error("Get collaboration workspace actions failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取协作空间功能权限失败")
		c.JSON(status, resp)
		return
	}
	actions, err := h.actionRepo.GetByIDs(snapshot.EffectiveIDs)
	if err != nil {
		h.logger.Error("Get permission actions failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取功能权限失败")
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"action_ids": actionIDsToStrings(snapshot.EffectiveIDs),
		"actions":    actionListToMaps(actions),
	}))
}

func (h *CollaborationWorkspaceHandler) GetCollaborationWorkspaceMenus(c *gin.Context) {
	collaborationWorkspaceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的协作空间ID")
		c.JSON(status, resp)
		return
	}
	if err := h.requireTargetCollaborationWorkspace(c, collaborationWorkspaceID); err != nil {
		return
	}
	appKey, err := appctx.RequireRequestAppKey(c)
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrParamInvalid, "app_key 为必填项")
		c.JSON(status, resp)
		return
	}
	snapshot, err := h.boundaryService.GetMenuSnapshot(collaborationWorkspaceID, appKey)
	if err != nil {
		h.logger.Error("Get collaboration workspace menus failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取协作空间菜单边界失败")
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"menu_ids": actionIDsToStrings(snapshot.EffectiveIDs),
	}))
}

func (h *CollaborationWorkspaceHandler) GetCollaborationWorkspaceActionOrigins(c *gin.Context) {
	collaborationWorkspaceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的协作空间ID")
		c.JSON(status, resp)
		return
	}
	if err := h.requireTargetCollaborationWorkspace(c, collaborationWorkspaceID); err != nil {
		return
	}
	h.respondCollaborationWorkspaceActionOrigins(c, collaborationWorkspaceID)
}

func (h *CollaborationWorkspaceHandler) GetCollaborationWorkspaceMenuOrigins(c *gin.Context) {
	collaborationWorkspaceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的协作空间ID")
		c.JSON(status, resp)
		return
	}
	if err := h.requireTargetCollaborationWorkspace(c, collaborationWorkspaceID); err != nil {
		return
	}
	h.respondCollaborationWorkspaceMenuOrigins(c, collaborationWorkspaceID)
}

func (h *CollaborationWorkspaceHandler) SetCollaborationWorkspaceActions(c *gin.Context) {
	collaborationWorkspaceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的协作空间ID")
		c.JSON(status, resp)
		return
	}
	if err := h.requireTargetCollaborationWorkspace(c, collaborationWorkspaceID); err != nil {
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
	actionIDs, parseErr := parseUUIDSlice(req.KeyIDs)
	if parseErr != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的功能权限ID")
		c.JSON(status, resp)
		return
	}
	if len(actionIDs) > 0 {
		actions, actionErr := h.actionRepo.GetByIDs(actionIDs)
		if actionErr != nil {
			h.logger.Error("Get collaboration workspace action detail failed", zap.Error(actionErr))
			status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取功能权限失败")
			c.JSON(status, resp)
			return
		}
		if len(actions) != len(actionIDs) {
			status, resp := errcode.ResponseWithMsg(errcode.ErrNotFound, "存在无效的功能权限")
			c.JSON(status, resp)
			return
		}
	}
	snapshot, err := h.boundaryService.GetSnapshot(collaborationWorkspaceID, appKey)
	if err != nil {
		h.logger.Error("Get derived collaboration workspace actions failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取功能包展开权限失败")
		c.JSON(status, resp)
		return
	}
	blockedIDs := excludeActionIDs(snapshot.DerivedIDs, actionIDs)
	if err := appscope.ReplaceCollaborationWorkspaceBlockedActionsInScope(database.DB, collaborationWorkspaceID, snapshot.DerivedIDs, blockedIDs); err != nil {
		h.logger.Error("Set collaboration workspace blocked actions failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "保存协作空间功能边界失败")
		c.JSON(status, resp)
		return
	}
	if h.refresher != nil {
		if err := h.refresher.RefreshCollaborationWorkspace(collaborationWorkspaceID); err != nil {
			h.logger.Error("Set collaboration workspace actions failed", zap.Error(err))
			status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "保存协作空间功能权限失败")
			c.JSON(status, resp)
			return
		}
	} else if _, err := h.boundaryService.RefreshSnapshot(collaborationWorkspaceID, appKey); err != nil {
		h.logger.Error("Set collaboration workspace actions failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "保存协作空间功能权限失败")
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(nil))
}

func (h *CollaborationWorkspaceHandler) SetCollaborationWorkspaceMenus(c *gin.Context) {
	collaborationWorkspaceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的协作空间ID")
		c.JSON(status, resp)
		return
	}
	if err := h.requireTargetCollaborationWorkspace(c, collaborationWorkspaceID); err != nil {
		return
	}
	var req dto.CollaborationWorkspaceMenuPermissionsRequest
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
	menuIDs, parseErr := parseUUIDSlice(req.MenuIDs)
	if parseErr != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的菜单ID")
		c.JSON(status, resp)
		return
	}
	snapshot, err := h.boundaryService.GetMenuSnapshot(collaborationWorkspaceID, appKey)
	if err != nil {
		h.logger.Error("Get derived collaboration workspace menus failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取功能包展开菜单失败")
		c.JSON(status, resp)
		return
	}
	blockedIDs := excludeActionIDs(snapshot.DerivedIDs, menuIDs)
	if err := appscope.ReplaceCollaborationWorkspaceBlockedMenusInApp(database.DB, collaborationWorkspaceID, appKey, blockedIDs); err != nil {
		h.logger.Error("Set collaboration workspace blocked menus failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "保存协作空间菜单边界失败")
		c.JSON(status, resp)
		return
	}
	if h.refresher != nil {
		if err := h.refresher.RefreshCollaborationWorkspace(collaborationWorkspaceID); err != nil {
			h.logger.Error("Set collaboration workspace menus failed", zap.Error(err))
			status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "保存协作空间菜单边界失败")
			c.JSON(status, resp)
			return
		}
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(nil))
}

func (h *CollaborationWorkspaceHandler) GetCurrentCollaborationWorkspaceBoundaryPackages(c *gin.Context) {
	member, err := h.resolveCollaborationWorkspaceMember(c)
	if err != nil {
		if err == gorm.ErrRecordNotFound || err == ErrCollaborationWorkspaceMemberNotFound {
			status, resp := errcode.Response(errcode.ErrNoCurrentCollaborationWorkspace)
			c.JSON(status, resp)
			return
		}
		h.logger.Error("Resolve current collaboration workspace failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取我的协作空间失败")
		c.JSON(status, resp)
		return
	}

	appKey, err := appctx.RequireRequestAppKey(c)
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrParamInvalid, "app_key 为必填项")
		c.JSON(status, resp)
		return
	}
	packageIDs, err := appscope.PackageIDsByCollaborationWorkspace(database.DB, member.CollaborationWorkspaceID, appKey)
	if err != nil {
		h.logger.Error("Get current collaboration workspace packages failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取协作空间功能包失败")
		c.JSON(status, resp)
		return
	}
	if len(packageIDs) == 0 {
		c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
			"package_ids": []string{},
			"packages":    []gin.H{},
		}))
		return
	}

	packages, err := h.featurePkgRepo.GetByIDs(packageIDs)
	if err != nil {
		h.logger.Error("Get current collaboration workspace package details failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取功能包详情失败")
		c.JSON(status, resp)
		return
	}

	filtered := make([]user.FeaturePackage, 0, len(packages))
	for _, item := range packages {
		if strings.TrimSpace(item.Status) != "" && item.Status != "normal" {
			continue
		}
		if item.ContextType != "" && item.ContextType != "collaboration" && item.ContextType != "common" {
			continue
		}
		filtered = append(filtered, item)
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"package_ids": actionIDsToStrings(packageIDs),
		"packages":    featurePackageListToMaps(filtered),
	}))
}

func (h *CollaborationWorkspaceHandler) GetCurrentCollaborationWorkspaceActions(c *gin.Context) {
	member, err := h.resolveCollaborationWorkspaceMember(c)
	if err != nil {
		if err == gorm.ErrRecordNotFound || err == ErrCollaborationWorkspaceMemberNotFound {
			status, resp := errcode.Response(errcode.ErrNoCurrentCollaborationWorkspace)
			c.JSON(status, resp)
			return
		}
		h.logger.Error("Resolve current collaboration workspace failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取我的协作空间失败")
		c.JSON(status, resp)
		return
	}
	appKey, err := appctx.RequireRequestAppKey(c)
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrParamInvalid, "app_key 为必填项")
		c.JSON(status, resp)
		return
	}
	snapshot, err := h.boundaryService.GetSnapshot(member.CollaborationWorkspaceID, appKey)
	if err != nil {
		h.logger.Error("Get current collaboration workspace actions failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取协作空间功能权限失败")
		c.JSON(status, resp)
		return
	}
	actions, err := h.actionRepo.GetByIDs(snapshot.EffectiveIDs)
	if err != nil {
		h.logger.Error("Get current collaboration workspace permission actions failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取功能权限失败")
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"action_ids": actionIDsToStrings(snapshot.EffectiveIDs),
		"actions":    actionListToMaps(actions),
	}))
}

func (h *CollaborationWorkspaceHandler) GetCurrentCollaborationWorkspaceMenus(c *gin.Context) {
	member, err := h.resolveCollaborationWorkspaceMember(c)
	if err != nil {
		if err == gorm.ErrRecordNotFound || err == ErrCollaborationWorkspaceMemberNotFound {
			status, resp := errcode.Response(errcode.ErrNoCurrentCollaborationWorkspace)
			c.JSON(status, resp)
			return
		}
		h.logger.Error("Resolve current collaboration workspace failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取我的协作空间失败")
		c.JSON(status, resp)
		return
	}
	appKey, err := appctx.RequireRequestAppKey(c)
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrParamInvalid, "app_key 为必填项")
		c.JSON(status, resp)
		return
	}
	snapshot, err := h.boundaryService.GetMenuSnapshot(member.CollaborationWorkspaceID, appKey)
	if err != nil {
		h.logger.Error("Get current collaboration workspace menus failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取协作空间菜单边界失败")
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"menu_ids": actionIDsToStrings(snapshot.EffectiveIDs),
	}))
}

func (h *CollaborationWorkspaceHandler) GetCurrentCollaborationWorkspaceActionOrigins(c *gin.Context) {
	member, err := h.resolveCollaborationWorkspaceMember(c)
	if err != nil {
		if err == gorm.ErrRecordNotFound || err == ErrCollaborationWorkspaceMemberNotFound {
			c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
				"derived_action_ids": []string{},
				"derived_sources":    []gin.H{},
				"blocked_action_ids": []string{},
			}))
			return
		}
		h.logger.Error("Resolve current collaboration workspace failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取我的协作空间失败")
		c.JSON(status, resp)
		return
	}
	h.respondCollaborationWorkspaceActionOrigins(c, member.CollaborationWorkspaceID)
}

func (h *CollaborationWorkspaceHandler) GetCurrentCollaborationWorkspaceMenuOrigins(c *gin.Context) {
	member, err := h.resolveCollaborationWorkspaceMember(c)
	if err != nil {
		if err == gorm.ErrRecordNotFound || err == ErrCollaborationWorkspaceMemberNotFound {
			c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
				"derived_menu_ids": []string{},
				"derived_sources":  []gin.H{},
				"blocked_menu_ids": []string{},
			}))
			return
		}
		h.logger.Error("Resolve current collaboration workspace failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取我的协作空间失败")
		c.JSON(status, resp)
		return
	}
	h.respondCollaborationWorkspaceMenuOrigins(c, member.CollaborationWorkspaceID)
}

func (h *CollaborationWorkspaceHandler) respondCollaborationWorkspaceActionOrigins(c *gin.Context, collaborationWorkspaceID uuid.UUID) {
	snapshot, err := h.boundaryService.GetSnapshot(collaborationWorkspaceID)
	if err != nil {
		h.logger.Error("Get collaboration workspace action origins failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取协作空间功能权限来源失败")
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"derived_action_ids": actionIDsToStrings(snapshot.DerivedIDs),
		"derived_sources":    buildDerivedSourceMaps(snapshot.DerivedMap),
		"blocked_action_ids": actionIDsToStrings(snapshot.BlockedIDs),
	}))
}

func (h *CollaborationWorkspaceHandler) respondCollaborationWorkspaceMenuOrigins(c *gin.Context, collaborationWorkspaceID uuid.UUID) {
	snapshot, err := h.boundaryService.GetMenuSnapshot(collaborationWorkspaceID)
	if err != nil {
		h.logger.Error("Get collaboration workspace menu origins failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取协作空间菜单来源失败")
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"derived_menu_ids": actionIDsToStrings(snapshot.DerivedIDs),
		"derived_sources":  buildMenuSourceMaps(snapshot.DerivedMap),
		"blocked_menu_ids": actionIDsToStrings(snapshot.BlockedIDs),
	}))
}

func excludeActionIDs(source []uuid.UUID, excluded []uuid.UUID) []uuid.UUID {
	excludedSet := make(map[uuid.UUID]struct{}, len(excluded))
	for _, actionID := range excluded {
		excludedSet[actionID] = struct{}{}
	}
	result := make([]uuid.UUID, 0, len(source))
	seen := make(map[uuid.UUID]struct{}, len(source))
	for _, actionID := range source {
		if _, skip := excludedSet[actionID]; skip {
			continue
		}
		if _, ok := seen[actionID]; ok {
			continue
		}
		seen[actionID] = struct{}{}
		result = append(result, actionID)
	}
	return result
}

func mergeActionIDs(groups ...[]uuid.UUID) []uuid.UUID {
	result := make([]uuid.UUID, 0)
	seen := make(map[uuid.UUID]struct{})
	for _, group := range groups {
		for _, actionID := range group {
			if _, ok := seen[actionID]; ok {
				continue
			}
			seen[actionID] = struct{}{}
			result = append(result, actionID)
		}
	}
	return result
}

func (h *CollaborationWorkspaceHandler) getCollaborationWorkspaceEnabledActionSet(collaborationWorkspaceID uuid.UUID) (map[uuid.UUID]bool, error) {
	snapshot, err := h.boundaryService.GetSnapshot(collaborationWorkspaceID)
	if err != nil {
		return nil, err
	}
	result := make(map[uuid.UUID]bool, len(snapshot.EffectiveIDs))
	for _, actionID := range snapshot.EffectiveIDs {
		result[actionID] = true
	}
	return result, nil
}

func (h *CollaborationWorkspaceHandler) getCollaborationWorkspaceEnabledMenuSet(collaborationWorkspaceID uuid.UUID) (map[uuid.UUID]bool, bool, error) {
	snapshot, err := h.boundaryService.GetMenuSnapshot(collaborationWorkspaceID)
	if err != nil {
		return nil, false, err
	}
	if len(snapshot.EffectiveIDs) == 0 {
		return map[uuid.UUID]bool{}, false, nil
	}
	result := make(map[uuid.UUID]bool, len(snapshot.EffectiveIDs))
	for _, menuID := range snapshot.EffectiveIDs {
		result[menuID] = true
	}
	return result, true, nil
}

func (h *CollaborationWorkspaceHandler) getRoleSnapshot(collaborationWorkspaceID uuid.UUID, role *user.Role, appKey string) (*collaborationworkspaceboundary.RoleSnapshot, error) {
	if role == nil {
		return &collaborationworkspaceboundary.RoleSnapshot{
			PackageIDs:         []uuid.UUID{},
			ExpandedPackageIDs: []uuid.UUID{},
			AvailableActionIDs: []uuid.UUID{},
			DisabledActionIDs:  []uuid.UUID{},
			ActionIDs:          []uuid.UUID{},
			ActionSourceMap:    map[uuid.UUID][]uuid.UUID{},
			AvailableMenuIDs:   []uuid.UUID{},
			HiddenMenuIDs:      []uuid.UUID{},
			MenuIDs:            []uuid.UUID{},
			MenuSourceMap:      map[uuid.UUID][]uuid.UUID{},
		}, nil
	}
	inheritAll := role.CollaborationWorkspaceID == nil
	return h.boundaryService.GetRoleSnapshot(collaborationWorkspaceID, role.ID, inheritAll, appKey)
}

func uuidSliceToSet(ids []uuid.UUID) map[uuid.UUID]bool {
	result := make(map[uuid.UUID]bool, len(ids))
	for _, id := range ids {
		result[id] = true
	}
	return result
}

func isAssignableRoleForCollaborationWorkspace(role user.Role, collaborationWorkspaceID uuid.UUID) bool {
	if role.CollaborationWorkspaceID != nil {
		return *role.CollaborationWorkspaceID == collaborationWorkspaceID
	}
	return role.Code == "collaboration_workspace_admin" || role.Code == "collaboration_workspace_member"
}

func (h *CollaborationWorkspaceHandler) ListMyCollaborationWorkspaces(c *gin.Context) {
	userID, err := h.mustUserID(c)
	if err != nil {
		status, resp := errcode.Response(errcode.ErrUnauthorized)
		c.JSON(status, resp)
		return
	}

	collaboration_workspaces, err := h.collaborationWorkspaceMemberRepo.GetCollaborationWorkspacesByUserID(userID)
	if err != nil {
		h.logger.Error("List my collaboration workspaces failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取我的协作空间列表失败")
		c.JSON(status, resp)
		return
	}

	records := make([]gin.H, 0, len(collaboration_workspaces))
	for _, t := range collaboration_workspaces {
		adminUsers, _ := h.collaborationWorkspaceMemberRepo.GetAdminUsersByCollaborationWorkspaceID(t.ID)
		member, memberErr := h.collaborationWorkspaceMemberRepo.GetByUserAndCollaborationWorkspace(userID, t.ID)
		record := h.collaborationWorkspaceToMap(&t, adminUsers)
		if memberErr == nil {
			record["current_role_code"] = member.RoleCode
			record["member_status"] = member.Status
		}
		records = append(records, record)
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(records))
}

func (h *CollaborationWorkspaceHandler) mustUserID(c *gin.Context) (uuid.UUID, error) {
	userID, ok := c.Get("user_id")
	if !ok {
		return uuid.Nil, errors.New("unauthorized")
	}
	userIDStr, ok := userID.(string)
	if !ok {
		return uuid.Nil, errors.New("unauthorized")
	}
	return uuid.Parse(userIDStr)
}

func (h *CollaborationWorkspaceHandler) resolveCollaborationWorkspaceMember(c *gin.Context) (*user.CollaborationWorkspaceMember, error) {
	userID, err := h.mustUserID(c)
	if err != nil {
		return nil, err
	}

	if collaborationWorkspaceID, ok := parseCollaborationWorkspaceIDFromContext(c); ok {
		member, memberErr := h.collaborationWorkspaceMemberRepo.GetByUserAndCollaborationWorkspace(userID, collaborationWorkspaceID)
		if memberErr != nil {
			if memberErr == gorm.ErrRecordNotFound {
				return nil, ErrCollaborationWorkspaceMemberNotFound
			}
			return nil, memberErr
		}
		return member, nil
	}

	member, err := h.collaborationWorkspaceMemberRepo.GetByUserID(userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrCollaborationWorkspaceMemberNotFound
		}
		return nil, err
	}
	return member, nil
}

func (h *CollaborationWorkspaceHandler) resolveCurrentCollaborationWorkspaceRole(c *gin.Context) (*user.CollaborationWorkspaceMember, *user.Role, error) {
	member, err := h.resolveCollaborationWorkspaceMember(c)
	if err != nil {
		return nil, nil, err
	}
	roleID, err := uuid.Parse(c.Param("roleId"))
	if err != nil {
		return member, nil, err
	}
	role, err := h.roleRepo.GetByID(roleID)
	if err != nil {
		return member, nil, err
	}
	if !isAssignableRoleForCollaborationWorkspace(*role, member.CollaborationWorkspaceID) {
		return member, nil, ErrCurrentCollaborationWorkspaceRoleForbidden
	}
	return member, role, nil
}

func (h *CollaborationWorkspaceHandler) resolveEditableCurrentCollaborationWorkspaceRole(c *gin.Context) (*user.CollaborationWorkspaceMember, *user.Role, error) {
	member, role, err := h.resolveCurrentCollaborationWorkspaceRole(c)
	if err != nil {
		return member, role, err
	}
	if role.CollaborationWorkspaceID == nil {
		return member, role, ErrCurrentCollaborationWorkspaceRoleForbidden
	}
	return member, role, nil
}

func (h *CollaborationWorkspaceHandler) respondCurrentCollaborationWorkspaceRoleError(c *gin.Context, err error) {
	switch {
	case err == gorm.ErrRecordNotFound:
		status, resp := errcode.Response(errcode.ErrRoleNotFound)
		c.JSON(status, resp)
	case err == ErrCollaborationWorkspaceMemberNotFound:
		status, resp := errcode.Response(errcode.ErrNoCurrentCollaborationWorkspace)
		c.JSON(status, resp)
	case errors.Is(err, ErrCurrentCollaborationWorkspaceRoleForbidden):
		status, resp := errcode.ResponseWithMsg(errcode.ErrForbidden, "仅支持操作当前协作空间自定义角色")
		c.JSON(status, resp)
	default:
		if _, parseErr := uuid.Parse(c.Param("roleId")); parseErr != nil {
			status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的角色ID")
			c.JSON(status, resp)
			return
		}
		h.logger.Error("Resolve current collaboration workspace role failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取协作空间角色失败")
		c.JSON(status, resp)
	}
}

func uuidPtrToString(value *uuid.UUID) string {
	if value == nil {
		return ""
	}
	return value.String()
}

func normalizeRoleStatus(status string) string {
	value := strings.TrimSpace(status)
	if value == "" {
		return "normal"
	}
	return value
}

func defaultString(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func parseCollaborationWorkspaceIDFromContext(c *gin.Context) (uuid.UUID, bool) {
	if value, ok := c.Get("collaboration_workspace_id"); ok {
		switch typed := value.(type) {
		case string:
			if collaborationWorkspaceID, err := uuid.Parse(strings.TrimSpace(typed)); err == nil {
				return collaborationWorkspaceID, true
			}
		case uuid.UUID:
			if typed != uuid.Nil {
				return typed, true
			}
		}
	}

	candidates := []string{
		strings.TrimSpace(c.Query("collaboration_workspace_id")),
		strings.TrimSpace(c.GetHeader(collaborationWorkspaceContextHeader)),
	}
	for _, candidate := range candidates {
		if candidate == "" {
			continue
		}
		if collaborationWorkspaceID, err := uuid.Parse(candidate); err == nil {
			return collaborationWorkspaceID, true
		}
	}
	return uuid.Nil, false
}

func (h *CollaborationWorkspaceHandler) requireTargetCollaborationWorkspace(c *gin.Context, collaborationWorkspaceID uuid.UUID) error {
	if h.authz == nil {
		return nil
	}
	authCtx, err := authorization.ResolveContext(c)
	if err != nil {
		h.authz.RespondAuthError(c, err, "collaboration_workspace.manage")
		return err
	}
	if _, err := h.authz.RequirePersonalWorkspaceTargetWorkspace(authCtx, collaborationWorkspaceID); err != nil {
		h.authz.RespondAuthError(c, err, "collaboration_workspace.manage")
		return err
	}
	return nil
}

func (h *CollaborationWorkspaceHandler) collaborationWorkspaceToMap(t *user.CollaborationWorkspace, ownerUsers []user.User) gin.H {
	m := gin.H{
		"id":          t.ID.String(),
		"name":        t.Name,
		"remark":      t.Remark,
		"logo_url":    t.LogoURL,
		"plan":        t.Plan,
		"owner_id":    t.OwnerID.String(),
		"max_members": t.MaxMembers,
		"status":      t.Status,
		"created_at":  t.CreatedAt.Format("2006-01-02 15:04:05"),
		"updated_at":  t.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
	if h.workspaceService != nil {
		if workspace, err := h.workspaceService.GetCollaborationWorkspaceByCollaborationWorkspaceID(t.ID); err == nil && workspace != nil {
			m["workspace_id"] = workspace.ID.String()
			m["collaboration_workspace_id"] = t.ID.String()
			m["workspace_type"] = workspace.WorkspaceType
		}
	}
	if len(ownerUsers) > 0 {
		admins := make([]gin.H, 0, len(ownerUsers))
		for _, u := range ownerUsers {
			admins = append(admins, gin.H{
				"user_id":   u.ID.String(),
				"user_name": u.Username,
				"nick_name": u.Nickname,
			})
		}
		m["admin_users"] = admins
	}
	return m
}

func (h *CollaborationWorkspaceHandler) memberToMap(m *user.CollaborationWorkspaceMember, userInfo *user.User) gin.H {
	result := gin.H{
		"id":                         m.ID.String(),
		"collaboration_workspace_id": m.CollaborationWorkspaceID.String(),
		"user_id":                    m.UserID.String(),
		"role_code":                  m.RoleCode,
		"member_type":                h.resolveCollaborationWorkspaceMemberType(m.CollaborationWorkspaceID, m.UserID, m),
		"status":                     m.Status,
		"joined_at":                  m.JoinedAt.Format("2006-01-02 15:04:05"),
	}
	if h.workspaceService != nil {
		if workspace, err := h.workspaceService.GetCollaborationWorkspaceByCollaborationWorkspaceID(m.CollaborationWorkspaceID); err == nil && workspace != nil {
			result["workspace_id"] = workspace.ID.String()
			result["collaboration_workspace_id"] = m.CollaborationWorkspaceID.String()
			result["workspace_type"] = workspace.WorkspaceType
		}
	}
	if userInfo != nil {
		result["user_name"] = userInfo.Username
		result["nick_name"] = userInfo.Nickname
		result["user_email"] = userInfo.Email
		result["user_phone"] = userInfo.Phone
		result["avatar"] = userInfo.AvatarURL
	}
	if m.InvitedBy != nil {
		result["invited_by"] = m.InvitedBy.String()
	}
	return result
}

func (h *CollaborationWorkspaceHandler) resolveCollaborationWorkspaceRoleBindingMeta(collaborationWorkspaceID, userID uuid.UUID, collaborationWorkspaceMember *user.CollaborationWorkspaceMember) (string, string, string) {
	memberType := h.resolveCollaborationWorkspaceMemberType(collaborationWorkspaceID, userID, collaborationWorkspaceMember)
	if h.workspaceService == nil {
		return "", "", memberType
	}
	workspace, err := h.workspaceService.GetCollaborationWorkspaceByCollaborationWorkspaceID(collaborationWorkspaceID)
	if err != nil || workspace == nil {
		return "", "", memberType
	}
	return workspace.ID.String(), workspace.WorkspaceType, memberType
}

func (h *CollaborationWorkspaceHandler) resolveCollaborationWorkspaceMemberType(collaborationWorkspaceID, userID uuid.UUID, collaborationWorkspaceMember *user.CollaborationWorkspaceMember) string {
	if h.workspaceService != nil {
		if workspace, err := h.workspaceService.GetCollaborationWorkspaceByCollaborationWorkspaceID(collaborationWorkspaceID); err == nil && workspace != nil {
			if member, err := h.workspaceService.GetMember(workspace.ID, userID); err == nil && member != nil && strings.TrimSpace(member.MemberType) != "" {
				return member.MemberType
			}
		}
	}
	if collaborationWorkspaceMember == nil {
		return ""
	}
	switch strings.ToLower(strings.TrimSpace(collaborationWorkspaceMember.RoleCode)) {
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

func actionMapToMap(action *user.PermissionKey) gin.H {
	return gin.H{
		"id":             action.ID.String(),
		"module_code":    action.ModuleCode,
		"context_type":   action.ContextType,
		"permission_key": action.PermissionKey,
		"feature_kind":   action.FeatureKind,
		"name":           action.Name,
		"description":    action.Description,
		"status":         action.Status,
		"sort_order":     action.SortOrder,
		"created_at":     action.CreatedAt.Format("2006-01-02 15:04:05"),
		"updated_at":     action.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

func actionListToMaps(actions []user.PermissionKey) []gin.H {
	items := make([]gin.H, 0, len(actions))
	for _, action := range actions {
		items = append(items, actionMapToMap(&action))
	}
	return items
}

func featurePackageListToMaps(items []user.FeaturePackage) []gin.H {
	result := make([]gin.H, 0, len(items))
	for _, item := range items {
		result = append(result, gin.H{
			"id":           item.ID.String(),
			"package_key":  item.PackageKey,
			"name":         item.Name,
			"description":  item.Description,
			"context_type": item.ContextType,
			"status":       item.Status,
			"sort_order":   item.SortOrder,
			"created_at":   item.CreatedAt.Format("2006-01-02 15:04:05"),
			"updated_at":   item.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	return result
}

func actionIDsToStrings(ids []uuid.UUID) []string {
	items := make([]string, 0, len(ids))
	for _, id := range ids {
		items = append(items, id.String())
	}
	return items
}

func buildDerivedSourceMaps(sourceMap map[uuid.UUID][]uuid.UUID) []gin.H {
	if len(sourceMap) == 0 {
		return []gin.H{}
	}
	items := make([]gin.H, 0, len(sourceMap))
	for actionID, packageIDs := range sourceMap {
		items = append(items, gin.H{
			"action_id":   actionID.String(),
			"package_ids": actionIDsToStrings(packageIDs),
		})
	}
	return items
}

func buildMenuSourceMaps(sourceMap map[uuid.UUID][]uuid.UUID) []gin.H {
	if len(sourceMap) == 0 {
		return []gin.H{}
	}
	items := make([]gin.H, 0, len(sourceMap))
	for menuID, packageIDs := range sourceMap {
		items = append(items, gin.H{
			"menu_id":     menuID.String(),
			"package_ids": actionIDsToStrings(packageIDs),
		})
	}
	return items
}

func mergeUUIDSourceLists(current []uuid.UUID, incoming []uuid.UUID) []uuid.UUID {
	result := make([]uuid.UUID, 0, len(current)+len(incoming))
	seen := make(map[uuid.UUID]struct{}, len(current)+len(incoming))
	for _, item := range current {
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		result = append(result, item)
	}
	for _, item := range incoming {
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		result = append(result, item)
	}
	return result
}

func parseUUIDSlice(items []string) ([]uuid.UUID, error) {
	result := make([]uuid.UUID, 0, len(items))
	seen := make(map[uuid.UUID]struct{}, len(items))
	for _, item := range items {
		if strings.TrimSpace(item) == "" {
			continue
		}
		parsed, err := uuid.Parse(item)
		if err != nil {
			return nil, err
		}
		if _, ok := seen[parsed]; ok {
			continue
		}
		seen[parsed] = struct{}{}
		result = append(result, parsed)
	}
	return result, nil
}

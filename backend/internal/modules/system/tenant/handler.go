package tenant

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
	appctx "github.com/gg-ecommerce/backend/internal/pkg/appctx"
	"github.com/gg-ecommerce/backend/internal/pkg/appscope"
	"github.com/gg-ecommerce/backend/internal/pkg/database"
	"github.com/gg-ecommerce/backend/internal/pkg/teamboundary"
)

const tenantContextHeader = "X-Tenant-ID"

var ErrMyTeamRoleForbidden = errors.New("team role forbidden")

type TenantHandler struct {
	tenantService          TenantService
	tenantMemberRepo       user.TenantMemberRepository
	userRepo               user.UserRepository
	roleRepo               user.RoleRepository
	roleHiddenMenuRepo     user.RoleHiddenMenuRepository
	roleDisabledActionRepo user.RoleDisabledActionRepository
	userRoleRepo           user.UserRoleRepository
	actionRepo             user.PermissionKeyRepository
	blockedMenuRepo        user.TeamBlockedMenuRepository
	blockedActionRepo      user.TeamBlockedActionRepository
	teamPackageRepo        user.TeamFeaturePackageRepository
	rolePackageRepo        user.RoleFeaturePackageRepository
	featurePkgRepo         user.FeaturePackageRepository
	packageActionRepo      user.FeaturePackageKeyRepository
	packageMenuRepo        user.FeaturePackageMenuRepository
	boundaryService        teamboundary.Service
	refresher              interface {
		RefreshTeam(teamID uuid.UUID) error
	}
	logger *zap.Logger
}

func NewTenantHandler(tenantService TenantService, tenantMemberRepo user.TenantMemberRepository, userRepo user.UserRepository, roleRepo user.RoleRepository, roleHiddenMenuRepo user.RoleHiddenMenuRepository, roleDisabledActionRepo user.RoleDisabledActionRepository, userRoleRepo user.UserRoleRepository, actionRepo user.PermissionKeyRepository, blockedMenuRepo user.TeamBlockedMenuRepository, blockedActionRepo user.TeamBlockedActionRepository, teamPackageRepo user.TeamFeaturePackageRepository, rolePackageRepo user.RoleFeaturePackageRepository, featurePkgRepo user.FeaturePackageRepository, packageActionRepo user.FeaturePackageKeyRepository, packageMenuRepo user.FeaturePackageMenuRepository, boundaryService teamboundary.Service, refresher interface {
	RefreshTeam(teamID uuid.UUID) error
}, logger *zap.Logger) *TenantHandler {
	return &TenantHandler{
		tenantService:          tenantService,
		tenantMemberRepo:       tenantMemberRepo,
		userRepo:               userRepo,
		roleRepo:               roleRepo,
		roleHiddenMenuRepo:     roleHiddenMenuRepo,
		roleDisabledActionRepo: roleDisabledActionRepo,
		userRoleRepo:           userRoleRepo,
		actionRepo:             actionRepo,
		blockedMenuRepo:        blockedMenuRepo,
		blockedActionRepo:      blockedActionRepo,
		teamPackageRepo:        teamPackageRepo,
		rolePackageRepo:        rolePackageRepo,
		featurePkgRepo:         featurePkgRepo,
		packageActionRepo:      packageActionRepo,
		packageMenuRepo:        packageMenuRepo,
		boundaryService:        boundaryService,
		refresher:              refresher,
		logger:                 logger,
	}
}

func (h *TenantHandler) List(c *gin.Context) {
	var req dto.TenantListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		status, resp := errcode.Response(errcode.ErrParamInvalid)
		c.JSON(status, resp)
		return
	}
	list, total, err := h.tenantService.List(&req)
	if err != nil {
		h.logger.Error("Tenant list failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取团队列表失败")
		c.JSON(status, resp)
		return
	}
	records := make([]gin.H, 0, len(list))
	for _, t := range list {
		adminUsers, _ := h.tenantMemberRepo.GetAdminUsersByTenantID(t.ID)
		records = append(records, tenantToMap(&t, adminUsers))
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"records": records,
		"total":   total,
		"current": req.Current,
		"size":    req.Size,
	}))
}

func (h *TenantHandler) ListOptions(c *gin.Context) {
	var req dto.TenantListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		status, resp := errcode.Response(errcode.ErrParamInvalid)
		c.JSON(status, resp)
		return
	}
	list, err := h.tenantService.ListOptions(&req)
	if err != nil {
		h.logger.Error("Tenant options failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取团队候选失败")
		c.JSON(status, resp)
		return
	}
	records := make([]gin.H, 0, len(list))
	for _, t := range list {
		tenant := t
		records = append(records, tenantToMap(&tenant, nil))
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"records": records,
		"total":   len(records),
	}))
}

func (h *TenantHandler) Get(c *gin.Context) {
	idStr := c.Param("id")
	if strings.TrimSpace(strings.ToLower(idStr)) == "my-team" {
		h.GetMyTeam(c)
		return
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的团队ID")
		c.JSON(status, resp)
		return
	}
	t, err := h.tenantService.Get(id)
	if err != nil {
		if err == ErrTenantNotFound {
			status, resp := errcode.Response(errcode.ErrTenantNotFound)
			c.JSON(status, resp)
			return
		}
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取团队失败")
		c.JSON(status, resp)
		return
	}
	adminUsers, _ := h.tenantMemberRepo.GetAdminUsersByTenantID(id)
	c.JSON(http.StatusOK, dto.SuccessResponse(tenantToMap(t, adminUsers)))
}

func (h *TenantHandler) Create(c *gin.Context) {
	var req dto.TenantCreateRequest
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
	t, err := h.tenantService.Create(&req, ownerID)
	if err != nil {
		h.logger.Error("Tenant create failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, err.Error())
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{"id": t.ID.String()}))
}

func (h *TenantHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的团队ID")
		c.JSON(status, resp)
		return
	}
	var req dto.TenantUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		status, resp := errcode.Response(errcode.ErrParamInvalid)
		c.JSON(status, resp)
		return
	}
	if err := h.tenantService.Update(id, &req); err != nil {
		if err == ErrTenantNotFound {
			status, resp := errcode.Response(errcode.ErrTenantNotFound)
			c.JSON(status, resp)
			return
		}
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "更新团队失败")
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(nil))
}

func (h *TenantHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的团队ID")
		c.JSON(status, resp)
		return
	}
	if err := h.tenantService.Delete(id); err != nil {
		if err == ErrTenantNotFound {
			status, resp := errcode.Response(errcode.ErrTenantNotFound)
			c.JSON(status, resp)
			return
		}
		h.logger.Error("Tenant delete failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "删除团队失败")
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(nil))
}

func (h *TenantHandler) ListMembers(c *gin.Context) {
	tenantIDStr := c.Param("id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的团队ID")
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
	members, err := h.tenantService.ListMembers(tenantID, searchParams)
	if err != nil {
		h.logger.Error("List members failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取成员列表失败")
		c.JSON(status, resp)
		return
	}
	records := make([]gin.H, 0, len(members))
	for _, m := range members {
		userInfo, _ := h.userRepo.GetByID(m.UserID)
		records = append(records, memberToMap(&m, userInfo))
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(records))
}

func (h *TenantHandler) AddMember(c *gin.Context) {
	tenantIDStr := c.Param("id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的团队ID")
		c.JSON(status, resp)
		return
	}
	var req dto.TenantAddMemberRequest
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
	if err := h.tenantService.AddMember(tenantID, &req, invitedBy); err != nil {
		if err == ErrTenantMemberExists {
			status, resp := errcode.Response(errcode.ErrTenantMemberExists)
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

func (h *TenantHandler) RemoveMember(c *gin.Context) {
	tenantIDStr := c.Param("id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的团队ID")
		c.JSON(status, resp)
		return
	}
	userIDStr := c.Param("userId")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的用户ID")
		c.JSON(status, resp)
		return
	}
	if err := h.tenantService.RemoveMember(tenantID, userID); err != nil {
		if err == ErrTenantMemberNotFound {
			status, resp := errcode.Response(errcode.ErrTenantMemberNotFound)
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

func (h *TenantHandler) UpdateMemberRole(c *gin.Context) {
	tenantIDStr := c.Param("id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的团队ID")
		c.JSON(status, resp)
		return
	}
	userIDStr := c.Param("userId")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的用户ID")
		c.JSON(status, resp)
		return
	}
	var req dto.TenantUpdateMemberRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		status, resp := errcode.Response(errcode.ErrParamInvalid)
		c.JSON(status, resp)
		return
	}
	roleCode := strings.TrimSpace(req.RoleCode)
	if err := h.tenantService.UpdateMemberRole(tenantID, userID, roleCode); err != nil {
		if err == ErrTenantMemberNotFound {
			status, resp := errcode.Response(errcode.ErrTenantMemberNotFound)
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

func (h *TenantHandler) GetMyTeam(c *gin.Context) {
	member, err := h.resolveTenantMember(c)
	if err != nil {
		if err == gorm.ErrRecordNotFound || err == ErrTenantMemberNotFound {
			status, resp := errcode.Response(errcode.ErrNoTeam)
			c.JSON(status, resp)
			return
		}
		h.logger.Error("Get my team failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取我的团队失败")
		c.JSON(status, resp)
		return
	}

	tenant, err := h.tenantService.Get(member.TenantID)
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取团队信息失败")
		c.JSON(status, resp)
		return
	}
	adminUsers, _ := h.tenantMemberRepo.GetAdminUsersByTenantID(tenant.ID)
	c.JSON(http.StatusOK, dto.SuccessResponse(tenantToMap(tenant, adminUsers)))
}

func (h *TenantHandler) ListMyMembers(c *gin.Context) {
	member, err := h.resolveTenantMember(c)
	if err != nil {
		if err == gorm.ErrRecordNotFound || err == ErrTenantMemberNotFound {
			c.JSON(http.StatusOK, dto.SuccessResponse([]gin.H{}))
			return
		}
		h.logger.Error("Get my team member failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取我的团队成员失败")
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
	members, err := h.tenantService.ListMembers(member.TenantID, searchParams)
	if err != nil {
		h.logger.Error("List my members failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取团队成员列表失败")
		c.JSON(status, resp)
		return
	}

	records := make([]gin.H, 0, len(members))
	for _, m := range members {
		userInfo, _ := h.userRepo.GetByID(m.UserID)
		records = append(records, memberToMap(&m, userInfo))
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(records))
}

func (h *TenantHandler) AddMyMember(c *gin.Context) {
	uid, err := h.mustUserID(c)
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的用户ID")
		c.JSON(status, resp)
		return
	}

	member, err := h.resolveTenantMember(c)
	if err != nil {
		if err == gorm.ErrRecordNotFound || err == ErrTenantMemberNotFound {
			status, resp := errcode.Response(errcode.ErrNoTeam)
			c.JSON(status, resp)
			return
		}
		h.logger.Error("Get my team member failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取我的团队失败")
		c.JSON(status, resp)
		return
	}

	var req dto.TenantAddMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		status, resp := errcode.Response(errcode.ErrParamInvalid)
		c.JSON(status, resp)
		return
	}
	invitedBy := &uid
	if err := h.tenantService.AddMember(member.TenantID, &req, invitedBy); err != nil {
		if err == ErrTenantMemberExists {
			status, resp := errcode.Response(errcode.ErrTenantMemberExists)
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

func (h *TenantHandler) RemoveMyMember(c *gin.Context) {
	member, err := h.resolveTenantMember(c)
	if err != nil {
		if err == gorm.ErrRecordNotFound || err == ErrTenantMemberNotFound {
			status, resp := errcode.Response(errcode.ErrNoTeam)
			c.JSON(status, resp)
			return
		}
		h.logger.Error("Get my team member failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取我的团队失败")
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

	if err := h.tenantService.RemoveMember(member.TenantID, targetUserID); err != nil {
		if err == ErrTenantMemberNotFound {
			status, resp := errcode.Response(errcode.ErrTenantMemberNotFound)
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

func (h *TenantHandler) UpdateMyMemberRole(c *gin.Context) {
	member, err := h.resolveTenantMember(c)
	if err != nil {
		if err == gorm.ErrRecordNotFound || err == ErrTenantMemberNotFound {
			status, resp := errcode.Response(errcode.ErrNoTeam)
			c.JSON(status, resp)
			return
		}
		h.logger.Error("Get my team member failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取我的团队失败")
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

	var req dto.TenantUpdateMemberRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		status, resp := errcode.Response(errcode.ErrParamInvalid)
		c.JSON(status, resp)
		return
	}
	roleCode := strings.TrimSpace(req.RoleCode)
	if err := h.tenantService.UpdateMemberRole(member.TenantID, targetUserID, roleCode); err != nil {
		if err == ErrTenantMemberNotFound {
			status, resp := errcode.Response(errcode.ErrTenantMemberNotFound)
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

func (h *TenantHandler) GetMyTeamMemberRoles(c *gin.Context) {
	member, err := h.resolveTenantMember(c)
	if err != nil {
		if err == gorm.ErrRecordNotFound || err == ErrTenantMemberNotFound {
			c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
				"role_ids": []string{},
			}))
			return
		}
		h.logger.Error("Get my team member failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取我的团队失败")
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
	if _, err := h.tenantMemberRepo.GetByUserAndTenant(targetUserID, member.TenantID); err != nil {
		if err == gorm.ErrRecordNotFound {
			status, resp := errcode.Response(errcode.ErrTenantMemberNotFound)
			c.JSON(status, resp)
			return
		}
		h.logger.Error("Get tenant member failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取成员失败")
		c.JSON(status, resp)
		return
	}

	roleIDs, err := h.userRoleRepo.GetRoleIDsByUserAndTenant(targetUserID, &member.TenantID, h.tenantMemberRepo)
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
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"role_ids": roleIDsStr,
		"roles":    roleList,
	}))
}

func (h *TenantHandler) SetMyTeamMemberRoles(c *gin.Context) {
	member, err := h.resolveTenantMember(c)
	if err != nil {
		if err == gorm.ErrRecordNotFound || err == ErrTenantMemberNotFound {
			status, resp := errcode.Response(errcode.ErrNoTeam)
			c.JSON(status, resp)
			return
		}
		h.logger.Error("Get my team member failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取我的团队失败")
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

	var req dto.TenantSetMemberRolesRequest
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

	memberRecord, err := h.tenantMemberRepo.GetByUserAndTenant(targetUserID, member.TenantID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			status, resp := errcode.Response(errcode.ErrTenantMemberNotFound)
			c.JSON(status, resp)
			return
		}
		h.logger.Error("Get tenant member failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取成员失败")
		c.JSON(status, resp)
		return
	}

	allRoles, err := h.roleRepo.ListTeamRoles(member.TenantID)
	if err != nil {
		h.logger.Error("Get team roles failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取团队角色失败")
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

	if err := h.userRoleRepo.SetUserRoles(targetUserID, filteredRoleIDs, &member.TenantID); err != nil {
		h.logger.Error("Set user roles failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "设置角色失败")
		c.JSON(status, resp)
		return
	}
	if h.refresher != nil {
		if err := h.refresher.RefreshTeam(member.TenantID); err != nil {
			h.logger.Error("Refresh team after setting member roles failed", zap.Error(err))
			status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "刷新团队权限快照失败")
			c.JSON(status, resp)
			return
		}
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(nil))
}

func (h *TenantHandler) ListMyTeamRoles(c *gin.Context) {
	member, err := h.resolveTenantMember(c)
	if err != nil {
		if err == gorm.ErrRecordNotFound || err == ErrTenantMemberNotFound {
			c.JSON(http.StatusOK, dto.SuccessResponse([]gin.H{}))
			return
		}
		h.logger.Error("Get my team member failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取我的团队失败")
		c.JSON(status, resp)
		return
	}

	_ = member.TenantID
	allRoles, err := h.roleRepo.ListTeamRoles(member.TenantID)
	if err != nil {
		h.logger.Error("List roles failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取角色列表失败")
		c.JSON(status, resp)
		return
	}

	roleList := make([]gin.H, 0, len(allRoles))
	for _, r := range allRoles {
		roleList = append(roleList, gin.H{
			"id":          r.ID.String(),
			"code":        r.Code,
			"name":        r.Name,
			"description": r.Description,
			"status":      r.Status,
			"is_system":   r.IsSystem,
			"tenant_id":   uuidPtrToString(r.TenantID),
			"is_global":   r.TenantID == nil,
			"create_time": r.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(roleList))
}

func (h *TenantHandler) ListTenantRoles(c *gin.Context) {
	tenantIDStr := c.Param("id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的团队ID")
		c.JSON(status, resp)
		return
	}

	if _, err := h.tenantService.Get(tenantID); err != nil {
		if err == ErrTenantNotFound {
			status, resp := errcode.Response(errcode.ErrTenantNotFound)
			c.JSON(status, resp)
			return
		}
		h.logger.Error("Get tenant failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取团队失败")
		c.JSON(status, resp)
		return
	}

	allRoles, err := h.roleRepo.ListTeamRoles(tenantID)
	if err != nil {
		h.logger.Error("List tenant roles failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取团队角色失败")
		c.JSON(status, resp)
		return
	}

	roleList := make([]gin.H, 0, len(allRoles))
	for _, r := range allRoles {
		roleList = append(roleList, gin.H{
			"id":          r.ID.String(),
			"code":        r.Code,
			"name":        r.Name,
			"description": r.Description,
			"status":      r.Status,
			"is_system":   r.IsSystem,
			"tenant_id":   uuidPtrToString(r.TenantID),
			"is_global":   r.TenantID == nil,
			"create_time": r.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(roleList))
}

func (h *TenantHandler) CreateMyTeamRole(c *gin.Context) {
	member, err := h.resolveTenantMember(c)
	if err != nil {
		if err == gorm.ErrRecordNotFound || err == ErrTenantMemberNotFound {
			status, resp := errcode.Response(errcode.ErrNoTeam)
			c.JSON(status, resp)
			return
		}
		h.logger.Error("Get my team member failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取我的团队失败")
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
		h.logger.Error("Find team role code failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "校验角色编码失败")
		c.JSON(status, resp)
		return
	}
	for _, existing := range existingRoles {
		if existing.TenantID != nil && *existing.TenantID == member.TenantID {
			status, resp := errcode.Response(errcode.ErrRoleCodeExists)
			c.JSON(status, resp)
			return
		}
	}

	role := &user.Role{
		TenantID:    &member.TenantID,
		Code:        code,
		Name:        strings.TrimSpace(req.Name),
		Description: strings.TrimSpace(req.Description),
		SortOrder:   req.SortOrder,
		Priority:    req.Priority,
		Status:      normalizeRoleStatus(req.Status),
	}
	if err := h.roleRepo.Create(role); err != nil {
		h.logger.Error("Create team role failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "创建团队角色失败")
		c.JSON(status, resp)
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{"roleId": role.ID.String()}))
}

func (h *TenantHandler) UpdateMyTeamRole(c *gin.Context) {
	member, role, err := h.resolveEditableMyTeamRole(c)
	if err != nil {
		h.respondMyTeamRoleError(c, err)
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
			h.logger.Error("Find team role code failed", zap.Error(findErr))
			status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "校验角色编码失败")
			c.JSON(status, resp)
			return
		}
		for _, existing := range existingRoles {
			if existing.ID == role.ID {
				continue
			}
			if existing.TenantID != nil && *existing.TenantID == member.TenantID {
				status, resp := errcode.Response(errcode.ErrRoleCodeExists)
				c.JSON(status, resp)
				return
			}
		}
		updates["code"] = code
	}

	if err := h.roleRepo.UpdateWithMap(role.ID, updates); err != nil {
		h.logger.Error("Update team role failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "更新团队角色失败")
		c.JSON(status, resp)
		return
	}
	if h.refresher != nil {
		if err := h.refresher.RefreshTeam(member.TenantID); err != nil {
			h.logger.Error("Refresh team after updating role failed", zap.Error(err))
			status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "刷新团队权限快照失败")
			c.JSON(status, resp)
			return
		}
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(nil))
}

func (h *TenantHandler) DeleteMyTeamRole(c *gin.Context) {
	member, role, err := h.resolveEditableMyTeamRole(c)
	if err != nil {
		h.respondMyTeamRoleError(c, err)
		return
	}

	if err := database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("role_id = ? AND tenant_id = ?", role.ID, member.TenantID).Delete(&user.UserRole{}).Error; err != nil {
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
		if err := tx.Where("team_id = ? AND role_id = ?", member.TenantID, role.ID).Delete(&models.TeamRoleAccessSnapshot{}).Error; err != nil {
			return err
		}
		return tx.Delete(&user.Role{}, role.ID).Error
	}); err != nil {
		h.logger.Error("Delete team role failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "删除团队角色失败")
		c.JSON(status, resp)
		return
	}
	if h.refresher != nil {
		if err := h.refresher.RefreshTeam(member.TenantID); err != nil {
			h.logger.Error("Refresh team after deleting role failed", zap.Error(err))
			status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "刷新团队权限快照失败")
			c.JSON(status, resp)
			return
		}
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(nil))
}

func (h *TenantHandler) GetMyTeamRolePackages(c *gin.Context) {
	member, role, err := h.resolveMyTeamRole(c)
	if err != nil {
		h.respondMyTeamRoleError(c, err)
		return
	}

	appKey, err := appctx.RequireRequestAppKey(c)
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrParamInvalid, "app_key 为必填项")
		c.JSON(status, resp)
		return
	}
	snapshot, err := h.getRoleSnapshot(member.TenantID, role, appKey)
	if err != nil {
		h.logger.Error("Get team role packages failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取团队角色功能包失败")
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

func (h *TenantHandler) SetMyTeamRolePackages(c *gin.Context) {
	member, role, err := h.resolveEditableMyTeamRole(c)
	if err != nil {
		h.respondMyTeamRoleError(c, err)
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

	teamPackageIDs, err := h.teamPackageRepo.GetPackageIDsByTeamID(member.TenantID)
	if err != nil {
		h.logger.Error("Get team feature packages failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取团队功能包失败")
		c.JSON(status, resp)
		return
	}
	allowedPackageSet := make(map[uuid.UUID]struct{}, len(teamPackageIDs))
	for _, packageID := range teamPackageIDs {
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
			if item.ContextType != "" && item.ContextType != "team" && item.ContextType != "common" {
				status, resp := errcode.ResponseWithMsg(errcode.ErrForbidden, "仅支持绑定团队上下文可用的功能包")
				c.JSON(status, resp)
				return
			}
		}
	}
	for _, packageID := range packageIDs {
		if _, ok := allowedPackageSet[packageID]; !ok {
			status, resp := errcode.ResponseWithMsg(errcode.ErrForbidden, "存在未向当前团队开通的功能包")
			c.JSON(status, resp)
			return
		}
	}

	userID, _ := h.mustUserID(c)
	if err := appscope.ReplaceRolePackagesInApp(database.DB, role.ID, appKey, packageIDs, &userID); err != nil {
		h.logger.Error("Set team role packages failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "保存团队角色功能包失败")
		c.JSON(status, resp)
		return
	}
	if h.refresher != nil {
		if err := h.refresher.RefreshTeam(member.TenantID); err != nil {
			h.logger.Error("Refresh team after setting role packages failed", zap.Error(err))
			status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "刷新团队权限快照失败")
			c.JSON(status, resp)
			return
		}
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(nil))
}

func (h *TenantHandler) GetMyTeamRoleMenus(c *gin.Context) {
	member, role, err := h.resolveMyTeamRole(c)
	if err != nil {
		h.respondMyTeamRoleError(c, err)
		return
	}

	appKey, err := appctx.RequireRequestAppKey(c)
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrParamInvalid, "app_key 为必填项")
		c.JSON(status, resp)
		return
	}
	snapshot, err := h.getRoleSnapshot(member.TenantID, role, appKey)
	if err != nil {
		h.logger.Error("Get role menu boundary failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取团队角色菜单范围失败")
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

func (h *TenantHandler) SetMyTeamRoleMenus(c *gin.Context) {
	member, role, err := h.resolveEditableMyTeamRole(c)
	if err != nil {
		h.respondMyTeamRoleError(c, err)
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

	snapshot, err := h.getRoleSnapshot(member.TenantID, role, appKey)
	if err != nil {
		h.logger.Error("Get role menu boundary failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取团队角色菜单范围失败")
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
		h.logger.Error("Set team role hidden menus failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "保存团队角色菜单权限失败")
		c.JSON(status, resp)
		return
	}
	if h.refresher != nil {
		if err := h.refresher.RefreshTeam(member.TenantID); err != nil {
			h.logger.Error("Refresh team after setting role menus failed", zap.Error(err))
			status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "刷新团队权限快照失败")
			c.JSON(status, resp)
			return
		}
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(nil))
}

func (h *TenantHandler) GetMyTeamRoleActions(c *gin.Context) {
	member, role, err := h.resolveMyTeamRole(c)
	if err != nil {
		h.respondMyTeamRoleError(c, err)
		return
	}

	appKey, err := appctx.RequireRequestAppKey(c)
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrParamInvalid, "app_key 为必填项")
		c.JSON(status, resp)
		return
	}
	snapshot, err := h.getRoleSnapshot(member.TenantID, role, appKey)
	if err != nil {
		h.logger.Error("Get role action boundary failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取团队角色功能范围失败")
		c.JSON(status, resp)
		return
	}
	actions, err := h.actionRepo.GetByIDs(snapshot.AvailableActionIDs)
	if err != nil {
		h.logger.Error("Get role boundary actions failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取团队角色功能范围失败")
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

func (h *TenantHandler) SetMyTeamRoleActions(c *gin.Context) {
	member, role, err := h.resolveEditableMyTeamRole(c)
	if err != nil {
		h.respondMyTeamRoleError(c, err)
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

	snapshot, err := h.getRoleSnapshot(member.TenantID, role, appKey)
	if err != nil {
		h.logger.Error("Get role action boundary failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取团队角色功能范围失败")
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
		h.logger.Error("Set team role disabled actions failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "保存团队角色功能权限失败")
		c.JSON(status, resp)
		return
	}
	if h.refresher != nil {
		if err := h.refresher.RefreshTeam(member.TenantID); err != nil {
			h.logger.Error("Refresh team after setting role actions failed", zap.Error(err))
			status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "刷新团队权限快照失败")
			c.JSON(status, resp)
			return
		}
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(nil))
}

func (h *TenantHandler) GetTenantActions(c *gin.Context) {
	tenantID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的团队ID")
		c.JSON(status, resp)
		return
	}
	appKey, err := appctx.RequireRequestAppKey(c)
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrParamInvalid, "app_key 为必填项")
		c.JSON(status, resp)
		return
	}
	snapshot, err := h.boundaryService.GetSnapshot(tenantID, appKey)
	if err != nil {
		h.logger.Error("Get tenant actions failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取团队功能权限失败")
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

func (h *TenantHandler) GetTenantMenus(c *gin.Context) {
	tenantID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的团队ID")
		c.JSON(status, resp)
		return
	}
	appKey, err := appctx.RequireRequestAppKey(c)
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrParamInvalid, "app_key 为必填项")
		c.JSON(status, resp)
		return
	}
	snapshot, err := h.boundaryService.GetMenuSnapshot(tenantID, appKey)
	if err != nil {
		h.logger.Error("Get tenant menus failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取团队菜单边界失败")
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"menu_ids": actionIDsToStrings(snapshot.EffectiveIDs),
	}))
}

func (h *TenantHandler) GetTenantActionOrigins(c *gin.Context) {
	tenantID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的团队ID")
		c.JSON(status, resp)
		return
	}
	h.respondTenantActionOrigins(c, tenantID)
}

func (h *TenantHandler) GetTenantMenuOrigins(c *gin.Context) {
	tenantID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的团队ID")
		c.JSON(status, resp)
		return
	}
	h.respondTenantMenuOrigins(c, tenantID)
}

func (h *TenantHandler) SetTenantActions(c *gin.Context) {
	tenantID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的团队ID")
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
	actionIDs, parseErr := parseUUIDSlice(req.KeyIDs)
	if parseErr != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的功能权限ID")
		c.JSON(status, resp)
		return
	}
	if len(actionIDs) > 0 {
		actions, actionErr := h.actionRepo.GetByIDs(actionIDs)
		if actionErr != nil {
			h.logger.Error("Get tenant action detail failed", zap.Error(actionErr))
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
	snapshot, err := h.boundaryService.GetSnapshot(tenantID, appKey)
	if err != nil {
		h.logger.Error("Get derived team actions failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取功能包展开权限失败")
		c.JSON(status, resp)
		return
	}
	blockedIDs := excludeActionIDs(snapshot.DerivedIDs, actionIDs)
	if err := appscope.ReplaceTeamBlockedActionsInScope(database.DB, tenantID, snapshot.DerivedIDs, blockedIDs); err != nil {
		h.logger.Error("Set team blocked actions failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "保存团队功能边界失败")
		c.JSON(status, resp)
		return
	}
	if h.refresher != nil {
		if err := h.refresher.RefreshTeam(tenantID); err != nil {
			h.logger.Error("Set tenant actions failed", zap.Error(err))
			status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "保存团队功能权限失败")
			c.JSON(status, resp)
			return
		}
	} else if _, err := h.boundaryService.RefreshSnapshot(tenantID, appKey); err != nil {
		h.logger.Error("Set tenant actions failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "保存团队功能权限失败")
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(nil))
}

func (h *TenantHandler) SetTenantMenus(c *gin.Context) {
	tenantID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的团队ID")
		c.JSON(status, resp)
		return
	}
	var req dto.TenantMenuPermissionsRequest
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
	snapshot, err := h.boundaryService.GetMenuSnapshot(tenantID, appKey)
	if err != nil {
		h.logger.Error("Get derived team menus failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取功能包展开菜单失败")
		c.JSON(status, resp)
		return
	}
	blockedIDs := excludeActionIDs(snapshot.DerivedIDs, menuIDs)
	if err := appscope.ReplaceTeamBlockedMenusInApp(database.DB, tenantID, appKey, blockedIDs); err != nil {
		h.logger.Error("Set team blocked menus failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "保存团队菜单边界失败")
		c.JSON(status, resp)
		return
	}
	if h.refresher != nil {
		if err := h.refresher.RefreshTeam(tenantID); err != nil {
			h.logger.Error("Set tenant menus failed", zap.Error(err))
			status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "保存团队菜单边界失败")
			c.JSON(status, resp)
			return
		}
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(nil))
}

func (h *TenantHandler) GetMyTeamBoundaryPackages(c *gin.Context) {
	member, err := h.resolveTenantMember(c)
	if err != nil {
		if err == gorm.ErrRecordNotFound || err == ErrTenantMemberNotFound {
			status, resp := errcode.Response(errcode.ErrNoTeam)
			c.JSON(status, resp)
			return
		}
		h.logger.Error("Resolve my team failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取我的团队失败")
		c.JSON(status, resp)
		return
	}

	appKey, err := appctx.RequireRequestAppKey(c)
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrParamInvalid, "app_key 为必填项")
		c.JSON(status, resp)
		return
	}
	packageIDs, err := appscope.PackageIDsByTeam(database.DB, member.TenantID, appKey)
	if err != nil {
		h.logger.Error("Get my team packages failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取团队功能包失败")
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
		h.logger.Error("Get my team package details failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取功能包详情失败")
		c.JSON(status, resp)
		return
	}

	filtered := make([]user.FeaturePackage, 0, len(packages))
	for _, item := range packages {
		if strings.TrimSpace(item.Status) != "" && item.Status != "normal" {
			continue
		}
		if item.ContextType != "" && item.ContextType != "team" && item.ContextType != "common" {
			continue
		}
		filtered = append(filtered, item)
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"package_ids": actionIDsToStrings(packageIDs),
		"packages":    featurePackageListToMaps(filtered),
	}))
}

func (h *TenantHandler) GetMyTeamActions(c *gin.Context) {
	member, err := h.resolveTenantMember(c)
	if err != nil {
		if err == gorm.ErrRecordNotFound || err == ErrTenantMemberNotFound {
			status, resp := errcode.Response(errcode.ErrNoTeam)
			c.JSON(status, resp)
			return
		}
		h.logger.Error("Resolve my team failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取我的团队失败")
		c.JSON(status, resp)
		return
	}
	appKey, err := appctx.RequireRequestAppKey(c)
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrParamInvalid, "app_key 为必填项")
		c.JSON(status, resp)
		return
	}
	snapshot, err := h.boundaryService.GetSnapshot(member.TenantID, appKey)
	if err != nil {
		h.logger.Error("Get my team actions failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取团队功能权限失败")
		c.JSON(status, resp)
		return
	}
	actions, err := h.actionRepo.GetByIDs(snapshot.EffectiveIDs)
	if err != nil {
		h.logger.Error("Get my team permission actions failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取功能权限失败")
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"action_ids": actionIDsToStrings(snapshot.EffectiveIDs),
		"actions":    actionListToMaps(actions),
	}))
}

func (h *TenantHandler) GetMyTeamMenus(c *gin.Context) {
	member, err := h.resolveTenantMember(c)
	if err != nil {
		if err == gorm.ErrRecordNotFound || err == ErrTenantMemberNotFound {
			status, resp := errcode.Response(errcode.ErrNoTeam)
			c.JSON(status, resp)
			return
		}
		h.logger.Error("Resolve my team failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取我的团队失败")
		c.JSON(status, resp)
		return
	}
	appKey, err := appctx.RequireRequestAppKey(c)
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrParamInvalid, "app_key 为必填项")
		c.JSON(status, resp)
		return
	}
	snapshot, err := h.boundaryService.GetMenuSnapshot(member.TenantID, appKey)
	if err != nil {
		h.logger.Error("Get my team menus failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取团队菜单边界失败")
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"menu_ids": actionIDsToStrings(snapshot.EffectiveIDs),
	}))
}

func (h *TenantHandler) GetMyTeamActionOrigins(c *gin.Context) {
	member, err := h.resolveTenantMember(c)
	if err != nil {
		if err == gorm.ErrRecordNotFound || err == ErrTenantMemberNotFound {
			c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
				"derived_action_ids": []string{},
				"derived_sources":    []gin.H{},
				"blocked_action_ids": []string{},
			}))
			return
		}
		h.logger.Error("Resolve my team failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取我的团队失败")
		c.JSON(status, resp)
		return
	}
	h.respondTenantActionOrigins(c, member.TenantID)
}

func (h *TenantHandler) GetMyTeamMenuOrigins(c *gin.Context) {
	member, err := h.resolveTenantMember(c)
	if err != nil {
		if err == gorm.ErrRecordNotFound || err == ErrTenantMemberNotFound {
			c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
				"derived_menu_ids": []string{},
				"derived_sources":  []gin.H{},
				"blocked_menu_ids": []string{},
			}))
			return
		}
		h.logger.Error("Resolve my team failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取我的团队失败")
		c.JSON(status, resp)
		return
	}
	h.respondTenantMenuOrigins(c, member.TenantID)
}

func (h *TenantHandler) respondTenantActionOrigins(c *gin.Context, tenantID uuid.UUID) {
	snapshot, err := h.boundaryService.GetSnapshot(tenantID)
	if err != nil {
		h.logger.Error("Get tenant action origins failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取团队功能权限来源失败")
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"derived_action_ids": actionIDsToStrings(snapshot.DerivedIDs),
		"derived_sources":    buildDerivedSourceMaps(snapshot.DerivedMap),
		"blocked_action_ids": actionIDsToStrings(snapshot.BlockedIDs),
	}))
}

func (h *TenantHandler) respondTenantMenuOrigins(c *gin.Context, tenantID uuid.UUID) {
	snapshot, err := h.boundaryService.GetMenuSnapshot(tenantID)
	if err != nil {
		h.logger.Error("Get tenant menu origins failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取团队菜单来源失败")
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

func (h *TenantHandler) getTenantEnabledActionSet(tenantID uuid.UUID) (map[uuid.UUID]bool, error) {
	snapshot, err := h.boundaryService.GetSnapshot(tenantID)
	if err != nil {
		return nil, err
	}
	result := make(map[uuid.UUID]bool, len(snapshot.EffectiveIDs))
	for _, actionID := range snapshot.EffectiveIDs {
		result[actionID] = true
	}
	return result, nil
}

func (h *TenantHandler) getTenantEnabledMenuSet(tenantID uuid.UUID) (map[uuid.UUID]bool, bool, error) {
	snapshot, err := h.boundaryService.GetMenuSnapshot(tenantID)
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

func (h *TenantHandler) getRoleSnapshot(teamID uuid.UUID, role *user.Role, appKey string) (*teamboundary.RoleSnapshot, error) {
	if role == nil {
		return &teamboundary.RoleSnapshot{
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
	inheritAll := role.TenantID == nil
	return h.boundaryService.GetRoleSnapshot(teamID, role.ID, inheritAll, appKey)
}

func uuidSliceToSet(ids []uuid.UUID) map[uuid.UUID]bool {
	result := make(map[uuid.UUID]bool, len(ids))
	for _, id := range ids {
		result[id] = true
	}
	return result
}

func isAssignableTeamRoleForTenant(role user.Role, tenantID uuid.UUID) bool {
	if role.TenantID != nil {
		return *role.TenantID == tenantID
	}
	return role.Code == "team_admin" || role.Code == "team_member"
}

func (h *TenantHandler) ListMyTeams(c *gin.Context) {
	userID, err := h.mustUserID(c)
	if err != nil {
		status, resp := errcode.Response(errcode.ErrUnauthorized)
		c.JSON(status, resp)
		return
	}

	tenants, err := h.tenantMemberRepo.GetTenantsByUserID(userID)
	if err != nil {
		h.logger.Error("List my teams failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取我的团队列表失败")
		c.JSON(status, resp)
		return
	}

	records := make([]gin.H, 0, len(tenants))
	for _, t := range tenants {
		adminUsers, _ := h.tenantMemberRepo.GetAdminUsersByTenantID(t.ID)
		member, memberErr := h.tenantMemberRepo.GetByUserAndTenant(userID, t.ID)
		record := tenantToMap(&t, adminUsers)
		if memberErr == nil {
			record["current_role_code"] = member.RoleCode
			record["member_status"] = member.Status
		}
		records = append(records, record)
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(records))
}

func (h *TenantHandler) mustUserID(c *gin.Context) (uuid.UUID, error) {
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

func (h *TenantHandler) resolveTenantMember(c *gin.Context) (*user.TenantMember, error) {
	userID, err := h.mustUserID(c)
	if err != nil {
		return nil, err
	}

	if tenantID, ok := parseTenantIDFromContext(c); ok {
		member, memberErr := h.tenantMemberRepo.GetByUserAndTenant(userID, tenantID)
		if memberErr != nil {
			if memberErr == gorm.ErrRecordNotFound {
				return nil, ErrTenantMemberNotFound
			}
			return nil, memberErr
		}
		return member, nil
	}

	member, err := h.tenantMemberRepo.GetByUserID(userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrTenantMemberNotFound
		}
		return nil, err
	}
	return member, nil
}

func (h *TenantHandler) resolveMyTeamRole(c *gin.Context) (*user.TenantMember, *user.Role, error) {
	member, err := h.resolveTenantMember(c)
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
	if !isAssignableTeamRoleForTenant(*role, member.TenantID) {
		return member, nil, ErrMyTeamRoleForbidden
	}
	return member, role, nil
}

func (h *TenantHandler) resolveEditableMyTeamRole(c *gin.Context) (*user.TenantMember, *user.Role, error) {
	member, role, err := h.resolveMyTeamRole(c)
	if err != nil {
		return member, role, err
	}
	if role.TenantID == nil {
		return member, role, ErrMyTeamRoleForbidden
	}
	return member, role, nil
}

func (h *TenantHandler) respondMyTeamRoleError(c *gin.Context, err error) {
	switch {
	case err == gorm.ErrRecordNotFound:
		status, resp := errcode.Response(errcode.ErrRoleNotFound)
		c.JSON(status, resp)
	case err == ErrTenantMemberNotFound:
		status, resp := errcode.Response(errcode.ErrNoTeam)
		c.JSON(status, resp)
	case errors.Is(err, ErrMyTeamRoleForbidden):
		status, resp := errcode.ResponseWithMsg(errcode.ErrForbidden, "仅支持操作当前团队自定义角色")
		c.JSON(status, resp)
	default:
		if _, parseErr := uuid.Parse(c.Param("roleId")); parseErr != nil {
			status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的角色ID")
			c.JSON(status, resp)
			return
		}
		h.logger.Error("Resolve my team role failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取团队角色失败")
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

func parseTenantIDFromContext(c *gin.Context) (uuid.UUID, bool) {
	if value, ok := c.Get("tenant_id"); ok {
		switch typed := value.(type) {
		case string:
			if tenantID, err := uuid.Parse(strings.TrimSpace(typed)); err == nil {
				return tenantID, true
			}
		case uuid.UUID:
			if typed != uuid.Nil {
				return typed, true
			}
		}
	}

	candidates := []string{
		strings.TrimSpace(c.Query("tenant_id")),
		strings.TrimSpace(c.GetHeader(tenantContextHeader)),
	}
	for _, candidate := range candidates {
		if candidate == "" {
			continue
		}
		if tenantID, err := uuid.Parse(candidate); err == nil {
			return tenantID, true
		}
	}
	return uuid.Nil, false
}

func tenantToMap(t *user.Tenant, ownerUsers []user.User) gin.H {
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

func memberToMap(m *user.TenantMember, userInfo *user.User) gin.H {
	result := gin.H{
		"id":        m.ID.String(),
		"tenant_id": m.TenantID.String(),
		"user_id":   m.UserID.String(),
		"role_code": m.RoleCode,
		"status":    m.Status,
		"joined_at": m.JoinedAt.Format("2006-01-02 15:04:05"),
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

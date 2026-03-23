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
	"github.com/gg-ecommerce/backend/internal/modules/system/user"
	"github.com/gg-ecommerce/backend/internal/pkg/database"
)

const tenantContextHeader = "X-Tenant-ID"

var ErrMyTeamRoleForbidden = errors.New("team role forbidden")

type TenantHandler struct {
	tenantService     TenantService
	tenantMemberRepo  user.TenantMemberRepository
	userRepo          user.UserRepository
	roleRepo          user.RoleRepository
	roleMenuRepo      user.RoleMenuRepository
	roleActionRepo    user.RoleActionPermissionRepository
	userRoleRepo      user.UserRoleRepository
	actionRepo        user.PermissionActionRepository
	tenantActionRepo  user.TenantActionPermissionRepository
	manualActionRepo  user.TeamManualActionPermissionRepository
	userActionRepo    user.UserActionPermissionRepository
	teamPackageRepo   user.TeamFeaturePackageRepository
	packageActionRepo user.FeaturePackageActionRepository
	logger            *zap.Logger
}

func NewTenantHandler(tenantService TenantService, tenantMemberRepo user.TenantMemberRepository, userRepo user.UserRepository, roleRepo user.RoleRepository, roleMenuRepo user.RoleMenuRepository, roleActionRepo user.RoleActionPermissionRepository, userRoleRepo user.UserRoleRepository, actionRepo user.PermissionActionRepository, tenantActionRepo user.TenantActionPermissionRepository, manualActionRepo user.TeamManualActionPermissionRepository, userActionRepo user.UserActionPermissionRepository, teamPackageRepo user.TeamFeaturePackageRepository, packageActionRepo user.FeaturePackageActionRepository, logger *zap.Logger) *TenantHandler {
	return &TenantHandler{
		tenantService:     tenantService,
		tenantMemberRepo:  tenantMemberRepo,
		userRepo:          userRepo,
		roleRepo:          roleRepo,
		roleMenuRepo:      roleMenuRepo,
		roleActionRepo:    roleActionRepo,
		userRoleRepo:      userRoleRepo,
		actionRepo:        actionRepo,
		tenantActionRepo:  tenantActionRepo,
		manualActionRepo:  manualActionRepo,
		userActionRepo:    userActionRepo,
		teamPackageRepo:   teamPackageRepo,
		packageActionRepo: packageActionRepo,
		logger:            logger,
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
	} else if v := c.Query("role"); v != "" {
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
	if strings.TrimSpace(req.RoleCode) == "" {
		req.RoleCode = req.Role
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
	if roleCode == "" {
		roleCode = strings.TrimSpace(req.Role)
	}
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
			status, resp := errcode.Response(errcode.ErrNoTeam)
			c.JSON(status, resp)
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
	} else if v := c.Query("role"); v != "" {
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
	if strings.TrimSpace(req.RoleCode) == "" {
		req.RoleCode = req.Role
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
	if roleCode == "" {
		roleCode = strings.TrimSpace(req.Role)
	}
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

	c.JSON(http.StatusOK, dto.SuccessResponse(nil))
}

func (h *TenantHandler) ListMyTeamRoles(c *gin.Context) {
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

	c.JSON(http.StatusOK, dto.SuccessResponse(nil))
}

func (h *TenantHandler) DeleteMyTeamRole(c *gin.Context) {
	member, role, err := h.resolveEditableMyTeamRole(c)
	if err != nil {
		h.respondMyTeamRoleError(c, err)
		return
	}

	if err := database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ? AND role_id = ? AND tenant_id = ?", member.UserID, role.ID, member.TenantID).Delete(&user.UserRole{}).Error; err != nil {
			return err
		}
		if err := tx.Where("role_id = ?", role.ID).Delete(&user.RoleMenu{}).Error; err != nil {
			return err
		}
		if err := tx.Where("role_id = ?", role.ID).Delete(&user.RoleActionPermission{}).Error; err != nil {
			return err
		}
		if err := tx.Where("role_id = ?", role.ID).Delete(&user.RoleDataPermission{}).Error; err != nil {
			return err
		}
		return tx.Delete(&user.Role{}, role.ID).Error
	}); err != nil {
		h.logger.Error("Delete team role failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "删除团队角色失败")
		c.JSON(status, resp)
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(nil))
}

func (h *TenantHandler) GetMyTeamRoleMenus(c *gin.Context) {
	_, role, err := h.resolveMyTeamRole(c)
	if err != nil {
		h.respondMyTeamRoleError(c, err)
		return
	}

	menuIDs, err := h.roleMenuRepo.GetMenuIDsByRoleID(role.ID)
	if err != nil {
		h.logger.Error("Get team role menus failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取团队角色菜单权限失败")
		c.JSON(status, resp)
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{"menu_ids": actionIDsToStrings(menuIDs)}))
}

func (h *TenantHandler) SetMyTeamRoleMenus(c *gin.Context) {
	_, role, err := h.resolveEditableMyTeamRole(c)
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

	menuIDs, parseErr := parseUUIDSlice(req.MenuIDs)
	if parseErr != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的菜单ID")
		c.JSON(status, resp)
		return
	}

	if err := h.roleMenuRepo.SetRoleMenus(role.ID, menuIDs); err != nil {
		h.logger.Error("Set team role menus failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "保存团队角色菜单权限失败")
		c.JSON(status, resp)
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(nil))
}

func (h *TenantHandler) GetMyTeamRoleActions(c *gin.Context) {
	member, role, err := h.resolveMyTeamRole(c)
	if err != nil {
		h.respondMyTeamRoleError(c, err)
		return
	}

	records, err := h.roleActionRepo.GetByRoleID(role.ID)
	if err != nil {
		h.logger.Error("Get team role actions failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取团队角色功能权限失败")
		c.JSON(status, resp)
		return
	}
	enabledActionSet, err := h.getTenantEnabledActionSet(member.TenantID)
	if err != nil {
		h.logger.Error("Get team enabled actions failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取团队功能边界失败")
		c.JSON(status, resp)
		return
	}

	actionIDs := make([]string, 0, len(records))
	for _, record := range records {
		if !enabledActionSet[record.ActionID] {
			continue
		}
		actionIDs = append(actionIDs, record.ActionID.String())
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{"action_ids": actionIDs}))
}

func (h *TenantHandler) SetMyTeamRoleActions(c *gin.Context) {
	member, role, err := h.resolveEditableMyTeamRole(c)
	if err != nil {
		h.respondMyTeamRoleError(c, err)
		return
	}

	var req dto.RoleActionPermissionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		status, resp := errcode.Response(errcode.ErrParamInvalid)
		c.JSON(status, resp)
		return
	}

	actionIDs, parseErr := parseUUIDSlice(req.ActionIDs)
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

	enabledActionSet, err := h.getTenantEnabledActionSet(member.TenantID)
	if err != nil {
		h.logger.Error("Get team enabled actions failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取团队功能边界失败")
		c.JSON(status, resp)
		return
	}
	for _, actionID := range actionIDs {
		if !enabledActionSet[actionID] {
			status, resp := errcode.ResponseWithMsg(errcode.ErrForbidden, "存在超出当前团队已开通能力范围的功能权限")
			c.JSON(status, resp)
			return
		}
	}

	actions := make([]user.RoleActionPermission, 0, len(actionIDs))
	for _, actionID := range actionIDs {
		actions = append(actions, user.RoleActionPermission{
			RoleID:   role.ID,
			ActionID: actionID,
		})
	}
	if err := h.roleActionRepo.SetRoleActions(role.ID, actions); err != nil {
		h.logger.Error("Set team role actions failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "保存团队角色功能权限失败")
		c.JSON(status, resp)
		return
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
	actionIDs, err := h.tenantActionRepo.GetEnabledActionIDsByTenantID(tenantID)
	if err != nil {
		h.logger.Error("Get tenant actions failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取团队功能权限失败")
		c.JSON(status, resp)
		return
	}
	actions, err := h.actionRepo.GetByIDs(actionIDs)
	if err != nil {
		h.logger.Error("Get permission actions failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取功能权限失败")
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"action_ids": actionIDsToStrings(actionIDs),
		"actions":    actionListToMaps(actions),
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

func (h *TenantHandler) SetTenantActions(c *gin.Context) {
	tenantID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的团队ID")
		c.JSON(status, resp)
		return
	}
	var req dto.TenantActionPermissionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		status, resp := errcode.Response(errcode.ErrParamInvalid)
		c.JSON(status, resp)
		return
	}
	actionIDs, parseErr := parseUUIDSlice(req.ActionIDs)
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
	derivedIDs, err := h.getDerivedTeamActionIDs(tenantID)
	if err != nil {
		h.logger.Error("Get derived team actions failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取功能包展开权限失败")
		c.JSON(status, resp)
		return
	}
	manualIDs := excludeActionIDs(actionIDs, derivedIDs)
	effectiveIDs := mergeActionIDs(derivedIDs, manualIDs)
	if err := h.manualActionRepo.ReplaceTenantActions(tenantID, manualIDs); err != nil {
		h.logger.Error("Set manual tenant actions failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "保存团队补充权限失败")
		c.JSON(status, resp)
		return
	}
	if err := h.tenantActionRepo.ReplaceTenantActions(tenantID, effectiveIDs); err != nil {
		h.logger.Error("Set tenant actions failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "保存团队功能权限失败")
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(nil))
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
	actionIDs, err := h.tenantActionRepo.GetEnabledActionIDsByTenantID(member.TenantID)
	if err != nil {
		h.logger.Error("Get my team actions failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取团队功能权限失败")
		c.JSON(status, resp)
		return
	}
	actions, err := h.actionRepo.GetByIDs(actionIDs)
	if err != nil {
		h.logger.Error("Get my team permission actions failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取功能权限失败")
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"action_ids": actionIDsToStrings(actionIDs),
		"actions":    actionListToMaps(actions),
	}))
}

func (h *TenantHandler) GetMyTeamActionOrigins(c *gin.Context) {
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
	h.respondTenantActionOrigins(c, member.TenantID)
}

func (h *TenantHandler) respondTenantActionOrigins(c *gin.Context, tenantID uuid.UUID) {
	actionIDs, err := h.tenantActionRepo.GetEnabledActionIDsByTenantID(tenantID)
	if err != nil {
		h.logger.Error("Get tenant action ids failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取团队功能权限来源失败")
		c.JSON(status, resp)
		return
	}
	packageIDs, err := h.teamPackageRepo.GetPackageIDsByTeamID(tenantID)
	if err != nil {
		h.logger.Error("Get team packages failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取团队功能包失败")
		c.JSON(status, resp)
		return
	}
	derivedIDs, err := h.getDerivedTeamActionIDs(tenantID)
	if err != nil {
		h.logger.Error("Get derived actions failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取功能包权限失败")
		c.JSON(status, resp)
		return
	}
	manualIDs, err := h.manualActionRepo.GetEnabledActionIDsByTenantID(tenantID)
	if err != nil {
		h.logger.Error("Get manual actions failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取团队补充权限失败")
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"package_ids":          actionIDsToStrings(packageIDs),
		"derived_action_ids":   actionIDsToStrings(derivedIDs),
		"manual_action_ids":    actionIDsToStrings(manualIDs),
		"effective_action_ids": actionIDsToStrings(actionIDs),
	}))
}

func (h *TenantHandler) getDerivedTeamActionIDs(tenantID uuid.UUID) ([]uuid.UUID, error) {
	packageIDs, err := h.teamPackageRepo.GetPackageIDsByTeamID(tenantID)
	if err != nil {
		return nil, err
	}
	result := make([]uuid.UUID, 0)
	seen := make(map[uuid.UUID]struct{})
	for _, packageID := range packageIDs {
		actionIDs, err := h.packageActionRepo.GetActionIDsByPackageID(packageID)
		if err != nil {
			return nil, err
		}
		for _, actionID := range actionIDs {
			if _, ok := seen[actionID]; ok {
				continue
			}
			seen[actionID] = struct{}{}
			result = append(result, actionID)
		}
	}
	return result, nil
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
	actionIDs, err := h.tenantActionRepo.GetEnabledActionIDsByTenantID(tenantID)
	if err != nil {
		return nil, err
	}
	result := make(map[uuid.UUID]bool, len(actionIDs))
	for _, actionID := range actionIDs {
		result[actionID] = true
	}
	return result, nil
}

func isAssignableTeamRoleForTenant(role user.Role, tenantID uuid.UUID) bool {
	if role.TenantID != nil {
		return *role.TenantID == tenantID
	}
	return role.Code == "team_admin" || role.Code == "team_member"
}

func (h *TenantHandler) GetMyTeamMemberActionPermissions(c *gin.Context) {
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
	targetUserID, err := uuid.Parse(c.Param("userId"))
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
	records, err := h.userActionRepo.GetByUserAndTenant(targetUserID, &member.TenantID)
	if err != nil {
		h.logger.Error("Get member action overrides failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取成员功能权限失败")
		c.JSON(status, resp)
		return
	}
	enabledActionSet, err := h.getTenantEnabledActionSet(member.TenantID)
	if err != nil {
		h.logger.Error("Get team enabled actions failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取团队功能边界失败")
		c.JSON(status, resp)
		return
	}
	actionIDs := make([]uuid.UUID, 0, len(records))
	for _, record := range records {
		if !enabledActionSet[record.ActionID] {
			continue
		}
		actionIDs = append(actionIDs, record.ActionID)
	}
	actions, err := h.actionRepo.GetByIDs(actionIDs)
	if err != nil {
		h.logger.Error("Get action detail failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取功能权限失败")
		c.JSON(status, resp)
		return
	}
	actionMap := make(map[uuid.UUID]user.PermissionAction, len(actions))
	for _, action := range actions {
		actionMap[action.ID] = action
	}
	items := make([]gin.H, 0, len(records))
	for _, record := range records {
		if !enabledActionSet[record.ActionID] {
			continue
		}
		item := gin.H{
			"action_id": record.ActionID.String(),
			"effect":    record.Effect,
		}
		if action, ok := actionMap[record.ActionID]; ok {
			item["action"] = actionMapToMap(&action)
		}
		items = append(items, item)
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{"actions": items}))
}

func (h *TenantHandler) SetMyTeamMemberActionPermissions(c *gin.Context) {
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
	targetUserID, err := uuid.Parse(c.Param("userId"))
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
	var req dto.UserActionPermissionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		status, resp := errcode.Response(errcode.ErrParamInvalid)
		c.JSON(status, resp)
		return
	}
	actions := make([]user.UserActionPermission, 0, len(req.Actions))
	actionIDs := make([]uuid.UUID, 0, len(req.Actions))
	for _, item := range req.Actions {
		actionID, parseErr := uuid.Parse(item.ActionID)
		if parseErr != nil {
			status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的功能权限ID")
			c.JSON(status, resp)
			return
		}
		actionIDs = append(actionIDs, actionID)
		actions = append(actions, user.UserActionPermission{
			UserID:   targetUserID,
			ActionID: actionID,
			TenantID: &member.TenantID,
			Effect:   item.Effect,
		})
	}
	if len(actionIDs) > 0 {
		actionList, actionErr := h.actionRepo.GetByIDs(actionIDs)
		if actionErr != nil {
			h.logger.Error("Get member action detail failed", zap.Error(actionErr))
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
	enabledActionSet, err := h.getTenantEnabledActionSet(member.TenantID)
	if err != nil {
		h.logger.Error("Get team enabled actions failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取团队功能边界失败")
		c.JSON(status, resp)
		return
	}
	for _, actionID := range actionIDs {
		if !enabledActionSet[actionID] {
			status, resp := errcode.ResponseWithMsg(errcode.ErrForbidden, "存在超出当前团队已开通能力范围的功能权限")
			c.JSON(status, resp)
			return
		}
	}
	if err := h.userActionRepo.ReplaceUserActions(targetUserID, &member.TenantID, actions); err != nil {
		h.logger.Error("Set member action overrides failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "保存成员功能权限失败")
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(nil))
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

	return h.tenantMemberRepo.GetByUserID(userID)
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

func actionMapToMap(action *user.PermissionAction) gin.H {
	return gin.H{
		"id":            action.ID.String(),
		"resource_code": action.ResourceCode,
		"action_code":   action.ActionCode,
		"name":          action.Name,
		"description":   action.Description,
		"status":        action.Status,
		"sort_order":    action.SortOrder,
		"created_at":    action.CreatedAt.Format("2006-01-02 15:04:05"),
		"updated_at":    action.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

func actionListToMaps(actions []user.PermissionAction) []gin.H {
	items := make([]gin.H, 0, len(actions))
	for _, action := range actions {
		items = append(items, actionMapToMap(&action))
	}
	return items
}

func actionIDsToStrings(ids []uuid.UUID) []string {
	items := make([]string, 0, len(ids))
	for _, id := range ids {
		items = append(items, id.String())
	}
	return items
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

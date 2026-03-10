package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/gg-ecommerce/backend/internal/api/dto"
	"github.com/gg-ecommerce/backend/internal/api/errcode"
	"github.com/gg-ecommerce/backend/internal/model"
	"github.com/gg-ecommerce/backend/internal/repository"
	"github.com/gg-ecommerce/backend/internal/service"
)

// TenantHandler 团队管理处理器
type TenantHandler struct {
	tenantService    service.TenantService
	tenantMemberRepo repository.TenantMemberRepository
	userRepo         repository.UserRepository
	roleRepo         repository.RoleRepository
	userRoleRepo     repository.UserRoleRepository
	logger           *zap.Logger
}

// NewTenantHandler 创建团队处理器
func NewTenantHandler(tenantService service.TenantService, tenantMemberRepo repository.TenantMemberRepository, userRepo repository.UserRepository, roleRepo repository.RoleRepository, userRoleRepo repository.UserRoleRepository, logger *zap.Logger) *TenantHandler {
	return &TenantHandler{tenantService: tenantService, tenantMemberRepo: tenantMemberRepo, userRepo: userRepo, roleRepo: roleRepo, userRoleRepo: userRoleRepo, logger: logger}
}

// List 团队列表
func (h *TenantHandler) List(c *gin.Context) {
	var req dto.TenantListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		status, resp := errcode.Response(errcode.ErrParamInvalid); c.JSON(status, resp)
		return
	}
	list, total, err := h.tenantService.List(&req)
	if err != nil {
		h.logger.Error("Tenant list failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取团队列表失败"); c.JSON(status, resp)
		return
	}
	records := make([]gin.H, 0, len(list))
	for _, t := range list {
		records = append(records, tenantToMap(&t))
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"records": records,
		"total":   total,
		"current": req.Current,
		"size":    req.Size,
	}))
}

// Get 团队详情
func (h *TenantHandler) Get(c *gin.Context) {
	idStr := c.Param("id")
	if strings.TrimSpace(strings.ToLower(idStr)) == "my-team" {
		h.GetMyTeam(c)
		return
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的团队ID"); c.JSON(status, resp)
		return
	}
	t, err := h.tenantService.Get(id)
	if err != nil {
		if err == service.ErrTenantNotFound {
			status, resp := errcode.Response(errcode.ErrTenantNotFound); c.JSON(status, resp)
			return
		}
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取团队失败"); c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(tenantToMap(t)))
}

// Create 创建团队
func (h *TenantHandler) Create(c *gin.Context) {
	var req dto.TenantCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		status, resp := errcode.Response(errcode.ErrParamInvalid); c.JSON(status, resp)
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
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, err.Error()); c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{"id": t.ID.String()}))
}

// Update 更新团队
func (h *TenantHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的团队ID"); c.JSON(status, resp)
		return
	}
	var req dto.TenantUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		status, resp := errcode.Response(errcode.ErrParamInvalid); c.JSON(status, resp)
		return
	}
	if err := h.tenantService.Update(id, &req); err != nil {
		if err == service.ErrTenantNotFound {
			status, resp := errcode.Response(errcode.ErrTenantNotFound); c.JSON(status, resp)
			return
		}
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "更新团队失败"); c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(nil))
}

// Delete 删除团队
func (h *TenantHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的团队ID"); c.JSON(status, resp)
		return
	}
	if err := h.tenantService.Delete(id); err != nil {
		if err == service.ErrTenantNotFound {
			status, resp := errcode.Response(errcode.ErrTenantNotFound); c.JSON(status, resp)
			return
		}
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "删除团队失败"); c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(nil))
}

// ListMembers 团队成员列表（带用户信息）
func (h *TenantHandler) ListMembers(c *gin.Context) {
	idStr := c.Param("id")
	if strings.TrimSpace(strings.ToLower(idStr)) == "my-team" {
		h.ListMyMembers(c)
		return
	}
	tenantID, err := uuid.Parse(idStr)
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的团队ID"); c.JSON(status, resp)
		return
	}
	members, err := h.tenantService.ListMembers(tenantID)
	if err != nil {
		if err == service.ErrTenantNotFound {
			status, resp := errcode.Response(errcode.ErrTenantNotFound); c.JSON(status, resp)
			return
		}
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取成员列表失败"); c.JSON(status, resp)
		return
	}
	roleCodes, _ := h.userRoleRepo.GetScopeTeamRoleCodesByTenant(tenantID)
	records := make([]gin.H, 0, len(members))
	for _, m := range members {
		role := roleCodes[m.UserID]
		if role == "" {
			role = "team_member"
		}
		item := gin.H{
			"id":         m.ID.String(),
			"userId":     m.UserID.String(),
			"role":       role,
			"status":     m.Status,
			"joinedAt":   nil,
			"userName":   "",
			"nickName":   "",
			"userEmail":  "",
		}
		if m.JoinedAt != nil {
			item["joinedAt"] = m.JoinedAt.Format("2006-01-02 15:04:05")
		}
		user, err := h.userRepo.GetByID(m.UserID)
		if err == nil {
			item["userName"] = user.Username
			item["nickName"] = user.Nickname
			item["userEmail"] = user.Email
		}
		records = append(records, item)
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{"records": records}))
}

// AddMember 添加团队成员
func (h *TenantHandler) AddMember(c *gin.Context) {
	idStr := c.Param("id")
	if strings.TrimSpace(strings.ToLower(idStr)) == "my-team" {
		h.AddMyMember(c)
		return
	}
	tenantID, err := uuid.Parse(idStr)
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的团队ID"); c.JSON(status, resp)
		return
	}
	var req dto.TenantAddMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		status, resp := errcode.Response(errcode.ErrParamInvalid); c.JSON(status, resp)
		return
	}
	var invitedBy *uuid.UUID
	if uid, ok := c.Get("user_id"); ok {
		if s, ok := uid.(string); ok {
			if parsed, err := uuid.Parse(s); err == nil {
				invitedBy = &parsed
			}
		}
	}
	if err := h.tenantService.AddMember(tenantID, &req, invitedBy); err != nil {
		if err == service.ErrTenantNotFound {
			status, resp := errcode.Response(errcode.ErrTenantNotFound); c.JSON(status, resp)
			return
		}
		if err == service.ErrTenantMemberExists {
			status, resp := errcode.Response(errcode.ErrMemberExists); c.JSON(status, resp)
			return
		}
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, err.Error()); c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(nil))
}

// RemoveMember 移除团队成员
func (h *TenantHandler) RemoveMember(c *gin.Context) {
	idStr := c.Param("id")
	if strings.TrimSpace(strings.ToLower(idStr)) == "my-team" {
		h.RemoveMyMember(c)
		return
	}
	tenantID, err := uuid.Parse(idStr)
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的团队ID"); c.JSON(status, resp)
		return
	}
	userIDStr := c.Param("userId")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的用户ID"); c.JSON(status, resp)
		return
	}
	if err := h.tenantService.RemoveMember(tenantID, userID); err != nil {
		if err == service.ErrTenantMemberNotFound {
			status, resp := errcode.Response(errcode.ErrMemberNotFound); c.JSON(status, resp)
			return
		}
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "移除成员失败"); c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(nil))
}

// UpdateMemberRole 更新成员角色
func (h *TenantHandler) UpdateMemberRole(c *gin.Context) {
	idStr := c.Param("id")
	if strings.TrimSpace(strings.ToLower(idStr)) == "my-team" {
		h.UpdateMyMemberRole(c)
		return
	}
	tenantID, err := uuid.Parse(idStr)
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的团队ID"); c.JSON(status, resp)
		return
	}
	userIDStr := c.Param("userId")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的用户ID"); c.JSON(status, resp)
		return
	}
	var req dto.TenantMemberRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		status, resp := errcode.Response(errcode.ErrParamInvalid); c.JSON(status, resp)
		return
	}
	if err := h.tenantService.UpdateMemberRole(tenantID, userID, req.Role); err != nil {
		if err == service.ErrTenantMemberNotFound {
			status, resp := errcode.Response(errcode.ErrMemberNotFound); c.JSON(status, resp)
			return
		}
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "更新角色失败"); c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(nil))
}

// getCurrentUserID 从 context 获取当前用户 ID，失败返回 nil
func (h *TenantHandler) getCurrentUserID(c *gin.Context) *uuid.UUID {
	uid, ok := c.Get("user_id")
	if !ok {
		return nil
	}
	s, ok := uid.(string)
	if !ok || s == "" {
		return nil
	}
	parsed, err := uuid.Parse(s)
	if err != nil {
		return nil
	}
	return &parsed
}

// resolveMyTeamID 解析当前用户有管理权限的团队 ID，失败返回 false。
// 若用户是超级管理员且不是任何团队的 team_admin，则使用系统中第一个团队，以便管理员仍可操作「我的团队」相关接口（如角色菜单权限）。
func (h *TenantHandler) resolveMyTeamID(c *gin.Context) (uuid.UUID, bool) {
	userID := h.getCurrentUserID(c)
	if userID == nil {
		return uuid.Nil, false
	}
	tenantID, err := h.tenantMemberRepo.GetFirstManagedTenantID(*userID)
	if err == nil && tenantID != uuid.Nil {
		return tenantID, true
	}
	// 超级管理员：若无已管理的团队，则取系统中第一个团队作为可操作上下文
	user, err := h.userRepo.GetByID(*userID)
	if err != nil || user == nil || !user.IsSuperAdmin {
		return uuid.Nil, false
	}
	list, _, err := h.tenantService.List(&dto.TenantListRequest{Current: 1, Size: 1})
	if err != nil || len(list) == 0 {
		return uuid.Nil, false
	}
	return list[0].ID, true
}

// GetMyTeam 获取当前用户有管理权限的团队（team_admin 用）
func (h *TenantHandler) GetMyTeam(c *gin.Context) {
	tenantID, ok := h.resolveMyTeamID(c)
	if !ok {
		status, resp := errcode.Response(errcode.ErrNoManagedTeam); c.JSON(status, resp)
		return
	}
	t, err := h.tenantService.Get(tenantID)
	if err != nil {
		if err == service.ErrTenantNotFound {
			status, resp := errcode.Response(errcode.ErrTenantNotFound); c.JSON(status, resp)
			return
		}
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取团队失败"); c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(tenantToMap(t)))
}

// ListMyMembers 获取「我的团队」成员列表
func (h *TenantHandler) ListMyMembers(c *gin.Context) {
	tenantID, ok := h.resolveMyTeamID(c)
	if !ok {
		status, resp := errcode.Response(errcode.ErrNoManagedTeam); c.JSON(status, resp)
		return
	}
	members, err := h.tenantService.ListMembers(tenantID)
	if err != nil {
		if err == service.ErrTenantNotFound {
			status, resp := errcode.Response(errcode.ErrTenantNotFound); c.JSON(status, resp)
			return
		}
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取成员列表失败"); c.JSON(status, resp)
		return
	}
	roleCodes, _ := h.userRoleRepo.GetScopeTeamRoleCodesByTenant(tenantID)
	records := make([]gin.H, 0, len(members))
	for _, m := range members {
		role := roleCodes[m.UserID]
		if role == "" {
			role = "team_member"
		}
		item := gin.H{
			"id":        m.ID.String(),
			"userId":    m.UserID.String(),
			"role":      role,
			"status":    m.Status,
			"joinedAt":  nil,
			"userName":  "",
			"nickName":  "",
			"userEmail": "",
		}
		if m.JoinedAt != nil {
			item["joinedAt"] = m.JoinedAt.Format("2006-01-02 15:04:05")
		}
		user, err := h.userRepo.GetByID(m.UserID)
		if err == nil {
			item["userName"] = user.Username
			item["nickName"] = user.Nickname
			item["userEmail"] = user.Email
		}
		records = append(records, item)
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{"records": records}))
}

// AddMyMember 向「我的团队」添加成员
func (h *TenantHandler) AddMyMember(c *gin.Context) {
	tenantID, ok := h.resolveMyTeamID(c)
	if !ok {
		status, resp := errcode.Response(errcode.ErrNoManagedTeam); c.JSON(status, resp)
		return
	}
	var req dto.TenantAddMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		status, resp := errcode.Response(errcode.ErrParamInvalid); c.JSON(status, resp)
		return
	}
	var invitedBy *uuid.UUID
	if uid := h.getCurrentUserID(c); uid != nil {
		invitedBy = uid
	}
	if err := h.tenantService.AddMember(tenantID, &req, invitedBy); err != nil {
		if err == service.ErrTenantNotFound {
			status, resp := errcode.Response(errcode.ErrTenantNotFound); c.JSON(status, resp)
			return
		}
		if err == service.ErrTenantMemberExists {
			status, resp := errcode.Response(errcode.ErrMemberExists); c.JSON(status, resp)
			return
		}
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, err.Error()); c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(nil))
}

// RemoveMyMember 从「我的团队」移除成员
func (h *TenantHandler) RemoveMyMember(c *gin.Context) {
	tenantID, ok := h.resolveMyTeamID(c)
	if !ok {
		status, resp := errcode.Response(errcode.ErrNoManagedTeam); c.JSON(status, resp)
		return
	}
	userIDStr := c.Param("userId")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的用户ID"); c.JSON(status, resp)
		return
	}
	if err := h.tenantService.RemoveMember(tenantID, userID); err != nil {
		if err == service.ErrTenantMemberNotFound {
			status, resp := errcode.Response(errcode.ErrMemberNotFound); c.JSON(status, resp)
			return
		}
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "移除成员失败"); c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(nil))
}

// UpdateMyMemberRole 更新「我的团队」成员角色
func (h *TenantHandler) UpdateMyMemberRole(c *gin.Context) {
	tenantID, ok := h.resolveMyTeamID(c)
	if !ok {
		status, resp := errcode.Response(errcode.ErrNoManagedTeam); c.JSON(status, resp)
		return
	}
	userIDStr := c.Param("userId")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的用户ID"); c.JSON(status, resp)
		return
	}
	var req dto.TenantMemberRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		status, resp := errcode.Response(errcode.ErrParamInvalid); c.JSON(status, resp)
		return
	}
	if err := h.tenantService.UpdateMemberRole(tenantID, userID, req.Role); err != nil {
		if err == service.ErrTenantMemberNotFound {
			status, resp := errcode.Response(errcode.ErrMemberNotFound); c.JSON(status, resp)
			return
		}
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "更新角色失败"); c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(nil))
}

// GetMyTeamMemberRoles 获取「我的团队」某成员在本团队内的角色（仅全局 scope=team 角色）
func (h *TenantHandler) GetMyTeamMemberRoles(c *gin.Context) {
	tenantID, ok := h.resolveMyTeamID(c)
	if !ok {
		status, resp := errcode.Response(errcode.ErrNoManagedTeam); c.JSON(status, resp)
		return
	}
	userIDStr := c.Param("userId")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的用户ID"); c.JSON(status, resp)
		return
	}
	globalRoleIDs, _ := h.userRoleRepo.GetRoleIDsByUserAndTenant(userID, &tenantID)
	gIds := make([]string, 0, len(globalRoleIDs))
	for _, id := range globalRoleIDs {
		gIds = append(gIds, id.String())
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"global_role_ids": gIds,
		"role_ids":        gIds,
	}))
}

// SetMyTeamMemberRoles 设置「我的团队」某成员在本团队内的角色（仅支持全局 scope=team 角色）
func (h *TenantHandler) SetMyTeamMemberRoles(c *gin.Context) {
	tenantID, ok := h.resolveMyTeamID(c)
	if !ok {
		status, resp := errcode.Response(errcode.ErrNoManagedTeam); c.JSON(status, resp)
		return
	}
	userIDStr := c.Param("userId")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的用户ID"); c.JSON(status, resp)
		return
	}
	var req dto.TenantMemberRolesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		status, resp := errcode.Response(errcode.ErrParamInvalid); c.JSON(status, resp)
		return
	}
	var globalRoleIDs []uuid.UUID
	for _, s := range req.RoleIDs {
		if s == "" {
			continue
		}
		id, err := uuid.Parse(s)
		if err != nil {
			status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的角色ID: "+s); c.JSON(status, resp)
			return
		}
		if role, _ := h.roleRepo.GetByID(id); role != nil && role.Scope.Code == "team" {
			globalRoleIDs = append(globalRoleIDs, id)
		}
	}
	if err := h.userRoleRepo.ReplaceRolesForTenant(userID, tenantID, globalRoleIDs); err != nil {
		h.logger.Error("Set my-team member roles failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "设置成员角色失败"); c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(nil))
}

// ListMyTeamRoles 获取「我的团队」可见的角色：仅返回全局 scope=team 角色（不再支持团队自建角色）
func (h *TenantHandler) ListMyTeamRoles(c *gin.Context) {
	_, ok := h.resolveMyTeamID(c)
	if !ok {
		status, resp := errcode.Response(errcode.ErrNoManagedTeam); c.JSON(status, resp)
		return
	}
	globalRoles, _, err := h.roleRepo.ListByScope("team", 0, 1000, "", "", "", "", "", nil)
	if err != nil {
		h.logger.Error("List my-team roles failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取角色列表失败"); c.JSON(status, resp)
		return
	}
	records := make([]gin.H, 0, len(globalRoles))
	for _, r := range globalRoles {
		scopeCode := ""
		scopeName := ""
		scopeId := ""
		if r.Scope.ID != (uuid.UUID{}) {
			scopeCode = r.Scope.Code
			scopeName = r.Scope.Name
			scopeId = r.Scope.ID.String()
		}
		records = append(records, gin.H{
			"roleId":            r.ID.String(),
			"roleName":          r.Name,
			"roleCode":          r.Code,
			"description":       r.Description,
			"enabled":           r.Enabled,
			"createTime":        r.CreatedAt.Format("2006-01-02 15:04:05"),
			"isGlobal":          true,
			"scopeId":           scopeId,
			"scopeCode":         scopeCode,
			"scopeName":         scopeName,
			"scope":             scopeCode,
			"canEditPermission": false,
			"canEdit":           false,
			"canDelete":         false,
		})
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{"records": records}))
}

func tenantIDPtrToStr(p *uuid.UUID) interface{} {
	if p == nil {
		return nil
	}
	return p.String()
}

func tenantToMap(t *model.Tenant) gin.H {
	m := gin.H{
		"id":          t.ID.String(),
		"name":        t.Name,
		"remark":      t.Remark,
		"logoUrl":     t.LogoURL,
		"plan":        t.Plan,
		"maxMembers":  t.MaxMembers,
		"maxProducts": t.MaxProducts,
		"status":      t.Status,
		"createTime":  t.CreatedAt.Format("2006-01-02 15:04:05"),
		"updateTime":  t.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
	if t.OwnerID != nil {
		m["ownerId"] = t.OwnerID.String()
	}
	return m
}

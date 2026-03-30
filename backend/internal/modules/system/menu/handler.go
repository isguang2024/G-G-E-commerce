package menu

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
	spaceutil "github.com/gg-ecommerce/backend/internal/modules/system/space"
	"github.com/gg-ecommerce/backend/internal/modules/system/user"
	"github.com/gg-ecommerce/backend/internal/pkg/platformaccess"
	"github.com/gg-ecommerce/backend/internal/pkg/teamboundary"
)

const tenantContextHeader = "X-Tenant-ID"

type MenuHandler struct {
	db          *gorm.DB
	menuService MenuService
	userRepo    user.UserRepository
	menuRepo    interface {
		ListAll() ([]user.Menu, error)
	}
	roleRepo        user.RoleRepository
	userRoleRepo    user.UserRoleRepository
	platformService platformaccess.Service
	boundaryService teamboundary.Service
	authzService    menuAuthzService
	logger          *zap.Logger
}

type menuAuthzService interface {
	Authorize(userID uuid.UUID, tenantID *uuid.UUID, permissionKey string, legacy ...string) (bool, *models.PermissionKey, error)
}

func NewMenuHandler(db *gorm.DB, menuService MenuService, userRepo user.UserRepository, menuRepo interface {
	ListAll() ([]user.Menu, error)
}, roleRepo user.RoleRepository, userRoleRepo user.UserRoleRepository, boundaryService teamboundary.Service, authzService menuAuthzService, platformService platformaccess.Service, logger *zap.Logger) *MenuHandler {
	return &MenuHandler{
		db:              db,
		menuService:     menuService,
		userRepo:        userRepo,
		menuRepo:        menuRepo,
		roleRepo:        roleRepo,
		userRoleRepo:    userRoleRepo,
		platformService: platformService,
		boundaryService: boundaryService,
		authzService:    authzService,
		logger:          logger,
	}
}

func (h *MenuHandler) GetTree(c *gin.Context) {
	all := c.Query("all") == "1" || c.Query("all") == "true"
	spaceKey := ""
	if all && h.authzService != nil {
		userID, err := currentUserID(c)
		if err != nil {
			status, resp := errcode.Response(errcode.ErrUnauthorized)
			c.JSON(status, resp)
			return
		}
		tenantID, err := currentTenantID(c)
		if err != nil {
			status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的团队ID")
			c.JSON(status, resp)
			return
		}
		allowed, _, authErr := h.authzService.Authorize(userID, tenantID, "system.menu.manage")
		if authErr != nil || !allowed {
			status, resp := errcode.ResponseWithMsg(errcode.ErrForbidden, "无权限查看全部菜单")
			c.JSON(status, resp)
			return
		}
	}
	if all {
		spaceKey = spaceutil.RequestSpaceKey(c)
		if strings.TrimSpace(spaceKey) == "" {
			spaceKey = spaceutil.DefaultMenuSpaceKey
		}
	}

	var allowedMenuIDs []uuid.UUID
	if !all {
		userID, err := currentUserID(c)
		if err == nil {
			tenantID, tenantErr := currentTenantID(c)
			if tenantErr == nil {
				allowedMenuIDs, _ = h.getAllowedMenuIDs(userID, tenantID)
			}
		}
		spaceKey = spaceutil.ResolveSpaceKey(h.db, c)
	}
	tree, err := h.menuService.GetTree(all, allowedMenuIDs, spaceKey)
	if err != nil {
		h.logger.Error("Menu tree failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取菜单失败")
		c.JSON(status, resp)
		return
	}
	out := make([]gin.H, 0, len(tree))
	for _, node := range tree {
		if all {
			out = append(out, menuToMap(node))
			continue
		}
		out = append(out, menuToRuntimeMap(node))
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(out))
}

func (h *MenuHandler) getAllowedMenuIDs(userID uuid.UUID, tenantID *uuid.UUID) ([]uuid.UUID, error) {
	if tenantID != nil && h.userRoleRepo != nil {
		roleIDs, err := h.userRoleRepo.GetEffectiveActiveRoleIDsByUserAndTenant(userID, tenantID)
		if err != nil {
			return nil, err
		}
		return h.getTeamContextAllowedMenuIDs(*tenantID, roleIDs)
	}
	if tenantID == nil && h.platformService != nil {
		snapshot, err := h.platformService.GetSnapshot(userID)
		if err != nil {
			return nil, err
		}
		return snapshot.MenuIDs, nil
	}
	return []uuid.UUID{}, nil
}

func isMenuEnabled(menu user.Menu) bool {
	if menu.Meta == nil {
		return true
	}
	if enabled, ok := menu.Meta["isEnable"].(bool); ok {
		return enabled
	}
	return true
}

func (h *MenuHandler) getTeamContextAllowedMenuIDs(teamID uuid.UUID, roleIDs []uuid.UUID) ([]uuid.UUID, error) {
	if len(roleIDs) == 0 || h.roleRepo == nil || h.boundaryService == nil {
		return []uuid.UUID{}, nil
	}
	roles, err := h.roleRepo.GetByIDs(roleIDs)
	if err != nil {
		return nil, err
	}
	roleMap := make(map[uuid.UUID]user.Role, len(roles))
	for _, role := range roles {
		roleMap[role.ID] = role
	}
	allowedSet := make(map[uuid.UUID]struct{})
	for _, roleID := range roleIDs {
		role, ok := roleMap[roleID]
		if !ok {
			continue
		}
		snapshot, snapshotErr := h.boundaryService.GetRoleSnapshot(teamID, role.ID, role.TenantID == nil)
		if snapshotErr != nil {
			return nil, snapshotErr
		}
		for _, menuID := range snapshot.MenuIDs {
			allowedSet[menuID] = struct{}{}
		}
	}
	if len(allowedSet) == 0 {
		return []uuid.UUID{}, nil
	}
	allowedMenuIDs := make([]uuid.UUID, 0, len(allowedSet))
	for menuID := range allowedSet {
		allowedMenuIDs = append(allowedMenuIDs, menuID)
	}
	return allowedMenuIDs, nil
}

func (h *MenuHandler) Create(c *gin.Context) {
	var req dto.MenuCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		status, resp := errcode.Response(errcode.ErrParamInvalid)
		c.JSON(status, resp)
		return
	}
	m, err := h.menuService.Create(&req)
	if err != nil {
		h.logger.Error("Menu create failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "创建菜单失败")
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{"id": m.ID.String()}))
}

func (h *MenuHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的菜单ID")
		c.JSON(status, resp)
		return
	}
	var req dto.MenuUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Menu update bind error", zap.Error(err))
		status, resp := errcode.Response(errcode.ErrParamInvalid)
		c.JSON(status, resp)
		return
	}
	if err := h.menuService.Update(id, &req); err != nil {
		if err == ErrMenuNotFound {
			status, resp := errcode.Response(errcode.ErrMenuNotFound)
			c.JSON(status, resp)
			return
		}
		if err.Error() == "不能将上级设为自己" || err.Error() == "不能将上级设为自身子级（会造成循环）" {
			status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidParent, err.Error())
			c.JSON(status, resp)
			return
		}
		h.logger.Error("Menu update failed", zap.String("error", err.Error()), zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "更新菜单失败: "+err.Error())
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(nil))
}

func (h *MenuHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的菜单ID")
		c.JSON(status, resp)
		return
	}
	if err := h.menuService.Delete(id); err != nil {
		if err == ErrMenuNotFound {
			status, resp := errcode.Response(errcode.ErrMenuNotFound)
			c.JSON(status, resp)
			return
		}
		if err == ErrMenuSystemProtected {
			status, resp := errcode.Response(errcode.ErrMenuSystemProtected)
			c.JSON(status, resp)
			return
		}
		h.logger.Error("Menu delete failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "删除菜单失败: "+err.Error())
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(nil))
}

func (h *MenuHandler) ListGroups(c *gin.Context) {
	groups, err := h.menuService.ListGroups()
	if err != nil {
		h.logger.Error("List menu groups failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取菜单分组失败")
		c.JSON(status, resp)
		return
	}
	out := make([]gin.H, 0, len(groups))
	for _, item := range groups {
		out = append(out, gin.H{
			"id":         item.ID.String(),
			"name":       item.Name,
			"sort_order": item.SortOrder,
			"status":     item.Status,
			"created_at": item.CreatedAt,
			"updated_at": item.UpdatedAt,
		})
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(out))
}

func (h *MenuHandler) CreateGroup(c *gin.Context) {
	var req dto.MenuManageGroupCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		status, resp := errcode.Response(errcode.ErrParamInvalid)
		c.JSON(status, resp)
		return
	}
	group, err := h.menuService.CreateGroup(&req)
	if err != nil {
		h.logger.Error("Create menu group failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "创建菜单分组失败")
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{"id": group.ID.String()}))
}

func (h *MenuHandler) UpdateGroup(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的菜单分组ID")
		c.JSON(status, resp)
		return
	}
	var req dto.MenuManageGroupUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		status, resp := errcode.Response(errcode.ErrParamInvalid)
		c.JSON(status, resp)
		return
	}
	if err := h.menuService.UpdateGroup(id, &req); err != nil {
		if err == ErrMenuGroupNotFound {
			status, resp := errcode.ResponseWithMsg(errcode.ErrNotFound, "菜单分组不存在")
			c.JSON(status, resp)
			return
		}
		h.logger.Error("Update menu group failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "更新菜单分组失败")
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(nil))
}

func (h *MenuHandler) DeleteGroup(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的菜单分组ID")
		c.JSON(status, resp)
		return
	}
	if err := h.menuService.DeleteGroup(id); err != nil {
		if err == ErrMenuGroupNotFound {
			status, resp := errcode.ResponseWithMsg(errcode.ErrNotFound, "菜单分组不存在")
			c.JSON(status, resp)
			return
		}
		if err == ErrMenuGroupInUse {
			status, resp := errcode.ResponseWithMsg(errcode.ErrConflict, "菜单分组下仍有关联菜单，无法删除")
			c.JSON(status, resp)
			return
		}
		h.logger.Error("Delete menu group failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "删除菜单分组失败")
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(nil))
}

// 菜单备份相关接口
func (h *MenuHandler) CreateBackup(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		status, resp := errcode.Response(errcode.ErrParamInvalid)
		c.JSON(status, resp)
		return
	}

	// 获取当前用户ID
	var createdBy *uuid.UUID
	if userIDStr, ok := c.Get("user_id"); ok {
		if idStr, ok := userIDStr.(string); ok {
			if userID, err := uuid.Parse(idStr); err == nil {
				createdBy = &userID
			}
		}
	}

	if err := h.menuService.CreateBackup(req.Name, req.Description, createdBy); err != nil {
		h.logger.Error("Create backup failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "创建备份失败")
		c.JSON(status, resp)
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(nil))
}

func (h *MenuHandler) ListBackups(c *gin.Context) {
	backups, err := h.menuService.ListBackups()
	if err != nil {
		h.logger.Error("List backups failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取备份列表失败")
		c.JSON(status, resp)
		return
	}

	out := make([]gin.H, 0, len(backups))
	for _, backup := range backups {
		out = append(out, gin.H{
			"id":          backup.ID.String(),
			"name":        backup.Name,
			"description": backup.Description,
			"created_at":  backup.CreatedAt,
			"created_by":  backup.CreatedBy,
		})
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(out))
}

func (h *MenuHandler) DeleteBackup(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的备份ID")
		c.JSON(status, resp)
		return
	}

	if err := h.menuService.DeleteBackup(id); err != nil {
		h.logger.Error("Delete backup failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "删除备份失败: "+err.Error())
		c.JSON(status, resp)
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(nil))
}

func (h *MenuHandler) RestoreBackup(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的备份ID")
		c.JSON(status, resp)
		return
	}

	if err := h.menuService.RestoreBackup(id); err != nil {
		h.logger.Error("Restore backup failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "恢复备份失败: "+err.Error())
		c.JSON(status, resp)
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(nil))
}

func menuToMap(m *user.Menu) gin.H {
	meta := gin.H{"title": m.Title}
	if m.Icon != "" {
		meta["icon"] = m.Icon
	}
	if m.Meta != nil {
		for k, v := range m.Meta {
			meta[k] = v
		}
	}
	node := gin.H{
		"id":         m.ID.String(),
		"space_key":  m.SpaceKey,
		"path":       m.Path,
		"name":       m.Name,
		"component":  m.Component,
		"meta":       meta,
		"sort_order": m.SortOrder,
	}
	if m.ParentID != nil {
		node["parent_id"] = m.ParentID.String()
	}
	if m.ManageGroupID != nil {
		node["manage_group_id"] = m.ManageGroupID.String()
	}
	if m.ManageGroup != nil {
		node["manage_group"] = gin.H{
			"id":         m.ManageGroup.ID.String(),
			"name":       m.ManageGroup.Name,
			"sort_order": m.ManageGroup.SortOrder,
			"status":     m.ManageGroup.Status,
		}
	}
	if len(m.Children) > 0 {
		children := make([]gin.H, 0, len(m.Children))
		for _, ch := range m.Children {
			children = append(children, menuToMap(ch))
		}
		node["children"] = children
	}
	return node
}

func menuToRuntimeMap(m *user.Menu) gin.H {
	meta := gin.H{
		"title": m.Title,
	}
	if m.Icon != "" {
		meta["icon"] = m.Icon
	}
	if m.Meta != nil {
		if accessMode := strings.TrimSpace(toStringValue(m.Meta["accessMode"])); accessMode != "" {
			meta["accessMode"] = accessMode
		}
		if link := strings.TrimSpace(toStringValue(m.Meta["link"])); link != "" {
			meta["link"] = link
		}
		if activePath := strings.TrimSpace(toStringValue(m.Meta["activePath"])); activePath != "" {
			meta["activePath"] = activePath
		}
		if roles := filterStringArray(m.Meta["roles"]); len(roles) > 0 {
			meta["roles"] = roles
		}
		copyBool(meta, "isEnable", m.Meta["isEnable"])
		copyTruthyBool(meta, "isHide", m.Meta["isHide"])
		copyTruthyBool(meta, "isIframe", m.Meta["isIframe"])
		copyTruthyBool(meta, "isHideTab", m.Meta["isHideTab"])
		copyTruthyBool(meta, "keepAlive", m.Meta["keepAlive"])
		copyTruthyBool(meta, "fixedTab", m.Meta["fixedTab"])
		copyTruthyBool(meta, "isFullPage", m.Meta["isFullPage"])
	}

	node := gin.H{
		"id":         m.ID.String(),
		"space_key":  m.SpaceKey,
		"path":       m.Path,
		"name":       m.Name,
		"component":  m.Component,
		"meta":       meta,
		"sort_order": m.SortOrder,
	}
	if m.ParentID != nil {
		node["parent_id"] = m.ParentID.String()
	}
	if len(m.Children) > 0 {
		children := make([]gin.H, 0, len(m.Children))
		for _, ch := range m.Children {
			children = append(children, menuToRuntimeMap(ch))
		}
		node["children"] = children
	}
	return node
}

func copyBool(target gin.H, key string, value any) {
	if flag, ok := value.(bool); ok {
		target[key] = flag
	}
}

func copyTruthyBool(target gin.H, key string, value any) {
	if flag, ok := value.(bool); ok && flag {
		target[key] = true
	}
}

func toStringValue(value any) string {
	text, _ := value.(string)
	return text
}

func filterStringArray(value any) []string {
	raw, ok := value.([]any)
	if !ok {
		if typed, ok := value.([]string); ok {
			result := make([]string, 0, len(typed))
			for _, item := range typed {
				if trimmed := strings.TrimSpace(item); trimmed != "" {
					result = append(result, trimmed)
				}
			}
			return result
		}
		return nil
	}
	result := make([]string, 0, len(raw))
	for _, item := range raw {
		text, ok := item.(string)
		if !ok {
			continue
		}
		if trimmed := strings.TrimSpace(text); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func currentUserID(c *gin.Context) (uuid.UUID, error) {
	value, ok := c.Get("user_id")
	if !ok {
		return uuid.Nil, errors.New("unauthorized")
	}
	userIDStr, ok := value.(string)
	if !ok {
		return uuid.Nil, errors.New("unauthorized")
	}
	return uuid.Parse(userIDStr)
}

func currentTenantID(c *gin.Context) (*uuid.UUID, error) {
	tenantIDStr := strings.TrimSpace(c.Query("tenant_id"))
	if tenantIDStr == "" {
		tenantIDStr = strings.TrimSpace(c.GetHeader(tenantContextHeader))
	}
	if tenantIDStr == "" {
		return nil, nil
	}
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		return nil, err
	}
	return &tenantID, nil
}

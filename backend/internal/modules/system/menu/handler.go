package menu

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/gg-ecommerce/backend/internal/api/dto"
	"github.com/gg-ecommerce/backend/internal/api/errcode"
	"github.com/gg-ecommerce/backend/internal/modules/system/models"
	"github.com/gg-ecommerce/backend/internal/modules/system/user"
)

const tenantContextHeader = "X-Tenant-ID"

type MenuHandler struct {
	menuService      MenuService
	userRepo         user.UserRepository
	roleMenuRepo     user.RoleMenuRepository
	userRoleRepo     user.UserRoleRepository
	tenantMemberRepo user.TenantMemberRepository
	authzService     interface {
		Authorize(userID uuid.UUID, tenantID *uuid.UUID, resourceCode, actionCode string, scopeCodes ...string) (bool, *models.PermissionAction, error)
	}
	logger *zap.Logger
}

func NewMenuHandler(menuService MenuService, userRepo user.UserRepository, roleMenuRepo user.RoleMenuRepository, userRoleRepo user.UserRoleRepository, tenantMemberRepo user.TenantMemberRepository, authzService interface {
	Authorize(userID uuid.UUID, tenantID *uuid.UUID, resourceCode, actionCode string, scopeCodes ...string) (bool, *models.PermissionAction, error)
}, logger *zap.Logger) *MenuHandler {
	return &MenuHandler{
		menuService:      menuService,
		userRepo:         userRepo,
		roleMenuRepo:     roleMenuRepo,
		userRoleRepo:     userRoleRepo,
		tenantMemberRepo: tenantMemberRepo,
		authzService:     authzService,
		logger:           logger,
	}
}

func (h *MenuHandler) GetTree(c *gin.Context) {
	all := c.Query("all") == "1" || c.Query("all") == "true"
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
		allowed, _, authErr := h.authzService.Authorize(userID, tenantID, "menu", "list", "global")
		if authErr != nil || !allowed {
			status, resp := errcode.ResponseWithMsg(errcode.ErrForbidden, "无权限查看全部菜单")
			c.JSON(status, resp)
			return
		}
	}

	var allowedMenuIDs []uuid.UUID
	if !all {
		userIDStr, ok := c.Get("user_id")
		if ok {
			if idStr, ok := userIDStr.(string); ok {
				if userID, err := uuid.Parse(idStr); err == nil {
					tenantIDStr := strings.TrimSpace(c.Query("tenant_id"))
					if tenantIDStr == "" {
						tenantIDStr = strings.TrimSpace(c.GetHeader(tenantContextHeader))
					}
					if tenantIDStr != "" && h.tenantMemberRepo != nil && h.roleMenuRepo != nil {
						if tid, err := uuid.Parse(tenantIDStr); err == nil {
							roleIDs, _ := h.userRoleRepo.GetEffectiveActiveRoleIDsByUserAndTenant(userID, &tid)
							allowedMenuIDs, _ = h.roleMenuRepo.GetMenuIDsByRoleIDs(roleIDs)
						}
					}
					if len(allowedMenuIDs) == 0 && h.userRepo != nil {
						user, err := h.userRepo.GetByID(userID)
						if err == nil && user.Roles != nil && len(user.Roles) > 0 {
							roleIDs := make([]uuid.UUID, 0, len(user.Roles))
							isAdmin := false
							for _, r := range user.Roles {
								roleIDs = append(roleIDs, r.ID)
								if r.Code == "admin" {
									isAdmin = true
								}
							}
							allowedMenuIDs, _ = h.roleMenuRepo.GetMenuIDsByRoleIDs(roleIDs)
							// 如果是admin角色且没有菜单关联，返回所有菜单
							if len(allowedMenuIDs) == 0 && isAdmin {
								all = true
							}
						}
					}
				}
			}
		}
	}
	tree, err := h.menuService.GetTree(all, allowedMenuIDs)
	if err != nil {
		h.logger.Error("Menu tree failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取菜单失败")
		c.JSON(status, resp)
		return
	}
	out := make([]gin.H, 0, len(tree))
	for _, node := range tree {
		out = append(out, menuToMap(node))
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(out))
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
			children = append(children, menuToMap(ch))
		}
		node["children"] = children
	}
	return node
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

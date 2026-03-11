package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/gg-ecommerce/backend/internal/api/dto"
	"github.com/gg-ecommerce/backend/internal/api/errcode"
	"github.com/gg-ecommerce/backend/internal/model"
	"github.com/gg-ecommerce/backend/internal/repository"
	"github.com/gg-ecommerce/backend/internal/service"
)

// MenuHandler 菜单管理处理器
type MenuHandler struct {
	menuService      service.MenuService
	userRepo         repository.UserRepository
	roleMenuRepo     repository.RoleMenuRepository
	userRoleRepo     repository.UserRoleRepository
	tenantMemberRepo repository.TenantMemberRepository
	logger           *zap.Logger
}

// NewMenuHandler 创建菜单处理器
func NewMenuHandler(menuService service.MenuService, userRepo repository.UserRepository, roleMenuRepo repository.RoleMenuRepository, userRoleRepo repository.UserRoleRepository, tenantMemberRepo repository.TenantMemberRepository, logger *zap.Logger) *MenuHandler {
	return &MenuHandler{menuService: menuService, userRepo: userRepo, roleMenuRepo: roleMenuRepo, userRoleRepo: userRoleRepo, tenantMemberRepo: tenantMemberRepo, logger: logger}
}

// GetTree 获取菜单树。query all=1 时返回全部；否则按角色权限过滤。
// query tenant_id=xxx 时在团队上下文中使用用户在该团队的团队角色 + 全局角色
func (h *MenuHandler) GetTree(c *gin.Context) {
	all := c.Query("all") == "1" || c.Query("all") == "true"

	var allowedMenuIDs []uuid.UUID
	if !all {
		userIDStr, ok := c.Get("user_id")
		if ok {
			if idStr, ok := userIDStr.(string); ok {
				if userID, err := uuid.Parse(idStr); err == nil {
					// 优先使用团队上下文中的团队角色
					tenantIDStr := c.Query("tenant_id")
					if tenantIDStr != "" && h.tenantMemberRepo != nil && h.roleMenuRepo != nil {
						if tid, err := uuid.Parse(tenantIDStr); err == nil {
							// 获取用户在团队中的角色（全局角色 + 团队角色）
							roleIDs, _ := h.userRoleRepo.GetRoleIDsByUserAndTenant(userID, &tid, h.tenantMemberRepo)
							allowedMenuIDs, _ = h.roleMenuRepo.GetMenuIDsByRoleIDs(roleIDs)
						}
					}
					// 如果没有团队上下文，使用全局角色
					if len(allowedMenuIDs) == 0 && h.userRepo != nil {
						user, err := h.userRepo.GetByID(userID)
						if err == nil && user.Roles != nil && len(user.Roles) > 0 {
							roleIDs := make([]uuid.UUID, 0, len(user.Roles))
							for _, r := range user.Roles {
								roleIDs = append(roleIDs, r.ID)
							}
							allowedMenuIDs, _ = h.roleMenuRepo.GetMenuIDsByRoleIDs(roleIDs)
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

// Create 创建菜单
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

// Update 更新菜单
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
		if err == service.ErrMenuNotFound {
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

// Delete 删除菜单
func (h *MenuHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的菜单ID")
		c.JSON(status, resp)
		return
	}
	if err := h.menuService.Delete(id); err != nil {
		if err == service.ErrMenuNotFound {
			status, resp := errcode.Response(errcode.ErrMenuNotFound)
			c.JSON(status, resp)
			return
		}
		if err == service.ErrMenuSystemProtected {
			status, resp := errcode.Response(errcode.ErrMenuSystemProtected)
			c.JSON(status, resp)
			return
		}
		h.logger.Error("Menu delete failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "删除菜单失败")
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(nil))
}

// UpdateSort 批量更新菜单排序
func (h *MenuHandler) UpdateSort(c *gin.Context) {
	var items []dto.MenuSortItem
	if err := c.ShouldBindJSON(&items); err != nil {
		status, resp := errcode.Response(errcode.ErrParamInvalid)
		c.JSON(status, resp)
		return
	}
	h.logger.Info("Menu sort update received", zap.Any("items", items))
	if err := h.menuService.UpdateSort(items); err != nil {
		h.logger.Error("Menu sort update failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "更新菜单排序失败")
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(nil))
}

// UpdateSortByParentID 根据父级ID更新子节点排序（全量重排）
func (h *MenuHandler) UpdateSortByParentID(c *gin.Context) {
	var req dto.MenuSortRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		status, resp := errcode.Response(errcode.ErrParamInvalid)
		c.JSON(status, resp)
		return
	}
	h.logger.Info("Menu sort update by parent", zap.Any("parent_id", req.ParentID), zap.Any("menu_ids", req.MenuIDs))
	if err := h.menuService.UpdateSortByParentID(req.ParentID, req.MenuIDs); err != nil {
		h.logger.Error("Menu sort update failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "更新菜单排序失败")
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(nil))
}

func menuToMap(m *model.Menu) gin.H {
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
		"is_system":  m.IsSystem,
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

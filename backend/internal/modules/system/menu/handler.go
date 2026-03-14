package menu

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/gg-ecommerce/backend/internal/api/dto"
	"github.com/gg-ecommerce/backend/internal/api/errcode"
	"github.com/gg-ecommerce/backend/internal/modules/system/user"
)

type MenuHandler struct {
	menuService      MenuService
	userRepo         user.UserRepository
	roleMenuRepo     user.RoleMenuRepository
	userRoleRepo     user.UserRoleRepository
	tenantMemberRepo user.TenantMemberRepository
	logger           *zap.Logger
}

func NewMenuHandler(menuService MenuService, userRepo user.UserRepository, roleMenuRepo user.RoleMenuRepository, userRoleRepo user.UserRoleRepository, tenantMemberRepo user.TenantMemberRepository, logger *zap.Logger) *MenuHandler {
	return &MenuHandler{
		menuService:      menuService,
		userRepo:         userRepo,
		roleMenuRepo:     roleMenuRepo,
		userRoleRepo:     userRoleRepo,
		tenantMemberRepo: tenantMemberRepo,
		logger:           logger,
	}
}

func (h *MenuHandler) GetTree(c *gin.Context) {
	all := c.Query("all") == "1" || c.Query("all") == "true"

	var allowedMenuIDs []uuid.UUID
	if !all {
		userIDStr, ok := c.Get("user_id")
		if ok {
			if idStr, ok := userIDStr.(string); ok {
				if userID, err := uuid.Parse(idStr); err == nil {
					tenantIDStr := c.Query("tenant_id")
					if tenantIDStr != "" && h.tenantMemberRepo != nil && h.roleMenuRepo != nil {
						if tid, err := uuid.Parse(tenantIDStr); err == nil {
							roleIDs, _ := h.userRoleRepo.GetRoleIDsByUserAndTenant(userID, &tid, h.tenantMemberRepo)
							allowedMenuIDs, _ = h.roleMenuRepo.GetMenuIDsByRoleIDs(roleIDs)
						}
					}
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
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "删除菜单失败")
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
		"id":          m.ID.String(),
		"path":        m.Path,
		"name":        m.Name,
		"component":   m.Component,
		"meta":        meta,
		"is_system":   m.IsSystem,
		"sort_order":  m.SortOrder,
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

package featurepackage

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/gg-ecommerce/backend/internal/api/dto"
	"github.com/gg-ecommerce/backend/internal/api/errcode"
	"github.com/gg-ecommerce/backend/internal/modules/system/user"
)

type Handler struct {
	service Service
	logger  *zap.Logger
}

func NewHandler(service Service, logger *zap.Logger) *Handler {
	return &Handler{service: service, logger: logger}
}

func (h *Handler) List(c *gin.Context) {
	var req dto.FeaturePackageListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		status, resp := errcode.Response(errcode.ErrParamInvalid)
		c.JSON(status, resp)
		return
	}
	list, total, err := h.service.List(&req)
	if err != nil {
		h.logger.Error("List feature packages failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取功能包列表失败")
		c.JSON(status, resp)
		return
	}
	packageIDs := make([]uuid.UUID, 0, len(list))
	for _, item := range list {
		packageIDs = append(packageIDs, item.ID)
	}
	actionCounts, menuCounts, teamCounts, err := h.service.GetPackageStats(packageIDs)
	if err != nil {
		h.logger.Error("Get feature package stats failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取功能包统计失败")
		c.JSON(status, resp)
		return
	}
	records := make([]gin.H, 0, len(list))
	for _, item := range list {
		records = append(records, packageToMapWithStats(&item, actionCounts[item.ID], menuCounts[item.ID], teamCounts[item.ID]))
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"records": records,
		"total":   total,
		"current": req.Current,
		"size":    req.Size,
	}))
}

func (h *Handler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的功能包ID")
		c.JSON(status, resp)
		return
	}
	item, err := h.service.Get(id)
	if err != nil {
		if err == ErrFeaturePackageNotFound {
			status, resp := errcode.ResponseWithMsg(errcode.ErrNotFound, "功能包不存在")
			c.JSON(status, resp)
			return
		}
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取功能包失败")
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(packageToMap(item)))
}

func (h *Handler) Create(c *gin.Context) {
	var req dto.FeaturePackageCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		status, resp := errcode.Response(errcode.ErrParamInvalid)
		c.JSON(status, resp)
		return
	}
	item, err := h.service.Create(&req)
	if err != nil {
		if err == ErrFeaturePackageExists {
			status, resp := errcode.ResponseWithMsg(errcode.ErrConflict, "功能包编码已存在")
			c.JSON(status, resp)
			return
		}
		h.logger.Error("Create feature package failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "创建功能包失败: "+err.Error())
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{"id": item.ID.String()}))
}

func (h *Handler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的功能包ID")
		c.JSON(status, resp)
		return
	}
	var req dto.FeaturePackageUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		status, resp := errcode.Response(errcode.ErrParamInvalid)
		c.JSON(status, resp)
		return
	}
	if err := h.service.Update(id, &req); err != nil {
		if err == ErrFeaturePackageNotFound {
			status, resp := errcode.ResponseWithMsg(errcode.ErrNotFound, "功能包不存在")
			c.JSON(status, resp)
			return
		}
		if err == ErrFeaturePackageExists {
			status, resp := errcode.ResponseWithMsg(errcode.ErrConflict, "功能包编码已存在")
			c.JSON(status, resp)
			return
		}
		h.logger.Error("Update feature package failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "更新功能包失败: "+err.Error())
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(nil))
}

func (h *Handler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的功能包ID")
		c.JSON(status, resp)
		return
	}
	if err := h.service.Delete(id); err != nil {
		if err == ErrFeaturePackageNotFound {
			status, resp := errcode.ResponseWithMsg(errcode.ErrNotFound, "功能包不存在")
			c.JSON(status, resp)
			return
		}
		if err == ErrFeaturePackageBuiltin {
			status, resp := errcode.ResponseWithMsg(errcode.ErrForbidden, "内置功能包不允许删除")
			c.JSON(status, resp)
			return
		}
		h.logger.Error("Delete feature package failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "删除功能包失败")
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(nil))
}

func (h *Handler) GetPackageActions(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的功能包ID")
		c.JSON(status, resp)
		return
	}
	actionIDs, actions, err := h.service.GetPackageActions(id)
	if err != nil {
		if err == ErrFeaturePackageNotFound {
			status, resp := errcode.ResponseWithMsg(errcode.ErrNotFound, "功能包不存在")
			c.JSON(status, resp)
			return
		}
		h.logger.Error("Get package actions failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取功能包权限失败")
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"action_ids": uuidListToStrings(actionIDs),
		"actions":    actionListToMaps(actions),
	}))
}

func (h *Handler) GetPackageChildren(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的功能包ID")
		c.JSON(status, resp)
		return
	}
	childPackageIDs, packages, err := h.service.GetPackageChildren(id)
	if err != nil {
		if err == ErrFeaturePackageNotFound {
			status, resp := errcode.ResponseWithMsg(errcode.ErrNotFound, "功能包不存在")
			c.JSON(status, resp)
			return
		}
		h.logger.Error("Get package children failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取组合包基础包失败")
		c.JSON(status, resp)
		return
	}
	items := make([]gin.H, 0, len(packages))
	for _, item := range packages {
		packageItem := item
		items = append(items, packageToMap(&packageItem))
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"child_package_ids": uuidListToStrings(childPackageIDs),
		"packages":          items,
	}))
}

func (h *Handler) SetPackageChildren(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的功能包ID")
		c.JSON(status, resp)
		return
	}
	var req dto.FeaturePackageChildSetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		status, resp := errcode.Response(errcode.ErrParamInvalid)
		c.JSON(status, resp)
		return
	}
	childPackageIDs, err := parseUUIDSlice(req.ChildPackageIDs)
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的基础包ID")
		c.JSON(status, resp)
		return
	}
	if err := h.service.SetPackageChildren(id, childPackageIDs); err != nil {
		if err == ErrFeaturePackageNotFound {
			status, resp := errcode.ResponseWithMsg(errcode.ErrNotFound, "功能包不存在")
			c.JSON(status, resp)
			return
		}
		h.logger.Error("Set package children failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "保存组合包基础包失败: "+err.Error())
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(nil))
}

func (h *Handler) SetPackageActions(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的功能包ID")
		c.JSON(status, resp)
		return
	}
	var req dto.FeaturePackageActionSetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		status, resp := errcode.Response(errcode.ErrParamInvalid)
		c.JSON(status, resp)
		return
	}
	actionIDs, err := parseUUIDSlice(req.ActionIDs)
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的功能权限ID")
		c.JSON(status, resp)
		return
	}
	if err := h.service.SetPackageActions(id, actionIDs); err != nil {
		if err == ErrFeaturePackageNotFound {
			status, resp := errcode.ResponseWithMsg(errcode.ErrNotFound, "功能包不存在")
			c.JSON(status, resp)
			return
		}
		h.logger.Error("Set package actions failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "保存功能包权限失败: "+err.Error())
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(nil))
}

func (h *Handler) GetPackageMenus(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的功能包ID")
		c.JSON(status, resp)
		return
	}
	menuIDs, menus, err := h.service.GetPackageMenus(id)
	if err != nil {
		if err == ErrFeaturePackageNotFound {
			status, resp := errcode.ResponseWithMsg(errcode.ErrNotFound, "功能包不存在")
			c.JSON(status, resp)
			return
		}
		h.logger.Error("Get package menus failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取功能包菜单失败")
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"menu_ids": uuidListToStrings(menuIDs),
		"menus":    menuListToMaps(menus),
	}))
}

func (h *Handler) SetPackageMenus(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的功能包ID")
		c.JSON(status, resp)
		return
	}
	var req dto.FeaturePackageMenuSetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		status, resp := errcode.Response(errcode.ErrParamInvalid)
		c.JSON(status, resp)
		return
	}
	menuIDs, err := parseUUIDSlice(req.MenuIDs)
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的菜单ID")
		c.JSON(status, resp)
		return
	}
	if err := h.service.SetPackageMenus(id, menuIDs); err != nil {
		if err == ErrFeaturePackageNotFound {
			status, resp := errcode.ResponseWithMsg(errcode.ErrNotFound, "功能包不存在")
			c.JSON(status, resp)
			return
		}
		h.logger.Error("Set package menus failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "保存功能包菜单失败: "+err.Error())
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(nil))
}

func (h *Handler) GetPackageTeams(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的功能包ID")
		c.JSON(status, resp)
		return
	}
	teamIDs, err := h.service.GetPackageTeams(id)
	if err != nil {
		if err == ErrFeaturePackageNotFound {
			status, resp := errcode.ResponseWithMsg(errcode.ErrNotFound, "功能包不存在")
			c.JSON(status, resp)
			return
		}
		h.logger.Error("Get package teams failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取功能包团队失败")
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"team_ids": uuidListToStrings(teamIDs),
	}))
}

func (h *Handler) SetPackageTeams(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的功能包ID")
		c.JSON(status, resp)
		return
	}
	var req dto.FeaturePackageTeamSetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		status, resp := errcode.Response(errcode.ErrParamInvalid)
		c.JSON(status, resp)
		return
	}
	teamIDs, err := parseUUIDSlice(req.TeamIDs)
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的团队ID")
		c.JSON(status, resp)
		return
	}
	grantedBy, _ := currentUserID(c)
	if err := h.service.SetPackageTeams(id, teamIDs, grantedBy); err != nil {
		if err == ErrFeaturePackageNotFound {
			status, resp := errcode.ResponseWithMsg(errcode.ErrNotFound, "功能包不存在")
			c.JSON(status, resp)
			return
		}
		h.logger.Error("Set package teams failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "保存功能包团队失败: "+err.Error())
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(nil))
}

func (h *Handler) GetTeamPackages(c *gin.Context) {
	teamID, err := uuid.Parse(c.Param("teamId"))
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的团队ID")
		c.JSON(status, resp)
		return
	}
	packageIDs, items, err := h.service.GetTeamPackages(teamID)
	if err != nil {
		h.logger.Error("Get team packages failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取团队功能包失败")
		c.JSON(status, resp)
		return
	}
	records := make([]gin.H, 0, len(items))
	for _, item := range items {
		records = append(records, packageToMap(&item))
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"package_ids": uuidListToStrings(packageIDs),
		"packages":    records,
	}))
}

func (h *Handler) SetTeamPackages(c *gin.Context) {
	teamID, err := uuid.Parse(c.Param("teamId"))
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的团队ID")
		c.JSON(status, resp)
		return
	}
	var req dto.TeamFeaturePackageSetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		status, resp := errcode.Response(errcode.ErrParamInvalid)
		c.JSON(status, resp)
		return
	}
	packageIDs, err := parseUUIDSlice(req.PackageIDs)
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的功能包ID")
		c.JSON(status, resp)
		return
	}
	grantedBy, _ := currentUserID(c)
	if err := h.service.SetTeamPackages(teamID, packageIDs, grantedBy); err != nil {
		if err == ErrFeaturePackageNotFound {
			status, resp := errcode.ResponseWithMsg(errcode.ErrNotFound, "存在无效的功能包")
			c.JSON(status, resp)
			return
		}
		h.logger.Error("Set team packages failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "保存团队功能包失败: "+err.Error())
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(nil))
}

func packageToMap(item *user.FeaturePackage) gin.H {
	return gin.H{
		"id":           item.ID.String(),
		"package_key":  item.PackageKey,
		"package_type": item.PackageType,
		"name":         item.Name,
		"description":  item.Description,
		"context_type": item.ContextType,
		"is_builtin":   item.IsBuiltin,
		"status":       item.Status,
		"sort_order":   item.SortOrder,
		"created_at":   item.CreatedAt.Format("2006-01-02 15:04:05"),
		"updated_at":   item.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

func packageToMapWithStats(item *user.FeaturePackage, actionCount, menuCount, teamCount int64) gin.H {
	result := packageToMap(item)
	result["action_count"] = actionCount
	result["menu_count"] = menuCount
	result["team_count"] = teamCount
	return result
}

func actionListToMaps(actions []user.PermissionAction) []gin.H {
	items := make([]gin.H, 0, len(actions))
	for _, action := range actions {
		items = append(items, gin.H{
			"id":             action.ID.String(),
			"permission_key": strings.TrimSpace(action.PermissionKey),
			"resource_code":  action.ResourceCode,
			"action_code":    action.ActionCode,
			"module_code":    action.ModuleCode,
			"context_type":   action.ContextType,
			"source":         action.Source,
			"feature_kind":   action.FeatureKind,
			"name":           action.Name,
			"description":    action.Description,
			"status":         action.Status,
			"sort_order":     action.SortOrder,
		})
	}
	return items
}

func menuListToMaps(menus []user.Menu) []gin.H {
	items := make([]gin.H, 0, len(menus))
	for _, menu := range menus {
		items = append(items, gin.H{
			"id":         menu.ID.String(),
			"parent_id":  uuidPtrToString(menu.ParentID),
			"path":       menu.Path,
			"name":       menu.Name,
			"component":  menu.Component,
			"title":      menu.Title,
			"icon":       menu.Icon,
			"hidden":     menu.Hidden,
			"sort_order": menu.SortOrder,
			"meta":       menu.Meta,
		})
	}
	return items
}

func uuidPtrToString(value *uuid.UUID) string {
	if value == nil {
		return ""
	}
	return value.String()
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

func uuidListToStrings(items []uuid.UUID) []string {
	result := make([]string, 0, len(items))
	for _, item := range items {
		result = append(result, item.String())
	}
	return result
}

func currentUserID(c *gin.Context) (*uuid.UUID, bool) {
	value, ok := c.Get("user_id")
	if !ok {
		return nil, false
	}
	userIDStr, ok := value.(string)
	if !ok {
		return nil, false
	}
	userID, err := uuid.Parse(strings.TrimSpace(userIDStr))
	if err != nil {
		return nil, false
	}
	return &userID, true
}

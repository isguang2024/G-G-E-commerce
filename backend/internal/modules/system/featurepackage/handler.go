package featurepackage

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/gg-ecommerce/backend/internal/api/dto"
	"github.com/gg-ecommerce/backend/internal/api/errcode"
	"github.com/gg-ecommerce/backend/internal/modules/system/user"
	"github.com/gg-ecommerce/backend/internal/pkg/appctx"
	"github.com/gg-ecommerce/backend/internal/pkg/permissionrefresh"
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
	resolvedAppKey, err := appctx.ResolveManagedAppKey(req.AppKey, c)
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrParamInvalid, "app_key 为必填项")
		c.JSON(status, resp)
		return
	}
	req.AppKey = resolvedAppKey
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

func (h *Handler) ListOptions(c *gin.Context) {
	var req dto.FeaturePackageListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		status, resp := errcode.Response(errcode.ErrParamInvalid)
		c.JSON(status, resp)
		return
	}
	resolvedAppKey, err := appctx.ResolveManagedAppKey(req.AppKey, c)
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrParamInvalid, "app_key 为必填项")
		c.JSON(status, resp)
		return
	}
	req.AppKey = resolvedAppKey
	list, err := h.service.ListOptions(&req)
	if err != nil {
		h.logger.Error("List feature package options failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取功能包候选失败")
		c.JSON(status, resp)
		return
	}
	records := make([]gin.H, 0, len(list))
	for _, item := range list {
		packageItem := item
		records = append(records, packageToMap(&packageItem))
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"records": records,
		"total":   len(records),
	}))
}

func (h *Handler) GetRelationTree(c *gin.Context) {
	contextType := strings.TrimSpace(c.Query("context_type"))
	keyword := strings.TrimSpace(c.Query("keyword"))
	appKey, err := appctx.RequireRequestAppKey(c)
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrParamInvalid, "app_key 为必填项")
		c.JSON(status, resp)
		return
	}
	result, err := h.service.GetRelationTree(appKey, contextType, keyword)
	if err != nil {
		h.logger.Error("Get feature package relation tree failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取功能包关系树失败")
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(result))
}

func (h *Handler) GetImpactPreview(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的功能包ID")
		c.JSON(status, resp)
		return
	}
	result, err := h.service.GetImpactPreview(id)
	if err != nil {
		if err == ErrFeaturePackageNotFound {
			status, resp := errcode.ResponseWithMsg(errcode.ErrNotFound, "功能包不存在")
			c.JSON(status, resp)
			return
		}
		h.logger.Error("Get feature package impact preview failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取影响预览失败")
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"package_id":   result.PackageID.String(),
		"role_count":   result.RoleCount,
		"team_count":   result.TeamCount,
		"user_count":   result.UserCount,
		"menu_count":   result.MenuCount,
		"action_count": result.ActionCount,
	}))
}

func (h *Handler) ListVersions(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的功能包ID")
		c.JSON(status, resp)
		return
	}
	current := parsePositiveInt(c.Query("current"), 1)
	size := parsePositiveInt(c.Query("size"), 20)
	items, total, err := h.service.ListVersions(id, current, size)
	if err != nil {
		if err == ErrFeaturePackageNotFound {
			status, resp := errcode.ResponseWithMsg(errcode.ErrNotFound, "功能包不存在")
			c.JSON(status, resp)
			return
		}
		h.logger.Error("List feature package versions failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取版本历史失败")
		c.JSON(status, resp)
		return
	}
	records := make([]gin.H, 0, len(items))
	for _, item := range items {
		records = append(records, gin.H{
			"id":          item.ID.String(),
			"package_id":  item.PackageID.String(),
			"version_no":  item.VersionNo,
			"change_type": item.ChangeType,
			"snapshot":    item.Snapshot,
			"operator_id": uuidPtrToString(item.OperatorID),
			"request_id":  item.RequestID,
			"created_at":  item.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"records": records,
		"total":   total,
		"current": current,
		"size":    size,
	}))
}

func (h *Handler) Rollback(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的功能包ID")
		c.JSON(status, resp)
		return
	}
	var req struct {
		VersionID string `json:"version_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		status, resp := errcode.Response(errcode.ErrParamInvalid)
		c.JSON(status, resp)
		return
	}
	versionID, err := uuid.Parse(strings.TrimSpace(req.VersionID))
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的版本ID")
		c.JSON(status, resp)
		return
	}
	operatorID, _ := currentUserID(c)
	stats, err := h.service.Rollback(id, versionID, operatorID, strings.TrimSpace(c.GetHeader("X-Request-ID")))
	if err != nil {
		if err == ErrFeaturePackageNotFound {
			status, resp := errcode.ResponseWithMsg(errcode.ErrNotFound, "功能包不存在")
			c.JSON(status, resp)
			return
		}
		h.logger.Error("Rollback feature package failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "回滚功能包版本失败: "+err.Error())
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"refresh_stats": refreshStatsToMap(stats),
	}))
}

func (h *Handler) ListRiskAudits(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的功能包ID")
		c.JSON(status, resp)
		return
	}
	current := parsePositiveInt(c.Query("current"), 1)
	size := parsePositiveInt(c.Query("size"), 20)
	items, total, err := h.service.ListRiskAudits(id, current, size)
	if err != nil {
		h.logger.Error("List feature package risk audits failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取最近变更失败")
		c.JSON(status, resp)
		return
	}
	records := make([]gin.H, 0, len(items))
	for _, item := range items {
		records = append(records, gin.H{
			"id":             item.ID.String(),
			"operator_id":    uuidPtrToString(item.OperatorID),
			"object_type":    item.ObjectType,
			"object_id":      item.ObjectID,
			"operation_type": item.OperationType,
			"before_summary": item.BeforeSummary,
			"after_summary":  item.AfterSummary,
			"impact_summary": item.ImpactSummary,
			"request_id":     item.RequestID,
			"created_at":     item.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"records": records,
		"total":   total,
		"current": current,
		"size":    size,
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
	resolvedAppKey, err := appctx.ResolveManagedAppKey(req.AppKey, c)
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrParamInvalid, "app_key 为必填项")
		c.JSON(status, resp)
		return
	}
	req.AppKey = resolvedAppKey
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
	resolvedAppKey, err := appctx.ResolveManagedAppKey(req.AppKey, c)
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrParamInvalid, "app_key 为必填项")
		c.JSON(status, resp)
		return
	}
	req.AppKey = resolvedAppKey
	stats, err := h.service.Update(id, &req)
	if err != nil {
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
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"refresh_stats": refreshStatsToMap(stats),
	}))
}

func (h *Handler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的功能包ID")
		c.JSON(status, resp)
		return
	}
	stats, err := h.service.Delete(id)
	if err != nil {
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
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"refresh_stats": refreshStatsToMap(stats),
	}))
}

func (h *Handler) GetPackageKeys(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的功能包ID")
		c.JSON(status, resp)
		return
	}
	appKey, err := appctx.RequireRequestAppKey(c)
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrParamInvalid, "app_key 为必填项")
		c.JSON(status, resp)
		return
	}
	actionIDs, actions, err := h.service.GetPackageKeys(id, appKey)
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
	appKey, err := appctx.RequireRequestAppKey(c)
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrParamInvalid, "app_key 为必填项")
		c.JSON(status, resp)
		return
	}
	childPackageIDs, packages, err := h.service.GetPackageChildren(id, appKey)
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
	resolvedAppKey, err := appctx.ResolveManagedAppKey(req.AppKey, c)
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrParamInvalid, "app_key 为必填项")
		c.JSON(status, resp)
		return
	}
	stats, err := h.service.SetPackageChildren(id, childPackageIDs, resolvedAppKey)
	if err != nil {
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
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"refresh_stats": refreshStatsToMap(stats),
	}))
}

func (h *Handler) SetPackageKeys(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的功能包ID")
		c.JSON(status, resp)
		return
	}
	var req dto.FeaturePackageKeySetRequest
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
	resolvedAppKey, err := appctx.ResolveManagedAppKey(req.AppKey, c)
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrParamInvalid, "app_key 为必填项")
		c.JSON(status, resp)
		return
	}
	stats, err := h.service.SetPackageKeys(id, actionIDs, resolvedAppKey)
	if err != nil {
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
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"refresh_stats": refreshStatsToMap(stats),
	}))
}

func (h *Handler) GetPackageMenus(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的功能包ID")
		c.JSON(status, resp)
		return
	}
	appKey, err := appctx.RequireRequestAppKey(c)
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrParamInvalid, "app_key 为必填项")
		c.JSON(status, resp)
		return
	}
	menuIDs, menus, err := h.service.GetPackageMenus(id, appKey)
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
	resolvedAppKey, err := appctx.ResolveManagedAppKey(req.AppKey, c)
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrParamInvalid, "app_key 为必填项")
		c.JSON(status, resp)
		return
	}
	stats, err := h.service.SetPackageMenus(id, menuIDs, resolvedAppKey)
	if err != nil {
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
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"refresh_stats": refreshStatsToMap(stats),
	}))
}

func (h *Handler) GetPackageTeams(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的功能包ID")
		c.JSON(status, resp)
		return
	}
	appKey, err := appctx.RequireRequestAppKey(c)
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrParamInvalid, "app_key 为必填项")
		c.JSON(status, resp)
		return
	}
	teamIDs, err := h.service.GetPackageTeams(id, appKey)
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
	resolvedAppKey, err := appctx.ResolveManagedAppKey(req.AppKey, c)
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrParamInvalid, "app_key 为必填项")
		c.JSON(status, resp)
		return
	}
	stats, err := h.service.SetPackageTeams(id, teamIDs, grantedBy, resolvedAppKey)
	if err != nil {
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
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"refresh_stats": refreshStatsToMap(stats),
	}))
}

func (h *Handler) GetTeamPackages(c *gin.Context) {
	teamID, err := uuid.Parse(c.Param("teamId"))
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
	packageIDs, items, err := h.service.GetTeamPackages(teamID, appKey)
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
	resolvedAppKey, err := appctx.ResolveManagedAppKey(req.AppKey, c)
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrParamInvalid, "app_key 为必填项")
		c.JSON(status, resp)
		return
	}
	stats, err := h.service.SetTeamPackages(teamID, packageIDs, grantedBy, resolvedAppKey)
	if err != nil {
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
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"refresh_stats": refreshStatsToMap(stats),
	}))
}

func packageToMap(item *user.FeaturePackage) gin.H {
	return gin.H{
		"id":           item.ID.String(),
		"app_key":      item.AppKey,
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

func actionListToMaps(actions []user.PermissionKey) []gin.H {
	items := make([]gin.H, 0, len(actions))
	for _, action := range actions {
		moduleGroup := gin.H(nil)
		if action.ModuleGroup != nil {
			moduleGroup = gin.H{
				"id":         action.ModuleGroup.ID.String(),
				"group_type": action.ModuleGroup.GroupType,
				"code":       action.ModuleGroup.Code,
				"name":       action.ModuleGroup.Name,
				"name_en":    action.ModuleGroup.NameEn,
				"status":     action.ModuleGroup.Status,
				"sort_order": action.ModuleGroup.SortOrder,
				"is_builtin": action.ModuleGroup.IsBuiltin,
			}
		}
		featureGroup := gin.H(nil)
		if action.FeatureGroup != nil {
			featureGroup = gin.H{
				"id":         action.FeatureGroup.ID.String(),
				"group_type": action.FeatureGroup.GroupType,
				"code":       action.FeatureGroup.Code,
				"name":       action.FeatureGroup.Name,
				"name_en":    action.FeatureGroup.NameEn,
				"status":     action.FeatureGroup.Status,
				"sort_order": action.FeatureGroup.SortOrder,
				"is_builtin": action.FeatureGroup.IsBuiltin,
			}
		}
		items = append(items, gin.H{
			"id":               action.ID.String(),
			"permission_key":   strings.TrimSpace(action.PermissionKey),
			"module_code":      action.ModuleCode,
			"module_group_id":  uuidPtrToString(action.ModuleGroupID),
			"feature_group_id": uuidPtrToString(action.FeatureGroupID),
			"module_group":     moduleGroup,
			"feature_group":    featureGroup,
			"context_type":     action.ContextType,
			"feature_kind":     action.FeatureKind,
			"name":             action.Name,
			"description":      action.Description,
			"status":           action.Status,
			"sort_order":       action.SortOrder,
			"is_builtin":       action.IsBuiltin,
		})
	}
	return items
}

func menuListToMaps(menus []user.Menu) []gin.H {
	items := make([]gin.H, 0, len(menus))
	for _, menu := range menus {
		items = append(items, gin.H{
			"id":         menu.ID.String(),
			"app_key":    menu.AppKey,
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

func parsePositiveInt(value string, fallback int) int {
	target := strings.TrimSpace(value)
	if target == "" {
		return fallback
	}
	var parsed int
	if _, err := fmt.Sscanf(target, "%d", &parsed); err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
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

func refreshStatsToMap(stats *permissionrefresh.RefreshStats) gin.H {
	if stats == nil {
		return gin.H{}
	}
	return gin.H{
		"requested_package_count": stats.RequestedPackageCount,
		"impacted_package_count":  stats.ImpactedPackageCount,
		"role_count":              stats.RoleCount,
		"team_count":              stats.TeamCount,
		"user_count":              stats.UserCount,
		"elapsed_milliseconds":    stats.ElapsedMilliseconds,
		"finished_at":             stats.FinishedAt,
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if target := strings.TrimSpace(value); target != "" {
			return target
		}
	}
	return ""
}

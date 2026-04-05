package page

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/gg-ecommerce/backend/internal/api/dto"
	"github.com/gg-ecommerce/backend/internal/api/errcode"
	apppkg "github.com/gg-ecommerce/backend/internal/modules/system/app"
	spaceutil "github.com/gg-ecommerce/backend/internal/modules/system/space"
	appctx "github.com/gg-ecommerce/backend/internal/pkg/appctx"
)

type Handler struct {
	service Service
	logger  *zap.Logger
}

func NewHandler(service Service, logger *zap.Logger) *Handler {
	return &Handler{service: service, logger: logger}
}

func (h *Handler) List(c *gin.Context) {
	var req ListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		status, resp := errcode.Response(errcode.ErrParamInvalid)
		c.JSON(status, resp)
		return
	}
	appKey, appErr := appctx.RequireRequestAppKey(c)
	if appErr != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrParamInvalid, "app_key is required")
		c.JSON(status, resp)
		return
	}
	req.AppKey = appKey
	list, total, err := h.service.List(&req)
	if err != nil {
		h.logger.Error("List pages failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取页面列表失败")
		c.JSON(status, resp)
		return
	}
	current, size := normalizePageAndSize(&req)
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"records": list,
		"total":   total,
		"current": current,
		"size":    size,
	}))
}

func (h *Handler) ListRuntime(c *gin.Context) {
	list, err := h.service.ListRuntime(apppkg.CurrentAppKey(c), spaceutil.RequestHost(c), spaceutil.RequestSpaceKey(c), pageContextUserID(c), pageContextTenantID(c))
	if err != nil {
		h.logger.Error("List runtime pages failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取运行时页面注册表失败")
		c.JSON(status, resp)
		return
	}
	records := buildRuntimePageRecords(list)
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"records": records,
		"total":   len(records),
	}))
}

func (h *Handler) ListRuntimePublic(c *gin.Context) {
	list, err := h.service.ListRuntimePublic(apppkg.CurrentAppKey(c), spaceutil.RequestHost(c), spaceutil.RequestSpaceKey(c), pageContextUserID(c), pageContextTenantID(c))
	if err != nil {
		h.logger.Error("List public runtime pages failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取公开运行时页面注册表失败")
		c.JSON(status, resp)
		return
	}
	records := buildRuntimePageRecords(list)
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"records": records,
		"total":   len(records),
	}))
}

func (h *Handler) ListUnregistered(c *gin.Context) {
	appKey, appErr := appctx.RequireRequestAppKey(c)
	if appErr != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrParamInvalid, "app_key is required")
		c.JSON(status, resp)
		return
	}
	items, err := h.service.ListUnregistered(appKey)
	if err != nil {
		h.logger.Error("List unregistered pages failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取未注册页面失败")
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"records": items,
		"total":   len(items),
	}))
}

func (h *Handler) Sync(c *gin.Context) {
	appKey, appErr := appctx.RequireRequestAppKey(c)
	if appErr != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrParamInvalid, "app_key is required")
		c.JSON(status, resp)
		return
	}
	result, err := h.service.Sync(appKey)
	if err != nil {
		h.respondServiceError(c, err, "同步页面注册表失败")
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(result))
}

func (h *Handler) ListMenuOptions(c *gin.Context) {
	appKey, appErr := appctx.RequireRequestAppKey(c)
	if appErr != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrParamInvalid, "app_key is required")
		c.JSON(status, resp)
		return
	}
	items, err := h.service.ListMenuOptions(appKey, spaceutil.RequestSpaceKey(c))
	if err != nil {
		h.logger.Error("List page menu options failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取上级菜单候选失败")
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"records": items,
		"total":   len(items),
	}))
}

func (h *Handler) ListPageOptions(c *gin.Context) {
	appKey, appErr := appctx.RequireRequestAppKey(c)
	if appErr != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrParamInvalid, "app_key is required")
		c.JSON(status, resp)
		return
	}
	items, err := h.service.ListOptions(appKey, spaceutil.RequestSpaceKey(c))
	if err != nil {
		h.logger.Error("List page options failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取页面候选失败")
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"records": items,
		"total":   len(items),
	}))
}

func (h *Handler) GetAccessTrace(c *gin.Context) {
	var req AccessTraceRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		status, resp := errcode.Response(errcode.ErrParamInvalid)
		c.JSON(status, resp)
		return
	}
	appKey, appErr := appctx.RequireRequestAppKey(c)
	if appErr != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrParamInvalid, "app_key is required")
		c.JSON(status, resp)
		return
	}
	req.AppKey = appKey
	result, err := h.service.GetAccessTrace(req.AppKey, &req)
	if err != nil {
		h.respondServiceError(c, err, "鑾峰彇椤甸潰璁块棶閾捐矾澶辫触")
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(result))
}

func (h *Handler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的页面ID")
		c.JSON(status, resp)
		return
	}
	appKey, appErr := appctx.RequireRequestAppKey(c)
	if appErr != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrParamInvalid, "app_key is required")
		c.JSON(status, resp)
		return
	}
	item, getErr := h.service.Get(id, appKey)
	if getErr != nil {
		h.respondServiceError(c, getErr, "获取页面详情失败")
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(item))
}

func (h *Handler) PreviewBreadcrumb(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的页面ID")
		c.JSON(status, resp)
		return
	}
	appKey, appErr := appctx.RequireRequestAppKey(c)
	if appErr != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrParamInvalid, "app_key is required")
		c.JSON(status, resp)
		return
	}
	items, previewErr := h.service.PreviewBreadcrumb(id, appKey)
	if previewErr != nil {
		h.respondServiceError(c, previewErr, "预览页面面包屑失败")
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"items": items,
		"total": len(items),
	}))
}

func (h *Handler) Create(c *gin.Context) {
	var req SaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		status, resp := errcode.Response(errcode.ErrParamInvalid)
		c.JSON(status, resp)
		return
	}
	appKey, appErr := appctx.RequireRequestAppKey(c)
	if appErr != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrParamInvalid, "app_key is required")
		c.JSON(status, resp)
		return
	}
	req.AppKey = appKey
	item, createErr := h.service.Create(&req)
	if createErr != nil {
		h.respondServiceError(c, createErr, "创建页面失败")
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(item))
}

func (h *Handler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的页面ID")
		c.JSON(status, resp)
		return
	}
	var req SaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		status, resp := errcode.Response(errcode.ErrParamInvalid)
		c.JSON(status, resp)
		return
	}
	appKey, appErr := appctx.RequireRequestAppKey(c)
	if appErr != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrParamInvalid, "app_key is required")
		c.JSON(status, resp)
		return
	}
	req.AppKey = appKey
	item, updateErr := h.service.Update(id, &req)
	if updateErr != nil {
		h.respondServiceError(c, updateErr, "更新页面失败")
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(item))
}

func (h *Handler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的页面ID")
		c.JSON(status, resp)
		return
	}
	appKey, appErr := appctx.RequireRequestAppKey(c)
	if appErr != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrParamInvalid, "app_key is required")
		c.JSON(status, resp)
		return
	}
	if deleteErr := h.service.Delete(id, appKey); deleteErr != nil {
		h.respondServiceError(c, deleteErr, "删除页面失败")
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(nil))
}

func (h *Handler) respondServiceError(c *gin.Context, err error, fallback string) {
	switch err {
	case ErrPageNotFound:
		status, resp := errcode.ResponseWithMsg(errcode.ErrNotFound, "页面不存在")
		c.JSON(status, resp)
	case ErrPageKeyExists:
		status, resp := errcode.ResponseWithMsg(errcode.ErrConflict, "页面标识已存在")
		c.JSON(status, resp)
	case ErrRouteNameExists:
		status, resp := errcode.ResponseWithMsg(errcode.ErrConflict, "路由名称已存在")
		c.JSON(status, resp)
	case ErrParentMenuInvalid:
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidParent, "无效的上级菜单")
		c.JSON(status, resp)
	case ErrParentPageInvalid:
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidParent, "无效的上级页面")
		c.JSON(status, resp)
	case ErrDisplayGroupInvalid:
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidParent, "无效的普通分组")
		c.JSON(status, resp)
	case ErrPageHasChildren:
		status, resp := errcode.ResponseWithMsg(errcode.ErrConflict, "当前页面下仍有子页面或分组，不能直接删除")
		c.JSON(status, resp)
	default:
		if strings.Contains(err.Error(), ErrPageValidation.Error()) {
			status, resp := errcode.ResponseWithMsg(errcode.ErrParamInvalid, strings.TrimPrefix(err.Error(), ErrPageValidation.Error()+": "))
			c.JSON(status, resp)
			return
		}
		h.logger.Error(fallback, zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, fallback)
		c.JSON(status, resp)
	}
}

func pageContextUserID(c *gin.Context) *uuid.UUID {
	raw, ok := c.Get("user_id")
	if !ok {
		return nil
	}
	value, ok := raw.(string)
	if !ok {
		return nil
	}
	id, err := uuid.Parse(strings.TrimSpace(value))
	if err != nil {
		return nil
	}
	return &id
}

func pageContextTenantID(c *gin.Context) *uuid.UUID {
	raw, ok := c.Get("tenant_id")
	if ok {
		value, ok := raw.(string)
		if ok && strings.TrimSpace(value) != "" {
			id, err := uuid.Parse(strings.TrimSpace(value))
			if err == nil {
				return &id
			}
		}
	}
	headerValue := strings.TrimSpace(c.GetHeader("X-Tenant-ID"))
	if headerValue == "" {
		return nil
	}
	id, err := uuid.Parse(headerValue)
	if err != nil {
		return nil
	}
	return &id
}

func buildRuntimePageRecords(items []Record) []gin.H {
	if len(items) == 0 {
		return []gin.H{}
	}
	pageMap := make(map[string]Record, len(items))
	for _, item := range items {
		pageKey := strings.TrimSpace(item.PageKey)
		if pageKey == "" {
			continue
		}
		pageMap[pageKey] = item
	}

	result := make([]gin.H, 0, len(items))
	for _, item := range items {
		if normalizePageType(item.PageType) == "group" {
			continue
		}
		result = append(result, buildRuntimePageRecord(flattenRuntimePageRecord(item, pageMap)))
	}
	return result
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if target := strings.TrimSpace(value); target != "" {
			return target
		}
	}
	return ""
}

func buildRuntimePageRecord(item Record) gin.H {
	node := gin.H{
		"page_key":   item.PageKey,
		"name":       item.Name,
		"route_path": item.RoutePath,
	}
	if item.Meta != nil {
		if values, ok := item.Meta["spaceKeys"].([]string); ok && len(values) > 0 {
			node["space_keys"] = values
		} else if values, ok := item.Meta["spaceKeys"].([]interface{}); ok && len(values) > 0 {
			node["space_keys"] = values
		}
		if scope, ok := item.Meta["spaceScope"].(string); ok && strings.TrimSpace(scope) != "" {
			node["space_scope"] = scope
		}
	}
	if scope := strings.TrimSpace(item.VisibilityScope); scope != "" && scope != "inherit" {
		node["visibility_scope"] = scope
	}

	if routeName := strings.TrimSpace(item.RouteName); routeName != "" && routeName != strings.TrimSpace(item.PageKey) {
		node["route_name"] = routeName
	}
	if component := strings.TrimSpace(item.Component); component != "" {
		node["component"] = component
	}
	if pageType := strings.TrimSpace(item.PageType); pageType != "" && pageType != "inner" {
		node["page_type"] = pageType
	}
	if item.ParentMenuID != nil {
		node["parent_menu_id"] = item.ParentMenuID.String()
	}
	if parentPageKey := strings.TrimSpace(item.ParentPageKey); parentPageKey != "" {
		node["parent_page_key"] = parentPageKey
	}
	if activeMenuPath := strings.TrimSpace(item.ActiveMenuPath); activeMenuPath != "" {
		node["active_menu_path"] = activeMenuPath
	}
	if breadcrumbMode := strings.TrimSpace(item.BreadcrumbMode); breadcrumbMode != "" && breadcrumbMode != "inherit_menu" {
		node["breadcrumb_mode"] = breadcrumbMode
	}
	if accessMode := strings.TrimSpace(item.AccessMode); accessMode != "" && accessMode != "inherit" {
		node["access_mode"] = accessMode
	}
	if permissionKey := strings.TrimSpace(item.PermissionKey); permissionKey != "" {
		node["permission_key"] = permissionKey
	}
	if item.KeepAlive {
		node["keep_alive"] = true
	}
	if item.IsFullPage {
		node["is_full_page"] = true
	}
	if status := strings.TrimSpace(item.Status); status != "" && status != "normal" {
		node["status"] = status
	}

	meta := gin.H{}
	if item.Meta != nil {
		if value, ok := item.Meta["isIframe"].(bool); ok && value {
			meta["isIframe"] = true
		}
		if value, ok := item.Meta["isHideTab"].(bool); ok && value {
			meta["isHideTab"] = true
		}
		if value, ok := item.Meta["link"].(string); ok && strings.TrimSpace(value) != "" {
			meta["link"] = strings.TrimSpace(value)
		}
	}
	if len(meta) > 0 {
		node["meta"] = meta
	}

	return node
}

func flattenRuntimePageRecord(item Record, pageMap map[string]Record) Record {
	flattened := item
	flattened.RoutePath = resolveRuntimeOutputRoutePath(item, pageMap, map[string]struct{}{})
	flattened.ParentPageKey = resolveNearestRuntimeParentPageKey(item, pageMap)
	if mode, permissionKey, ok := resolveRuntimeGroupAccessOverride(item, pageMap); ok {
		flattened.AccessMode = mode
		flattened.PermissionKey = permissionKey
	}
	return flattened
}

func resolveRuntimeOutputRoutePath(
	page Record,
	pageMap map[string]Record,
	seen map[string]struct{},
) string {
	pageKey := strings.TrimSpace(page.PageKey)
	if pageKey != "" {
		if _, ok := seen[pageKey]; ok {
			return ""
		}
		seen[pageKey] = struct{}{}
		defer delete(seen, pageKey)
	}

	rawRoutePath := strings.TrimSpace(page.RoutePath)
	if rawRoutePath == "" {
		return resolveRuntimeOutputBasePath(page, pageMap, seen)
	}
	if strings.HasPrefix(rawRoutePath, "http://") || strings.HasPrefix(rawRoutePath, "https://") {
		return rawRoutePath
	}
	if strings.HasPrefix(rawRoutePath, "/") && !isSingleSegmentRuntimePath(rawRoutePath) {
		return normalizeRoutePath(rawRoutePath)
	}

	basePath := resolveRuntimeOutputBasePath(page, pageMap, seen)
	segment := strings.TrimLeft(rawRoutePath, "/")
	if basePath != "" && !strings.HasPrefix(basePath, "http://") && !strings.HasPrefix(basePath, "https://") {
		return buildMenuFullPath(segment, basePath)
	}
	return normalizeRoutePath(segment)
}

func resolveRuntimeOutputBasePath(
	page Record,
	pageMap map[string]Record,
	seen map[string]struct{},
) string {
	if activeMenuPath := normalizeRoutePath(page.ActiveMenuPath); activeMenuPath != "" {
		return activeMenuPath
	}
	parentPageKey := strings.TrimSpace(page.ParentPageKey)
	if parentPageKey == "" {
		return ""
	}
	parentPage, ok := pageMap[parentPageKey]
	if !ok {
		return ""
	}
	return resolveRuntimeOutputRoutePath(parentPage, pageMap, seen)
}

func resolveNearestRuntimeParentPageKey(page Record, pageMap map[string]Record) string {
	parentPageKey := strings.TrimSpace(page.ParentPageKey)
	for parentPageKey != "" {
		parentPage, ok := pageMap[parentPageKey]
		if !ok {
			return ""
		}
		if normalizePageType(parentPage.PageType) != "group" {
			return parentPage.PageKey
		}
		parentPageKey = strings.TrimSpace(parentPage.ParentPageKey)
	}
	return ""
}

func resolveRuntimeGroupAccessOverride(
	page Record,
	pageMap map[string]Record,
) (string, string, bool) {
	if normalizeAccessMode(page.AccessMode) != "inherit" {
		return "", "", false
	}

	parentPageKey := strings.TrimSpace(page.ParentPageKey)
	for parentPageKey != "" {
		parentPage, ok := pageMap[parentPageKey]
		if !ok {
			return "", "", false
		}
		if normalizePageType(parentPage.PageType) != "group" {
			return "", "", false
		}

		mode := normalizeAccessMode(parentPage.AccessMode)
		switch mode {
		case "public", "jwt":
			return mode, "", true
		case "permission":
			return mode, strings.TrimSpace(parentPage.PermissionKey), true
		default:
			parentPageKey = strings.TrimSpace(parentPage.ParentPageKey)
		}
	}
	return "", "", false
}

func isSingleSegmentRuntimePath(path string) bool {
	normalized := strings.Trim(strings.TrimSpace(path), "/")
	return normalized != "" && !strings.Contains(normalized, "/")
}

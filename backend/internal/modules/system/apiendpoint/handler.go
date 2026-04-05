package apiendpoint

import (
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/gg-ecommerce/backend/internal/api/dto"
	"github.com/gg-ecommerce/backend/internal/api/errcode"
	"github.com/gg-ecommerce/backend/internal/modules/system/user"
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
	var req struct {
		Current           int    `form:"current"`
		Size              int    `form:"size"`
		AppKey            string `form:"app_key"`
		AppScope          string `form:"app_scope"`
		PermissionKey     string `form:"permission_key"`
		PermissionPattern string `form:"permission_pattern"`
		Keyword           string `form:"keyword"`
		Method            string `form:"method"`
		Path              string `form:"path"`
		CategoryID        string `form:"category_id"`
		ContextScope      string `form:"context_scope"`
		Source            string `form:"source"`
		FeatureKind       string `form:"feature_kind"`
		Status            string `form:"status"`
		HasPermission     *bool  `form:"has_permission_key"`
		HasCategory       *bool  `form:"has_category"`
	}
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
	list, total, err := h.service.List(&ListRequest{
		Current:           req.Current,
		Size:              req.Size,
		AppKey:            appKey,
		AppScope:          req.AppScope,
		PermissionKey:     req.PermissionKey,
		PermissionPattern: req.PermissionPattern,
		Keyword:           req.Keyword,
		Method:            req.Method,
		Path:              req.Path,
		CategoryID:        req.CategoryID,
		ContextScope:      req.ContextScope,
		Source:            req.Source,
		FeatureKind:       req.FeatureKind,
		Status:            req.Status,
		HasPermission:     req.HasPermission,
		HasCategory:       req.HasCategory,
	})
	if err != nil {
		h.logger.Error("List api endpoints failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取 API 列表失败")
		c.JSON(status, resp)
		return
	}
	categories, _ := h.service.ListCategories()
	categoryMap := make(map[uuid.UUID]user.APIEndpointCategory, len(categories))
	for _, category := range categories {
		categoryMap[category.ID] = category
	}
	runtimeStateMap := h.service.ListRuntimeStates(list)
	records := make([]gin.H, 0, len(list))
	endpointCodes := make([]string, 0, len(list))
	for _, endpoint := range list {
		if code := strings.TrimSpace(endpoint.Code); code != "" {
			endpointCodes = append(endpointCodes, code)
		}
	}
	bindings, _ := h.service.ListBindingsByEndpointCodes(endpointCodes)
	bindingsMap := make(map[string][]user.APIEndpointPermissionBinding, len(endpointCodes))
	for _, item := range bindings {
		bindingsMap[item.EndpointCode] = append(bindingsMap[item.EndpointCode], item)
	}
	for _, endpoint := range list {
		records = append(records, endpointToMap(&endpoint, bindingsMap[endpoint.Code], categoryMap, runtimeStateMap[endpoint.ID]))
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"records": records,
		"total":   total,
		"current": maxInt(req.Current, 1),
		"size":    maxInt(req.Size, 20),
	}))
}

func (h *Handler) Overview(c *gin.Context) {
	appKey, appErr := appctx.RequireRequestAppKey(c)
	if appErr != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrParamInvalid, "app_key is required")
		c.JSON(status, resp)
		return
	}
	overview, err := h.service.Overview(appKey)
	if err != nil {
		h.logger.Error("Get api endpoint overview failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取 API 概览失败")
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(overview))
}

func (h *Handler) ListStale(c *gin.Context) {
	var req struct {
		Current int `form:"current"`
		Size    int `form:"size"`
		AppKey  string `form:"app_key"`
	}
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
	list, total, err := h.service.ListStale(&StaleListRequest{
		Current: req.Current,
		Size:    req.Size,
		AppKey:  appKey,
	})
	if err != nil {
		h.logger.Error("List stale api endpoints failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取失效 API 失败")
		c.JSON(status, resp)
		return
	}
	categories, _ := h.service.ListCategories()
	categoryMap := make(map[uuid.UUID]user.APIEndpointCategory, len(categories))
	for _, category := range categories {
		categoryMap[category.ID] = category
	}
	runtimeStateMap := h.service.ListRuntimeStates(list)
	records := make([]gin.H, 0, len(list))
	for _, endpoint := range list {
		records = append(records, endpointToMap(&endpoint, nil, categoryMap, runtimeStateMap[endpoint.ID]))
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"records": records,
		"total":   total,
		"current": maxInt(req.Current, 1),
		"size":    maxInt(req.Size, 20),
	}))
}

func (h *Handler) ListUnregistered(c *gin.Context) {
	var req struct {
		Current    int    `form:"current"`
		Size       int    `form:"size"`
		Method     string `form:"method"`
		Path       string `form:"path"`
		Keyword    string `form:"keyword"`
		OnlyNoMeta bool   `form:"only_no_meta"`
	}
	if err := c.ShouldBindQuery(&req); err != nil {
		status, resp := errcode.Response(errcode.ErrParamInvalid)
		c.JSON(status, resp)
		return
	}
	list, total, err := h.service.ListUnregisteredRoutes(&UnregisteredRouteListRequest{
		Current:    req.Current,
		Size:       req.Size,
		Method:     req.Method,
		Path:       req.Path,
		Keyword:    req.Keyword,
		OnlyNoMeta: req.OnlyNoMeta,
	})
	if err != nil {
		h.logger.Error("List unregistered api routes failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取未注册 API 失败")
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"records": list,
		"total":   total,
		"current": maxInt(req.Current, 1),
		"size":    maxInt(req.Size, 20),
	}))
}

func (h *Handler) GetUnregisteredScanConfig(c *gin.Context) {
	config, err := h.service.GetUnregisteredScanConfig()
	if err != nil {
		h.logger.Error("Get unregistered scan config failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取未注册扫描配置失败")
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(config))
}

func (h *Handler) SaveUnregisteredScanConfig(c *gin.Context) {
	var req struct {
		Enabled              *bool  `json:"enabled"`
		FrequencyMinutes     int    `json:"frequency_minutes"`
		DefaultCategoryID    string `json:"default_category_id"`
		DefaultPermissionKey string `json:"default_permission_key"`
		MarkAsNoPermission   *bool  `json:"mark_as_no_permission"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		status, resp := errcode.Response(errcode.ErrParamInvalid)
		c.JSON(status, resp)
		return
	}
	current, err := h.service.GetUnregisteredScanConfig()
	if err != nil {
		h.logger.Error("Get current unregistered scan config failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "保存未注册扫描配置失败")
		c.JSON(status, resp)
		return
	}
	target := UnregisteredScanConfig{
		Enabled:              current.Enabled,
		FrequencyMinutes:     req.FrequencyMinutes,
		DefaultCategoryID:    strings.TrimSpace(req.DefaultCategoryID),
		DefaultPermissionKey: strings.TrimSpace(req.DefaultPermissionKey),
		MarkAsNoPermission:   current.MarkAsNoPermission,
	}
	if req.Enabled != nil {
		target.Enabled = *req.Enabled
	}
	if req.FrequencyMinutes <= 0 {
		target.FrequencyMinutes = current.FrequencyMinutes
	}
	if req.MarkAsNoPermission != nil {
		target.MarkAsNoPermission = *req.MarkAsNoPermission
	}
	saved, saveErr := h.service.SaveUnregisteredScanConfig(target)
	if saveErr != nil {
		h.logger.Error("Save unregistered scan config failed", zap.Error(saveErr))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "保存未注册扫描配置失败")
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(saved))
}

func (h *Handler) ListCategories(c *gin.Context) {
	items, err := h.service.ListCategories()
	if err != nil {
		h.logger.Error("List api endpoint categories failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取 API 分类失败")
		c.JSON(status, resp)
		return
	}
	records := make([]gin.H, 0, len(items))
	for _, item := range items {
		records = append(records, categoryToMap(&item))
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"records": records,
		"total":   len(records),
	}))
}

func (h *Handler) Sync(c *gin.Context) {
	if err := h.service.Sync(); err != nil {
		h.logger.Error("Sync api endpoints failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "同步 API 注册表失败")
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(nil))
}

func (h *Handler) CleanupStale(c *gin.Context) {
	var req struct {
		IDs []string `json:"ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		status, resp := errcode.Response(errcode.ErrParamInvalid)
		c.JSON(status, resp)
		return
	}
	if len(req.IDs) == 0 {
		status, resp := errcode.ResponseWithMsg(errcode.ErrParamInvalid, "请选择要清理的失效 API")
		c.JSON(status, resp)
		return
	}
	endpointIDs := make([]uuid.UUID, 0, len(req.IDs))
	seen := make(map[uuid.UUID]struct{}, len(req.IDs))
	for _, rawID := range req.IDs {
		endpointID, parseErr := uuid.Parse(strings.TrimSpace(rawID))
		if parseErr != nil {
			status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "存在无效的API ID")
			c.JSON(status, resp)
			return
		}
		if _, ok := seen[endpointID]; ok {
			continue
		}
		seen[endpointID] = struct{}{}
		endpointIDs = append(endpointIDs, endpointID)
	}

	appKey, appErr := appctx.RequireRequestAppKey(c)
	if appErr != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrParamInvalid, "app_key is required")
		c.JSON(status, resp)
		return
	}
	deletedCount, err := h.service.CleanupStale(endpointIDs, appKey)
	if err != nil {
		if errors.Is(err, ErrNoStaleCleanupSelection) || errors.Is(err, ErrStaleCleanupTargetGone) {
			status, resp := errcode.ResponseWithMsg(errcode.ErrParamInvalid, err.Error())
			c.JSON(status, resp)
			return
		}
		h.logger.Error("Cleanup stale api endpoints failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "清理失效 API 失败")
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"deleted_count": deletedCount,
	}))
}

func (h *Handler) Create(c *gin.Context) {
	var req struct {
		Code           string   `json:"code"`
		AppKey         string   `json:"app_key"`
		AppScope       string   `json:"app_scope"`
		Method         string   `json:"method" binding:"required"`
		Path           string   `json:"path" binding:"required"`
		Summary        string   `json:"summary"`
		FeatureKind    string   `json:"feature_kind"`
		CategoryID     string   `json:"category_id"`
		ContextScope   string   `json:"context_scope"`
		Source         string   `json:"source"`
		Status         string   `json:"status"`
		Handler        string   `json:"handler"`
		PermissionKeys []string `json:"permission_keys"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		status, resp := errcode.Response(errcode.ErrParamInvalid)
		c.JSON(status, resp)
		return
	}
	categoryID, parseErr := parseOptionalUUID(req.CategoryID)
	if parseErr != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的分类ID")
		c.JSON(status, resp)
		return
	}
	endpoint := &user.APIEndpoint{
		Code:         strings.TrimSpace(req.Code),
		AppKey:       strings.TrimSpace(req.AppKey),
		AppScope:     strings.TrimSpace(req.AppScope),
		Method:       strings.TrimSpace(req.Method),
		Path:         strings.TrimSpace(req.Path),
		Summary:      strings.TrimSpace(req.Summary),
		FeatureKind:  strings.TrimSpace(req.FeatureKind),
		CategoryID:   categoryID,
		ContextScope: strings.TrimSpace(req.ContextScope),
		Source:       strings.TrimSpace(req.Source),
		Status:       strings.TrimSpace(req.Status),
		Handler:      strings.TrimSpace(req.Handler),
	}
	appKey, appErr := appctx.RequireRequestAppKey(c)
	if appErr != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrParamInvalid, "app_key is required")
		c.JSON(status, resp)
		return
	}
	saved, err := h.service.Save(endpoint, req.PermissionKeys, appKey)
	if err != nil {
		h.logger.Error("Create api endpoint failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, err.Error())
		c.JSON(status, resp)
		return
	}
	bindings, _ := h.service.ListBindings(saved.Code)
	categories, _ := h.service.ListCategories()
	categoryMap := make(map[uuid.UUID]user.APIEndpointCategory, len(categories))
	for _, category := range categories {
		categoryMap[category.ID] = category
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(endpointToMap(saved, bindings, categoryMap, EndpointRuntimeState{})))
}

func (h *Handler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的API ID")
		c.JSON(status, resp)
		return
	}
	var req struct {
		Code           string   `json:"code"`
		AppKey         string   `json:"app_key"`
		AppScope       string   `json:"app_scope"`
		Method         string   `json:"method" binding:"required"`
		Path           string   `json:"path" binding:"required"`
		Summary        string   `json:"summary"`
		FeatureKind    string   `json:"feature_kind"`
		CategoryID     string   `json:"category_id"`
		ContextScope   string   `json:"context_scope"`
		Source         string   `json:"source"`
		Status         string   `json:"status"`
		Handler        string   `json:"handler"`
		PermissionKeys []string `json:"permission_keys"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		status, resp := errcode.Response(errcode.ErrParamInvalid)
		c.JSON(status, resp)
		return
	}
	categoryID, parseErr := parseOptionalUUID(req.CategoryID)
	if parseErr != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的分类ID")
		c.JSON(status, resp)
		return
	}
	endpoint := &user.APIEndpoint{
		ID:           id,
		Code:         strings.TrimSpace(req.Code),
		AppKey:       strings.TrimSpace(req.AppKey),
		AppScope:     strings.TrimSpace(req.AppScope),
		Method:       strings.TrimSpace(req.Method),
		Path:         strings.TrimSpace(req.Path),
		Summary:      strings.TrimSpace(req.Summary),
		FeatureKind:  strings.TrimSpace(req.FeatureKind),
		CategoryID:   categoryID,
		ContextScope: strings.TrimSpace(req.ContextScope),
		Source:       strings.TrimSpace(req.Source),
		Status:       strings.TrimSpace(req.Status),
		Handler:      strings.TrimSpace(req.Handler),
	}
	appKey, appErr := appctx.RequireRequestAppKey(c)
	if appErr != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrParamInvalid, "app_key is required")
		c.JSON(status, resp)
		return
	}
	saved, saveErr := h.service.Save(endpoint, req.PermissionKeys, appKey)
	if saveErr != nil {
		h.logger.Error("Update api endpoint failed", zap.Error(saveErr))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, saveErr.Error())
		c.JSON(status, resp)
		return
	}
	bindings, _ := h.service.ListBindings(saved.Code)
	categories, _ := h.service.ListCategories()
	categoryMap := make(map[uuid.UUID]user.APIEndpointCategory, len(categories))
	for _, category := range categories {
		categoryMap[category.ID] = category
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(endpointToMap(saved, bindings, categoryMap, EndpointRuntimeState{})))
}

func (h *Handler) SaveCategory(c *gin.Context) {
	var req struct {
		Code      string `json:"code" binding:"required"`
		Name      string `json:"name" binding:"required"`
		NameEn    string `json:"name_en" binding:"required"`
		SortOrder int    `json:"sort_order"`
		Status    string `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		status, resp := errcode.Response(errcode.ErrParamInvalid)
		c.JSON(status, resp)
		return
	}
	item := &user.APIEndpointCategory{
		Code:      strings.TrimSpace(req.Code),
		Name:      strings.TrimSpace(req.Name),
		NameEn:    strings.TrimSpace(req.NameEn),
		SortOrder: req.SortOrder,
		Status:    strings.TrimSpace(req.Status),
	}
	saved, err := h.service.SaveCategory(item)
	if err != nil {
		h.logger.Error("Create api endpoint category failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, err.Error())
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(categoryToMap(saved)))
}

func (h *Handler) UpdateCategory(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的分类ID")
		c.JSON(status, resp)
		return
	}
	var req struct {
		Code      string `json:"code" binding:"required"`
		Name      string `json:"name" binding:"required"`
		NameEn    string `json:"name_en" binding:"required"`
		SortOrder int    `json:"sort_order"`
		Status    string `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		status, resp := errcode.Response(errcode.ErrParamInvalid)
		c.JSON(status, resp)
		return
	}
	item := &user.APIEndpointCategory{
		ID:        id,
		Code:      strings.TrimSpace(req.Code),
		Name:      strings.TrimSpace(req.Name),
		NameEn:    strings.TrimSpace(req.NameEn),
		SortOrder: req.SortOrder,
		Status:    strings.TrimSpace(req.Status),
	}
	saved, saveErr := h.service.SaveCategory(item)
	if saveErr != nil {
		h.logger.Error("Update api endpoint category failed", zap.Error(saveErr))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, saveErr.Error())
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(categoryToMap(saved)))
}

func (h *Handler) UpdateContextScope(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的API ID")
		c.JSON(status, resp)
		return
	}
	var req struct {
		ContextScope string `json:"context_scope" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		status, resp := errcode.Response(errcode.ErrParamInvalid)
		c.JSON(status, resp)
		return
	}
	saved, updateErr := h.service.UpdateContextScope(id, req.ContextScope)
	if updateErr != nil {
		h.logger.Error("Update api endpoint context scope failed", zap.Error(updateErr))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, updateErr.Error())
		c.JSON(status, resp)
		return
	}
	bindings, _ := h.service.ListBindings(saved.Code)
	categories, _ := h.service.ListCategories()
	categoryMap := make(map[uuid.UUID]user.APIEndpointCategory, len(categories))
	for _, category := range categories {
		categoryMap[category.ID] = category
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(endpointToMap(saved, bindings, categoryMap, EndpointRuntimeState{})))
}

func endpointToMap(endpoint *user.APIEndpoint, bindings []user.APIEndpointPermissionBinding, categoryMap map[uuid.UUID]user.APIEndpointCategory, runtimeState EndpointRuntimeState) gin.H {
	permissionKeys := make([]string, 0, len(bindings))
	for _, item := range bindings {
		permissionKeys = append(permissionKeys, item.PermissionKey)
	}
	profile := buildPermissionProfile(endpoint.Path, permissionKeys)
	authMode := deriveEndpointAuthMode(endpoint.Path, profile.Keys)
	var category gin.H
	if endpoint.CategoryID != nil {
		if item, ok := categoryMap[*endpoint.CategoryID]; ok {
			category = categoryToMap(&item)
		}
	}
	return gin.H{
		"id":                      endpoint.ID.String(),
		"code":                    endpoint.Code,
		"app_key":                 endpoint.AppKey,
		"app_scope":               endpoint.AppScope,
		"method":                  endpoint.Method,
		"path":                    endpoint.Path,
		"spec":                    endpoint.Method + " " + endpoint.Path,
		"feature_kind":            endpoint.FeatureKind,
		"handler":                 endpoint.Handler,
		"summary":                 endpoint.Summary,
		"permission_key":          profile.PrimaryKey,
		"permission_keys":         profile.Keys,
		"permission_contexts":     profile.Contexts,
		"permission_binding_mode": profile.BindingMode,
		"shared_across_contexts":  profile.SharedAcrossContexts,
		"permission_note":         profile.Note,
		"auth_mode":               authMode,
		"category_id":             stringifyUUIDPointer(endpoint.CategoryID),
		"category":                category,
		"context_scope":           endpoint.ContextScope,
		"source":                  endpoint.Source,
		"status":                  endpoint.Status,
		"runtime_exists":          runtimeState.RuntimeExists,
		"stale":                   runtimeState.Stale,
		"stale_reason":            runtimeState.StaleReason,
		"created_at":              endpoint.CreatedAt.Format("2006-01-02 15:04:05"),
		"updated_at":              endpoint.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

func categoryToMap(item *user.APIEndpointCategory) gin.H {
	if item == nil {
		return gin.H{}
	}
	return gin.H{
		"id":         item.ID.String(),
		"code":       item.Code,
		"name":       item.Name,
		"name_en":    item.NameEn,
		"sort_order": item.SortOrder,
		"status":     item.Status,
	}
}

func parseOptionalUUID(value string) (*uuid.UUID, error) {
	target := strings.TrimSpace(value)
	if target == "" {
		return nil, nil
	}
	id, err := uuid.Parse(target)
	if err != nil {
		return nil, err
	}
	return &id, nil
}

func stringifyUUIDPointer(value *uuid.UUID) string {
	if value == nil {
		return ""
	}
	return value.String()
}

func maxInt(value, fallback int) int {
	if value > 0 {
		return value
	}
	return fallback
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if target := strings.TrimSpace(value); target != "" {
			return target
		}
	}
	return ""
}

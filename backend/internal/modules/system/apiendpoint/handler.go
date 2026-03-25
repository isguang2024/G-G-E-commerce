package apiendpoint

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
	var req struct {
		Current       int    `form:"current"`
		Size          int    `form:"size"`
		PermissionKey string `form:"permission_key"`
		Keyword       string `form:"keyword"`
		Method        string `form:"method"`
		Path          string `form:"path"`
		CategoryID    string `form:"category_id"`
		ContextScope  string `form:"context_scope"`
		Source        string `form:"source"`
		FeatureKind   string `form:"feature_kind"`
		Status        string `form:"status"`
		HasPermission *bool  `form:"has_permission_key"`
		HasCategory   *bool  `form:"has_category"`
	}
	if err := c.ShouldBindQuery(&req); err != nil {
		status, resp := errcode.Response(errcode.ErrParamInvalid)
		c.JSON(status, resp)
		return
	}
	list, total, err := h.service.List(&ListRequest{
		Current:       req.Current,
		Size:          req.Size,
		PermissionKey: req.PermissionKey,
		Keyword:       req.Keyword,
		Method:        req.Method,
		Path:          req.Path,
		CategoryID:    req.CategoryID,
		ContextScope:  req.ContextScope,
		Source:        req.Source,
		FeatureKind:   req.FeatureKind,
		Status:        req.Status,
		HasPermission: req.HasPermission,
		HasCategory:   req.HasCategory,
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
	records := make([]gin.H, 0, len(list))
	for _, endpoint := range list {
		bindings, _ := h.service.ListBindings(endpoint.ID)
		records = append(records, endpointToMap(&endpoint, bindings, categoryMap))
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

func (h *Handler) Create(c *gin.Context) {
	var req struct {
		Code           string   `json:"code"`
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
	saved, err := h.service.Save(endpoint, req.PermissionKeys)
	if err != nil {
		h.logger.Error("Create api endpoint failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, err.Error())
		c.JSON(status, resp)
		return
	}
	bindings, _ := h.service.ListBindings(saved.ID)
	categories, _ := h.service.ListCategories()
	categoryMap := make(map[uuid.UUID]user.APIEndpointCategory, len(categories))
	for _, category := range categories {
		categoryMap[category.ID] = category
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(endpointToMap(saved, bindings, categoryMap)))
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
	saved, saveErr := h.service.Save(endpoint, req.PermissionKeys)
	if saveErr != nil {
		h.logger.Error("Update api endpoint failed", zap.Error(saveErr))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, saveErr.Error())
		c.JSON(status, resp)
		return
	}
	bindings, _ := h.service.ListBindings(saved.ID)
	categories, _ := h.service.ListCategories()
	categoryMap := make(map[uuid.UUID]user.APIEndpointCategory, len(categories))
	for _, category := range categories {
		categoryMap[category.ID] = category
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(endpointToMap(saved, bindings, categoryMap)))
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
	bindings, _ := h.service.ListBindings(saved.ID)
	categories, _ := h.service.ListCategories()
	categoryMap := make(map[uuid.UUID]user.APIEndpointCategory, len(categories))
	for _, category := range categories {
		categoryMap[category.ID] = category
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(endpointToMap(saved, bindings, categoryMap)))
}

func endpointToMap(endpoint *user.APIEndpoint, bindings []user.APIEndpointPermissionBinding, categoryMap map[uuid.UUID]user.APIEndpointCategory) gin.H {
	permissionKeys := make([]string, 0, len(bindings))
	for _, item := range bindings {
		permissionKeys = append(permissionKeys, item.PermissionKey)
	}
	primaryPermissionKey := ""
	if len(permissionKeys) > 0 {
		primaryPermissionKey = permissionKeys[0]
	}
	authMode := "jwt"
	switch {
	case endpoint.Path == "/health":
		authMode = "public"
	case endpoint.Path == "/api/v1/auth/login" || endpoint.Path == "/api/v1/auth/register" || endpoint.Path == "/api/v1/auth/refresh":
		authMode = "public"
	case len(endpoint.Path) >= len("/open/v1/") && endpoint.Path[:len("/open/v1/")] == "/open/v1/":
		authMode = "api_key"
	case len(permissionKeys) > 0:
		authMode = "permission"
	}
	var category gin.H
	if endpoint.CategoryID != nil {
		if item, ok := categoryMap[*endpoint.CategoryID]; ok {
			category = categoryToMap(&item)
		}
	}
	return gin.H{
		"id":              endpoint.ID.String(),
		"code":            endpoint.Code,
		"method":          endpoint.Method,
		"path":            endpoint.Path,
		"spec":            endpoint.Method + " " + endpoint.Path,
		"feature_kind":    endpoint.FeatureKind,
		"handler":         endpoint.Handler,
		"summary":         endpoint.Summary,
		"permission_key":  primaryPermissionKey,
		"permission_keys": permissionKeys,
		"auth_mode":       authMode,
		"category_id":     stringifyUUIDPointer(endpoint.CategoryID),
		"category":        category,
		"context_scope":   endpoint.ContextScope,
		"source":          endpoint.Source,
		"status":          endpoint.Status,
		"created_at":      endpoint.CreatedAt.Format("2006-01-02 15:04:05"),
		"updated_at":      endpoint.UpdatedAt.Format("2006-01-02 15:04:05"),
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

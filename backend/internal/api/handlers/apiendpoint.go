// apiendpoint.go — Phase 5 ogen handlers for the api-endpoints domain.
package handlers

import (
	"context"
	"errors"
	"io"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/gg-ecommerce/backend/api/gen"
	"github.com/gg-ecommerce/backend/internal/modules/system/apiendpoint"
	"github.com/gg-ecommerce/backend/internal/modules/system/user"
	"github.com/gg-ecommerce/backend/internal/pkg/permissionseed"
)

// ─── local mapping helpers ────────────────────────────────────────────────────

// epAccessModeMap is a lazily-initialised cache of the openapi seed lookup.
// Populated on first call; safe for concurrent reads after init.
var epAccessModeMap map[string]string

func getEpAccessModeMap() map[string]string {
	if epAccessModeMap != nil {
		return epAccessModeMap
	}
	seed, err := permissionseed.LoadOpenAPISeed()
	if err != nil {
		return map[string]string{}
	}
	epAccessModeMap = seed.AccessModeByMethodPath()
	return epAccessModeMap
}

func epToMap(endpoint *user.APIEndpoint, bindings []user.APIEndpointPermissionBinding, categoryMap map[uuid.UUID]user.APIEndpointCategory, runtimeState apiendpoint.EndpointRuntimeState) map[string]interface{} {
	permissionKeys := make([]string, 0, len(bindings))
	for _, b := range bindings {
		permissionKeys = append(permissionKeys, b.PermissionKey)
	}
	var category map[string]interface{}
	if endpoint.CategoryID != nil {
		if item, ok := categoryMap[*endpoint.CategoryID]; ok {
			category = catToMap(&item)
		}
	}
	catID := ""
	if endpoint.CategoryID != nil {
		catID = endpoint.CategoryID.String()
	}

	// Resolve auth_mode from the embedded spec seed — single source of truth.
	seedKey := strings.ToUpper(strings.TrimSpace(endpoint.Method)) + " " + strings.TrimSpace(endpoint.Path)
	authMode := getEpAccessModeMap()[seedKey]
	if authMode == "authenticated" {
		authMode = "jwt"
	}
	if authMode == "" {
		// Fallback: has binding → permission, otherwise jwt.
		if len(permissionKeys) > 0 {
			authMode = "permission"
		} else {
			authMode = "jwt"
		}
	}

	return map[string]interface{}{
		"id":              endpoint.ID.String(),
		"code":            endpoint.Code,
		"method":          endpoint.Method,
		"path":            endpoint.Path,
		"spec":            endpoint.Method + " " + endpoint.Path,
		"handler":         endpoint.Handler,
		"summary":         endpoint.Summary,
		"permission_keys": permissionKeys,
		"auth_mode":       authMode,
		"category_id":     catID,
		"category":        category,
		"status":          endpoint.Status,
		"runtime_exists":  runtimeState.RuntimeExists,
		"stale":           runtimeState.Stale,
		"stale_reason":    runtimeState.StaleReason,
		"created_at":      endpoint.CreatedAt.Format("2006-01-02 15:04:05"),
		"updated_at":      endpoint.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

func catToMap(item *user.APIEndpointCategory) map[string]interface{} {
	if item == nil {
		return map[string]interface{}{}
	}
	return map[string]interface{}{
		"id":         item.ID.String(),
		"code":       item.Code,
		"name":       item.Name,
		"name_en":    item.NameEn,
		"sort_order": item.SortOrder,
		"status":     item.Status,
	}
}

func parseMaybeUUID(value string) (*uuid.UUID, error) {
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

// ─── listApiEndpoints ─────────────────────────────────────────────────────────

func (h *APIHandler) ListApiEndpoints(ctx context.Context, params gen.ListApiEndpointsParams) (gen.ListApiEndpointsRes, error) {
	var hasPermission *bool
	if params.HasPermissionKey.Set {
		v := params.HasPermissionKey.Value
		hasPermission = &v
	}
	var hasCategory *bool
	if params.HasCategory.Set {
		v := params.HasCategory.Value
		hasCategory = &v
	}
	list, total, err := h.apiEndpointSvc.List(&apiendpoint.ListRequest{
		Current:           optInt(params.Current, 1),
		Size:              optInt(params.Size, 20),
		PermissionKey:     optString(params.PermissionKey),
		PermissionPattern: optString(params.PermissionPattern),
		Keyword:           optString(params.Keyword),
		Method:            optString(params.Method),
		Path:              optString(params.Path),
		CategoryID:        optString(params.CategoryID),
		Status:            optString(params.Status),
		HasPermission:     hasPermission,
		HasCategory:       hasCategory,
	})
	if err != nil {
		h.logger.Error("list api endpoints failed", zap.Error(err))
		return nil, err
	}
	categories, _ := h.apiEndpointSvc.ListCategories()
	categoryMap := make(map[uuid.UUID]user.APIEndpointCategory, len(categories))
	for _, cat := range categories {
		categoryMap[cat.ID] = cat
	}
	runtimeStateMap := h.apiEndpointSvc.ListRuntimeStates(list)
	endpointCodes := make([]string, 0, len(list))
	for _, ep := range list {
		if code := strings.TrimSpace(ep.Code); code != "" {
			endpointCodes = append(endpointCodes, code)
		}
	}
	bindings, _ := h.apiEndpointSvc.ListBindingsByEndpointCodes(endpointCodes)
	bindingsMap := make(map[string][]user.APIEndpointPermissionBinding, len(endpointCodes))
	for _, b := range bindings {
		bindingsMap[b.EndpointCode] = append(bindingsMap[b.EndpointCode], b)
	}
	records := make([]interface{}, 0, len(list))
	for _, ep := range list {
		records = append(records, epToMap(&ep, bindingsMap[ep.Code], categoryMap, runtimeStateMap[ep.ID]))
	}
	obj := marshalAnyObject(map[string]interface{}{
		"records": records,
		"total":   total,
		"current": optInt(params.Current, 1),
		"size":    optInt(params.Size, 20),
	})
	return &obj, nil
}

// ─── getApiEndpointOverview ───────────────────────────────────────────────────

func (h *APIHandler) GetApiEndpointOverview(ctx context.Context, params gen.GetApiEndpointOverviewParams) (gen.GetApiEndpointOverviewRes, error) {
	overview, err := h.apiEndpointSvc.Overview(optString(params.AppKey))
	if err != nil {
		h.logger.Error("get api endpoint overview failed", zap.Error(err))
		return nil, err
	}
	obj := marshalAnyObject(overview)
	return &obj, nil
}

// ─── listStaleApiEndpoints ────────────────────────────────────────────────────

func (h *APIHandler) ListStaleApiEndpoints(ctx context.Context, params gen.ListStaleApiEndpointsParams) (gen.ListStaleApiEndpointsRes, error) {
	list, total, err := h.apiEndpointSvc.ListStale(&apiendpoint.StaleListRequest{
		Current: optInt(params.Current, 1),
		Size:    optInt(params.Size, 20),
	})
	if err != nil {
		h.logger.Error("list stale api endpoints failed", zap.Error(err))
		return nil, err
	}
	categories, _ := h.apiEndpointSvc.ListCategories()
	categoryMap := make(map[uuid.UUID]user.APIEndpointCategory, len(categories))
	for _, cat := range categories {
		categoryMap[cat.ID] = cat
	}
	runtimeStateMap := h.apiEndpointSvc.ListRuntimeStates(list)
	records := make([]interface{}, 0, len(list))
	for _, ep := range list {
		records = append(records, epToMap(&ep, nil, categoryMap, runtimeStateMap[ep.ID]))
	}
	obj := marshalAnyObject(map[string]interface{}{
		"records": records,
		"total":   total,
		"current": optInt(params.Current, 1),
		"size":    optInt(params.Size, 20),
	})
	return &obj, nil
}

// ─── listUnregisteredApiEndpoints ────────────────────────────────────────────

func (h *APIHandler) ListUnregisteredApiEndpoints(ctx context.Context, params gen.ListUnregisteredApiEndpointsParams) (gen.ListUnregisteredApiEndpointsRes, error) {
	list, total, err := h.apiEndpointSvc.ListUnregisteredRoutes(&apiendpoint.UnregisteredRouteListRequest{
		Current:    optInt(params.Current, 1),
		Size:       optInt(params.Size, 20),
		Method:     optString(params.Method),
		Path:       optString(params.Path),
		Keyword:    optString(params.Keyword),
		OnlyNoMeta: optBool(params.OnlyNoMeta),
	})
	if err != nil {
		h.logger.Error("list unregistered api routes failed", zap.Error(err))
		return nil, err
	}
	obj := marshalAnyObject(map[string]interface{}{
		"records": list,
		"total":   total,
		"current": optInt(params.Current, 1),
		"size":    optInt(params.Size, 20),
	})
	return &obj, nil
}

// ─── listApiEndpointCategories ────────────────────────────────────────────────

func (h *APIHandler) ListApiEndpointCategories(ctx context.Context) (gen.ListApiEndpointCategoriesRes, error) {
	items, err := h.apiEndpointSvc.ListCategories()
	if err != nil {
		h.logger.Error("list api endpoint categories failed", zap.Error(err))
		return nil, err
	}
	records := make([]interface{}, 0, len(items))
	for _, item := range items {
		records = append(records, catToMap(&item))
	}
	obj := marshalAnyObject(map[string]interface{}{
		"records": records,
		"total":   len(records),
	})
	return &obj, nil
}

// ─── syncApiEndpoints ─────────────────────────────────────────────────────────

func (h *APIHandler) SyncApiEndpoints(ctx context.Context) (gen.SyncApiEndpointsRes, error) {
	if err := h.apiEndpointSvc.Sync(); err != nil {
		h.logger.Error("sync api endpoints failed", zap.Error(err))
		return nil, err
	}
	return ok(), nil
}

// ─── cleanupStaleApiEndpoints ─────────────────────────────────────────────────

func (h *APIHandler) CleanupStaleApiEndpoints(ctx context.Context, req gen.AnyObject) (gen.CleanupStaleApiEndpointsRes, error) {
	var body struct {
		IDs []string `json:"ids"`
	}
	if err := unmarshalAnyObject(req, &body); err != nil && !errors.Is(err, io.EOF) {
		return nil, err
	}
	endpointIDs := make([]uuid.UUID, 0, len(body.IDs))
	seen := make(map[uuid.UUID]struct{}, len(body.IDs))
	for _, rawID := range body.IDs {
		id, err := uuid.Parse(strings.TrimSpace(rawID))
		if err != nil {
			return nil, err
		}
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		endpointIDs = append(endpointIDs, id)
	}
	deletedCount, err := h.apiEndpointSvc.CleanupStale(endpointIDs, "")
	if err != nil {
		h.logger.Error("cleanup stale api endpoints failed", zap.Error(err))
		return nil, err
	}
	obj := marshalAnyObject(map[string]interface{}{
		"deleted_count": deletedCount,
	})
	return &obj, nil
}

// ─── updateApiEndpoint ────────────────────────────────────────────────────────

func (h *APIHandler) UpdateApiEndpoint(ctx context.Context, req gen.AnyObject, params gen.UpdateApiEndpointParams) (gen.UpdateApiEndpointRes, error) {
	return h.saveEndpointFromBody(ctx, params.ID, req)
}

// ─── createApiEndpointCategory ────────────────────────────────────────────────

func (h *APIHandler) CreateApiEndpointCategory(ctx context.Context, req gen.AnyObject) (gen.CreateApiEndpointCategoryRes, error) {
	return h.saveCategoryFromBody(ctx, uuid.Nil, req)
}

// ─── updateApiEndpointCategory ────────────────────────────────────────────────

func (h *APIHandler) UpdateApiEndpointCategory(ctx context.Context, req gen.AnyObject, params gen.UpdateApiEndpointCategoryParams) (gen.UpdateApiEndpointCategoryRes, error) {
	return h.saveCategoryFromBody(ctx, params.ID, req)
}

// ─── internal save helpers ────────────────────────────────────────────────────

func (h *APIHandler) saveEndpointFromBody(_ context.Context, id uuid.UUID, req gen.AnyObject) (*gen.AnyObject, error) {
	var body struct {
		Code           string   `json:"code"`
		Method         string   `json:"method"`
		Path           string   `json:"path"`
		Summary        string   `json:"summary"`
		CategoryID     string   `json:"category_id"`
		Status         string   `json:"status"`
		Handler        string   `json:"handler"`
		PermissionKeys []string `json:"permission_keys"`
	}
	if err := unmarshalAnyObject(req, &body); err != nil {
		return nil, err
	}
	categoryID, err := parseMaybeUUID(body.CategoryID)
	if err != nil {
		return nil, err
	}
	endpoint := &user.APIEndpoint{
		ID:         id,
		Code:       strings.TrimSpace(body.Code),
		Method:     strings.TrimSpace(body.Method),
		Path:       strings.TrimSpace(body.Path),
		Summary:    strings.TrimSpace(body.Summary),
		CategoryID: categoryID,
		Status:     strings.TrimSpace(body.Status),
		Handler:    strings.TrimSpace(body.Handler),
	}
	saved, err := h.apiEndpointSvc.Save(endpoint, body.PermissionKeys, "")
	if err != nil {
		h.logger.Error("save api endpoint failed", zap.Error(err))
		return nil, err
	}
	bindings, _ := h.apiEndpointSvc.ListBindings(saved.Code)
	categories, _ := h.apiEndpointSvc.ListCategories()
	categoryMap := make(map[uuid.UUID]user.APIEndpointCategory, len(categories))
	for _, cat := range categories {
		categoryMap[cat.ID] = cat
	}
	obj := marshalAnyObject(epToMap(saved, bindings, categoryMap, apiendpoint.EndpointRuntimeState{}))
	return &obj, nil
}

func (h *APIHandler) saveCategoryFromBody(_ context.Context, id uuid.UUID, req gen.AnyObject) (*gen.AnyObject, error) {
	var body struct {
		Code      string `json:"code"`
		Name      string `json:"name"`
		NameEn    string `json:"name_en"`
		SortOrder int    `json:"sort_order"`
		Status    string `json:"status"`
	}
	if err := unmarshalAnyObject(req, &body); err != nil {
		return nil, err
	}
	item := &user.APIEndpointCategory{
		ID:        id,
		Code:      strings.TrimSpace(body.Code),
		Name:      strings.TrimSpace(body.Name),
		NameEn:    strings.TrimSpace(body.NameEn),
		SortOrder: body.SortOrder,
		Status:    strings.TrimSpace(body.Status),
	}
	saved, err := h.apiEndpointSvc.SaveCategory(item)
	if err != nil {
		h.logger.Error("save api endpoint category failed", zap.Error(err))
		return nil, err
	}
	obj := marshalAnyObject(catToMap(saved))
	return &obj, nil
}

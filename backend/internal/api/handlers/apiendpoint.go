// apiendpoint.go — Phase 5 ogen handlers for the api-endpoints domain.
package handlers

import (
	"context"
	"errors"
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

func apiEndpointCategoryItemFromModel(item *user.APIEndpointCategory) gen.ApiEndpointCategoryItem {
	if item == nil {
		return gen.ApiEndpointCategoryItem{}
	}
	return gen.ApiEndpointCategoryItem{
		ID:        item.ID,
		Code:      item.Code,
		Name:      item.Name,
		NameEn:    item.NameEn,
		SortOrder: item.SortOrder,
		Status:    item.Status,
	}
}

func apiEndpointItemFromModel(endpoint *user.APIEndpoint, bindings []user.APIEndpointPermissionBinding, categoryMap map[uuid.UUID]user.APIEndpointCategory, runtimeState apiendpoint.EndpointRuntimeState) gen.ApiEndpointItem {
	permissionKeys := make([]string, 0, len(bindings))
	for _, b := range bindings {
		permissionKeys = append(permissionKeys, b.PermissionKey)
	}

	catID := ""
	category := gen.OptNilApiEndpointCategoryItem{}
	category.SetToNull()
	if endpoint.CategoryID != nil {
		catID = endpoint.CategoryID.String()
		if item, ok := categoryMap[*endpoint.CategoryID]; ok {
			category.SetTo(apiEndpointCategoryItemFromModel(&item))
		}
	}

	seedKey := strings.ToUpper(strings.TrimSpace(endpoint.Method)) + " " + strings.TrimSpace(endpoint.Path)
	authMode := getEpAccessModeMap()[seedKey]
	if authMode == "authenticated" {
		authMode = "jwt"
	}
	if authMode == "" {
		if len(permissionKeys) > 0 {
			authMode = "permission"
		} else {
			authMode = "jwt"
		}
	}

	return gen.ApiEndpointItem{
		ID:             endpoint.ID,
		Code:           endpoint.Code,
		Method:         endpoint.Method,
		Path:           endpoint.Path,
		Spec:           endpoint.Method + " " + endpoint.Path,
		Handler:        endpoint.Handler,
		Summary:        endpoint.Summary,
		PermissionKeys: permissionKeys,
		AuthMode:       authMode,
		CategoryID:     catID,
		Category:       category,
		Status:         endpoint.Status,
		RuntimeExists:  runtimeState.RuntimeExists,
		Stale:          runtimeState.Stale,
		StaleReason:    runtimeState.StaleReason,
		CreatedAt:      endpoint.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:      endpoint.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

func unregisteredMetaFromMap(meta map[string]interface{}) gen.UnregisteredApiEndpointMeta {
	result := gen.UnregisteredApiEndpointMeta{
		PermissionKeys: []string{},
	}
	if len(meta) == 0 {
		return result
	}

	if v, ok := meta["summary"].(string); ok {
		result.Summary = gen.NewOptString(v)
	}
	if v, ok := meta["category_code"].(string); ok {
		result.CategoryCode = gen.NewOptString(v)
	}
	if v, ok := meta["permission_keys"].([]string); ok {
		result.PermissionKeys = append(result.PermissionKeys, v...)
		return result
	}
	if v, ok := meta["permission_keys"].([]interface{}); ok {
		for _, item := range v {
			if key, ok := item.(string); ok {
				result.PermissionKeys = append(result.PermissionKeys, key)
			}
		}
	}
	return result
}

func unregisteredItemFromModel(item apiendpoint.UnregisteredRouteItem) gen.UnregisteredApiEndpointItem {
	return gen.UnregisteredApiEndpointItem{
		Method:  item.Method,
		Path:    item.Path,
		Spec:    item.Spec,
		Handler: item.Handler,
		HasMeta: item.HasMeta,
		Meta:    unregisteredMetaFromMap(item.Meta),
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
	records := make([]gen.ApiEndpointItem, 0, len(list))
	for _, ep := range list {
		records = append(records, apiEndpointItemFromModel(&ep, bindingsMap[ep.Code], categoryMap, runtimeStateMap[ep.ID]))
	}
	return &gen.ApiEndpointList{
		Records: records,
		Total:   total,
		Current: optInt(params.Current, 1),
		Size:    optInt(params.Size, 20),
	}, nil
}

// ─── getApiEndpointOverview ───────────────────────────────────────────────────

func (h *APIHandler) GetApiEndpointOverview(ctx context.Context, params gen.GetApiEndpointOverviewParams) (gen.GetApiEndpointOverviewRes, error) {
	overview, err := h.apiEndpointSvc.Overview(optString(params.AppKey))
	if err != nil {
		h.logger.Error("get api endpoint overview failed", zap.Error(err))
		return nil, err
	}
	return &gen.ApiEndpointOverview{
		TotalCount:             overview.TotalCount,
		UncategorizedCount:     overview.UncategorizedCount,
		StaleCount:             overview.StaleCount,
		NoPermissionCount:     overview.NoPermissionCount,
		SharedPermissionCount: overview.SharedPermissionCount,
		CrossContextSharedCount: overview.CrossContextSharedCount,
		CategoryCounts:         apiEndpointOverviewCategoryCountsFromModel(overview.CategoryCounts),
	}, nil
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
	records := make([]gen.ApiEndpointItem, 0, len(list))
	for _, ep := range list {
		records = append(records, apiEndpointItemFromModel(&ep, nil, categoryMap, runtimeStateMap[ep.ID]))
	}
	return &gen.StaleApiEndpointList{
		Records: records,
		Total:   total,
		Current: optInt(params.Current, 1),
		Size:    optInt(params.Size, 20),
	}, nil
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
	records := make([]gen.UnregisteredApiEndpointItem, 0, len(list))
	for _, item := range list {
		records = append(records, unregisteredItemFromModel(item))
	}
	return &gen.UnregisteredApiEndpointList{
		Records: records,
		Total:   total,
		Current: optInt(params.Current, 1),
		Size:    optInt(params.Size, 20),
	}, nil
}

// ─── listApiEndpointCategories ────────────────────────────────────────────────

func (h *APIHandler) ListApiEndpointCategories(ctx context.Context) (gen.ListApiEndpointCategoriesRes, error) {
	items, err := h.apiEndpointSvc.ListCategories()
	if err != nil {
		h.logger.Error("list api endpoint categories failed", zap.Error(err))
		return nil, err
	}
	records := make([]gen.ApiEndpointCategoryItem, 0, len(items))
	for _, item := range items {
		records = append(records, apiEndpointCategoryItemFromModel(&item))
	}
	return &gen.ApiEndpointCategoryList{
		Records: records,
		Total:   int64(len(records)),
	}, nil
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

func (h *APIHandler) CleanupStaleApiEndpoints(ctx context.Context, req *gen.CleanupStaleRequest) (gen.CleanupStaleApiEndpointsRes, error) {
	if req == nil {
		return nil, errors.New("request body required")
	}
	if len(req.Ids) == 0 {
		return &gen.CleanupStaleResult{DeletedCount: 0}, nil
	}
	endpointIDs := make([]uuid.UUID, 0, len(req.Ids))
	seen := make(map[uuid.UUID]struct{}, len(req.Ids))
	for _, id := range req.Ids {
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
	return &gen.CleanupStaleResult{DeletedCount: int64(deletedCount)}, nil
}

// ─── updateApiEndpoint ────────────────────────────────────────────────────────

func (h *APIHandler) UpdateApiEndpoint(ctx context.Context, req *gen.ApiEndpointSaveRequest, params gen.UpdateApiEndpointParams) (gen.UpdateApiEndpointRes, error) {
	return h.saveEndpointFromBody(ctx, params.ID, req)
}

// ─── createApiEndpointCategory ────────────────────────────────────────────────

func (h *APIHandler) UpdateApiEndpointContextScope(ctx context.Context, req *gen.ApiEndpointSaveRequest, params gen.UpdateApiEndpointContextScopeParams) (gen.UpdateApiEndpointContextScopeRes, error) {
	return h.saveEndpointFromBody(ctx, params.ID, req)
}

func (h *APIHandler) CreateApiEndpointCategory(ctx context.Context, req *gen.ApiEndpointCategorySaveRequest) (gen.CreateApiEndpointCategoryRes, error) {
	return h.saveCategoryFromBody(ctx, uuid.Nil, req)
}

// ─── updateApiEndpointCategory ────────────────────────────────────────────────

func (h *APIHandler) UpdateApiEndpointCategory(ctx context.Context, req *gen.ApiEndpointCategorySaveRequest, params gen.UpdateApiEndpointCategoryParams) (gen.UpdateApiEndpointCategoryRes, error) {
	return h.saveCategoryFromBody(ctx, params.ID, req)
}

// ─── internal save helpers ────────────────────────────────────────────────────

func (h *APIHandler) saveEndpointFromBody(_ context.Context, id uuid.UUID, req *gen.ApiEndpointSaveRequest) (*gen.ApiEndpointItem, error) {
	if req == nil {
		return nil, errors.New("request body required")
	}
	var categoryID *uuid.UUID
	if req.CategoryID.Set {
		parsed, err := uuid.Parse(req.CategoryID.Value)
		if err != nil {
			return nil, err
		}
		categoryID = &parsed
	}
	endpoint := &user.APIEndpoint{
		ID:         id,
		Code:       strings.TrimSpace(req.Code),
		Method:     strings.TrimSpace(req.Method),
		Path:       strings.TrimSpace(req.Path),
		Summary:    strings.TrimSpace(req.Summary),
		CategoryID: categoryID,
		Status:     strings.TrimSpace(req.Status),
		Handler:    strings.TrimSpace(req.Handler),
	}
	saved, err := h.apiEndpointSvc.Save(endpoint, req.PermissionKeys, "")
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
	out := apiEndpointItemFromModel(saved, bindings, categoryMap, apiendpoint.EndpointRuntimeState{})
	return &out, nil
}

func (h *APIHandler) saveCategoryFromBody(_ context.Context, id uuid.UUID, req *gen.ApiEndpointCategorySaveRequest) (*gen.ApiEndpointCategoryItem, error) {
	if req == nil {
		return nil, errors.New("request body required")
	}
	item := &user.APIEndpointCategory{
		ID:        id,
		Code:      strings.TrimSpace(req.Code),
		Name:      strings.TrimSpace(req.Name),
		NameEn:    strings.TrimSpace(req.NameEn),
		SortOrder: req.SortOrder,
		Status:    strings.TrimSpace(req.Status),
	}
	saved, err := h.apiEndpointSvc.SaveCategory(item)
	if err != nil {
		h.logger.Error("save api endpoint category failed", zap.Error(err))
		return nil, err
	}
	out := apiEndpointCategoryItemFromModel(saved)
	return &out, nil
}

func apiEndpointOverviewCategoryCountsFromModel(items []apiendpoint.EndpointCategoryCount) []gen.ApiEndpointOverviewCategoryCountsItem {
	out := make([]gen.ApiEndpointOverviewCategoryCountsItem, 0, len(items))
	for _, item := range items {
		out = append(out, gen.ApiEndpointOverviewCategoryCountsItem{
			CategoryID: item.CategoryID,
			Count:      item.Count,
		})
	}
	return out
}

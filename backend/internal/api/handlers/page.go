// page.go: ogen handler implementations for /pages/* and runtime/sync.
package handlers

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/maben/backend/api/gen"
	"github.com/maben/backend/internal/modules/observability/audit"
	"github.com/maben/backend/internal/modules/system/page"
)

func (h *pageAPIHandler) ListPages(ctx context.Context, params gen.ListPagesParams) (*gen.PageListResponse, error) {
	req := &page.ListRequest{
		Current:  optInt(params.Current, 1),
		Size:     optInt(params.Size, 20),
		Keyword:  optString(params.Keyword),
		AppKey:   params.AppKey,
		SpaceKey: optString(params.SpaceKey),
		Status:   optString(params.Status),
	}
	list, total, err := h.pageSvc.List(req)
	if err != nil {
		h.logger.Error("list pages failed", zap.Error(err))
		return nil, err
	}
	return &gen.PageListResponse{Records: pageRecordsFromModels(list), Total: int(total)}, nil
}

func (h *pageAPIHandler) ListPageOptions(ctx context.Context, params gen.ListPageOptionsParams) (*gen.PageListResponse, error) {
	pages, err := h.pageSvc.ListOptions(params.AppKey, optString(params.SpaceKey))
	if err != nil {
		h.logger.Error("list page options failed", zap.Error(err))
		return nil, err
	}
	records := make([]page.Record, len(pages))
	for i, p := range pages {
		records[i] = page.Record{UIPage: p}
	}
	return &gen.PageListResponse{Records: pageRecordsFromModels(records), Total: len(records)}, nil
}

func (h *pageAPIHandler) ListPageMenuOptions(ctx context.Context, params gen.ListPageMenuOptionsParams) (*gen.PageMenuOptionsResponse, error) {
	list, err := h.pageSvc.ListMenuOptions(params.AppKey, optString(params.SpaceKey))
	if err != nil {
		h.logger.Error("list page menu options failed", zap.Error(err))
		return nil, err
	}
	return &gen.PageMenuOptionsResponse{Records: pageMenuOptionItemsFromModels(list), Total: len(list)}, nil
}

func (h *pageAPIHandler) ListRuntimePages(ctx context.Context, params gen.ListRuntimePagesParams) (*gen.PageListResponse, error) {
	var userID *uuid.UUID
	if uid, ok := userIDFromContext(ctx); ok {
		userID = &uid
	}
	cwID, _ := collaborationWorkspaceIDFromContext(ctx)
	list, err := h.pageSvc.ListRuntime(
		optString(params.AppKey),
		requestHostFromCtx(ctx),
		optString(params.SpaceKey),
		userID,
		cwID,
	)
	if err != nil {
		h.logger.Error("list runtime pages failed", zap.Error(err))
		return nil, err
	}
	return &gen.PageListResponse{Records: pageRecordsFromModels(list), Total: len(list)}, nil
}

func (h *pageAPIHandler) ListPublicRuntimePages(ctx context.Context, params gen.ListPublicRuntimePagesParams) (*gen.PageListResponse, error) {
	list, err := h.pageSvc.ListRuntimePublic(
		optString(params.AppKey),
		requestHostFromCtx(ctx),
		optString(params.SpaceKey),
		nil,
		nil,
	)
	if err != nil {
		h.logger.Error("list public runtime pages failed", zap.Error(err))
		return nil, err
	}
	return &gen.PageListResponse{Records: pageRecordsFromModels(list), Total: len(list)}, nil
}

func (h *pageAPIHandler) ListUnregisteredPages(ctx context.Context, params gen.ListUnregisteredPagesParams) (*gen.PageUnregisteredListResponse, error) {
	list, err := h.pageSvc.ListUnregistered(params.AppKey)
	if err != nil {
		h.logger.Error("list unregistered pages failed", zap.Error(err))
		return nil, err
	}
	return &gen.PageUnregisteredListResponse{Records: pageUnregisteredItemsFromModels(list), Total: len(list)}, nil
}

func (h *pageAPIHandler) SyncPages(ctx context.Context, params gen.SyncPagesParams) (*gen.PageSyncResult, error) {
	result, err := h.pageSvc.Sync(params.AppKey)
	if err != nil {
		h.logger.Error("sync pages failed", zap.Error(err))
		h.audit.Record(ctx, audit.Event{
			Action:       "system.page.sync",
			ResourceType: "app",
			ResourceID:   params.AppKey,
			Outcome:      audit.OutcomeError,
			ErrorCode:    errorCodeOf(err),
		})
		return nil, err
	}
	h.audit.Record(ctx, audit.Event{
		Action:       "system.page.sync",
		ResourceType: "app",
		ResourceID:   params.AppKey,
		Outcome:      audit.OutcomeSuccess,
		Metadata: map[string]any{
			"created_count": result.CreatedCount,
			"skipped_count": result.SkippedCount,
			"created_keys":  result.CreatedKeys,
		},
	})
	return &gen.PageSyncResult{
		CreatedCount: result.CreatedCount,
		SkippedCount: result.SkippedCount,
		CreatedKeys:  result.CreatedKeys,
	}, nil
}

func (h *pageAPIHandler) PreviewPageBreadcrumb(ctx context.Context, params gen.PreviewPageBreadcrumbParams) (*gen.PageBreadcrumbPreviewListResponse, error) {
	list, err := h.pageSvc.PreviewBreadcrumb(params.ID, params.AppKey)
	if err != nil {
		h.logger.Error("preview page breadcrumb failed", zap.Error(err))
		return nil, err
	}
	return &gen.PageBreadcrumbPreviewListResponse{
		Records: pageBreadcrumbPreviewItemsFromModels(list),
		Total:   len(list),
	}, nil
}

func pageBreadcrumbPreviewItemsFromModels(items []page.BreadcrumbPreviewItem) []gen.PageBreadcrumbPreviewItem {
	out := make([]gen.PageBreadcrumbPreviewItem, 0, len(items))
	for _, item := range items {
		record := gen.PageBreadcrumbPreviewItem{
			Type:  item.Type,
			Title: item.Title,
			Path:  item.Path,
		}
		if item.PageKey != "" {
			record.PageKey = gen.NewOptString(item.PageKey)
		}
		out = append(out, record)
	}
	return out
}

func pageRecordsFromModels(items []page.Record) []gen.PageListItem {
	out := make([]gen.PageListItem, 0, len(items))
	for i := range items {
		out = append(out, pageListItemFromModel(&items[i]))
	}
	return out
}

func pageMenuOptionItemsFromModels(items []page.MenuOption) []gen.PageMenuOptionItem {
	out := make([]gen.PageMenuOptionItem, 0, len(items))
	for i := range items {
		out = append(out, pageMenuOptionItemFromModel(items[i]))
	}
	return out
}

func pageMenuOptionItemFromModel(item page.MenuOption) gen.PageMenuOptionItem {
	children := make([]gen.PageMenuOptionItem, 0, len(item.Children))
	for i := range item.Children {
		children = append(children, pageMenuOptionItemFromModel(item.Children[i]))
	}
	return gen.PageMenuOptionItem{
		ID:       item.ID,
		Name:     item.Name,
		Title:    item.Title,
		Path:     item.Path,
		Children: children,
	}
}

func pageUnregisteredItemsFromModels(items []page.UnregisteredRecord) []gen.PageUnregisteredItem {
	out := make([]gen.PageUnregisteredItem, 0, len(items))
	for i := range items {
		item := items[i]
		out = append(out, gen.PageUnregisteredItem{
			FilePath:        item.FilePath,
			Component:       item.Component,
			PageKey:         item.PageKey,
			Name:            item.Name,
			RouteName:       item.RouteName,
			RoutePath:       item.RoutePath,
			PageType:        item.PageType,
			VisibilityScope: item.VisibilityScope,
			ModuleKey:       item.ModuleKey,
			ParentMenuID:    item.ParentMenuID,
			ParentMenuName:  item.ParentMenuName,
			ActiveMenuPath:  item.ActiveMenuPath,
		})
	}
	return out
}

func (h *pageAPIHandler) GetPage(ctx context.Context, params gen.GetPageParams) (*gen.PageSaveResult, error) {
	rec, err := h.pageSvc.Get(params.ID, params.AppKey)
	if err != nil {
		h.logger.Error("get page failed", zap.Error(err))
		return nil, err
	}
	result := pageSaveResultFromModel(rec)
	return &result, nil
}

func (h *pageAPIHandler) CreatePage(ctx context.Context, req *gen.PageSaveRequest, params gen.CreatePageParams) (*gen.PageSaveResult, error) {
	if req == nil {
		return nil, errors.New("request body required")
	}
	saveReq := pageSaveRequestFromGen(req, params.AppKey)
	rec, err := h.pageSvc.Create(saveReq)
	if err != nil {
		h.logger.Error("create page failed", zap.Error(err))
		h.audit.Record(ctx, audit.Event{
			Action:       "system.page.create",
			ResourceType: "page",
			Outcome:      audit.OutcomeError,
			ErrorCode:    errorCodeOf(err),
			Metadata: map[string]any{
				"app_key":  params.AppKey,
				"page_key": saveReq.PageKey,
			},
		})
		return nil, err
	}
	var resourceID string
	if rec != nil {
		resourceID = rec.ID.String()
	}
	h.audit.Record(ctx, audit.Event{
		Action:       "system.page.create",
		ResourceType: "page",
		ResourceID:   resourceID,
		Outcome:      audit.OutcomeSuccess,
		After:        saveReq,
	})
	result := pageSaveResultFromModel(rec)
	return &result, nil
}

func (h *pageAPIHandler) UpdatePage(ctx context.Context, req *gen.PageSaveRequest, params gen.UpdatePageParams) (*gen.PageSaveResult, error) {
	if req == nil {
		return nil, errors.New("request body required")
	}
	saveReq := pageSaveRequestFromGen(req, params.AppKey)
	rec, err := h.pageSvc.Update(params.ID, saveReq)
	if err != nil {
		h.logger.Error("update page failed", zap.Error(err))
		h.audit.Record(ctx, audit.Event{
			Action:       "system.page.update",
			ResourceType: "page",
			ResourceID:   params.ID.String(),
			Outcome:      audit.OutcomeError,
			ErrorCode:    errorCodeOf(err),
		})
		return nil, err
	}
	h.audit.Record(ctx, audit.Event{
		Action:       "system.page.update",
		ResourceType: "page",
		ResourceID:   params.ID.String(),
		Outcome:      audit.OutcomeSuccess,
		After:        saveReq,
	})
	result := pageSaveResultFromModel(rec)
	return &result, nil
}

func (h *pageAPIHandler) DeletePage(ctx context.Context, params gen.DeletePageParams) (*gen.MutationResult, error) {
	if err := h.pageSvc.Delete(params.ID, params.AppKey); err != nil {
		h.logger.Error("delete page failed", zap.Error(err))
		h.audit.Record(ctx, audit.Event{
			Action:       "system.page.delete",
			ResourceType: "page",
			ResourceID:   params.ID.String(),
			Outcome:      audit.OutcomeError,
			ErrorCode:    errorCodeOf(err),
			Metadata:     map[string]any{"app_key": params.AppKey},
		})
		return nil, err
	}
	h.audit.Record(ctx, audit.Event{
		Action:       "system.page.delete",
		ResourceType: "page",
		ResourceID:   params.ID.String(),
		Outcome:      audit.OutcomeSuccess,
		Metadata:     map[string]any{"app_key": params.AppKey},
	})
	return ok(), nil
}

func pageSaveRequestFromGen(req *gen.PageSaveRequest, appKey string) *page.SaveRequest {
	saveReq := &page.SaveRequest{
		AppKey:          appKey,
		PageKey:         req.PageKey,
		Name:            req.Name,
		RouteName:       req.RouteName,
		RoutePath:       req.RoutePath,
		Component:       req.Component,
		SpaceKeys:       req.SpaceKeys,
		PageType:        optString(req.PageType),
		VisibilityScope: optString(req.VisibilityScope),
		Source:          optString(req.Source),
		ModuleKey:       optString(req.ModuleKey),
		SortOrder:       optInt(req.SortOrder, 0),
		ParentMenuID:    optString(req.ParentMenuID),
		ParentPageKey:   optString(req.ParentPageKey),
		DisplayGroupKey: optString(req.DisplayGroupKey),
		ActiveMenuPath:  optString(req.ActiveMenuPath),
		BreadcrumbMode:  optString(req.BreadcrumbMode),
		AccessMode:      optString(req.AccessMode),
		PermissionKey:   optString(req.PermissionKey),
		Status:          optString(req.Status),
	}
	if req.InheritPermission.Set {
		value := req.InheritPermission.Value
		saveReq.InheritPermission = &value
	}
	if req.KeepAlive.Set {
		value := req.KeepAlive.Value
		saveReq.KeepAlive = &value
	}
	if req.IsFullPage.Set {
		value := req.IsFullPage.Value
		saveReq.IsFullPage = &value
	}
	if req.Meta.Set {
		saveReq.Meta = pageMetaToMap(req.Meta.Value)
	}
	if req.RemoteBinding.Set {
		saveReq.RemoteBinding = pageRemoteBindingSaveFromGen(req.RemoteBinding.Value)
	}
	return saveReq
}

func pageSaveResultFromModel(record *page.Record) gen.PageSaveResult {
	if record == nil {
		return gen.PageSaveResult{}
	}
	result := gen.PageSaveResult{
		ID:                record.ID,
		AppKey:            record.AppKey,
		PageKey:           record.PageKey,
		Name:              record.Name,
		RouteName:         record.RouteName,
		RoutePath:         record.RoutePath,
		Component:         record.Component,
		SpaceKey:          record.SpaceKey,
		SpaceKeys:         pageSaveResultSpaceKeys(record.Meta),
		PageType:          record.PageType,
		VisibilityScope:   record.VisibilityScope,
		Source:            record.Source,
		ModuleKey:         record.ModuleKey,
		SortOrder:         record.SortOrder,
		ParentPageKey:     record.ParentPageKey,
		DisplayGroupKey:   record.DisplayGroupKey,
		ActiveMenuPath:    record.ActiveMenuPath,
		BreadcrumbMode:    record.BreadcrumbMode,
		AccessMode:        record.AccessMode,
		PermissionKey:     record.PermissionKey,
		InheritPermission: record.InheritPermission,
		KeepAlive:         record.KeepAlive,
		IsFullPage:        record.IsFullPage,
		Status:            record.Status,
		Meta:              pageMetaFromMap(record.Meta),
		CreatedAt:         record.CreatedAt,
		UpdatedAt:         record.UpdatedAt,
	}
	if record.ParentMenuID != nil {
		result.ParentMenuID = gen.NewOptUUID(*record.ParentMenuID)
	}
	if record.ParentMenuName != "" {
		result.ParentMenuName = gen.NewOptString(record.ParentMenuName)
	}
	if record.ParentPageName != "" {
		result.ParentPageName = gen.NewOptString(record.ParentPageName)
	}
	if record.DisplayGroupName != "" {
		result.DisplayGroupName = gen.NewOptString(record.DisplayGroupName)
	}
	return result
}

func pageListItemFromModel(record *page.Record) gen.PageListItem {
	if record == nil {
		return gen.PageListItem{}
	}
	item := gen.PageListItem{
		ID:                gen.NewOptUUID(record.ID),
		AppKey:            gen.NewOptString(record.AppKey),
		PageKey:           record.PageKey,
		Name:              record.Name,
		RouteName:         gen.NewOptString(record.RouteName),
		RoutePath:         record.RoutePath,
		Component:         gen.NewOptString(record.Component),
		SpaceKey:          gen.NewOptString(record.SpaceKey),
		SpaceKeys:         pageSaveResultSpaceKeys(record.Meta),
		PageType:          gen.NewOptString(record.PageType),
		VisibilityScope:   gen.NewOptString(record.VisibilityScope),
		Source:            gen.NewOptString(record.Source),
		ModuleKey:         gen.NewOptString(record.ModuleKey),
		SortOrder:         gen.NewOptInt(record.SortOrder),
		ParentPageKey:     gen.NewOptString(record.ParentPageKey),
		DisplayGroupKey:   gen.NewOptString(record.DisplayGroupKey),
		ActiveMenuPath:    gen.NewOptString(record.ActiveMenuPath),
		BreadcrumbMode:    gen.NewOptString(record.BreadcrumbMode),
		AccessMode:        gen.NewOptString(record.AccessMode),
		PermissionKey:     gen.NewOptString(record.PermissionKey),
		InheritPermission: gen.NewOptBool(record.InheritPermission),
		KeepAlive:         gen.NewOptBool(record.KeepAlive),
		IsFullPage:        gen.NewOptBool(record.IsFullPage),
		Status:            gen.NewOptString(record.Status),
		Meta:              gen.NewOptPageMeta(pageMetaFromMap(record.Meta)),
		CreatedAt:         gen.NewOptDateTime(record.CreatedAt),
		UpdatedAt:         gen.NewOptDateTime(record.UpdatedAt),
	}
	if remoteBinding, ok := pageRemoteBindingFromMap(record.Meta); ok {
		item.RemoteBinding = gen.NewOptPageRemoteBinding(remoteBinding)
	}
	if record.ParentMenuID != nil {
		item.ParentMenuID = gen.NewOptUUID(*record.ParentMenuID)
	}
	if record.ParentMenuName != "" {
		item.ParentMenuName = gen.NewOptString(record.ParentMenuName)
	}
	if record.ParentPageName != "" {
		item.ParentPageName = gen.NewOptString(record.ParentPageName)
	}
	if record.DisplayGroupName != "" {
		item.DisplayGroupName = gen.NewOptString(record.DisplayGroupName)
	}
	return item
}

func pageMetaFromMap(meta map[string]interface{}) gen.PageMeta {
	out := gen.PageMeta{}
	if len(meta) == 0 {
		return out
	}
	if values := pageSaveResultSpaceKeys(meta); len(values) > 0 {
		out.SpaceKeys = values
	}
	if value, ok := meta["spaceScope"].(string); ok && value != "" {
		out.SpaceScope = gen.NewOptString(value)
	}
	if value, ok := meta["visibilityScope"].(string); ok && value != "" {
		out.VisibilityScope = gen.NewOptString(value)
	}
	if value, ok := meta["link"].(string); ok && value != "" {
		out.Link = gen.NewOptString(value)
	}
	if value, ok := meta["isIframe"].(bool); ok {
		out.IsIframe = gen.NewOptBool(value)
	}
	if value, ok := meta["isHideTab"].(bool); ok {
		out.IsHideTab = gen.NewOptBool(value)
	}
	if value, ok := meta["requiredAction"].(string); ok && value != "" {
		out.RequiredAction = gen.NewOptString(value)
	}
	if values := filterStringArray(meta["requiredActions"]); len(values) > 0 {
		out.RequiredActions = values
	}
	if value, ok := meta["actionMatchMode"].(string); ok && value != "" {
		out.ActionMatchMode = gen.NewOptString(value)
	}
	if value, ok := meta["actionVisibilityMode"].(string); ok && value != "" {
		out.ActionVisibilityMode = gen.NewOptString(value)
	}
	if value, ok := meta["customParent"].(string); ok && value != "" {
		out.CustomParent = gen.NewOptString(value)
	}
	if values := filterStringArray(meta["breadcrumbChain"]); len(values) > 0 {
		out.BreadcrumbChain = values
	}
	if value, ok := meta["hostKey"].(string); ok && value != "" {
		out.HostKey = gen.NewOptString(value)
	}
	if value, ok := meta["spaceType"].(string); ok && value != "" {
		out.SpaceType = gen.NewOptString(value)
	}
	return out
}

func pageRemoteBindingFromMap(meta map[string]interface{}) (gen.PageRemoteBinding, bool) {
	if len(meta) == 0 {
		return gen.PageRemoteBinding{}, false
	}
	readValue := func(keys ...string) string {
		for _, key := range keys {
			if value := strings.TrimSpace(fmt.Sprint(meta[key])); value != "" && value != "<nil>" {
				return value
			}
		}
		return ""
	}

	binding := gen.PageRemoteBinding{}
	hasValue := false
	assign := func(target *gen.OptString, value string) {
		if value == "" {
			return
		}
		*target = gen.NewOptString(value)
		hasValue = true
	}

	assign(&binding.ManifestURL, readValue("manifest_url", "manifestUrl"))
	assign(&binding.RemoteAppKey, readValue("remote_app_key", "remoteAppKey"))
	assign(&binding.RemotePageKey, readValue("remote_page_key", "remotePageKey"))
	assign(&binding.RemoteEntryURL, readValue("remote_entry_url", "remoteEntryUrl"))
	assign(&binding.RemoteRoutePath, readValue("remote_route_path", "remoteRoutePath"))
	assign(&binding.RemoteModule, readValue("remote_module", "remoteModule"))
	assign(&binding.RemoteModuleName, readValue("remote_module_name", "remoteModuleName"))
	assign(&binding.RemoteURL, readValue("remote_url", "remoteUrl"))
	assign(&binding.RuntimeVersion, readValue("runtime_version", "runtimeVersion", "version"))
	assign(&binding.HealthCheckURL, readValue("health_check_url", "healthCheckUrl"))

	return binding, hasValue
}

func pageRemoteBindingSaveFromGen(binding gen.PageRemoteBinding) *page.RemoteBinding {
	return &page.RemoteBinding{
		ManifestURL:      optString(binding.ManifestURL),
		RemoteAppKey:     optString(binding.RemoteAppKey),
		RemotePageKey:    optString(binding.RemotePageKey),
		RemoteEntryURL:   optString(binding.RemoteEntryURL),
		RemoteRoutePath:  optString(binding.RemoteRoutePath),
		RemoteModule:     optString(binding.RemoteModule),
		RemoteModuleName: optString(binding.RemoteModuleName),
		RemoteURL:        optString(binding.RemoteURL),
		RuntimeVersion:   optString(binding.RuntimeVersion),
		HealthCheckURL:   optString(binding.HealthCheckURL),
	}
}

func pageMetaToMap(meta gen.PageMeta) map[string]interface{} {
	out := map[string]interface{}{}
	if len(meta.SpaceKeys) > 0 {
		out["spaceKeys"] = meta.SpaceKeys
	}
	if meta.SpaceScope.Set {
		out["spaceScope"] = meta.SpaceScope.Value
	}
	if meta.VisibilityScope.Set {
		out["visibilityScope"] = meta.VisibilityScope.Value
	}
	if meta.Link.Set {
		out["link"] = meta.Link.Value
	}
	if meta.IsIframe.Set {
		out["isIframe"] = meta.IsIframe.Value
	}
	if meta.IsHideTab.Set {
		out["isHideTab"] = meta.IsHideTab.Value
	}
	if meta.RequiredAction.Set {
		out["requiredAction"] = meta.RequiredAction.Value
	}
	if len(meta.RequiredActions) > 0 {
		out["requiredActions"] = meta.RequiredActions
	}
	if meta.ActionMatchMode.Set {
		out["actionMatchMode"] = meta.ActionMatchMode.Value
	}
	if meta.ActionVisibilityMode.Set {
		out["actionVisibilityMode"] = meta.ActionVisibilityMode.Value
	}
	if meta.CustomParent.Set {
		out["customParent"] = meta.CustomParent.Value
	}
	if len(meta.BreadcrumbChain) > 0 {
		out["breadcrumbChain"] = meta.BreadcrumbChain
	}
	if meta.HostKey.Set {
		out["hostKey"] = meta.HostKey.Value
	}
	if meta.SpaceType.Set {
		out["spaceType"] = meta.SpaceType.Value
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func pageSaveResultSpaceKeys(meta map[string]interface{}) []string {
	if len(meta) == 0 {
		return nil
	}
	raw, ok := meta["spaceKeys"]
	if !ok {
		return nil
	}
	switch values := raw.(type) {
	case []string:
		out := make([]string, 0, len(values))
		for _, value := range values {
			if value != "" {
				out = append(out, value)
			}
		}
		return out
	case []interface{}:
		out := make([]string, 0, len(values))
		for _, value := range values {
			if text, ok := value.(string); ok && text != "" {
				out = append(out, text)
			}
		}
		return out
	default:
		return nil
	}
}


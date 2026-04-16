package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/go-faster/jx"
	"go.uber.org/zap"

	"github.com/maben/backend/api/gen"
	"github.com/maben/backend/internal/modules/system/models"
	"github.com/maben/backend/internal/modules/system/siteconfig"
)

const siteConfigTenantFallback = "default"

// ─── ResolveSiteConfigs ──────────────────────────────────────────────────────

func (h *APIHandler) ResolveSiteConfigs(ctx context.Context, params gen.ResolveSiteConfigsParams) (gen.ResolveSiteConfigsRes, error) {
	req := siteconfig.ResolveRequest{
		AppKey:   optString(params.AppKey),
		Keys:     splitCSV(optString(params.Keys)),
		SetCodes: splitCSV(optString(params.SetCodes)),
	}
	result, err := h.siteConfigSvc.Resolve(ctx, req)
	if err != nil {
		h.logger.Error("resolve site configs", zap.Error(err))
		return &gen.Error{Code: 500, Message: "解析失败"}, nil
	}

	items := gen.SiteConfigResolveResponseItems{}
	for key, item := range result.Items {
		items[key] = gen.SiteConfigResolvedItem{
			Value:     gen.SiteConfigResolvedItemValue(metaToJxRaw(item.Value)),
			Source:    gen.SiteConfigResolvedItemSource(item.Source),
			ValueType: gen.SiteConfigResolvedItemValueType(valueTypeOrDefault(item.ValueType)),
			Sets:      ensureStringSlice(item.Sets),
		}
	}
	return &gen.SiteConfigResolveResponse{
		Items:   items,
		Version: result.Version,
	}, nil
}

// ─── ListSiteConfigs ─────────────────────────────────────────────────────────

func (h *APIHandler) ListSiteConfigs(ctx context.Context, params gen.ListSiteConfigsParams) (gen.ListSiteConfigsRes, error) {
	appKey := optString(params.AppKey)
	list, err := h.siteConfigSvc.ListConfigs(ctx, siteConfigTenantFallback, appKey)
	if err != nil {
		h.logger.Error("list site configs", zap.Error(err))
		return &gen.Error{Code: 500, Message: "查询失败"}, nil
	}
	records := make([]gen.SiteConfigSummary, len(list))
	for i, item := range list {
		records[i] = mapSiteConfigSummary(item)
	}
	return &gen.SiteConfigListResponse{
		Records: records,
		Total:   len(records),
	}, nil
}

// ─── UpsertSiteConfig ────────────────────────────────────────────────────────

func (h *APIHandler) UpsertSiteConfig(ctx context.Context, req *gen.SiteConfigSaveRequest) (gen.UpsertSiteConfigRes, error) {
	cfg, err := buildSiteConfigFromRequest(req, nil)
	if err != nil {
		return &gen.Error{Code: 400, Message: err.Error()}, nil
	}
	if err := h.siteConfigSvc.UpsertConfig(ctx, cfg); err != nil {
		h.logger.Error("upsert site config", zap.Error(err))
		return &gen.Error{Code: 400, Message: err.Error()}, nil
	}
	// 重新查一次以拿到最新的时间戳/id。
	saved, err := h.siteConfigSvc.GetConfig(ctx, cfg.TenantID, cfg.ID)
	if err != nil || saved == nil {
		summary := mapSiteConfigSummary(*cfg)
		return &summary, nil
	}
	summary := mapSiteConfigSummary(*saved)
	return &summary, nil
}

// ─── UpdateSiteConfig ────────────────────────────────────────────────────────

func (h *APIHandler) UpdateSiteConfig(ctx context.Context, req *gen.SiteConfigSaveRequest, params gen.UpdateSiteConfigParams) (gen.UpdateSiteConfigRes, error) {
	existing, err := h.siteConfigSvc.GetConfig(ctx, siteConfigTenantFallback, params.ID)
	if err != nil {
		if errors.Is(err, siteconfig.ErrConfigNotFound) {
			return &gen.Error{Code: 404, Message: err.Error()}, nil
		}
		h.logger.Error("get site config", zap.Error(err))
		return &gen.Error{Code: 404, Message: "查询失败"}, nil
	}
	// PUT 以 id 为主：保留既有的 app_key / config_key；只更新可变字段。
	req.AppKey = gen.NewOptString(existing.AppKey)
	req.ConfigKey = existing.ConfigKey
	cfg, err := buildSiteConfigFromRequest(req, existing)
	if err != nil {
		return &gen.Error{Code: 404, Message: err.Error()}, nil
	}
	if err := h.siteConfigSvc.UpsertConfig(ctx, cfg); err != nil {
		h.logger.Error("update site config", zap.Error(err))
		return &gen.Error{Code: 404, Message: err.Error()}, nil
	}
	saved, err := h.siteConfigSvc.GetConfig(ctx, cfg.TenantID, cfg.ID)
	if err != nil || saved == nil {
		summary := mapSiteConfigSummary(*cfg)
		return &summary, nil
	}
	summary := mapSiteConfigSummary(*saved)
	return &summary, nil
}

// ─── DeleteSiteConfig ────────────────────────────────────────────────────────

func (h *APIHandler) DeleteSiteConfig(ctx context.Context, params gen.DeleteSiteConfigParams) (gen.DeleteSiteConfigRes, error) {
	if err := h.siteConfigSvc.DeleteConfig(ctx, siteConfigTenantFallback, params.ID); err != nil {
		if errors.Is(err, siteconfig.ErrConfigNotFound) {
			e := gen.DeleteSiteConfigNotFound(gen.Error{Code: 404, Message: err.Error()})
			return &e, nil
		}
		h.logger.Error("delete site config", zap.Error(err))
		e := gen.DeleteSiteConfigNotFound(gen.Error{Code: 500, Message: "删除失败"})
		return &e, nil
	}
	ok := gen.DeleteSiteConfigOK(gen.Error{Code: 200, Message: "ok"})
	return &ok, nil
}

// ─── ListSiteConfigSets ──────────────────────────────────────────────────────

func (h *APIHandler) ListSiteConfigSets(ctx context.Context) (gen.ListSiteConfigSetsRes, error) {
	sets, err := h.siteConfigSvc.ListSets(ctx, siteConfigTenantFallback)
	if err != nil {
		h.logger.Error("list site config sets", zap.Error(err))
		return &gen.Error{Code: 500, Message: "查询失败"}, nil
	}
	records := make([]gen.SiteConfigSetSummary, len(sets))
	for i, item := range sets {
		records[i] = mapSiteConfigSetSummary(item.Set, item.ConfigKeys)
	}
	return &gen.SiteConfigSetListResponse{
		Records: records,
		Total:   len(records),
	}, nil
}

// ─── UpsertSiteConfigSet ─────────────────────────────────────────────────────

func (h *APIHandler) UpsertSiteConfigSet(ctx context.Context, req *gen.SiteConfigSetSaveRequest) (gen.UpsertSiteConfigSetRes, error) {
	set := buildSiteConfigSetFromRequest(req, nil)
	if err := h.siteConfigSvc.UpsertSet(ctx, set); err != nil {
		h.logger.Error("upsert site config set", zap.Error(err))
		return &gen.Error{Code: 400, Message: err.Error()}, nil
	}
	saved, err := h.siteConfigSvc.GetSet(ctx, set.TenantID, set.ID)
	if err != nil || saved == nil {
		summary := mapSiteConfigSetSummary(*set, nil)
		return &summary, nil
	}
	summary := mapSiteConfigSetSummary(*saved, nil)
	return &summary, nil
}

// ─── UpdateSiteConfigSet ─────────────────────────────────────────────────────

func (h *APIHandler) UpdateSiteConfigSet(ctx context.Context, req *gen.SiteConfigSetSaveRequest, params gen.UpdateSiteConfigSetParams) (gen.UpdateSiteConfigSetRes, error) {
	existing, err := h.siteConfigSvc.GetSet(ctx, siteConfigTenantFallback, params.ID)
	if err != nil {
		if errors.Is(err, siteconfig.ErrSetNotFound) {
			return &gen.Error{Code: 404, Message: err.Error()}, nil
		}
		h.logger.Error("get site config set", zap.Error(err))
		return &gen.Error{Code: 404, Message: "查询失败"}, nil
	}
	// 保留既有 set_code，仅更新可变字段。
	req.SetCode = existing.SetCode
	set := buildSiteConfigSetFromRequest(req, existing)
	if err := h.siteConfigSvc.UpsertSet(ctx, set); err != nil {
		h.logger.Error("update site config set", zap.Error(err))
		return &gen.Error{Code: 404, Message: err.Error()}, nil
	}
	saved, err := h.siteConfigSvc.GetSet(ctx, set.TenantID, set.ID)
	if err != nil || saved == nil {
		summary := mapSiteConfigSetSummary(*set, nil)
		return &summary, nil
	}
	summary := mapSiteConfigSetSummary(*saved, nil)
	return &summary, nil
}

// ─── DeleteSiteConfigSet ─────────────────────────────────────────────────────

func (h *APIHandler) DeleteSiteConfigSet(ctx context.Context, params gen.DeleteSiteConfigSetParams) (gen.DeleteSiteConfigSetRes, error) {
	if err := h.siteConfigSvc.DeleteSet(ctx, siteConfigTenantFallback, params.ID); err != nil {
		if errors.Is(err, siteconfig.ErrSetNotFound) {
			e := gen.DeleteSiteConfigSetNotFound(gen.Error{Code: 404, Message: err.Error()})
			return &e, nil
		}
		h.logger.Error("delete site config set", zap.Error(err))
		e := gen.DeleteSiteConfigSetNotFound(gen.Error{Code: 500, Message: "删除失败"})
		return &e, nil
	}
	ok := gen.DeleteSiteConfigSetOK(gen.Error{Code: 200, Message: "ok"})
	return &ok, nil
}

// ─── UpdateSiteConfigSetItems ────────────────────────────────────────────────

func (h *APIHandler) UpdateSiteConfigSetItems(ctx context.Context, req *gen.SiteConfigSetItemsRequest, params gen.UpdateSiteConfigSetItemsParams) (gen.UpdateSiteConfigSetItemsRes, error) {
	keys := compactStringsHandler(req.ConfigKeys)
	if err := h.siteConfigSvc.UpdateSetItems(ctx, siteConfigTenantFallback, params.ID, keys); err != nil {
		if errors.Is(err, siteconfig.ErrSetNotFound) {
			return &gen.Error{Code: 404, Message: err.Error()}, nil
		}
		h.logger.Error("update site config set items", zap.Error(err))
		return &gen.Error{Code: 404, Message: err.Error()}, nil
	}
	saved, err := h.siteConfigSvc.GetSet(ctx, siteConfigTenantFallback, params.ID)
	if err != nil || saved == nil {
		return &gen.Error{Code: 404, Message: "set not found"}, nil
	}
	summary := mapSiteConfigSetSummary(*saved, keys)
	return &summary, nil
}

// ─── Mapping helpers ─────────────────────────────────────────────────────────

func mapSiteConfigSummary(cfg models.SiteConfig) gen.SiteConfigSummary {
	summary := gen.SiteConfigSummary{
		ID:        cfg.ID,
		TenantID:  cfg.TenantID,
		AppKey:    cfg.AppKey,
		ConfigKey: cfg.ConfigKey,
		ValueType: gen.SiteConfigSummaryValueType(valueTypeOrDefault(cfg.ValueType)),
		Status:    statusOrDefault(cfg.Status),
	}
	if raw := metaToJxRaw(cfg.ConfigValue); raw != nil {
		summary.ConfigValue = gen.NewOptSiteConfigSummaryConfigValue(gen.SiteConfigSummaryConfigValue(raw))
	} else {
		summary.ConfigValue = gen.NewOptSiteConfigSummaryConfigValue(gen.SiteConfigSummaryConfigValue{})
	}
	if cfg.Label != "" {
		summary.Label = gen.NewOptString(cfg.Label)
	}
	if cfg.Description != "" {
		summary.Description = gen.NewOptString(cfg.Description)
	}
	summary.SortOrder = gen.NewOptInt(cfg.SortOrder)
	summary.IsBuiltin = gen.NewOptBool(cfg.IsBuiltin)
	if !cfg.CreatedAt.IsZero() {
		summary.CreatedAt = gen.NewOptDateTime(cfg.CreatedAt)
	}
	if !cfg.UpdatedAt.IsZero() {
		summary.UpdatedAt = gen.NewOptDateTime(cfg.UpdatedAt)
	}
	return summary
}

func mapSiteConfigSetSummary(set models.SiteConfigSet, keys []string) gen.SiteConfigSetSummary {
	summary := gen.SiteConfigSetSummary{
		ID:         set.ID,
		SetCode:    set.SetCode,
		SetName:    set.SetName,
		Status:     statusOrDefault(set.Status),
		ConfigKeys: ensureStringSlice(keys),
	}
	if set.TenantID != "" {
		summary.TenantID = gen.NewOptString(set.TenantID)
	}
	if set.Description != "" {
		summary.Description = gen.NewOptString(set.Description)
	}
	summary.SortOrder = gen.NewOptInt(set.SortOrder)
	summary.IsBuiltin = gen.NewOptBool(set.IsBuiltin)
	if !set.CreatedAt.IsZero() {
		summary.CreatedAt = gen.NewOptDateTime(set.CreatedAt)
	}
	if !set.UpdatedAt.IsZero() {
		summary.UpdatedAt = gen.NewOptDateTime(set.UpdatedAt)
	}
	return summary
}

// buildSiteConfigFromRequest 构造 models.SiteConfig。existing 非 nil 时保留其 id/tenant。
func buildSiteConfigFromRequest(req *gen.SiteConfigSaveRequest, existing *models.SiteConfig) (*models.SiteConfig, error) {
	if req == nil {
		return nil, errors.New("request is nil")
	}
	configKey := strings.TrimSpace(req.ConfigKey)
	if configKey == "" {
		return nil, errors.New("config_key is required")
	}
	cfg := &models.SiteConfig{
		TenantID:  siteConfigTenantFallback,
		AppKey:    strings.TrimSpace(optString(req.AppKey)),
		ConfigKey: configKey,
	}
	if existing != nil {
		cfg.ID = existing.ID
		cfg.TenantID = existing.TenantID
		cfg.IsBuiltin = existing.IsBuiltin
	}
	// config_value
	if v, ok := req.ConfigValue.Get(); ok {
		cfg.ConfigValue = rawMapToSiteConfigMeta(map[string]jx.Raw(v))
	} else if existing != nil {
		cfg.ConfigValue = existing.ConfigValue
	} else {
		cfg.ConfigValue = models.MetaJSON{}
	}
	// value_type
	if v, ok := req.ValueType.Get(); ok {
		cfg.ValueType = string(v)
	} else if existing != nil {
		cfg.ValueType = existing.ValueType
	} else {
		cfg.ValueType = models.SiteConfigValueTypeString
	}
	// label
	if v, ok := req.Label.Get(); ok {
		cfg.Label = v
	} else if existing != nil {
		cfg.Label = existing.Label
	}
	// description
	if v, ok := req.Description.Get(); ok {
		cfg.Description = v
	} else if existing != nil {
		cfg.Description = existing.Description
	}
	// sort_order
	if v, ok := req.SortOrder.Get(); ok {
		cfg.SortOrder = v
	} else if existing != nil {
		cfg.SortOrder = existing.SortOrder
	}
	// status
	if v, ok := req.Status.Get(); ok {
		cfg.Status = string(v)
	} else if existing != nil {
		cfg.Status = existing.Status
	} else {
		cfg.Status = "normal"
	}
	return cfg, nil
}

// buildSiteConfigSetFromRequest 构造 models.SiteConfigSet。existing 非 nil 时保留其 id。
func buildSiteConfigSetFromRequest(req *gen.SiteConfigSetSaveRequest, existing *models.SiteConfigSet) *models.SiteConfigSet {
	set := &models.SiteConfigSet{
		TenantID: siteConfigTenantFallback,
		SetCode:  strings.TrimSpace(req.SetCode),
		SetName:  strings.TrimSpace(req.SetName),
	}
	if existing != nil {
		set.ID = existing.ID
		set.TenantID = existing.TenantID
		set.IsBuiltin = existing.IsBuiltin
	}
	if v, ok := req.Description.Get(); ok {
		set.Description = v
	} else if existing != nil {
		set.Description = existing.Description
	}
	if v, ok := req.SortOrder.Get(); ok {
		set.SortOrder = v
	} else if existing != nil {
		set.SortOrder = existing.SortOrder
	}
	if v, ok := req.Status.Get(); ok {
		set.Status = string(v)
	} else if existing != nil {
		set.Status = existing.Status
	} else {
		set.Status = "normal"
	}
	return set
}

// rawMapToSiteConfigMeta 始终返回非 nil 的 MetaJSON，即使入参为空（便于 jsonb 落库为 `{}`）。
func rawMapToSiteConfigMeta(src map[string]jx.Raw) models.MetaJSON {
	out := make(models.MetaJSON, len(src))
	for k, raw := range src {
		var v any
		if err := json.Unmarshal([]byte(raw), &v); err != nil {
			continue
		}
		out[k] = v
	}
	return out
}

// ─── Small helpers ───────────────────────────────────────────────────────────

func splitCSV(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func compactStringsHandler(in []string) []string {
	out := make([]string, 0, len(in))
	seen := make(map[string]struct{}, len(in))
	for _, s := range in {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		if _, dup := seen[s]; dup {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return out
}

func ensureStringSlice(in []string) []string {
	if in == nil {
		return []string{}
	}
	return in
}

func valueTypeOrDefault(v string) string {
	if v == "" {
		return models.SiteConfigValueTypeString
	}
	return v
}

func statusOrDefault(v string) string {
	if v == "" {
		return "normal"
	}
	return v
}


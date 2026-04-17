package handlers

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/maben/backend/api/gen"
	"github.com/maben/backend/internal/modules/system/dictionary"
	"github.com/maben/backend/internal/modules/system/models"
)

// ─── ListDictTypes ───────────────────────────────────────────────────────────

func (h *dictionaryAPIHandler) ListDictTypes(ctx context.Context, params gen.ListDictTypesParams) (gen.ListDictTypesRes, error) {
	current := 1
	if v, ok := params.Current.Get(); ok {
		current = v
	}
	size := 20
	if v, ok := params.Size.Get(); ok {
		size = v
	}
	keyword := ""
	if v, ok := params.Keyword.Get(); ok {
		keyword = v
	}
	status := ""
	if v, ok := params.Status.Get(); ok {
		status = v
	}

	types, total, err := h.dictSvc.ListTypes(ctx, current, size, keyword, status)
	if err != nil {
		h.logger.Error("list dict types", zap.Error(err))
		return &gen.Error{Code: 500, Message: "查询失败"}, nil
	}

	// Batch count items
	counts := map[uuid.UUID]int64{}
	if len(types) > 0 {
		ids := make([]uuid.UUID, len(types))
		for i, t := range types {
			ids[i] = t.ID
		}
		counts, _ = h.dictSvc.CountItemsBatch(ctx, ids)
	}

	records := make([]gen.DictTypeSummary, len(types))
	for i, t := range types {
		records[i] = mapDictTypeSummary(t, counts[t.ID])
	}

	return &gen.DictTypeListResponse{
		Records: records,
		Total:   int(total),
		Current: current,
		Size:    size,
	}, nil
}

// ─── CreateDictType ──────────────────────────────────────────────────────────

func (h *dictionaryAPIHandler) CreateDictType(ctx context.Context, req *gen.DictTypeSaveRequest) (gen.CreateDictTypeRes, error) {
	status := ""
	if v, ok := req.Status.Get(); ok {
		status = string(v)
	}
	sortOrder := 0
	if v, ok := req.SortOrder.Get(); ok {
		sortOrder = v
	}
	description := ""
	if v, ok := req.Description.Get(); ok {
		description = v
	}

	dt, err := h.dictSvc.CreateType(ctx, req.Code, req.Name, description, status, sortOrder)
	if err != nil {
		if errors.Is(err, dictionary.ErrCodeDuplicate) {
			return &gen.Error{Code: 400, Message: err.Error()}, nil
		}
		h.logger.Error("create dict type", zap.Error(err))
		return &gen.Error{Code: 500, Message: "创建失败"}, nil
	}

	s := mapDictTypeSummary(*dt, 0)
	return &s, nil
}

// ─── GetDictType ─────────────────────────────────────────────────────────────

func (h *dictionaryAPIHandler) GetDictType(ctx context.Context, params gen.GetDictTypeParams) (gen.GetDictTypeRes, error) {
	dt, items, err := h.dictSvc.GetType(ctx, params.ID)
	if err != nil {
		if errors.Is(err, dictionary.ErrTypeNotFound) {
			return &gen.Error{Code: 404, Message: err.Error()}, nil
		}
		h.logger.Error("get dict type", zap.Error(err))
		return &gen.Error{Code: 500, Message: "查询失败"}, nil
	}

	genItems := make([]gen.DictItemSummary, len(items))
	for i, item := range items {
		genItems[i] = mapDictItemSummary(item)
	}

	return &gen.DictTypeDetail{
		ID:          dt.ID,
		Code:        dt.Code,
		Name:        dt.Name,
		Description: gen.NewOptString(dt.Description),
		Status:      dt.Status,
		IsBuiltin:   dt.IsBuiltin,
		ItemCount:   len(items),
		SortOrder:   gen.NewOptInt(dt.SortOrder),
		CreatedAt:   gen.NewOptDateTime(dt.CreatedAt),
		UpdatedAt:   gen.NewOptDateTime(dt.UpdatedAt),
		Items:       genItems,
	}, nil
}

// ─── UpdateDictType ──────────────────────────────────────────────────────────

func (h *dictionaryAPIHandler) UpdateDictType(ctx context.Context, req *gen.DictTypeSaveRequest, params gen.UpdateDictTypeParams) (gen.UpdateDictTypeRes, error) {
	status := ""
	if v, ok := req.Status.Get(); ok {
		status = string(v)
	}
	sortOrder := 0
	if v, ok := req.SortOrder.Get(); ok {
		sortOrder = v
	}
	description := ""
	if v, ok := req.Description.Get(); ok {
		description = v
	}

	dt, err := h.dictSvc.UpdateType(ctx, params.ID, req.Name, description, status, sortOrder)
	if err != nil {
		if errors.Is(err, dictionary.ErrTypeNotFound) {
			e := gen.UpdateDictTypeNotFound(gen.Error{Code: 404, Message: err.Error()})
			return &e, nil
		}
		h.logger.Error("update dict type", zap.Error(err))
		e := gen.UpdateDictTypeBadRequest(gen.Error{Code: 500, Message: "更新失败"})
		return &e, nil
	}

	count, _ := h.dictSvc.CountItems(ctx, dt.ID)
	s := mapDictTypeSummary(*dt, count)
	return &s, nil
}

// ─── DeleteDictType ──────────────────────────────────────────────────────────

func (h *dictionaryAPIHandler) DeleteDictType(ctx context.Context, params gen.DeleteDictTypeParams) (gen.DeleteDictTypeRes, error) {
	if err := h.dictSvc.DeleteType(ctx, params.ID); err != nil {
		if errors.Is(err, dictionary.ErrTypeNotFound) {
			e := gen.DeleteDictTypeNotFound(gen.Error{Code: 404, Message: err.Error()})
			return &e, nil
		}
		if errors.Is(err, dictionary.ErrBuiltinReadonly) {
			e := gen.DeleteDictTypeBadRequest(gen.Error{Code: 400, Message: err.Error()})
			return &e, nil
		}
		h.logger.Error("delete dict type", zap.Error(err))
		e := gen.DeleteDictTypeBadRequest(gen.Error{Code: 500, Message: "删除失败"})
		return &e, nil
	}

	ok := gen.DeleteDictTypeOK(gen.Error{Code: 200, Message: "ok"})
	return &ok, nil
}

// ─── ListDictItems ───────────────────────────────────────────────────────────

func (h *dictionaryAPIHandler) ListDictItems(ctx context.Context, params gen.ListDictItemsParams) (gen.ListDictItemsRes, error) {
	items, err := h.dictSvc.ListItems(ctx, params.ID)
	if err != nil {
		if errors.Is(err, dictionary.ErrTypeNotFound) {
			return &gen.Error{Code: 404, Message: err.Error()}, nil
		}
		h.logger.Error("list dict items", zap.Error(err))
		return &gen.Error{Code: 500, Message: "查询失败"}, nil
	}

	result := make(gen.ListDictItemsOKApplicationJSON, len(items))
	for i, item := range items {
		result[i] = mapDictItemSummary(item)
	}
	return &result, nil
}

// ─── SaveDictItems ───────────────────────────────────────────────────────────

func (h *dictionaryAPIHandler) SaveDictItems(ctx context.Context, req *gen.DictItemsBatchSaveRequest, params gen.SaveDictItemsParams) (gen.SaveDictItemsRes, error) {
	inputs := make([]dictionary.DictItemInput, len(req.Items))
	for i, item := range req.Items {
		status := ""
		if v, ok := item.Status.Get(); ok {
			status = string(v)
		}
		sortOrder := 0
		if v, ok := item.SortOrder.Get(); ok {
			sortOrder = v
		}
		isDefault := false
		if v, ok := item.IsDefault.Get(); ok {
			isDefault = v
		}
		description := ""
		if v, ok := item.Description.Get(); ok {
			description = v
		}
		inputs[i] = dictionary.DictItemInput{
			Label:       item.Label,
			Value:       item.Value,
			Description: description,
			IsDefault:   isDefault,
			Status:      status,
			SortOrder:   sortOrder,
		}
	}

	items, err := h.dictSvc.SaveItems(ctx, params.ID, inputs)
	if err != nil {
		if errors.Is(err, dictionary.ErrTypeNotFound) {
			e := gen.SaveDictItemsNotFound(gen.Error{Code: 404, Message: err.Error()})
			return &e, nil
		}
		if errors.Is(err, dictionary.ErrItemLabelRequired) ||
			errors.Is(err, dictionary.ErrItemValueRequired) ||
			errors.Is(err, dictionary.ErrItemValueDuplicate) ||
			errors.Is(err, dictionary.ErrItemBuiltinReadonly) ||
			errors.Is(err, dictionary.ErrItemDeleteRequiresSuspended) ||
			errors.Is(err, dictionary.ErrItemBuiltinValueImmutable) {
			e := gen.SaveDictItemsBadRequest(gen.Error{Code: 400, Message: err.Error()})
			return &e, nil
		}
		h.logger.Error("save dict items", zap.Error(err))
		e := gen.SaveDictItemsBadRequest(gen.Error{Code: 500, Message: "保存失败"})
		return &e, nil
	}

	result := make(gen.SaveDictItemsOKApplicationJSON, len(items))
	for i, item := range items {
		result[i] = mapDictItemSummary(item)
	}
	return &result, nil
}

func (h *dictionaryAPIHandler) CreateDictItem(ctx context.Context, req *gen.DictItemSaveRequest, params gen.CreateDictItemParams) (gen.CreateDictItemRes, error) {
	item, err := h.dictSvc.CreateItem(ctx, params.ID, toDictItemInput(*req))
	if err != nil {
		if errors.Is(err, dictionary.ErrTypeNotFound) {
			e := gen.CreateDictItemNotFound(gen.Error{Code: 404, Message: err.Error()})
			return &e, nil
		}
		if errors.Is(err, dictionary.ErrItemLabelRequired) ||
			errors.Is(err, dictionary.ErrItemValueRequired) ||
			errors.Is(err, dictionary.ErrItemValueDuplicate) {
			e := gen.CreateDictItemBadRequest(gen.Error{Code: 400, Message: err.Error()})
			return &e, nil
		}
		h.logger.Error("create dict item", zap.Error(err))
		e := gen.CreateDictItemBadRequest(gen.Error{Code: 500, Message: "创建失败"})
		return &e, nil
	}
	summary := mapDictItemSummary(*item)
	return &summary, nil
}

func (h *dictionaryAPIHandler) UpdateDictItem(ctx context.Context, req *gen.DictItemSaveRequest, params gen.UpdateDictItemParams) (gen.UpdateDictItemRes, error) {
	item, err := h.dictSvc.UpdateItem(ctx, params.ID, params.ItemId, toDictItemInput(*req))
	if err != nil {
		if errors.Is(err, dictionary.ErrItemNotFound) || errors.Is(err, dictionary.ErrTypeNotFound) {
			e := gen.UpdateDictItemNotFound(gen.Error{Code: 404, Message: err.Error()})
			return &e, nil
		}
		if errors.Is(err, dictionary.ErrItemLabelRequired) ||
			errors.Is(err, dictionary.ErrItemValueRequired) ||
			errors.Is(err, dictionary.ErrItemValueDuplicate) ||
			errors.Is(err, dictionary.ErrItemBuiltinValueImmutable) {
			e := gen.UpdateDictItemBadRequest(gen.Error{Code: 400, Message: err.Error()})
			return &e, nil
		}
		h.logger.Error("update dict item", zap.Error(err))
		e := gen.UpdateDictItemBadRequest(gen.Error{Code: 500, Message: "更新失败"})
		return &e, nil
	}
	summary := mapDictItemSummary(*item)
	return &summary, nil
}

func (h *dictionaryAPIHandler) DeleteDictItem(ctx context.Context, params gen.DeleteDictItemParams) (gen.DeleteDictItemRes, error) {
	if err := h.dictSvc.DeleteItem(ctx, params.ID, params.ItemId); err != nil {
		if errors.Is(err, dictionary.ErrItemNotFound) || errors.Is(err, dictionary.ErrTypeNotFound) {
			e := gen.DeleteDictItemNotFound(gen.Error{Code: 404, Message: err.Error()})
			return &e, nil
		}
		if errors.Is(err, dictionary.ErrItemBuiltinReadonly) ||
			errors.Is(err, dictionary.ErrItemDeleteRequiresSuspended) {
			e := gen.DeleteDictItemBadRequest(gen.Error{Code: 400, Message: err.Error()})
			return &e, nil
		}
		h.logger.Error("delete dict item", zap.Error(err))
		e := gen.DeleteDictItemBadRequest(gen.Error{Code: 500, Message: "删除失败"})
		return &e, nil
	}
	ok := gen.DeleteDictItemOK(gen.Error{Code: 200, Message: "ok"})
	return &ok, nil
}

// ─── GetDictsByCodes ─────────────────────────────────────────────────────────

func (h *dictionaryAPIHandler) GetDictsByCodes(ctx context.Context, params gen.GetDictsByCodesParams) (gen.GetDictsByCodesRes, error) {
	codes := strings.Split(params.Codes, ",")
	// Trim whitespace
	cleaned := make([]string, 0, len(codes))
	for _, c := range codes {
		c = strings.TrimSpace(c)
		if c != "" {
			cleaned = append(cleaned, c)
		}
	}
	if len(cleaned) == 0 {
		empty := gen.DictsByCodesResponse{}
		return &empty, nil
	}

	result, err := h.dictSvc.GetByCodes(ctx, cleaned)
	if err != nil {
		h.logger.Error("get dicts by codes", zap.Error(err))
		return &gen.Error{Code: 500, Message: "查询失败"}, nil
	}

	resp := gen.DictsByCodesResponse{}
	for code, items := range result {
		genItems := make([]gen.DictItemSummary, len(items))
		for i, item := range items {
			genItems[i] = mapDictItemSummary(item)
		}
		resp[code] = genItems
	}
	return &resp, nil
}

// ─── Mapping helpers ─────────────────────────────────────────────────────────

func mapDictTypeSummary(t models.DictType, itemCount int64) gen.DictTypeSummary {
	return gen.DictTypeSummary{
		ID:          t.ID,
		Code:        t.Code,
		Name:        t.Name,
		Description: gen.NewOptString(t.Description),
		Status:      t.Status,
		IsBuiltin:   t.IsBuiltin,
		ItemCount:   int(itemCount),
		SortOrder:   gen.NewOptInt(t.SortOrder),
		CreatedAt:   gen.NewOptDateTime(t.CreatedAt),
		UpdatedAt:   gen.NewOptDateTime(t.UpdatedAt),
	}
}

func mapDictItemSummary(item models.DictItem) gen.DictItemSummary {
	return gen.DictItemSummary{
		ID:          item.ID,
		Label:       item.Label,
		Value:       item.Value,
		Description: gen.NewOptString(item.Description),
		IsBuiltin:   item.IsBuiltin,
		IsDefault:   gen.NewOptBool(item.IsDefault),
		Status:      item.Status,
		SortOrder:   gen.NewOptInt(item.SortOrder),
	}
}

func toDictItemInput(item gen.DictItemSaveRequest) dictionary.DictItemInput {
	status := ""
	if v, ok := item.Status.Get(); ok {
		status = string(v)
	}
	sortOrder := 0
	if v, ok := item.SortOrder.Get(); ok {
		sortOrder = v
	}
	isDefault := false
	if v, ok := item.IsDefault.Get(); ok {
		isDefault = v
	}
	description := ""
	if v, ok := item.Description.Get(); ok {
		description = v
	}
	return dictionary.DictItemInput{
		Label:       item.Label,
		Value:       item.Value,
		Description: description,
		IsDefault:   isDefault,
		Status:      status,
		SortOrder:   sortOrder,
	}
}


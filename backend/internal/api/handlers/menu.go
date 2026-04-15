// menu.go: ogen handler implementations for /menus/*.
package handlers

import (
	"context"
	"errors"
	"strings"

	"go.uber.org/zap"

	"github.com/gg-ecommerce/backend/api/gen"
	"github.com/gg-ecommerce/backend/internal/api/dto"
	"github.com/gg-ecommerce/backend/internal/modules/observability/audit"
	"github.com/gg-ecommerce/backend/internal/modules/system/user"
)

func (h *APIHandler) GetMenuTree(ctx context.Context, params gen.GetMenuTreeParams) (*gen.MenuTreeResponse, error) {
	all := optString(params.All) == "true" || optString(params.All) == "1"
	nodes, err := h.menuSvc.GetTree(all, nil, optString(params.AppKey), optString(params.SpaceKey))
	if err != nil {
		h.logger.Error("get menu tree failed", zap.Error(err))
		return nil, err
	}
	return &gen.MenuTreeResponse{Records: menuTreeItemsFromModels(nodes)}, nil
}

func (h *APIHandler) CreateMenu(ctx context.Context, req *gen.MenuSaveRequest) (*gen.MutationResult, error) {
	if req == nil {
		return nil, errors.New("request body required")
	}
	var parentID *string
	if req.ParentID.Set && !req.ParentID.Null {
		s := req.ParentID.Value.String()
		parentID = &s
	}
	dtoReq := &dto.MenuCreateRequest{
		AppKey:    optString(req.AppKey),
		ParentID:  parentID,
		SpaceKey:  optString(req.SpaceKey),
		Kind:      req.Kind,
		Path:      optString(req.Path),
		Component: optString(req.Component),
		Name:      req.Name,
		Title:     req.Name,
		Icon:      optString(req.Icon),
		SortOrder: optInt(req.SortOrder, 0),
	}
	created, err := h.menuSvc.Create(dtoReq)
	if err != nil {
		h.logger.Error("create menu failed", zap.Error(err))
		h.audit.Record(ctx, audit.Event{
			Action:       "system.menu.create",
			ResourceType: "menu",
			Outcome:      audit.OutcomeError,
			ErrorCode:    errorCodeOf(err),
			Metadata: map[string]any{
				"app_key":   dtoReq.AppKey,
				"space_key": dtoReq.SpaceKey,
				"name":      dtoReq.Name,
			},
		})
		return nil, err
	}
	var resourceID string
	if created != nil {
		resourceID = created.ID.String()
	}
	h.audit.Record(ctx, audit.Event{
		Action:       "system.menu.create",
		ResourceType: "menu",
		ResourceID:   resourceID,
		Outcome:      audit.OutcomeSuccess,
		After:        dtoReq,
	})
	return ok(), nil
}

func (h *APIHandler) UpdateMenu(ctx context.Context, req *gen.MenuSaveRequest, params gen.UpdateMenuParams) (*gen.MutationResult, error) {
	if req == nil {
		return nil, errors.New("request body required")
	}
	var parentID *string
	if req.ParentID.Set {
		if req.ParentID.Null {
			empty := ""
			parentID = &empty
		} else {
			s := req.ParentID.Value.String()
			parentID = &s
		}
	}
	dtoReq := &dto.MenuUpdateRequest{
		AppKey:    optString(req.AppKey),
		ParentID:  parentID,
		SpaceKey:  optString(req.SpaceKey),
		Kind:      req.Kind,
		Path:      optString(req.Path),
		Component: optString(req.Component),
		Name:      req.Name,
		Title:     req.Name,
		Icon:      optString(req.Icon),
		SortOrder: optInt(req.SortOrder, 0),
	}
	if err := h.menuSvc.Update(params.ID, dtoReq); err != nil {
		h.logger.Error("update menu failed", zap.Error(err))
		h.audit.Record(ctx, audit.Event{
			Action:       "system.menu.update",
			ResourceType: "menu",
			ResourceID:   params.ID.String(),
			Outcome:      audit.OutcomeError,
			ErrorCode:    errorCodeOf(err),
		})
		return nil, err
	}
	h.audit.Record(ctx, audit.Event{
		Action:       "system.menu.update",
		ResourceType: "menu",
		ResourceID:   params.ID.String(),
		Outcome:      audit.OutcomeSuccess,
		After:        dtoReq,
	})
	return ok(), nil
}

func (h *APIHandler) DeleteMenu(ctx context.Context, params gen.DeleteMenuParams) (*gen.MutationResult, error) {
	if err := h.menuSvc.Delete(params.ID, "", nil); err != nil {
		h.logger.Error("delete menu failed", zap.Error(err))
		h.audit.Record(ctx, audit.Event{
			Action:       "system.menu.delete",
			ResourceType: "menu",
			ResourceID:   params.ID.String(),
			Outcome:      audit.OutcomeError,
			ErrorCode:    errorCodeOf(err),
		})
		return nil, err
	}
	h.audit.Record(ctx, audit.Event{
		Action:       "system.menu.delete",
		ResourceType: "menu",
		ResourceID:   params.ID.String(),
		Outcome:      audit.OutcomeSuccess,
	})
	return ok(), nil
}

func menuTreeItemsFromModels(nodes []*user.Menu) []gen.MenuTreeItem {
	items := make([]gen.MenuTreeItem, 0, len(nodes))
	for _, node := range nodes {
		if node == nil {
			continue
		}
		items = append(items, menuTreeItemFromModel(node))
	}
	return items
}

func menuTreeItemFromModel(node *user.Menu) gen.MenuTreeItem {
	item := gen.MenuTreeItem{
		ID:        node.ID,
		AppKey:    node.AppKey,
		SpaceKey:  node.SpaceKey,
		Kind:      node.Kind,
		Path:      node.Path,
		Name:      node.Name,
		Component: node.Component,
		Meta:      menuTreeMetaFromModel(node),
		SortOrder: node.SortOrder,
		Children:  menuTreeItemsFromModels(node.Children),
	}
	if node.ParentID != nil {
		item.ParentID = gen.OptNilUUID{
			Value: *node.ParentID,
			Set:   true,
		}
	}
	if node.Title != "" {
		item.Title = gen.NewOptString(node.Title)
	}
	if node.Icon != "" {
		item.Icon = gen.NewOptString(node.Icon)
	}
	if node.Hidden {
		item.Hidden = gen.NewOptBool(true)
	}
	return item
}

func menuTreeMetaFromModel(node *user.Menu) gen.MenuTreeMeta {
	meta := gen.MenuTreeMeta{
		Title: node.Title,
	}
	if node.Icon != "" {
		meta.Icon = gen.NewOptString(node.Icon)
	}
	if node.Meta == nil {
		return meta
	}
	if accessMode := strings.TrimSpace(toStringValue(node.Meta["accessMode"])); accessMode != "" {
		meta.AccessMode = gen.NewOptString(accessMode)
	}
	if link := strings.TrimSpace(toStringValue(node.Meta["link"])); link != "" {
		meta.Link = gen.NewOptString(link)
	}
	if activePath := strings.TrimSpace(toStringValue(node.Meta["activePath"])); activePath != "" {
		meta.ActivePath = gen.NewOptString(activePath)
	}
	if roles := filterStringArray(node.Meta["roles"]); len(roles) > 0 {
		meta.Roles = append(meta.Roles, roles...)
	}
	if value, ok := node.Meta["isEnable"].(bool); ok {
		meta.IsEnable = gen.NewOptBool(value)
	}
	if value, ok := node.Meta["isHide"].(bool); ok && value {
		meta.IsHide = gen.NewOptBool(true)
	}
	if value, ok := node.Meta["isIframe"].(bool); ok && value {
		meta.IsIframe = gen.NewOptBool(true)
	}
	if value, ok := node.Meta["isHideTab"].(bool); ok && value {
		meta.IsHideTab = gen.NewOptBool(true)
	}
	if value, ok := node.Meta["keepAlive"].(bool); ok && value {
		meta.KeepAlive = gen.NewOptBool(true)
	}
	if value, ok := node.Meta["fixedTab"].(bool); ok && value {
		meta.FixedTab = gen.NewOptBool(true)
	}
	if value, ok := node.Meta["isFullPage"].(bool); ok && value {
		meta.IsFullPage = gen.NewOptBool(true)
	}
	return meta
}

func toStringValue(value any) string {
	text, _ := value.(string)
	return text
}

func filterStringArray(value any) []string {
	raw, ok := value.([]any)
	if !ok {
		if typed, ok := value.([]string); ok {
			result := make([]string, 0, len(typed))
			for _, item := range typed {
				if trimmed := strings.TrimSpace(item); trimmed != "" {
					result = append(result, trimmed)
				}
			}
			return result
		}
		return nil
	}
	result := make([]string, 0, len(raw))
	for _, item := range raw {
		text, ok := item.(string)
		if !ok {
			continue
		}
		if trimmed := strings.TrimSpace(text); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}


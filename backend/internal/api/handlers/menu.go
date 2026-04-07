// menu.go: ogen handler implementations for /menus/* and menu groups/backups.
package handlers

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/gg-ecommerce/backend/api/gen"
	"github.com/gg-ecommerce/backend/internal/api/dto"
)

func (h *APIHandler) GetMenuTree(ctx context.Context, params gen.GetMenuTreeParams) (*gen.MenuTreeResponse, error) {
	all := optString(params.All) == "true" || optString(params.All) == "1"
	nodes, err := h.menuSvc.GetTree(all, nil, optString(params.AppKey), optString(params.SpaceKey))
	if err != nil {
		h.logger.Error("get menu tree failed", zap.Error(err))
		return nil, err
	}
	return &gen.MenuTreeResponse{Records: marshalList(nodes)}, nil
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
		Name:      req.Name,
		Title:     req.Name,
		Icon:      optString(req.Icon),
		SortOrder: optInt(req.SortOrder, 0),
	}
	if _, err := h.menuSvc.Create(dtoReq); err != nil {
		h.logger.Error("create menu failed", zap.Error(err))
		return nil, err
	}
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
		Name:      req.Name,
		Title:     req.Name,
		Icon:      optString(req.Icon),
		SortOrder: optInt(req.SortOrder, 0),
	}
	if err := h.menuSvc.Update(params.ID, dtoReq); err != nil {
		h.logger.Error("update menu failed", zap.Error(err))
		return nil, err
	}
	return ok(), nil
}

func (h *APIHandler) DeleteMenu(ctx context.Context, params gen.DeleteMenuParams) (*gen.MutationResult, error) {
	if err := h.menuSvc.Delete(params.ID, "", nil); err != nil {
		h.logger.Error("delete menu failed", zap.Error(err))
		return nil, err
	}
	return ok(), nil
}

func (h *APIHandler) ListMenuGroups(ctx context.Context) (*gen.MenuGroupList, error) {
	list, err := h.menuSvc.ListGroups()
	if err != nil {
		h.logger.Error("list menu groups failed", zap.Error(err))
		return nil, err
	}
	return &gen.MenuGroupList{Records: marshalList(list), Total: len(list)}, nil
}

func (h *APIHandler) CreateMenuGroup(ctx context.Context, req *gen.MenuGroupSaveRequest) (*gen.MutationResult, error) {
	if req == nil {
		return nil, errors.New("request body required")
	}
	dtoReq := &dto.MenuManageGroupCreateRequest{
		Name:      req.Name,
		SortOrder: optInt(req.SortOrder, 0),
		Status:    optString(req.Status),
	}
	if _, err := h.menuSvc.CreateGroup(dtoReq); err != nil {
		h.logger.Error("create menu group failed", zap.Error(err))
		return nil, err
	}
	return ok(), nil
}

func (h *APIHandler) UpdateMenuGroup(ctx context.Context, req *gen.MenuGroupSaveRequest, params gen.UpdateMenuGroupParams) (*gen.MutationResult, error) {
	if req == nil {
		return nil, errors.New("request body required")
	}
	dtoReq := &dto.MenuManageGroupUpdateRequest{
		Name:      req.Name,
		SortOrder: optInt(req.SortOrder, 0),
		Status:    optString(req.Status),
	}
	if err := h.menuSvc.UpdateGroup(params.ID, dtoReq); err != nil {
		h.logger.Error("update menu group failed", zap.Error(err))
		return nil, err
	}
	return ok(), nil
}

func (h *APIHandler) DeleteMenuGroup(ctx context.Context, params gen.DeleteMenuGroupParams) (*gen.MutationResult, error) {
	if err := h.menuSvc.DeleteGroup(params.ID); err != nil {
		h.logger.Error("delete menu group failed", zap.Error(err))
		return nil, err
	}
	return ok(), nil
}

func (h *APIHandler) ListMenuBackups(ctx context.Context) (*gen.MenuBackupList, error) {
	list, err := h.menuSvc.ListBackups("", "")
	if err != nil {
		h.logger.Error("list menu backups failed", zap.Error(err))
		return nil, err
	}
	return &gen.MenuBackupList{Records: marshalList(list), Total: len(list)}, nil
}

func (h *APIHandler) CreateMenuBackup(ctx context.Context, req *gen.MenuBackupCreateRequest) (*gen.MutationResult, error) {
	if req == nil {
		return nil, errors.New("request body required")
	}
	var createdBy *uuid.UUID
	if uid, ok := userIDFromContext(ctx); ok {
		createdBy = &uid
	}
	if err := h.menuSvc.CreateBackup(req.Name, optString(req.Description), "", "", "", createdBy); err != nil {
		h.logger.Error("create menu backup failed", zap.Error(err))
		return nil, err
	}
	return ok(), nil
}

func (h *APIHandler) DeleteMenuBackup(ctx context.Context, params gen.DeleteMenuBackupParams) (*gen.MutationResult, error) {
	if err := h.menuSvc.DeleteBackup(params.ID, ""); err != nil {
		h.logger.Error("delete menu backup failed", zap.Error(err))
		return nil, err
	}
	return ok(), nil
}

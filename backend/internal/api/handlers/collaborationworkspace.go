// collaborationworkspace.go: ogen handler implementations for
// /collaboration-workspaces/*.
package handlers

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/maben/backend/api/gen"
	"github.com/maben/backend/internal/api/dto"
	"github.com/maben/backend/internal/modules/system/user"
)

func (h *APIHandler) ListCollaborationWorkspaces(ctx context.Context, params gen.ListCollaborationWorkspacesParams) (*gen.CollaborationWorkspaceList, error) {
	req := &dto.CollaborationWorkspaceListRequest{
		Current: optInt(params.Current, 1),
		Size:    optInt(params.Size, 20),
		Name:    optString(params.Keyword),
	}
	list, total, err := h.cwSvc.List(req)
	if err != nil {
		h.logger.Error("list collaboration workspaces failed", zap.Error(err))
		return nil, err
	}
	return &gen.CollaborationWorkspaceList{
		Records: collaborationWorkspaceItemsFromModels(list),
		Total:   int(total),
	}, nil
}

func (h *APIHandler) ListCollaborationWorkspaceOptions(ctx context.Context) (*gen.CollaborationWorkspaceList, error) {
	list, err := h.cwSvc.ListOptions(&dto.CollaborationWorkspaceListRequest{Current: 1, Size: 500})
	if err != nil {
		h.logger.Error("list collaboration workspace options failed", zap.Error(err))
		return nil, err
	}
	return &gen.CollaborationWorkspaceList{
		Records: collaborationWorkspaceItemsFromModels(list),
		Total:   len(list),
	}, nil
}

func (h *APIHandler) GetCollaborationWorkspace(ctx context.Context, params gen.GetCollaborationWorkspaceParams) (*gen.CollaborationWorkspaceItem, error) {
	cw, err := h.cwSvc.Get(params.ID)
	if err != nil {
		h.logger.Error("get collaboration workspace failed", zap.Error(err))
		return nil, err
	}
	item := collaborationWorkspaceItemFromModel(*cw)
	return &item, nil
}

func (h *APIHandler) CreateCollaborationWorkspace(ctx context.Context, req *gen.CollaborationWorkspaceSaveRequest) (*gen.IDResult, error) {
	if req == nil {
		return nil, errors.New("request body required")
	}
	dtoReq := dto.CollaborationWorkspaceCreateRequest{
		Name:       req.Name,
		Remark:     optString(req.Remark),
		LogoURL:    optString(req.LogoURL),
		Plan:       optString(req.Plan),
		MaxMembers: optInt(req.MaxMembers, 0),
		Status:     optString(req.Status),
	}
	var ownerID *uuid.UUID
	if uid, ok := userIDFromContext(ctx); ok {
		ownerID = &uid
	}
	created, err := h.cwSvc.Create(&dtoReq, ownerID)
	if err != nil {
		h.logger.Error("create collaboration workspace failed", zap.Error(err))
		return nil, err
	}
	return &gen.IDResult{ID: created.ID}, nil
}

func (h *APIHandler) UpdateCollaborationWorkspace(ctx context.Context, req *gen.CollaborationWorkspaceSaveRequest, params gen.UpdateCollaborationWorkspaceParams) (*gen.MutationResult, error) {
	if req == nil {
		return nil, errors.New("request body required")
	}
	dtoReq := dto.CollaborationWorkspaceUpdateRequest{
		Name:       req.Name,
		Remark:     optString(req.Remark),
		LogoURL:    optString(req.LogoURL),
		Plan:       optString(req.Plan),
		MaxMembers: optInt(req.MaxMembers, 0),
		Status:     optString(req.Status),
	}
	if err := h.cwSvc.Update(params.ID, &dtoReq); err != nil {
		h.logger.Error("update collaboration workspace failed", zap.Error(err))
		return nil, err
	}
	return ok(), nil
}

func (h *APIHandler) DeleteCollaborationWorkspace(ctx context.Context, params gen.DeleteCollaborationWorkspaceParams) (*gen.MutationResult, error) {
	if err := h.cwSvc.Delete(params.ID); err != nil {
		h.logger.Error("delete collaboration workspace failed", zap.Error(err))
		return nil, err
	}
	return ok(), nil
}

func (h *APIHandler) ListCollaborationWorkspaceMembers(ctx context.Context, params gen.ListCollaborationWorkspaceMembersParams) (*gen.CollaborationWorkspaceMemberList, error) {
	list, err := h.cwSvc.ListMembers(params.ID, &user.MemberSearchParams{})
	if err != nil {
		h.logger.Error("list collaboration workspace members failed", zap.Error(err))
		return nil, err
	}
	return &gen.CollaborationWorkspaceMemberList{
		Records: collaborationWorkspaceMemberItemsFromModels(list),
		Total:   len(list),
	}, nil
}

func (h *APIHandler) AddCollaborationWorkspaceMember(ctx context.Context, req *gen.CollaborationWorkspaceMemberAddRequest, params gen.AddCollaborationWorkspaceMemberParams) (*gen.MutationResult, error) {
	if req == nil {
		return nil, errors.New("request body required")
	}
	dtoReq := dto.CollaborationWorkspaceAddMemberRequest{
		UserID:   req.UserID.String(),
		RoleCode: optString(req.RoleCode),
	}
	var invitedBy *uuid.UUID
	if uid, ok := userIDFromContext(ctx); ok {
		invitedBy = &uid
	}
	if err := h.cwSvc.AddMember(params.ID, &dtoReq, invitedBy); err != nil {
		h.logger.Error("add collaboration workspace member failed", zap.Error(err))
		return nil, err
	}
	return ok(), nil
}

func (h *APIHandler) RemoveCollaborationWorkspaceMember(ctx context.Context, params gen.RemoveCollaborationWorkspaceMemberParams) (*gen.MutationResult, error) {
	if err := h.cwSvc.RemoveMember(params.ID, params.UserId); err != nil {
		h.logger.Error("remove collaboration workspace member failed", zap.Error(err))
		return nil, err
	}
	return ok(), nil
}

func (h *APIHandler) UpdateCollaborationWorkspaceMemberRole(ctx context.Context, req *gen.CollaborationWorkspaceMemberRoleRequest, params gen.UpdateCollaborationWorkspaceMemberRoleParams) (*gen.MutationResult, error) {
	if req == nil {
		return nil, errors.New("request body required")
	}
	if err := h.cwSvc.UpdateMemberRole(params.ID, params.UserId, req.RoleCode); err != nil {
		h.logger.Error("update collaboration workspace member role failed", zap.Error(err))
		return nil, err
	}
	return ok(), nil
}

func (h *APIHandler) GetCollaborationWorkspacePackages(ctx context.Context, params gen.GetCollaborationWorkspacePackagesParams) (*gen.FeaturePackageAssignmentResponse, error) {
	ids, pkgs, err := h.featurePkgSvc.GetCollaborationWorkspacePackages(params.CollaborationWorkspaceId, optString(params.AppKey))
	if err != nil {
		h.logger.Error("get collaboration workspace packages failed", zap.Error(err))
		return nil, err
	}
	return &gen.FeaturePackageAssignmentResponse{
		PackageIds: ids,
		Packages:   featurePackageRefsFromModels(pkgs),
	}, nil
}

func (h *APIHandler) SetCollaborationWorkspacePackages(ctx context.Context, req *gen.UUIDListRequest, params gen.SetCollaborationWorkspacePackagesParams) (*gen.FeaturePackageMutationResult, error) {
	var grantedBy *uuid.UUID
	if uid, ok := userIDFromContext(ctx); ok {
		grantedBy = &uid
	}
	stats, err := h.featurePkgSvc.SetCollaborationWorkspacePackages(params.CollaborationWorkspaceId, uuidIDsFromRequest(req), grantedBy, optString(params.AppKey))
	if err != nil {
		h.logger.Error("set collaboration workspace packages failed", zap.Error(err))
		return nil, err
	}
	return featurePackageMutationResultFromStats(stats), nil
}

var _ = errors.New


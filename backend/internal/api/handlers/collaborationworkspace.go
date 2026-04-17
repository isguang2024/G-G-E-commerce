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

func (h *cwAPIHandler) ListCollaborations(ctx context.Context, params gen.ListCollaborationsParams) (*gen.CollaborationList, error) {
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
	return &gen.CollaborationList{
		Records: collaborationWorkspaceItemsFromModels(list),
		Total:   int(total),
	}, nil
}

func (h *cwAPIHandler) ListCollaborationOptions(ctx context.Context) (*gen.CollaborationList, error) {
	list, err := h.cwSvc.ListOptions(&dto.CollaborationWorkspaceListRequest{Current: 1, Size: 500})
	if err != nil {
		h.logger.Error("list collaboration workspace options failed", zap.Error(err))
		return nil, err
	}
	return &gen.CollaborationList{
		Records: collaborationWorkspaceItemsFromModels(list),
		Total:   len(list),
	}, nil
}

func (h *cwAPIHandler) GetCollaboration(ctx context.Context, params gen.GetCollaborationParams) (*gen.CollaborationItem, error) {
	cw, err := h.cwSvc.Get(params.ID)
	if err != nil {
		h.logger.Error("get collaboration workspace failed", zap.Error(err))
		return nil, err
	}
	item := collaborationWorkspaceItemFromModel(*cw)
	return &item, nil
}

func (h *cwAPIHandler) CreateCollaboration(ctx context.Context, req *gen.CollaborationSaveRequest) (*gen.IDResult, error) {
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

func (h *cwAPIHandler) UpdateCollaboration(ctx context.Context, req *gen.CollaborationSaveRequest, params gen.UpdateCollaborationParams) (*gen.MutationResult, error) {
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

func (h *cwAPIHandler) DeleteCollaboration(ctx context.Context, params gen.DeleteCollaborationParams) (*gen.MutationResult, error) {
	if err := h.cwSvc.Delete(params.ID); err != nil {
		h.logger.Error("delete collaboration workspace failed", zap.Error(err))
		return nil, err
	}
	return ok(), nil
}

func (h *cwAPIHandler) ListCollaborationMembers(ctx context.Context, params gen.ListCollaborationMembersParams) (*gen.CollaborationMemberList, error) {
	list, err := h.cwSvc.ListMembers(params.ID, &user.MemberSearchParams{})
	if err != nil {
		h.logger.Error("list collaboration workspace members failed", zap.Error(err))
		return nil, err
	}
	return &gen.CollaborationMemberList{
		Records: collaborationWorkspaceMemberItemsFromModels(list),
		Total:   len(list),
	}, nil
}

func (h *cwAPIHandler) AddCollaborationMember(ctx context.Context, req *gen.CollaborationMemberAddRequest, params gen.AddCollaborationMemberParams) (*gen.MutationResult, error) {
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

func (h *cwAPIHandler) RemoveCollaborationMember(ctx context.Context, params gen.RemoveCollaborationMemberParams) (*gen.MutationResult, error) {
	if err := h.cwSvc.RemoveMember(params.ID, params.UserId); err != nil {
		h.logger.Error("remove collaboration workspace member failed", zap.Error(err))
		return nil, err
	}
	return ok(), nil
}

func (h *cwAPIHandler) UpdateCollaborationMemberRole(ctx context.Context, req *gen.CollaborationMemberRoleRequest, params gen.UpdateCollaborationMemberRoleParams) (*gen.MutationResult, error) {
	if req == nil {
		return nil, errors.New("request body required")
	}
	if err := h.cwSvc.UpdateMemberRole(params.ID, params.UserId, req.RoleCode); err != nil {
		h.logger.Error("update collaboration workspace member role failed", zap.Error(err))
		return nil, err
	}
	return ok(), nil
}

func (h *cwAPIHandler) GetCollaborationPackages(ctx context.Context, params gen.GetCollaborationPackagesParams) (*gen.FeaturePackageAssignmentResponse, error) {
	ids, pkgs, err := h.featurePkgSvc.GetCollaborationWorkspacePackages(params.WorkspaceId, optString(params.AppKey))
	if err != nil {
		h.logger.Error("get collaboration workspace packages failed", zap.Error(err))
		return nil, err
	}
	return &gen.FeaturePackageAssignmentResponse{
		PackageIds: ids,
		Packages:   featurePackageRefsFromModels(pkgs),
	}, nil
}

func (h *cwAPIHandler) SetCollaborationPackages(ctx context.Context, req *gen.UUIDListRequest, params gen.SetCollaborationPackagesParams) (*gen.FeaturePackageMutationResult, error) {
	var grantedBy *uuid.UUID
	if uid, ok := userIDFromContext(ctx); ok {
		grantedBy = &uid
	}
	stats, err := h.featurePkgSvc.SetCollaborationWorkspacePackages(params.WorkspaceId, uuidIDsFromRequest(req), grantedBy, optString(params.AppKey))
	if err != nil {
		h.logger.Error("set collaboration workspace packages failed", zap.Error(err))
		return nil, err
	}
	return featurePackageMutationResultFromStats(stats), nil
}

var _ = errors.New



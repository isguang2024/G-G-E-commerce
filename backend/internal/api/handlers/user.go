// user.go: ogen Handler implementations for the /users/* OpenAPI surface.
// Phase 4 slice 5 — user domain migration (step 1: read-only list + get).
// Legacy gin handlers under internal/modules/system/user/handler.go remain
// mounted for not-yet-migrated operations; they will be removed once every
// operation in this file is live and the frontend has switched to v5Client.
package handlers

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/maben/backend/api/gen"
	"github.com/maben/backend/internal/api/dto"
	"github.com/maben/backend/internal/modules/observability/audit"
	"github.com/maben/backend/internal/modules/system/models"
	"github.com/maben/backend/internal/modules/system/user"
)

const (
	userListDefaultCurrent = 1
	userListDefaultSize    = 20
	userListMaxSize        = 200
	userTimeLayout         = "2006-01-02 15:04:05"
)

func (h *APIHandler) ListUsers(ctx context.Context, params gen.ListUsersParams) (gen.ListUsersRes, error) {
	if _, ok := userIDFromContext(ctx); !ok {
		return &gen.ListUsersUnauthorized{Code: 401, Message: "未认证"}, nil
	}

	current := userListDefaultCurrent
	if params.Current.Set && params.Current.Value > 0 {
		current = params.Current.Value
	}
	size := userListDefaultSize
	if params.Size.Set && params.Size.Value > 0 {
		size = params.Size.Value
	}
	if size > userListMaxSize {
		size = userListMaxSize
	}
	offset := (current - 1) * size

	list, total, err := h.userRepo.List(
		offset,
		size,
		optString(params.UserName),
		optString(params.UserPhone),
		optString(params.UserEmail),
		optString(params.Status),
		optString(params.RoleID),
		optString(params.ID),
		optString(params.RegisterSource),
		optString(params.InvitedBy),
	)
	if err != nil {
		h.logger.Error("list users failed", zap.Error(err))
		return nil, err
	}

	inviterNames := h.resolveInviterNames(list)

	records := make([]gen.UserSummary, 0, len(list))
	for i := range list {
		records = append(records, userSummaryFromModel(&list[i], inviterNames))
	}

	return &gen.UserList{
		Records: records,
		Total:   total,
		Current: current,
		Size:    size,
	}, nil
}

func (h *APIHandler) GetUser(ctx context.Context, params gen.GetUserParams) (gen.GetUserRes, error) {
	if _, ok := userIDFromContext(ctx); !ok {
		return &gen.GetUserUnauthorized{Code: 401, Message: "未认证"}, nil
	}

	u, err := h.userRepo.GetByID(params.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) || errors.Is(err, user.ErrUserNotFound) {
			return &gen.GetUserNotFound{Code: 404, Message: "用户不存在"}, nil
		}
		h.logger.Error("get user failed", zap.Error(err))
		return nil, err
	}

	return userDetailFromModel(u), nil
}

// resolveInviterNames batch-loads display names for the invited_by references
// that appear in the page of users; missing inviters fall back to a sentinel
// label so the list surface keeps the legacy contract.
func (h *APIHandler) resolveInviterNames(list []models.User) map[uuid.UUID]string {
	ids := make([]uuid.UUID, 0, len(list))
	seen := make(map[uuid.UUID]struct{})
	for _, u := range list {
		if u.InvitedBy == nil {
			continue
		}
		if _, ok := seen[*u.InvitedBy]; ok {
			continue
		}
		seen[*u.InvitedBy] = struct{}{}
		ids = append(ids, *u.InvitedBy)
	}
	if len(ids) == 0 {
		return nil
	}
	inviters, err := h.userRepo.GetByIDs(ids)
	if err != nil {
		h.logger.Warn("load inviters failed", zap.Error(err))
		return nil
	}
	out := make(map[uuid.UUID]string, len(inviters))
	for i := range inviters {
		name := inviters[i].Nickname
		if name == "" {
			name = inviters[i].Username
		}
		out[inviters[i].ID] = name
	}
	return out
}

func (h *APIHandler) CreateUser(ctx context.Context, req *gen.UserCreateRequest) (gen.CreateUserRes, error) {
	if _, ok := userIDFromContext(ctx); !ok {
		return &gen.CreateUserUnauthorized{Code: 401, Message: "未认证"}, nil
	}
	if req == nil || req.Username == "" || req.Password == "" {
		return &gen.CreateUserBadRequest{Code: 400, Message: "用户名和密码必填"}, nil
	}
	dtoReq := userCreateRequestFromGen(req)
	created, err := h.userSvc.Create(dtoReq)
	if err != nil {
		if errors.Is(err, user.ErrUserExists) {
			h.audit.Record(ctx, audit.Event{
				Action:       "system.user.create",
				ResourceType: "user",
				Outcome:      audit.OutcomeError,
				ErrorCode:    "user_exists",
				Metadata:     map[string]any{"username": req.Username},
			})
			return &gen.CreateUserBadRequest{Code: 400, Message: "用户名已存在"}, nil
		}
		if errors.Is(err, user.ErrEmailExists) {
			h.audit.Record(ctx, audit.Event{
				Action:       "system.user.create",
				ResourceType: "user",
				Outcome:      audit.OutcomeError,
				ErrorCode:    "email_exists",
				Metadata:     map[string]any{"username": req.Username},
			})
			return &gen.CreateUserBadRequest{Code: 400, Message: "邮箱已存在"}, nil
		}
		h.logger.Error("create user failed", zap.Error(err))
		h.audit.Record(ctx, audit.Event{
			Action:       "system.user.create",
			ResourceType: "user",
			Outcome:      audit.OutcomeError,
			ErrorCode:    errorCodeOf(err),
			Metadata:     map[string]any{"username": req.Username},
		})
		return nil, err
	}
	h.audit.Record(ctx, audit.Event{
		Action:       "system.user.create",
		ResourceType: "user",
		ResourceID:   created.ID.String(),
		Outcome:      audit.OutcomeSuccess,
		Metadata:     map[string]any{"username": req.Username},
	})
	return &gen.UserCreateResult{ID: created.ID}, nil
}

func (h *APIHandler) UpdateUser(ctx context.Context, req *gen.UserUpdateRequest, params gen.UpdateUserParams) (gen.UpdateUserRes, error) {
	if _, ok := userIDFromContext(ctx); !ok {
		return &gen.UpdateUserUnauthorized{Code: 401, Message: "未认证"}, nil
	}
	if req == nil {
		return &gen.UpdateUserBadRequest{Code: 400, Message: "请求体不能为空"}, nil
	}
	dtoReq := userUpdateRequestFromGen(req)
	if err := h.userSvc.Update(params.ID, dtoReq); err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			h.audit.Record(ctx, audit.Event{
				Action:       "system.user.update",
				ResourceType: "user",
				ResourceID:   params.ID.String(),
				Outcome:      audit.OutcomeError,
				ErrorCode:    "not_found",
			})
			return &gen.UpdateUserNotFound{Code: 404, Message: "用户不存在"}, nil
		}
		if errors.Is(err, user.ErrEmailExists) {
			h.audit.Record(ctx, audit.Event{
				Action:       "system.user.update",
				ResourceType: "user",
				ResourceID:   params.ID.String(),
				Outcome:      audit.OutcomeError,
				ErrorCode:    "email_exists",
			})
			return &gen.UpdateUserBadRequest{Code: 400, Message: "邮箱已存在"}, nil
		}
		h.logger.Error("update user failed", zap.Error(err))
		h.audit.Record(ctx, audit.Event{
			Action:       "system.user.update",
			ResourceType: "user",
			ResourceID:   params.ID.String(),
			Outcome:      audit.OutcomeError,
			ErrorCode:    errorCodeOf(err),
		})
		return nil, err
	}
	h.audit.Record(ctx, audit.Event{
		Action:       "system.user.update",
		ResourceType: "user",
		ResourceID:   params.ID.String(),
		Outcome:      audit.OutcomeSuccess,
		After:        dtoReq,
	})
	return &gen.UserMutationResult{Success: true}, nil
}

func (h *APIHandler) DeleteUser(ctx context.Context, params gen.DeleteUserParams) (gen.DeleteUserRes, error) {
	if _, ok := userIDFromContext(ctx); !ok {
		return &gen.DeleteUserUnauthorized{Code: 401, Message: "未认证"}, nil
	}
	if err := h.userSvc.Delete(params.ID); err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			h.audit.Record(ctx, audit.Event{
				Action:       "system.user.delete",
				ResourceType: "user",
				ResourceID:   params.ID.String(),
				Outcome:      audit.OutcomeError,
				ErrorCode:    "not_found",
			})
			return &gen.DeleteUserNotFound{Code: 404, Message: "用户不存在"}, nil
		}
		h.logger.Error("delete user failed", zap.Error(err))
		h.audit.Record(ctx, audit.Event{
			Action:       "system.user.delete",
			ResourceType: "user",
			ResourceID:   params.ID.String(),
			Outcome:      audit.OutcomeError,
			ErrorCode:    errorCodeOf(err),
		})
		return nil, err
	}
	h.audit.Record(ctx, audit.Event{
		Action:       "system.user.delete",
		ResourceType: "user",
		ResourceID:   params.ID.String(),
		Outcome:      audit.OutcomeSuccess,
	})
	return &gen.UserMutationResult{Success: true}, nil
}

func (h *APIHandler) AssignUserRoles(ctx context.Context, req *gen.UserAssignRolesRequest, params gen.AssignUserRolesParams) (gen.AssignUserRolesRes, error) {
	if _, ok := userIDFromContext(ctx); !ok {
		return &gen.AssignUserRolesUnauthorized{Code: 401, Message: "未认证"}, nil
	}
	if req == nil {
		return &gen.AssignUserRolesBadRequest{Code: 400, Message: "请求体不能为空"}, nil
	}
	if err := h.userSvc.AssignRoles(params.ID, req.RoleIds); err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			h.audit.Record(ctx, audit.Event{
				Action:       "system.user.assign_roles",
				ResourceType: "user",
				ResourceID:   params.ID.String(),
				Outcome:      audit.OutcomeError,
				ErrorCode:    "not_found",
				Metadata:     map[string]any{"role_ids": req.RoleIds},
			})
			return &gen.AssignUserRolesNotFound{Code: 404, Message: "用户不存在"}, nil
		}
		h.logger.Error("assign user roles failed", zap.Error(err))
		h.audit.Record(ctx, audit.Event{
			Action:       "system.user.assign_roles",
			ResourceType: "user",
			ResourceID:   params.ID.String(),
			Outcome:      audit.OutcomeError,
			ErrorCode:    errorCodeOf(err),
			Metadata:     map[string]any{"role_ids": req.RoleIds},
		})
		return nil, err
	}
	h.audit.Record(ctx, audit.Event{
		Action:       "system.user.assign_roles",
		ResourceType: "user",
		ResourceID:   params.ID.String(),
		Outcome:      audit.OutcomeSuccess,
		Metadata:     map[string]any{"role_ids": req.RoleIds},
	})
	return &gen.UserMutationResult{Success: true}, nil
}

func userCreateRequestFromGen(req *gen.UserCreateRequest) *dto.UserCreateRequest {
	out := &dto.UserCreateRequest{
		Username: req.Username,
		Password: req.Password,
		Email:    optString(req.Email),
		Nickname: optString(req.Nickname),
		Phone:    optString(req.Phone),
		SystemRemark: optString(req.SystemRemark),
		Status:   optString(req.Status),
	}
	if len(req.RoleIds) > 0 {
		out.RoleIDs = append(out.RoleIDs, req.RoleIds...)
	}
	return out
}

func userUpdateRequestFromGen(req *gen.UserUpdateRequest) *dto.UserUpdateRequest {
	out := &dto.UserUpdateRequest{
		Email:        optString(req.Email),
		Nickname:     optString(req.Nickname),
		Phone:        optString(req.Phone),
		SystemRemark: optString(req.SystemRemark),
		Status:       optString(req.Status),
	}
	if req.RoleIds != nil {
		out.RoleIDs = append([]string{}, req.RoleIds...)
	}
	return out
}

func userSummaryFromModel(u *models.User, inviters map[uuid.UUID]string) gen.UserSummary {
	out := gen.UserSummary{
		ID:          u.ID,
		UserName:    u.Username,
		Status:      u.Status,
		CreateTime:  u.CreatedAt.Format(userTimeLayout),
		UpdateTime:  u.UpdatedAt.Format(userTimeLayout),
		UserRoles:   roleCodesFromModels(u.Roles),
		RoleDetails: roleRefsFromModels(u.Roles),
	}
	if u.Email != "" {
		out.UserEmail = gen.NewOptNilString(u.Email)
	}
	if u.Nickname != "" {
		out.NickName = gen.NewOptNilString(u.Nickname)
	}
	if u.Phone != "" {
		out.UserPhone = gen.NewOptNilString(u.Phone)
	}
	if u.SystemRemark != "" {
		out.SystemRemark = gen.NewOptNilString(u.SystemRemark)
	}
	if u.LastLoginAt != nil && !u.LastLoginAt.IsZero() {
		out.LastLoginTime = gen.NewOptNilString(u.LastLoginAt.Format(userTimeLayout))
	}
	if u.LastLoginIP != "" {
		out.LastLoginIP = gen.NewOptNilString(u.LastLoginIP)
	}
	if u.AvatarURL != "" {
		out.Avatar = gen.NewOptNilString(u.AvatarURL)
	}
	if u.RegisterSource != "" {
		out.RegisterSource = gen.NewOptNilString(u.RegisterSource)
	}
	if u.InvitedBy != nil {
		out.InvitedBy = gen.NewOptNilUUID(*u.InvitedBy)
		if name, ok := inviters[*u.InvitedBy]; ok && name != "" {
			out.InvitedByName = gen.NewOptNilString(name)
		} else {
			out.InvitedByName = gen.NewOptNilString("未知邀请人")
		}
	}
	return out
}

func userDetailFromModel(u *models.User) *gen.UserDetail {
	refs := roleRefsFromModels(u.Roles)
	out := &gen.UserDetail{
		ID:          u.ID,
		UserName:    u.Username,
		Status:      u.Status,
		CreateTime:  u.CreatedAt.Format(userTimeLayout),
		UpdateTime:  u.UpdatedAt.Format(userTimeLayout),
		Roles:       refs,
		UserRoles:   roleCodesFromModels(u.Roles),
		RoleDetails: refs,
	}
	if u.Email != "" {
		out.UserEmail = gen.NewOptNilString(u.Email)
	}
	if u.Nickname != "" {
		out.NickName = gen.NewOptNilString(u.Nickname)
	}
	if u.Phone != "" {
		out.UserPhone = gen.NewOptNilString(u.Phone)
	}
	if u.SystemRemark != "" {
		out.SystemRemark = gen.NewOptNilString(u.SystemRemark)
	}
	if u.LastLoginAt != nil && !u.LastLoginAt.IsZero() {
		out.LastLoginTime = gen.NewOptNilString(u.LastLoginAt.Format(userTimeLayout))
	}
	if u.LastLoginIP != "" {
		out.LastLoginIP = gen.NewOptNilString(u.LastLoginIP)
	}
	if u.AvatarURL != "" {
		out.Avatar = gen.NewOptNilString(u.AvatarURL)
	}
	return out
}

func roleCodesFromModels(roles []models.Role) []string {
	out := make([]string, 0, len(roles))
	for _, r := range roles {
		out = append(out, r.Code)
	}
	return out
}

func roleRefsFromModels(roles []models.Role) []gen.UserRoleRef {
	out := make([]gen.UserRoleRef, 0, len(roles))
	for _, r := range roles {
		out = append(out, gen.UserRoleRef{
			ID:   r.ID,
			Code: r.Code,
			Name: r.Name,
		})
	}
	return out
}



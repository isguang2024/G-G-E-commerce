// message.go — Phase 4: ogen handler implementations for the message domain.
// Covers 19 operations: inbox, templates, senders, recipient-groups, dispatch.
package handlers

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/gg-ecommerce/backend/api/gen"
	systemmod "github.com/gg-ecommerce/backend/internal/modules/system/system"
)

// ── helpers ─────────────────────────────────────────────────────────────────

// cwIDFromContext reads the collaboration workspace UUID from the request
// context (populated by the Gin auth middleware before the ogen bridge).
func cwIDFromContext(ctx context.Context) *uuid.UUID {
	raw := strings.TrimSpace(stringFromCtx(ctx, CtxCollaborationWorkspaceID))
	if raw == "" {
		return nil
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		return nil
	}
	return &id
}

// ── Inbox ────────────────────────────────────────────────────────────────────

func (h *APIHandler) GetInboxSummary(ctx context.Context) (gen.AnyObject, error) {
	userID, valid := userIDFromContext(ctx)
	if !valid {
		return gen.AnyObject{}, nil
	}
	summary, err := h.systemFacade.GetInboxSummary(userID)
	if err != nil {
		h.logger.Error("get inbox summary failed", zap.Error(err))
		return nil, err
	}
	return marshalAnyObject(summary), nil
}

func (h *APIHandler) ListInbox(ctx context.Context, params gen.ListInboxParams) (*gen.MessageListResponse, error) {
	userID, valid := userIDFromContext(ctx)
	if !valid {
		return &gen.MessageListResponse{}, nil
	}
	current := 1
	size := 20
	if params.Current.Set {
		current = params.Current.Value
	}
	if params.Size.Set {
		size = params.Size.Value
	}
	result, err := h.systemFacade.ListInbox(userID, systemmod.MessageInboxQuery{
		Current: current,
		Size:    size,
	})
	if err != nil {
		h.logger.Error("list inbox failed", zap.Error(err))
		return nil, err
	}
	return &gen.MessageListResponse{
		Records: marshalList(result.Records),
		Total:   int(result.Total),
	}, nil
}

func (h *APIHandler) GetInboxDetail(ctx context.Context, params gen.GetInboxDetailParams) (gen.AnyObject, error) {
	userID, valid := userIDFromContext(ctx)
	if !valid {
		return gen.AnyObject{}, nil
	}
	detail, err := h.systemFacade.GetInboxDetail(userID, params.DeliveryId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return gen.AnyObject{}, nil
		}
		h.logger.Error("get inbox detail failed", zap.Error(err))
		return nil, err
	}
	return marshalAnyObject(detail), nil
}

func (h *APIHandler) MarkInboxRead(ctx context.Context, params gen.MarkInboxReadParams) (*gen.MutationResult, error) {
	userID, valid := userIDFromContext(ctx)
	if !valid {
		return &gen.MutationResult{Success: false}, nil
	}
	if err := h.systemFacade.MarkRead(userID, params.DeliveryId); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &gen.MutationResult{Success: false}, nil
		}
		h.logger.Error("mark inbox read failed", zap.Error(err))
		return nil, err
	}
	return ok(), nil
}

func (h *APIHandler) MarkInboxReadAll(ctx context.Context) (*gen.MutationResult, error) {
	userID, valid := userIDFromContext(ctx)
	if !valid {
		return &gen.MutationResult{Success: false}, nil
	}
	if err := h.systemFacade.MarkAllRead(userID, ""); err != nil {
		h.logger.Error("mark inbox read all failed", zap.Error(err))
		return nil, err
	}
	return ok(), nil
}

func (h *APIHandler) HandleInboxTodo(ctx context.Context, req gen.AnyObject, params gen.HandleInboxTodoParams) (*gen.MutationResult, error) {
	userID, valid := userIDFromContext(ctx)
	if !valid {
		return &gen.MutationResult{Success: false}, nil
	}
	var body struct {
		Action string `json:"action"`
	}
	if err := unmarshalAnyObject(req, &body); err != nil {
		return &gen.MutationResult{Success: false}, nil
	}
	if err := h.systemFacade.UpdateTodoStatus(userID, params.DeliveryId, body.Action); err != nil {
		if strings.Contains(err.Error(), "invalid todo action") ||
			strings.Contains(err.Error(), "无效") {
			return &gen.MutationResult{Success: false}, nil
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &gen.MutationResult{Success: false}, nil
		}
		h.logger.Error("handle inbox todo failed", zap.Error(err))
		return nil, err
	}
	return ok(), nil
}

// ── Dispatch ─────────────────────────────────────────────────────────────────

func (h *APIHandler) GetMessageDispatchOptions(ctx context.Context) (gen.AnyObject, error) {
	userID, valid := userIDFromContext(ctx)
	if !valid {
		return gen.AnyObject{}, nil
	}
	cwID := cwIDFromContext(ctx)
	options, err := h.systemFacade.GetDispatchOptions(userID, cwID)
	if err != nil {
		h.logger.Error("get message dispatch options failed", zap.Error(err))
		return nil, err
	}
	return marshalAnyObject(options), nil
}

func (h *APIHandler) DispatchMessage(ctx context.Context, req gen.AnyObject) (*gen.MutationResult, error) {
	userID, valid := userIDFromContext(ctx)
	if !valid {
		return &gen.MutationResult{Success: false}, nil
	}
	cwID := cwIDFromContext(ctx)

	var body struct {
		SenderID                        string   `json:"sender_id"`
		TemplateID                      string   `json:"template_id"`
		TemplateKey                     string   `json:"template_key"`
		MessageType                     string   `json:"message_type"`
		AudienceType                    string   `json:"audience_type"`
		TargetCollaborationWorkspaceIDs []string `json:"target_collaboration_workspace_ids"`
		TargetUserIDs                   []string `json:"target_user_ids"`
		TargetGroupIDs                  []string `json:"target_group_ids"`
		Title                           string   `json:"title"`
		Summary                         string   `json:"summary"`
		Content                         string   `json:"content"`
		Priority                        string   `json:"priority"`
		ActionType                      string   `json:"action_type"`
		ActionTarget                    string   `json:"action_target"`
		BizType                         string   `json:"biz_type"`
		ExpiredAt                       string   `json:"expired_at"`
	}
	if err := unmarshalAnyObject(req, &body); err != nil {
		return &gen.MutationResult{Success: false}, nil
	}

	_, err := h.systemFacade.DispatchMessage(userID, cwID, systemmod.MessageDispatchRequest{
		SenderID:                        body.SenderID,
		TemplateID:                      body.TemplateID,
		TemplateKey:                     body.TemplateKey,
		MessageType:                     body.MessageType,
		AudienceType:                    body.AudienceType,
		TargetCollaborationWorkspaceIDs: body.TargetCollaborationWorkspaceIDs,
		TargetUserIDs:                   body.TargetUserIDs,
		TargetGroupIDs:                  body.TargetGroupIDs,
		Title:                           body.Title,
		Summary:                         body.Summary,
		Content:                         body.Content,
		Priority:                        body.Priority,
		ActionType:                      body.ActionType,
		ActionTarget:                    body.ActionTarget,
		BizType:                         body.BizType,
		ExpiredAt:                       body.ExpiredAt,
	})
	if err != nil {
		h.logger.Error("dispatch message failed", zap.Error(err))
		return nil, err
	}
	return ok(), nil
}

// ── Templates ────────────────────────────────────────────────────────────────

func (h *APIHandler) ListMessageTemplates(ctx context.Context) (*gen.MessageListResponse, error) {
	cwID := cwIDFromContext(ctx)
	result, err := h.systemFacade.ListTemplates(cwID, systemmod.MessageTemplateQuery{
		Current: 1,
		Size:    200,
	})
	if err != nil {
		h.logger.Error("list message templates failed", zap.Error(err))
		return nil, err
	}
	return &gen.MessageListResponse{
		Records: marshalList(result.Records),
		Total:   int(result.Total),
	}, nil
}

func (h *APIHandler) CreateMessageTemplate(ctx context.Context, req gen.AnyObject) (*gen.MutationResult, error) {
	cwID := cwIDFromContext(ctx)
	var body systemmod.MessageTemplateUpsertRequest
	if err := unmarshalAnyObject(req, &body); err != nil {
		return &gen.MutationResult{Success: false}, nil
	}
	if _, err := h.systemFacade.SaveTemplate("", cwID, body); err != nil {
		h.logger.Error("create message template failed", zap.Error(err))
		return nil, err
	}
	return ok(), nil
}

func (h *APIHandler) UpdateMessageTemplate(ctx context.Context, req gen.AnyObject, params gen.UpdateMessageTemplateParams) (*gen.MutationResult, error) {
	cwID := cwIDFromContext(ctx)
	var body systemmod.MessageTemplateUpsertRequest
	if err := unmarshalAnyObject(req, &body); err != nil {
		return &gen.MutationResult{Success: false}, nil
	}
	if _, err := h.systemFacade.SaveTemplate(params.TemplateId.String(), cwID, body); err != nil {
		h.logger.Error("update message template failed", zap.Error(err))
		return nil, err
	}
	return ok(), nil
}

// ── Senders ──────────────────────────────────────────────────────────────────

func (h *APIHandler) ListMessageSenders(ctx context.Context) (*gen.MessageListResponse, error) {
	cwID := cwIDFromContext(ctx)
	items, err := h.systemFacade.ListSenders(cwID)
	if err != nil {
		h.logger.Error("list message senders failed", zap.Error(err))
		return nil, err
	}
	return &gen.MessageListResponse{
		Records: marshalList(items),
		Total:   len(items),
	}, nil
}

func (h *APIHandler) CreateMessageSender(ctx context.Context, req gen.AnyObject) (*gen.MutationResult, error) {
	cwID := cwIDFromContext(ctx)
	var body systemmod.MessageSenderSaveRequest
	if err := unmarshalAnyObject(req, &body); err != nil {
		return &gen.MutationResult{Success: false}, nil
	}
	if _, err := h.systemFacade.SaveSender("", cwID, body); err != nil {
		h.logger.Error("create message sender failed", zap.Error(err))
		return nil, err
	}
	return ok(), nil
}

func (h *APIHandler) UpdateMessageSender(ctx context.Context, req gen.AnyObject, params gen.UpdateMessageSenderParams) (*gen.MutationResult, error) {
	cwID := cwIDFromContext(ctx)
	var body systemmod.MessageSenderSaveRequest
	if err := unmarshalAnyObject(req, &body); err != nil {
		return &gen.MutationResult{Success: false}, nil
	}
	if _, err := h.systemFacade.SaveSender(params.SenderId.String(), cwID, body); err != nil {
		h.logger.Error("update message sender failed", zap.Error(err))
		return nil, err
	}
	return ok(), nil
}

// ── Recipient Groups ─────────────────────────────────────────────────────────

func (h *APIHandler) ListMessageRecipientGroups(ctx context.Context) (*gen.MessageListResponse, error) {
	cwID := cwIDFromContext(ctx)
	items, err := h.systemFacade.ListRecipientGroups(cwID)
	if err != nil {
		h.logger.Error("list message recipient groups failed", zap.Error(err))
		return nil, err
	}
	return &gen.MessageListResponse{
		Records: marshalList(items),
		Total:   len(items),
	}, nil
}

func (h *APIHandler) CreateMessageRecipientGroup(ctx context.Context, req gen.AnyObject) (*gen.MutationResult, error) {
	cwID := cwIDFromContext(ctx)
	group, err := parseRecipientGroupRequest(req)
	if err != nil {
		return &gen.MutationResult{Success: false}, nil
	}
	if _, err := h.systemFacade.SaveRecipientGroup("", cwID, group); err != nil {
		h.logger.Error("create message recipient group failed", zap.Error(err))
		return nil, err
	}
	return ok(), nil
}

func (h *APIHandler) UpdateMessageRecipientGroup(ctx context.Context, req gen.AnyObject, params gen.UpdateMessageRecipientGroupParams) (*gen.MutationResult, error) {
	cwID := cwIDFromContext(ctx)
	group, err := parseRecipientGroupRequest(req)
	if err != nil {
		return &gen.MutationResult{Success: false}, nil
	}
	if _, err := h.systemFacade.SaveRecipientGroup(params.GroupId.String(), cwID, group); err != nil {
		h.logger.Error("update message recipient group failed", zap.Error(err))
		return nil, err
	}
	return ok(), nil
}

// parseRecipientGroupRequest decodes a recipient-group save request from AnyObject.
func parseRecipientGroupRequest(req gen.AnyObject) (systemmod.MessageRecipientGroupSaveRequest, error) {
	var body struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		MatchMode   string `json:"match_mode"`
		Status      string `json:"status"`
		Targets     []struct {
			TargetType               string `json:"target_type"`
			UserID                   string `json:"user_id"`
			CollaborationWorkspaceID string `json:"collaboration_workspace_id"`
			RoleCode                 string `json:"role_code"`
			PackageKey               string `json:"package_key"`
			SortOrder                int    `json:"sort_order"`
		} `json:"targets"`
	}
	if err := unmarshalAnyObject(req, &body); err != nil {
		return systemmod.MessageRecipientGroupSaveRequest{}, err
	}
	targets := make([]systemmod.MessageRecipientGroupTargetSaveRequest, 0, len(body.Targets))
	for _, t := range body.Targets {
		targets = append(targets, systemmod.MessageRecipientGroupTargetSaveRequest{
			TargetType:               t.TargetType,
			UserID:                   t.UserID,
			CollaborationWorkspaceID: t.CollaborationWorkspaceID,
			RoleCode:                 t.RoleCode,
			PackageKey:               t.PackageKey,
			SortOrder:                t.SortOrder,
		})
	}
	return systemmod.MessageRecipientGroupSaveRequest{
		Name:        body.Name,
		Description: body.Description,
		MatchMode:   body.MatchMode,
		Status:      body.Status,
		Targets:     targets,
	}, nil
}

// ── Dispatch Records ─────────────────────────────────────────────────────────

func (h *APIHandler) ListMessageDispatchRecords(ctx context.Context, params gen.ListMessageDispatchRecordsParams) (*gen.MessageListResponse, error) {
	cwID := cwIDFromContext(ctx)
	current := 1
	size := 20
	if params.Current.Set {
		current = params.Current.Value
	}
	if params.Size.Set {
		size = params.Size.Value
	}
	result, err := h.systemFacade.ListDispatchRecords(cwID, systemmod.DispatchRecordQuery{
		Current: current,
		Size:    size,
	})
	if err != nil {
		h.logger.Error("list message dispatch records failed", zap.Error(err))
		return nil, err
	}
	return &gen.MessageListResponse{
		Records: marshalList(result.Records),
		Total:   int(result.Total),
	}, nil
}

func (h *APIHandler) GetMessageDispatchRecord(ctx context.Context, params gen.GetMessageDispatchRecordParams) (gen.AnyObject, error) {
	cwID := cwIDFromContext(ctx)
	detail, err := h.systemFacade.GetDispatchRecordDetail(cwID, params.RecordId.String())
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return gen.AnyObject{}, nil
		}
		h.logger.Error("get message dispatch record failed", zap.Error(err))
		return nil, err
	}
	return marshalAnyObject(detail), nil
}

// message.go — Phase 4: ogen handler implementations for the message domain.
// Covers 19 operations: inbox, templates, senders, recipient-groups, dispatch.
package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/maben/backend/api/gen"
	"github.com/maben/backend/internal/modules/observability/audit"
	"github.com/maben/backend/internal/modules/system/models"
	systemmod "github.com/maben/backend/internal/modules/system/system"
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

func mapJSON[T any](input any) (T, error) {
	var out T
	raw, err := json.Marshal(input)
	if err != nil {
		return out, err
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return out, err
	}
	return out, nil
}

// ── Inbox ────────────────────────────────────────────────────────────────────

func (h *messageAPIHandler) GetInboxSummary(ctx context.Context) (*gen.InboxSummary, error) {
	userID, valid := userIDFromContext(ctx)
	if !valid {
		return &gen.InboxSummary{}, nil
	}
	summary, err := h.systemFacade.GetInboxSummary(userID)
	if err != nil {
		h.logger.Error("get inbox summary failed", zap.Error(err))
		return nil, err
	}
	return &gen.InboxSummary{
		UnreadTotal:  summary.UnreadTotal,
		NoticeCount:  summary.NoticeCount,
		MessageCount: summary.MessageCount,
		TodoCount:    summary.TodoCount,
	}, nil
}

func (h *messageAPIHandler) ListInbox(ctx context.Context, params gen.ListInboxParams) (*gen.InboxListResponse, error) {
	userID, valid := userIDFromContext(ctx)
	if !valid {
		return &gen.InboxListResponse{}, nil
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
	records, err := mapJSON[[]gen.InboxItem](result.Records)
	if err != nil {
		return nil, err
	}
	return &gen.InboxListResponse{
		Records: records,
		Total:   int(result.Total),
		Current: gen.NewOptInt(result.Current),
		Size:    gen.NewOptInt(result.Size),
	}, nil
}

func (h *messageAPIHandler) GetInboxDetail(ctx context.Context, params gen.GetInboxDetailParams) (*gen.InboxItem, error) {
	userID, valid := userIDFromContext(ctx)
	if !valid {
		return &gen.InboxItem{}, nil
	}
	detail, err := h.systemFacade.GetInboxDetail(userID, params.DeliveryId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &gen.InboxItem{}, nil
		}
		h.logger.Error("get inbox detail failed", zap.Error(err))
		return nil, err
	}
	return mapJSON[*gen.InboxItem](detail)
}

func (h *messageAPIHandler) MarkInboxRead(ctx context.Context, params gen.MarkInboxReadParams) (*gen.MutationResult, error) {
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

func (h *messageAPIHandler) MarkInboxReadAll(ctx context.Context) (*gen.MutationResult, error) {
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

func (h *messageAPIHandler) HandleInboxTodo(ctx context.Context, req *gen.InboxTodoActionRequest, params gen.HandleInboxTodoParams) (*gen.MutationResult, error) {
	userID, valid := userIDFromContext(ctx)
	if !valid {
		return &gen.MutationResult{Success: false}, nil
	}
	if req == nil {
		return &gen.MutationResult{Success: false}, nil
	}
	if err := h.systemFacade.UpdateTodoStatus(userID, params.DeliveryId, req.Action); err != nil {
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

func (h *messageAPIHandler) GetMessageDispatchOptions(ctx context.Context) (*gen.MessageDispatchOptions, error) {
	userID, valid := userIDFromContext(ctx)
	if !valid {
		return &gen.MessageDispatchOptions{}, nil
	}
	cwID := cwIDFromContext(ctx)
	options, err := h.systemFacade.GetDispatchOptions(userID, cwID)
	if err != nil {
		h.logger.Error("get message dispatch options failed", zap.Error(err))
		return nil, err
	}
	return mapJSON[*gen.MessageDispatchOptions](options)
}

func (h *messageAPIHandler) DispatchMessage(ctx context.Context, req *gen.MessageDispatchRequest) (*gen.MessageDispatchResult, error) {
	userID, valid := userIDFromContext(ctx)
	if !valid {
		return &gen.MessageDispatchResult{}, nil
	}
	cwID := cwIDFromContext(ctx)

	dispatchReq := systemmod.MessageDispatchRequest{
		SenderID:                        optString(req.SenderID),
		TemplateID:                      optString(req.TemplateID),
		TemplateKey:                     optString(req.TemplateKey),
		MessageType:                     optString(req.MessageType),
		AudienceType:                    optString(req.AudienceType),
		TargetCollaborationWorkspaceIDs: req.TargetCollaborationWorkspaceIds,
		TargetUserIDs:                   req.TargetUserIds,
		TargetGroupIDs:                  req.TargetGroupIds,
		Title:                           optString(req.Title),
		Summary:                         optString(req.Summary),
		Content:                         optString(req.Content),
		Priority:                        optString(req.Priority),
		ActionType:                      optString(req.ActionType),
		ActionTarget:                    optString(req.ActionTarget),
		BizType:                         optString(req.BizType),
		ExpiredAt:                       optString(req.ExpiredAt),
		DryRun:                          optBool(req.DryRun),
	}
	// 审计元数据只记可聚合字段，Title/Summary/Content 原文落到 After，由
	// redactor 负责剔除敏感 key。DryRun=true 也记一笔，排查误发演练需要。
	metadata := map[string]any{
		"message_type":  dispatchReq.MessageType,
		"audience_type": dispatchReq.AudienceType,
		"dry_run":       dispatchReq.DryRun,
		"template_id":   dispatchReq.TemplateID,
		"template_key":  dispatchReq.TemplateKey,
	}
	result, err := h.systemFacade.DispatchMessage(userID, cwID, dispatchReq)
	if err != nil {
		h.logger.Error("dispatch message failed", zap.Error(err))
		h.audit.Record(ctx, audit.Event{
			Action:       "system.message.dispatch",
			ResourceType: "message_dispatch",
			Outcome:      audit.OutcomeError,
			ErrorCode:    errorCodeOf(err),
			Metadata:     metadata,
			After:        dispatchReq,
		})
		return nil, err
	}
	h.audit.Record(ctx, audit.Event{
		Action:       "system.message.dispatch",
		ResourceType: "message_dispatch",
		Outcome:      audit.OutcomeSuccess,
		Metadata:     metadata,
		After:        result,
	})
	return mapJSON[*gen.MessageDispatchResult](result)
}

// ── Templates ────────────────────────────────────────────────────────────────

func (h *messageAPIHandler) ListMessageTemplates(ctx context.Context) (*gen.MessageTemplateListResponse, error) {
	cwID := cwIDFromContext(ctx)
	result, err := h.systemFacade.ListTemplates(cwID, systemmod.MessageTemplateQuery{
		Current: 1,
		Size:    200,
	})
	if err != nil {
		h.logger.Error("list message templates failed", zap.Error(err))
		return nil, err
	}
	records, err := mapJSON[[]gen.MessageTemplateItem](result.Records)
	if err != nil {
		return nil, err
	}
	return &gen.MessageTemplateListResponse{
		Records: records,
		Total:   int(result.Total),
		Current: gen.NewOptInt(result.Current),
		Size:    gen.NewOptInt(result.Size),
	}, nil
}

func (h *messageAPIHandler) CreateMessageTemplate(ctx context.Context, req *gen.MessageTemplateSaveRequest) (*gen.MessageTemplateItem, error) {
	cwID := cwIDFromContext(ctx)
	body := messageTemplateUpsertRequestFromGen(req)
	item, err := h.systemFacade.SaveTemplate("", cwID, body)
	if err != nil {
		h.logger.Error("create message template failed", zap.Error(err))
		return nil, err
	}
	return mapJSON[*gen.MessageTemplateItem](item)
}

func (h *messageAPIHandler) UpdateMessageTemplate(ctx context.Context, req *gen.MessageTemplateSaveRequest, params gen.UpdateMessageTemplateParams) (*gen.MessageTemplateItem, error) {
	cwID := cwIDFromContext(ctx)
	body := messageTemplateUpsertRequestFromGen(req)
	item, err := h.systemFacade.SaveTemplate(params.TemplateId.String(), cwID, body)
	if err != nil {
		h.logger.Error("update message template failed", zap.Error(err))
		return nil, err
	}
	return mapJSON[*gen.MessageTemplateItem](item)
}

// ── Senders ──────────────────────────────────────────────────────────────────

func (h *messageAPIHandler) ListMessageSenders(ctx context.Context) (*gen.MessageSenderListResponse, error) {
	cwID := cwIDFromContext(ctx)
	items, err := h.systemFacade.ListSenders(cwID)
	if err != nil {
		h.logger.Error("list message senders failed", zap.Error(err))
		return nil, err
	}
	records, err := mapJSON[[]gen.MessageSenderItem](items)
	if err != nil {
		return nil, err
	}
	return &gen.MessageSenderListResponse{Records: records}, nil
}

func (h *messageAPIHandler) CreateMessageSender(ctx context.Context, req *gen.MessageSenderSaveRequest) (*gen.MessageSenderItem, error) {
	cwID := cwIDFromContext(ctx)
	body := messageSenderSaveRequestFromGen(req)
	item, err := h.systemFacade.SaveSender("", cwID, body)
	if err != nil {
		h.logger.Error("create message sender failed", zap.Error(err))
		return nil, err
	}
	return mapJSON[*gen.MessageSenderItem](item)
}

func (h *messageAPIHandler) UpdateMessageSender(ctx context.Context, req *gen.MessageSenderSaveRequest, params gen.UpdateMessageSenderParams) (*gen.MessageSenderItem, error) {
	cwID := cwIDFromContext(ctx)
	body := messageSenderSaveRequestFromGen(req)
	item, err := h.systemFacade.SaveSender(params.SenderId.String(), cwID, body)
	if err != nil {
		h.logger.Error("update message sender failed", zap.Error(err))
		return nil, err
	}
	return mapJSON[*gen.MessageSenderItem](item)
}

// ── Recipient Groups ─────────────────────────────────────────────────────────

func (h *messageAPIHandler) ListMessageRecipientGroups(ctx context.Context) (*gen.MessageRecipientGroupListResponse, error) {
	cwID := cwIDFromContext(ctx)
	items, err := h.systemFacade.ListRecipientGroups(cwID)
	if err != nil {
		h.logger.Error("list message recipient groups failed", zap.Error(err))
		return nil, err
	}
	records, err := mapJSON[[]gen.MessageRecipientGroupItem](items)
	if err != nil {
		return nil, err
	}
	return &gen.MessageRecipientGroupListResponse{Records: records}, nil
}

func (h *messageAPIHandler) CreateMessageRecipientGroup(ctx context.Context, req *gen.MessageRecipientGroupSaveRequest) (*gen.MessageRecipientGroupItem, error) {
	cwID := cwIDFromContext(ctx)
	group, err := parseRecipientGroupRequest(req)
	if err != nil {
		return &gen.MessageRecipientGroupItem{}, nil
	}
	item, err := h.systemFacade.SaveRecipientGroup("", cwID, group)
	if err != nil {
		h.logger.Error("create message recipient group failed", zap.Error(err))
		return nil, err
	}
	return mapJSON[*gen.MessageRecipientGroupItem](item)
}

func (h *messageAPIHandler) UpdateMessageRecipientGroup(ctx context.Context, req *gen.MessageRecipientGroupSaveRequest, params gen.UpdateMessageRecipientGroupParams) (*gen.MessageRecipientGroupItem, error) {
	cwID := cwIDFromContext(ctx)
	group, err := parseRecipientGroupRequest(req)
	if err != nil {
		return &gen.MessageRecipientGroupItem{}, nil
	}
	item, err := h.systemFacade.SaveRecipientGroup(params.GroupId.String(), cwID, group)
	if err != nil {
		h.logger.Error("update message recipient group failed", zap.Error(err))
		return nil, err
	}
	return mapJSON[*gen.MessageRecipientGroupItem](item)
}

// parseRecipientGroupRequest decodes a recipient-group save request from AnyObject.
func parseRecipientGroupRequest(req *gen.MessageRecipientGroupSaveRequest) (systemmod.MessageRecipientGroupSaveRequest, error) {
	if req == nil {
		return systemmod.MessageRecipientGroupSaveRequest{}, errors.New("request body required")
	}
	body := systemmod.MessageRecipientGroupSaveRequest{
		Name:        req.Name,
		Description: optString(req.Description),
		MatchMode:   optString(req.MatchMode),
		Status:      optString(req.Status),
	}
	if req.Meta.Set {
		body.Meta = messageRecipientGroupMetaToJSON(req.Meta.Value)
	}
	targets := make([]systemmod.MessageRecipientGroupTargetSaveRequest, 0, len(req.Targets))
	for _, t := range req.Targets {
		targets = append(targets, systemmod.MessageRecipientGroupTargetSaveRequest{
			TargetType:               t.TargetType,
			UserID:                   optUUIDToString(t.UserID),
			CollaborationWorkspaceID: optUUIDToString(t.CollaborationWorkspaceID),
			RoleCode:                 optString(t.RoleCode),
			PackageKey:               optString(t.PackageKey),
			SortOrder:                optInt(t.SortOrder, 0),
			Meta: func() models.MetaJSON {
				if t.Meta.Set {
					return messageRecipientGroupTargetMetaToJSON(t.Meta.Value)
				}
				return nil
			}(),
		})
	}
	body.Targets = targets
	return body, nil
}

func messageTemplateUpsertRequestFromGen(req *gen.MessageTemplateSaveRequest) systemmod.MessageTemplateUpsertRequest {
	if req == nil {
		return systemmod.MessageTemplateUpsertRequest{}
	}
	return systemmod.MessageTemplateUpsertRequest{
		TemplateKey:     optString(req.TemplateKey),
		Name:            req.Name,
		Description:     optString(req.Description),
		MessageType:     req.MessageType,
		AudienceType:    req.AudienceType,
		TitleTemplate:   optString(req.TitleTemplate),
		SummaryTemplate: optString(req.SummaryTemplate),
		ContentTemplate: optString(req.ContentTemplate),
		Status:          optString(req.Status),
	}
}

func messageSenderSaveRequestFromGen(req *gen.MessageSenderSaveRequest) systemmod.MessageSenderSaveRequest {
	if req == nil {
		return systemmod.MessageSenderSaveRequest{}
	}
	body := systemmod.MessageSenderSaveRequest{
		Name:        req.Name,
		Description: optString(req.Description),
		AvatarURL:   optString(req.AvatarURL),
		IsDefault:   optBool(req.IsDefault),
		Status:      optString(req.Status),
	}
	if req.Meta.Set {
		body.Meta = messageSenderMetaToJSON(req.Meta.Value)
	}
	return body
}

func messageSenderMetaToJSON(meta gen.MessageSenderMeta) models.MetaJSON { return nil }

func messageRecipientGroupMetaToJSON(meta gen.MessageRecipientGroupMeta) models.MetaJSON { return nil }

func messageRecipientGroupTargetMetaToJSON(meta gen.MessageRecipientGroupTargetMeta) models.MetaJSON {
	return nil
}

func optUUIDToString(o gen.OptUUID) string {
	if !o.Set {
		return ""
	}
	return o.Value.String()
}

// ── Dispatch Records ─────────────────────────────────────────────────────────

func (h *messageAPIHandler) ListMessageDispatchRecords(ctx context.Context, params gen.ListMessageDispatchRecordsParams) (*gen.MessageDispatchRecordListResponse, error) {
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
	records, err := mapJSON[[]gen.DispatchRecordItem](result.Records)
	if err != nil {
		return nil, err
	}
	summary, err := mapJSON[gen.DispatchRecordSummary](result.Summary)
	if err != nil {
		return nil, err
	}
	return &gen.MessageDispatchRecordListResponse{
		Records: records,
		Total:   int(result.Total),
		Current: gen.NewOptInt(result.Current),
		Size:    gen.NewOptInt(result.Size),
		Summary: summary,
	}, nil
}

func (h *messageAPIHandler) GetMessageDispatchRecord(ctx context.Context, params gen.GetMessageDispatchRecordParams) (*gen.MessageDispatchRecord, error) {
	cwID := cwIDFromContext(ctx)
	detail, err := h.systemFacade.GetDispatchRecordDetail(cwID, params.RecordId.String())
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &gen.MessageDispatchRecord{}, nil
		}
		h.logger.Error("get message dispatch record failed", zap.Error(err))
		return nil, err
	}
	return mapJSON[*gen.MessageDispatchRecord](detail)
}


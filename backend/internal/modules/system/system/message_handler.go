package system

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/gg-ecommerce/backend/internal/api/dto"
	"github.com/gg-ecommerce/backend/internal/api/errcode"
	"github.com/gg-ecommerce/backend/internal/modules/system/models"
)

type messageTodoActionRequest struct {
	Action string `json:"action"`
}

type messageDispatchRequest struct {
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

type messageTemplateSaveRequest struct {
	TemplateKey     string `json:"template_key"`
	Name            string `json:"name"`
	Description     string `json:"description"`
	MessageType     string `json:"message_type"`
	AudienceType    string `json:"audience_type"`
	TitleTemplate   string `json:"title_template"`
	SummaryTemplate string `json:"summary_template"`
	ContentTemplate string `json:"content_template"`
	Status          string `json:"status"`
}

type messageSenderSavePayload struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	AvatarURL   string          `json:"avatar_url"`
	IsDefault   bool            `json:"is_default"`
	Status      string          `json:"status"`
	Meta        models.MetaJSON `json:"meta"`
}

type messageRecipientGroupTargetPayload struct {
	TargetType               string          `json:"target_type"`
	UserID                   string          `json:"user_id"`
	CollaborationWorkspaceID string          `json:"collaboration_workspace_id"`
	RoleCode                 string          `json:"role_code"`
	PackageKey               string          `json:"package_key"`
	SortOrder                int             `json:"sort_order"`
	Meta                     models.MetaJSON `json:"meta"`
}

type messageRecipientGroupSavePayload struct {
	Name        string                               `json:"name"`
	Description string                               `json:"description"`
	MatchMode   string                               `json:"match_mode"`
	Status      string                               `json:"status"`
	Meta        models.MetaJSON                      `json:"meta"`
	Targets     []messageRecipientGroupTargetPayload `json:"targets"`
}

func (h *SystemHandler) GetInboxSummary(c *gin.Context) {
	userID, err := currentAuthUserID(c)
	if err != nil {
		status, resp := errcode.Response(errcode.ErrUnauthorized)
		c.JSON(status, resp)
		return
	}
	summary, err := h.messageService.GetInboxSummary(userID)
	if err != nil {
		h.logger.Error("Get inbox summary failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取消息摘要失败")
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(summary))
}

func (h *SystemHandler) ListInbox(c *gin.Context) {
	userID, err := currentAuthUserID(c)
	if err != nil {
		status, resp := errcode.Response(errcode.ErrUnauthorized)
		c.JSON(status, resp)
		return
	}
	current, _ := strconv.Atoi(strings.TrimSpace(c.DefaultQuery("current", "1")))
	size, _ := strconv.Atoi(strings.TrimSpace(c.DefaultQuery("size", "20")))
	result, err := h.messageService.ListInbox(userID, inboxQuery{
		BoxType:    strings.TrimSpace(c.Query("box_type")),
		UnreadOnly: c.Query("unread_only") == "1" || strings.EqualFold(c.Query("unread_only"), "true"),
		Current:    current,
		Size:       size,
	})
	if err != nil {
		h.logger.Error("List inbox failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取消息列表失败")
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"records": result.Records,
		"current": result.Current,
		"size":    result.Size,
		"total":   result.Total,
	}))
}

func (h *SystemHandler) GetInboxDetail(c *gin.Context) {
	userID, err := currentAuthUserID(c)
	if err != nil {
		status, resp := errcode.Response(errcode.ErrUnauthorized)
		c.JSON(status, resp)
		return
	}
	deliveryID, err := uuid.Parse(c.Param("deliveryId"))
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的消息ID")
		c.JSON(status, resp)
		return
	}
	detail, err := h.messageService.GetInboxDetail(userID, deliveryID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			status, resp := errcode.ResponseWithMsg(errcode.ErrNotFound, "消息不存在")
			c.JSON(status, resp)
			return
		}
		h.logger.Error("Get inbox detail failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取消息详情失败")
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(detail))
}

func (h *SystemHandler) MarkInboxRead(c *gin.Context) {
	userID, err := currentAuthUserID(c)
	if err != nil {
		status, resp := errcode.Response(errcode.ErrUnauthorized)
		c.JSON(status, resp)
		return
	}
	deliveryID, err := uuid.Parse(c.Param("deliveryId"))
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的消息ID")
		c.JSON(status, resp)
		return
	}
	if err := h.messageService.MarkRead(userID, deliveryID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			status, resp := errcode.ResponseWithMsg(errcode.ErrNotFound, "消息不存在")
			c.JSON(status, resp)
			return
		}
		h.logger.Error("Mark inbox read failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "标记已读失败")
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(nil))
}

func (h *SystemHandler) MarkInboxReadAll(c *gin.Context) {
	userID, err := currentAuthUserID(c)
	if err != nil {
		status, resp := errcode.Response(errcode.ErrUnauthorized)
		c.JSON(status, resp)
		return
	}
	if err := h.messageService.MarkAllRead(userID, strings.TrimSpace(c.Query("box_type"))); err != nil {
		h.logger.Error("Mark inbox read all failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "批量已读失败")
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(nil))
}

func (h *SystemHandler) HandleInboxTodo(c *gin.Context) {
	userID, err := currentAuthUserID(c)
	if err != nil {
		status, resp := errcode.Response(errcode.ErrUnauthorized)
		c.JSON(status, resp)
		return
	}
	deliveryID, err := uuid.Parse(c.Param("deliveryId"))
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的消息ID")
		c.JSON(status, resp)
		return
	}
	var req messageTodoActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		status, resp := errcode.Response(errcode.ErrParamInvalid)
		c.JSON(status, resp)
		return
	}
	if err := h.messageService.UpdateTodoStatus(userID, deliveryID, req.Action); err != nil {
		if strings.Contains(err.Error(), "invalid todo action") {
			status, resp := errcode.Response(errcode.ErrParamInvalid)
			c.JSON(status, resp)
			return
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			status, resp := errcode.ResponseWithMsg(errcode.ErrNotFound, "待办不存在")
			c.JSON(status, resp)
			return
		}
		h.logger.Error("Handle inbox todo failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "处理待办失败")
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(nil))
}

func (h *SystemHandler) GetMessageDispatchOptions(c *gin.Context) {
	userID, err := currentAuthUserID(c)
	if err != nil {
		status, resp := errcode.Response(errcode.ErrUnauthorized)
		c.JSON(status, resp)
		return
	}
	tenantID, err := currentCollaborationWorkspaceID(c)
	if err != nil {
		status, resp := errcode.Response(errcode.ErrParamInvalid)
		c.JSON(status, resp)
		return
	}
	options, err := h.messageService.GetDispatchOptions(userID, tenantID)
	if err != nil {
		h.logger.Error("Get message dispatch options failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取发信配置失败")
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(options))
}

func (h *SystemHandler) DispatchMessage(c *gin.Context) {
	userID, err := currentAuthUserID(c)
	if err != nil {
		status, resp := errcode.Response(errcode.ErrUnauthorized)
		c.JSON(status, resp)
		return
	}
	tenantID, err := currentCollaborationWorkspaceID(c)
	if err != nil {
		status, resp := errcode.Response(errcode.ErrParamInvalid)
		c.JSON(status, resp)
		return
	}
	var req messageDispatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		status, resp := errcode.Response(errcode.ErrParamInvalid)
		c.JSON(status, resp)
		return
	}
	result, err := h.messageService.DispatchMessage(userID, tenantID, dispatchRequest{
		SenderID:                        req.SenderID,
		TemplateID:                      req.TemplateID,
		TemplateKey:                     req.TemplateKey,
		MessageType:                     req.MessageType,
		AudienceType:                    req.AudienceType,
		TargetCollaborationWorkspaceIDs: coalesceStringSlice(req.TargetCollaborationWorkspaceIDs, req.TargetCollaborationWorkspaceIDs),
		TargetUserIDs:                   req.TargetUserIDs,
		TargetGroupIDs:                  req.TargetGroupIDs,
		Title:                           req.Title,
		Summary:                         req.Summary,
		Content:                         req.Content,
		Priority:                        req.Priority,
		ActionType:                      req.ActionType,
		ActionTarget:                    req.ActionTarget,
		BizType:                         req.BizType,
		ExpiredAt:                       req.ExpiredAt,
	})
	if err != nil {
		if strings.Contains(err.Error(), "不能为空") ||
			strings.Contains(err.Error(), "不支持") ||
			strings.Contains(err.Error(), "无效") ||
			strings.Contains(err.Error(), "请选择") ||
			strings.Contains(err.Error(), "不存在") {
			status, resp := errcode.ResponseWithMsg(errcode.ErrParamInvalid, err.Error())
			c.JSON(status, resp)
			return
		}
		h.logger.Error("Dispatch message failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "发送消息失败")
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(result))
}

func (h *SystemHandler) ListMessageTemplates(c *gin.Context) {
	tenantID, err := currentCollaborationWorkspaceID(c)
	if err != nil {
		status, resp := errcode.Response(errcode.ErrParamInvalid)
		c.JSON(status, resp)
		return
	}
	current, _ := strconv.Atoi(strings.TrimSpace(c.DefaultQuery("current", "1")))
	size, _ := strconv.Atoi(strings.TrimSpace(c.DefaultQuery("size", "20")))
	result, err := h.messageService.ListTemplates(tenantID, messageTemplateQuery{
		Keyword: strings.TrimSpace(c.Query("keyword")),
		Current: current,
		Size:    size,
	})
	if err != nil {
		h.logger.Error("List message templates failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取消息模板失败")
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"records": result.Records,
		"current": result.Current,
		"size":    result.Size,
		"total":   result.Total,
	}))
}

func (h *SystemHandler) SaveMessageTemplate(c *gin.Context) {
	tenantID, err := currentCollaborationWorkspaceID(c)
	if err != nil {
		status, resp := errcode.Response(errcode.ErrParamInvalid)
		c.JSON(status, resp)
		return
	}

	var req messageTemplateSaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		status, resp := errcode.Response(errcode.ErrParamInvalid)
		c.JSON(status, resp)
		return
	}

	templateID := strings.TrimSpace(c.Param("templateId"))
	result, err := h.messageService.SaveTemplate(templateID, tenantID, messageTemplateUpsertRequest{
		TemplateKey:     req.TemplateKey,
		Name:            req.Name,
		Description:     req.Description,
		MessageType:     req.MessageType,
		AudienceType:    req.AudienceType,
		TitleTemplate:   req.TitleTemplate,
		SummaryTemplate: req.SummaryTemplate,
		ContentTemplate: req.ContentTemplate,
		Status:          req.Status,
	})
	if err != nil {
		if strings.Contains(err.Error(), "不能为空") ||
			strings.Contains(err.Error(), "无效") ||
			strings.Contains(err.Error(), "存在") ||
			strings.Contains(err.Error(), "不可编辑") {
			status, resp := errcode.ResponseWithMsg(errcode.ErrParamInvalid, err.Error())
			c.JSON(status, resp)
			return
		}
		h.logger.Error("Save message template failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "保存消息模板失败")
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(result))
}

func (h *SystemHandler) ListMessageSenders(c *gin.Context) {
	tenantID, err := currentCollaborationWorkspaceID(c)
	if err != nil {
		status, resp := errcode.Response(errcode.ErrParamInvalid)
		c.JSON(status, resp)
		return
	}
	result, err := h.messageService.ListSenders(tenantID)
	if err != nil {
		h.logger.Error("List message senders failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取发送人失败")
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"records": result,
	}))
}

func (h *SystemHandler) SaveMessageSender(c *gin.Context) {
	tenantID, err := currentCollaborationWorkspaceID(c)
	if err != nil {
		status, resp := errcode.Response(errcode.ErrParamInvalid)
		c.JSON(status, resp)
		return
	}
	var req messageSenderSavePayload
	if err := c.ShouldBindJSON(&req); err != nil {
		status, resp := errcode.Response(errcode.ErrParamInvalid)
		c.JSON(status, resp)
		return
	}
	senderID := strings.TrimSpace(c.Param("senderId"))
	result, err := h.messageService.SaveSender(senderID, tenantID, messageSenderSaveRequest{
		Name:        req.Name,
		Description: req.Description,
		AvatarURL:   req.AvatarURL,
		IsDefault:   req.IsDefault,
		Status:      req.Status,
		Meta:        req.Meta,
	})
	if err != nil {
		if strings.Contains(err.Error(), "不能为空") ||
			strings.Contains(err.Error(), "无效") ||
			strings.Contains(err.Error(), "不存在") ||
			strings.Contains(err.Error(), "不可编辑") {
			status, resp := errcode.ResponseWithMsg(errcode.ErrParamInvalid, err.Error())
			c.JSON(status, resp)
			return
		}
		h.logger.Error("Save message sender failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "保存发送人失败")
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(result))
}

func (h *SystemHandler) ListDispatchRecords(c *gin.Context) {
	tenantID, err := currentCollaborationWorkspaceID(c)
	if err != nil {
		status, resp := errcode.Response(errcode.ErrParamInvalid)
		c.JSON(status, resp)
		return
	}
	current, _ := strconv.Atoi(strings.TrimSpace(c.DefaultQuery("current", "1")))
	size, _ := strconv.Atoi(strings.TrimSpace(c.DefaultQuery("size", "20")))
	result, err := h.messageService.ListDispatchRecords(tenantID, dispatchRecordQuery{
		Keyword:      strings.TrimSpace(c.Query("keyword")),
		MessageType:  strings.TrimSpace(c.Query("message_type")),
		AudienceType: strings.TrimSpace(c.Query("audience_type")),
		Current:      current,
		Size:         size,
	})
	if err != nil {
		h.logger.Error("List dispatch records failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取发送记录失败")
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"records": result.Records,
		"current": result.Current,
		"size":    result.Size,
		"total":   result.Total,
		"summary": result.Summary,
	}))
}

func (h *SystemHandler) GetDispatchRecordDetail(c *gin.Context) {
	tenantID, err := currentCollaborationWorkspaceID(c)
	if err != nil {
		status, resp := errcode.Response(errcode.ErrParamInvalid)
		c.JSON(status, resp)
		return
	}
	result, err := h.messageService.GetDispatchRecordDetail(tenantID, strings.TrimSpace(c.Param("recordId")))
	if err != nil {
		if strings.Contains(err.Error(), "标识无效") {
			status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, err.Error())
			c.JSON(status, resp)
			return
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			status, resp := errcode.ResponseWithMsg(errcode.ErrNotFound, "发送记录不存在")
			c.JSON(status, resp)
			return
		}
		h.logger.Error("Get dispatch record detail failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取发送记录详情失败")
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(result))
}

func (h *SystemHandler) ListMessageRecipientGroups(c *gin.Context) {
	tenantID, err := currentCollaborationWorkspaceID(c)
	if err != nil {
		status, resp := errcode.Response(errcode.ErrParamInvalid)
		c.JSON(status, resp)
		return
	}
	result, err := h.messageService.ListRecipientGroups(tenantID)
	if err != nil {
		h.logger.Error("List message recipient groups failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取接收组失败")
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{"records": result}))
}

func (h *SystemHandler) SaveMessageRecipientGroup(c *gin.Context) {
	tenantID, err := currentCollaborationWorkspaceID(c)
	if err != nil {
		status, resp := errcode.Response(errcode.ErrParamInvalid)
		c.JSON(status, resp)
		return
	}
	var req messageRecipientGroupSavePayload
	if err := c.ShouldBindJSON(&req); err != nil {
		status, resp := errcode.Response(errcode.ErrParamInvalid)
		c.JSON(status, resp)
		return
	}
	targets := make([]messageRecipientGroupTargetSaveRequest, 0, len(req.Targets))
	for _, item := range req.Targets {
		resolvedLegacyCollaborationWorkspaceID, resolveErr := h.messageService.resolveLegacyCollaborationWorkspaceIDString(firstNonEmptyString(item.CollaborationWorkspaceID, item.CollaborationWorkspaceID))
		if resolveErr != nil {
			status, resp := errcode.ResponseWithMsg(errcode.ErrParamInvalid, resolveErr.Error())
			c.JSON(status, resp)
			return
		}
		targets = append(targets, messageRecipientGroupTargetSaveRequest{
			TargetType:               item.TargetType,
			UserID:                   item.UserID,
			CollaborationWorkspaceID: resolvedLegacyCollaborationWorkspaceID,
			RoleCode:                 item.RoleCode,
			PackageKey:               item.PackageKey,
			SortOrder:                item.SortOrder,
			Meta:                     item.Meta,
		})
	}
	groupID := strings.TrimSpace(c.Param("groupId"))
	result, err := h.messageService.SaveRecipientGroup(groupID, tenantID, messageRecipientGroupSaveRequest{
		Name:        req.Name,
		Description: req.Description,
		MatchMode:   req.MatchMode,
		Status:      req.Status,
		Meta:        req.Meta,
		Targets:     targets,
	})
	if err != nil {
		if strings.Contains(err.Error(), "不能为空") ||
			strings.Contains(err.Error(), "无效") ||
			strings.Contains(err.Error(), "不存在") ||
			strings.Contains(err.Error(), "不可编辑") ||
			strings.Contains(err.Error(), "请选择") {
			status, resp := errcode.ResponseWithMsg(errcode.ErrParamInvalid, err.Error())
			c.JSON(status, resp)
			return
		}
		h.logger.Error("Save message recipient group failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "保存接收组失败")
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(result))
}

func currentAuthUserID(c *gin.Context) (uuid.UUID, error) {
	raw, ok := c.Get("user_id")
	if !ok {
		return uuid.Nil, errors.New("unauthorized")
	}
	userIDStr, ok := raw.(string)
	if !ok {
		return uuid.Nil, errors.New("unauthorized")
	}
	return uuid.Parse(userIDStr)
}

func currentCollaborationWorkspaceID(c *gin.Context) (*uuid.UUID, error) {
	raw, ok := c.Get("collaboration_workspace_id")
	if ok {
		value, ok := raw.(string)
		if ok && strings.TrimSpace(value) != "" {
			parsedID, parseErr := uuid.Parse(strings.TrimSpace(value))
			if parseErr != nil {
				return nil, parseErr
			}
			return &parsedID, nil
		}
	}
	raw, ok = c.Get("collaboration_workspace_id")
	if ok {
		value, ok := raw.(string)
		if !ok {
			return nil, errors.New("invalid tenant context")
		}
		target := strings.TrimSpace(value)
		if target != "" {
			id, err := uuid.Parse(target)
			if err != nil {
				return nil, err
			}
			return &id, nil
		}
	}
	target := strings.TrimSpace(c.GetHeader("X-Collaboration-Workspace-Id"))
	if target == "" {
		return nil, nil
	}
	parsedID, parseErr := uuid.Parse(target)
	if parseErr != nil {
		return nil, parseErr
	}
	return &parsedID, nil
}

func coalesceStringSlice(primary, fallback []string) []string {
	if len(primary) > 0 {
		return primary
	}
	return fallback
}

func firstNonEmptyString(values ...string) string {
	for _, item := range values {
		if strings.TrimSpace(item) != "" {
			return strings.TrimSpace(item)
		}
	}
	return ""
}

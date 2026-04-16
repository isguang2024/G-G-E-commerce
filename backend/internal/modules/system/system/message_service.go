package system

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/maben/backend/internal/modules/system/models"
	"github.com/maben/backend/internal/pkg/workspacerolebinding"
)

type inboxSummary struct {
	UnreadTotal  int64 `json:"unread_total"`
	NoticeCount  int64 `json:"notice_count"`
	MessageCount int64 `json:"message_count"`
	TodoCount    int64 `json:"todo_count"`
}

type inboxQuery struct {
	BoxType    string
	UnreadOnly bool
	Current    int
	Size       int
}

type inboxListResult struct {
	Records []inboxListItem `json:"records"`
	Current int             `json:"current"`
	Size    int             `json:"size"`
	Total   int64           `json:"total"`
}

type inboxListItem struct {
	ID                                uuid.UUID       `json:"id"`
	MessageID                         uuid.UUID       `json:"message_id"`
	BoxType                           string          `json:"box_type"`
	DeliveryStatus                    string          `json:"delivery_status"`
	TodoStatus                        string          `json:"todo_status"`
	ReadAt                            *time.Time      `json:"read_at,omitempty"`
	DoneAt                            *time.Time      `json:"done_at,omitempty"`
	LastActionAt                      *time.Time      `json:"last_action_at,omitempty"`
	RecipientCollaborationWorkspaceID *uuid.UUID      `json:"recipient_collaboration_workspace_id,omitempty"`
	Title                             string          `json:"title"`
	Summary                           string          `json:"summary"`
	Content                           string          `json:"content"`
	Priority                          string          `json:"priority"`
	ActionType                        string          `json:"action_type"`
	ActionTarget                      string          `json:"action_target"`
	MessageType                       string          `json:"message_type"`
	BizType                           string          `json:"biz_type"`
	ScopeType                         string          `json:"scope_type"`
	ScopeID                           *uuid.UUID      `json:"scope_id,omitempty"`
	SenderType                        string          `json:"sender_type"`
	SenderNameSnapshot                string          `json:"sender_name_snapshot"`
	SenderAvatarSnapshot              string          `json:"sender_avatar_snapshot"`
	SenderServiceKey                  string          `json:"sender_service_key"`
	AudienceType                      string          `json:"audience_type"`
	AudienceScope                     string          `json:"audience_scope"`
	TargetCollaborationWorkspaceID    *uuid.UUID      `json:"target_collaboration_workspace_id,omitempty"`
	PublishedAt                       *time.Time      `json:"published_at,omitempty"`
	ExpiredAt                         *time.Time      `json:"expired_at,omitempty"`
	CreatedAt                         time.Time       `json:"created_at"`
	Meta                              models.MetaJSON `json:"meta,omitempty"`
}

type inboxDetail = inboxListItem

type dispatchAudienceOption struct {
	Value       string `json:"value"`
	Label       string `json:"label"`
	Description string `json:"description"`
}

type dispatchTemplateOption struct {
	ID              uuid.UUID `json:"id"`
	TemplateKey     string    `json:"template_key"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	MessageType     string    `json:"message_type"`
	OwnerScope      string    `json:"owner_scope"`
	AudienceType    string    `json:"audience_type"`
	TitleTemplate   string    `json:"title_template"`
	SummaryTemplate string    `json:"summary_template"`
	ContentTemplate string    `json:"content_template"`
}

type dispatchCollaborationWorkspaceOption struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

type dispatchSenderOption struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	AvatarURL   string    `json:"avatar_url"`
	IsDefault   bool      `json:"is_default"`
}

type dispatchUserOption struct {
	ID                         uuid.UUID  `json:"id"`
	Name                       string     `json:"name"`
	DisplayName                string     `json:"display_name"`
	Description                string     `json:"description"`
	CollaborationWorkspaceID   *uuid.UUID `json:"collaboration_workspace_id,omitempty"`
	CollaborationWorkspaceName string     `json:"collaboration_workspace_name,omitempty"`
}

type dispatchRecipientGroupOption struct {
	ID             uuid.UUID `json:"id"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	MatchMode      string    `json:"match_mode"`
	EstimatedCount int       `json:"estimated_count"`
}

type dispatchRoleOption struct {
	ID          uuid.UUID `json:"id"`
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
}

type dispatchFeaturePackageOption struct {
	ID          uuid.UUID `json:"id"`
	PackageKey  string    `json:"package_key"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
}

type dispatchOptions struct {
	SenderScope                       string                                 `json:"sender_scope"`
	CurrentCollaborationWorkspaceID   string                                 `json:"current_collaboration_workspace_id"`
	CurrentCollaborationWorkspaceName string                                 `json:"current_collaboration_workspace_name"`
	SenderOptions                     []dispatchSenderOption                 `json:"sender_options"`
	DefaultSenderID                   string                                 `json:"default_sender_id"`
	AudienceOptions                   []dispatchAudienceOption               `json:"audience_options"`
	TemplateOptions                   []dispatchTemplateOption               `json:"template_options"`
	CollaborationWorkspaces           []dispatchCollaborationWorkspaceOption `json:"collaboration_workspaces"`
	Users                             []dispatchUserOption                   `json:"users"`
	RecipientGroups                   []dispatchRecipientGroupOption         `json:"recipient_groups"`
	Roles                             []dispatchRoleOption                   `json:"roles"`
	FeaturePackages                   []dispatchFeaturePackageOption         `json:"feature_packages"`
	DefaultMessageType                string                                 `json:"default_message_type"`
	DefaultAudienceType               string                                 `json:"default_audience_type"`
	DefaultPriority                   string                                 `json:"default_priority"`
	SupportsExternalLink              bool                                   `json:"supports_external_link"`
}

type dispatchRequest struct {
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
	// DryRun 为 true 时走沙箱预览：完成校验但不写 messages 表、不入队、不触发投递。
	// 面向 E2E 深测 / QA 回归，返回 dispatch_status="preview"。
	DryRun bool `json:"dry_run"`
}

type dispatchResult struct {
	MessageID      uuid.UUID `json:"message_id"`
	DeliveryCount  int       `json:"delivery_count"`
	DispatchStatus string    `json:"dispatch_status"`
}

type dispatchRecipient struct {
	UserID                   uuid.UUID
	CollaborationWorkspaceID *uuid.UUID
	Username                 string
	SourceGroupID            *uuid.UUID
	SourceGroupName          string
	SourceRuleType           string
	SourceRuleLabel          string
	SourceTargetID           *uuid.UUID
	SourceTargetType         string
	SourceTargetValue        string
}

type dispatchRecipientUserRow struct {
	UserID   uuid.UUID `gorm:"column:user_id"`
	Username string    `gorm:"column:username"`
	Nickname string    `gorm:"column:nickname"`
}

type messageTemplateQuery struct {
	Keyword string
	Current int
	Size    int
}

type messageTemplateListResult struct {
	Records []messageTemplateListItem `json:"records"`
	Current int                       `json:"current"`
	Size    int                       `json:"size"`
	Total   int64                     `json:"total"`
}

type messageTemplateListItem struct {
	ID                              uuid.UUID       `json:"id"`
	TemplateKey                     string          `json:"template_key"`
	Name                            string          `json:"name"`
	Description                     string          `json:"description"`
	MessageType                     string          `json:"message_type"`
	OwnerScope                      string          `json:"owner_scope"`
	OwnerCollaborationWorkspaceID   *uuid.UUID      `json:"owner_collaboration_workspace_id"`
	OwnerCollaborationWorkspaceName string          `json:"owner_collaboration_workspace_name"`
	AudienceType                    string          `json:"audience_type"`
	TitleTemplate                   string          `json:"title_template"`
	SummaryTemplate                 string          `json:"summary_template"`
	ContentTemplate                 string          `json:"content_template"`
	Status                          string          `json:"status"`
	Editable                        bool            `json:"editable"`
	Meta                            models.MetaJSON `json:"meta"`
	CreatedAt                       time.Time       `json:"created_at"`
	UpdatedAt                       time.Time       `json:"updated_at"`
}

type messageTemplateUpsertRequest struct {
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

type messageSenderListItem struct {
	ID          uuid.UUID       `json:"id"`
	ScopeType   string          `json:"scope_type"`
	ScopeID     *uuid.UUID      `json:"scope_id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	AvatarURL   string          `json:"avatar_url"`
	IsDefault   bool            `json:"is_default"`
	Status      string          `json:"status"`
	Editable    bool            `json:"editable"`
	Meta        models.MetaJSON `json:"meta"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

type messageSenderSaveRequest struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	AvatarURL   string          `json:"avatar_url"`
	IsDefault   bool            `json:"is_default"`
	Status      string          `json:"status"`
	Meta        models.MetaJSON `json:"meta"`
}

type messageRecipientGroupTargetSaveRequest struct {
	TargetType               string          `json:"target_type"`
	UserID                   string          `json:"user_id"`
	CollaborationWorkspaceID string          `json:"collaboration_workspace_id"`
	RoleCode                 string          `json:"role_code"`
	PackageKey               string          `json:"package_key"`
	SortOrder                int             `json:"sort_order"`
	Meta                     models.MetaJSON `json:"meta"`
}

type messageRecipientGroupSaveRequest struct {
	Name        string                                   `json:"name"`
	Description string                                   `json:"description"`
	MatchMode   string                                   `json:"match_mode"`
	Status      string                                   `json:"status"`
	Meta        models.MetaJSON                          `json:"meta"`
	Targets     []messageRecipientGroupTargetSaveRequest `json:"targets"`
}

type messageRecipientGroupTargetItem struct {
	ID                         uuid.UUID       `json:"id"`
	TargetType                 string          `json:"target_type"`
	UserID                     *uuid.UUID      `json:"user_id"`
	UserName                   string          `json:"user_name"`
	CollaborationWorkspaceID   *uuid.UUID      `json:"collaboration_workspace_id,omitempty"`
	CollaborationWorkspaceName string          `json:"collaboration_workspace_name"`
	RoleCode                   string          `json:"role_code"`
	RoleName                   string          `json:"role_name"`
	PackageKey                 string          `json:"package_key"`
	PackageName                string          `json:"package_name"`
	SortOrder                  int             `json:"sort_order"`
	Meta                       models.MetaJSON `json:"meta"`
}

type messageRecipientGroupListItem struct {
	ID             uuid.UUID                         `json:"id"`
	ScopeType      string                            `json:"scope_type"`
	ScopeID        *uuid.UUID                        `json:"scope_id"`
	Name           string                            `json:"name"`
	Description    string                            `json:"description"`
	MatchMode      string                            `json:"match_mode"`
	Status         string                            `json:"status"`
	Editable       bool                              `json:"editable"`
	EstimatedCount int                               `json:"estimated_count"`
	Meta           models.MetaJSON                   `json:"meta"`
	Targets        []messageRecipientGroupTargetItem `json:"targets"`
	CreatedAt      time.Time                         `json:"created_at"`
	UpdatedAt      time.Time                         `json:"updated_at"`
}

type dispatchRecordQuery struct {
	Keyword      string
	MessageType  string
	AudienceType string
	Current      int
	Size         int
}

type dispatchRecordSummary struct {
	TotalMessages   int64 `json:"total_messages"`
	TotalDeliveries int64 `json:"total_deliveries"`
	ReadDeliveries  int64 `json:"read_deliveries"`
	TodoMessages    int64 `json:"todo_messages"`
}

type dispatchRecordListResult struct {
	Records []dispatchRecordListItem `json:"records"`
	Current int                      `json:"current"`
	Size    int                      `json:"size"`
	Total   int64                    `json:"total"`
	Summary dispatchRecordSummary    `json:"summary"`
}

type dispatchRecordListItem struct {
	ID                               uuid.UUID  `json:"id"`
	Title                            string     `json:"title"`
	Summary                          string     `json:"summary"`
	Content                          string     `json:"content"`
	MessageType                      string     `json:"message_type"`
	AudienceType                     string     `json:"audience_type"`
	ScopeType                        string     `json:"scope_type"`
	ScopeID                          *uuid.UUID `json:"scope_id"`
	TargetCollaborationWorkspaceID   *uuid.UUID `json:"target_collaboration_workspace_id"`
	TargetCollaborationWorkspaceName string     `json:"target_collaboration_workspace_name"`
	SenderName                       string     `json:"sender_name"`
	TemplateName                     string     `json:"template_name"`
	Priority                         string     `json:"priority"`
	Status                           string     `json:"status"`
	PublishedAt                      *time.Time `json:"published_at"`
	CreatedAt                        time.Time  `json:"created_at"`
	DeliveryCount                    int64      `json:"delivery_count"`
	ReadCount                        int64      `json:"read_count"`
	UnreadCount                      int64      `json:"unread_count"`
	PendingTodoCount                 int64      `json:"pending_todo_count"`
}

type dispatchRecordDeliveryItem struct {
	ID                                uuid.UUID  `json:"id"`
	RecipientUserID                   uuid.UUID  `json:"recipient_user_id"`
	RecipientName                     string     `json:"recipient_name"`
	RecipientCollaborationWorkspaceID *uuid.UUID `json:"recipient_collaboration_workspace_id"`
	RecipientCollaborationWorkspace   string     `json:"recipient_collaboration_workspace_name"`
	DeliveryStatus                    string     `json:"delivery_status"`
	TodoStatus                        string     `json:"todo_status"`
	ReadAt                            *time.Time `json:"read_at"`
	DoneAt                            *time.Time `json:"done_at"`
	LastActionAt                      *time.Time `json:"last_action_at"`
	SourceGroupID                     *uuid.UUID `json:"source_group_id"`
	SourceGroupName                   string     `json:"source_group_name"`
	SourceRuleType                    string     `json:"source_rule_type"`
	SourceRuleLabel                   string     `json:"source_rule_label"`
	SourceTargetID                    *uuid.UUID `json:"source_target_id"`
	SourceTargetType                  string     `json:"source_target_type"`
	SourceTargetValue                 string     `json:"source_target_value"`
}

type dispatchRecordDetail struct {
	dispatchRecordListItem
	Deliveries []dispatchRecordDeliveryItem `json:"deliveries"`
}

type messageService struct {
	db            *gorm.DB
	logger        *zap.Logger
	dispatchQueue chan uuid.UUID
}

func NewMessageService(db *gorm.DB, logger *zap.Logger) *messageService {
	service := &messageService{
		db:            db,
		logger:        logger,
		dispatchQueue: make(chan uuid.UUID, 256),
	}
	go service.runDispatchWorker()
	return service
}

func (s *messageService) enqueueDispatch(messageID uuid.UUID) {
	select {
	case s.dispatchQueue <- messageID:
	default:
		if s.logger != nil {
			s.logger.Warn("Message dispatch queue is full, fallback to scheduled scan", zap.String("message_id", messageID.String()))
		}
	}
}

func (s *messageService) runDispatchWorker() {
	ticker := time.NewTicker(12 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case messageID := <-s.dispatchQueue:
			s.processQueuedMessage(messageID)
		case <-ticker.C:
			s.scanQueuedMessages()
		}
	}
}

func (s *messageService) scanQueuedMessages() {
	staleAt := time.Now().Add(-5 * time.Minute)
	var messageIDs []uuid.UUID
	err := s.db.Model(&models.Message{}).
		Where("deleted_at IS NULL").
		Where("(status = ?) OR (status = ? AND updated_at < ?)", "queued", "processing", staleAt).
		Order("created_at ASC").
		Limit(20).
		Pluck("id", &messageIDs).Error
	if err != nil {
		if s.logger != nil {
			s.logger.Error("Scan queued messages failed", zap.Error(err))
		}
		return
	}
	for _, messageID := range messageIDs {
		s.processQueuedMessage(messageID)
	}
}

func (s *messageService) processQueuedMessage(messageID uuid.UUID) {
	var message models.Message
	if err := s.db.Where("id = ? AND deleted_at IS NULL", messageID).First(&message).Error; err != nil {
		if err != gorm.ErrRecordNotFound && s.logger != nil {
			s.logger.Error("Load queued message failed", zap.String("message_id", messageID.String()), zap.Error(err))
		}
		return
	}
	if message.Status != "queued" && message.Status != "processing" {
		return
	}

	message.Meta = cloneMetaJSON(message.Meta)
	message.Meta["dispatch_status"] = "processing"
	message.Meta["dispatch_error"] = ""
	if err := s.db.Model(&message).Updates(map[string]interface{}{
		"status": "processing",
		"meta":   message.Meta,
	}).Error; err != nil {
		if s.logger != nil {
			s.logger.Error("Mark message processing failed", zap.String("message_id", messageID.String()), zap.Error(err))
		}
		return
	}

	var collaborationWorkspaceID *uuid.UUID
	if (message.ScopeType == "collaboration") && message.ScopeID != nil {
		collaborationWorkspaceID = message.ScopeID
	}
	targetCollaborationWorkspaceIDs, err := parseMetaUUIDList(message.Meta["target_collaboration_workspace_ids"], "目标协作空间标识无效")
	if err != nil {
		s.markMessageDispatchFailed(&message, err)
		return
	}
	targetUserIDs, err := parseUUIDStrings(message.TargetUserIDs, "目标用户标识无效")
	if err != nil {
		s.markMessageDispatchFailed(&message, err)
		return
	}
	targetGroupIDs, err := parseUUIDStrings(message.TargetGroupIDs, "接收组标识无效")
	if err != nil {
		s.markMessageDispatchFailed(&message, err)
		return
	}

	recipients, err := s.resolveRecipients(message.AudienceType, collaborationWorkspaceID, targetCollaborationWorkspaceIDs, targetUserIDs, targetGroupIDs)
	if err != nil {
		s.markMessageDispatchFailed(&message, err)
		return
	}
	if len(recipients) == 0 {
		s.markMessageDispatchFailed(&message, errors.New("当前发送范围内没有可投递的接收人"))
		return
	}

	boxType := message.MessageType
	todoStatus := ""
	if message.MessageType == "todo" {
		boxType = "todo"
		todoStatus = "pending"
	}
	now := time.Now()
	err = s.db.Transaction(func(tx *gorm.DB) error {
		deliveries := make([]models.MessageDelivery, 0, len(recipients))
		for _, recipient := range recipients {
			deliveries = append(deliveries, models.MessageDelivery{
				MessageID:                         message.ID,
				RecipientUserID:                   recipient.UserID,
				RecipientCollaborationWorkspaceID: recipient.CollaborationWorkspaceID,
				BoxType:                           boxType,
				DeliveryStatus:                    "unread",
				TodoStatus:                        todoStatus,
				Meta: models.MetaJSON{
					"recipient_username":  recipient.Username,
					"source_group_id":     uuidString(recipient.SourceGroupID),
					"source_group_name":   recipient.SourceGroupName,
					"source_rule_type":    recipient.SourceRuleType,
					"source_rule_label":   recipient.SourceRuleLabel,
					"source_target_id":    uuidString(recipient.SourceTargetID),
					"source_target_type":  recipient.SourceTargetType,
					"source_target_value": recipient.SourceTargetValue,
				},
			})
		}
		if err := tx.Create(&deliveries).Error; err != nil {
			return err
		}
		nextMeta := cloneMetaJSON(message.Meta)
		nextMeta["dispatch_status"] = "published"
		nextMeta["dispatch_error"] = ""
		nextMeta["recipient_count"] = len(recipients)
		nextMeta["published_at"] = now.Format(time.RFC3339)
		return tx.Model(&models.Message{}).
			Where("id = ?", message.ID).
			Updates(map[string]interface{}{
				"status":       "published",
				"published_at": &now,
				"meta":         nextMeta,
			}).Error
	})
	if err != nil {
		s.markMessageDispatchFailed(&message, err)
		return
	}
	if s.logger != nil {
		s.logger.Info("Async message dispatch completed",
			zap.String("message_id", message.ID.String()),
			zap.Int("delivery_count", len(recipients)),
		)
	}
}

func (s *messageService) markMessageDispatchFailed(message *models.Message, dispatchErr error) {
	if message == nil {
		return
	}
	nextMeta := cloneMetaJSON(message.Meta)
	nextMeta["dispatch_status"] = "failed"
	nextMeta["dispatch_error"] = strings.TrimSpace(dispatchErr.Error())
	nextMeta["recipient_count"] = 0
	nextMeta["failed_at"] = time.Now().Format(time.RFC3339)
	if err := s.db.Model(&models.Message{}).
		Where("id = ?", message.ID).
		Updates(map[string]interface{}{
			"status": "failed",
			"meta":   nextMeta,
		}).Error; err != nil {
		if s.logger != nil {
			s.logger.Error("Mark message failed failed",
				zap.String("message_id", message.ID.String()),
				zap.Error(err),
			)
		}
		return
	}
	if s.logger != nil {
		s.logger.Warn("Async message dispatch failed",
			zap.String("message_id", message.ID.String()),
			zap.Error(dispatchErr),
		)
	}
}

func (s *messageService) GetInboxSummary(userID uuid.UUID) (inboxSummary, error) {
	var rows []struct {
		BoxType string
		Total   int64
	}

	now := time.Now()
	err := s.baseInboxQuery(userID).
		Where("message_deliveries.delivery_status = ?", "unread").
		Where("messages.status = ?", "published").
		Where("messages.expired_at IS NULL OR messages.expired_at > ?", now).
		Select("message_deliveries.box_type AS box_type, COUNT(*) AS total").
		Group("message_deliveries.box_type").
		Scan(&rows).Error
	if err != nil {
		return inboxSummary{}, err
	}

	result := inboxSummary{}
	for _, row := range rows {
		result.UnreadTotal += row.Total
		switch strings.TrimSpace(row.BoxType) {
		case "notice":
			result.NoticeCount = row.Total
		case "message":
			result.MessageCount = row.Total
		case "todo":
			result.TodoCount = row.Total
		}
	}
	return result, nil
}

func (s *messageService) ListInbox(userID uuid.UUID, query inboxQuery) (inboxListResult, error) {
	current := query.Current
	size := query.Size
	if current <= 0 {
		current = 1
	}
	if size <= 0 || size > 100 {
		size = 20
	}

	base := s.filteredInboxQuery(userID, query)
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return inboxListResult{}, err
	}

	var records []inboxListItem
	err := s.filteredInboxQuery(userID, query).
		Select(inboxSelectColumns()).
		Order("COALESCE(messages.published_at, message_deliveries.created_at) DESC").
		Order("message_deliveries.created_at DESC").
		Offset((current - 1) * size).
		Limit(size).
		Scan(&records).Error
	if err != nil {
		return inboxListResult{}, err
	}

	return inboxListResult{
		Records: records,
		Current: current,
		Size:    size,
		Total:   total,
	}, nil
}

func (s *messageService) GetInboxDetail(userID, deliveryID uuid.UUID) (inboxDetail, error) {
	var detail inboxDetail
	err := s.baseInboxQuery(userID).
		Where("message_deliveries.id = ?", deliveryID).
		Select(inboxSelectColumns()).
		Scan(&detail).Error
	if err != nil {
		return inboxDetail{}, err
	}
	if detail.ID == uuid.Nil {
		return inboxDetail{}, gorm.ErrRecordNotFound
	}
	return detail, nil
}

func (s *messageService) MarkRead(userID, deliveryID uuid.UUID) error {
	now := time.Now()
	result := s.db.Model(&models.MessageDelivery{}).
		Where("id = ? AND recipient_user_id = ? AND deleted_at IS NULL", deliveryID, userID).
		Updates(map[string]interface{}{
			"delivery_status": "read",
			"read_at":         &now,
			"last_action_at":  &now,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (s *messageService) MarkAllRead(userID uuid.UUID, boxType string) error {
	now := time.Now()
	tx := s.db.Model(&models.MessageDelivery{}).
		Where("recipient_user_id = ? AND deleted_at IS NULL", userID).
		Where("delivery_status = ?", "unread")
	if normalized := normalizeBoxType(boxType); normalized != "" {
		tx = tx.Where("box_type = ?", normalized)
	}
	return tx.Updates(map[string]interface{}{
		"delivery_status": "read",
		"read_at":         &now,
		"last_action_at":  &now,
	}).Error
}

func (s *messageService) UpdateTodoStatus(userID, deliveryID uuid.UUID, action string) error {
	normalizedAction := strings.TrimSpace(action)
	if normalizedAction != "done" && normalizedAction != "ignored" {
		return errors.New("无效的待办操作")
	}
	now := time.Now()
	updates := map[string]interface{}{
		"delivery_status": "read",
		"todo_status":     normalizedAction,
		"read_at":         &now,
		"last_action_at":  &now,
	}
	if normalizedAction == "done" {
		updates["done_at"] = &now
	}
	result := s.db.Model(&models.MessageDelivery{}).
		Where("id = ? AND recipient_user_id = ? AND deleted_at IS NULL", deliveryID, userID).
		Where("box_type = ?", "todo").
		Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (s *messageService) GetDispatchOptions(userID uuid.UUID, collaborationWorkspaceID *uuid.UUID) (dispatchOptions, error) {
	result := dispatchOptions{
		SenderOptions:           make([]dispatchSenderOption, 0, 4),
		AudienceOptions:         make([]dispatchAudienceOption, 0, 3),
		TemplateOptions:         make([]dispatchTemplateOption, 0),
		CollaborationWorkspaces: make([]dispatchCollaborationWorkspaceOption, 0),
		Users:                   make([]dispatchUserOption, 0),
		RecipientGroups:         make([]dispatchRecipientGroupOption, 0),
		Roles:                   make([]dispatchRoleOption, 0),
		FeaturePackages:         make([]dispatchFeaturePackageOption, 0),
		DefaultMessageType:      "notice",
		DefaultAudienceType:     "all_users",
		DefaultPriority:         "normal",
		SupportsExternalLink:    true,
	}

	if collaborationWorkspaceID != nil {
		result.SenderScope = "collaboration"
		result.DefaultAudienceType = "collaboration_workspace_users"
		var currentCollaborationWorkspace models.CollaborationWorkspace
		if err := s.db.Select("id", "name").Where("id = ?", *collaborationWorkspaceID).First(&currentCollaborationWorkspace).Error; err == nil {
			result.CurrentCollaborationWorkspaceID = currentCollaborationWorkspace.ID.String()
			result.CurrentCollaborationWorkspaceName = currentCollaborationWorkspace.Name
			result.CollaborationWorkspaces = append(result.CollaborationWorkspaces, dispatchCollaborationWorkspaceOption{
				ID:   currentCollaborationWorkspace.ID,
				Name: currentCollaborationWorkspace.Name,
			})
		}
		result.AudienceOptions = append(result.AudienceOptions, dispatchAudienceOption{
			Value:       "collaboration_workspace_users",
			Label:       "当前协作空间成员",
			Description: "给当前协作空间的全部有效成员发送消息。",
		})
		result.AudienceOptions = append(result.AudienceOptions,
			dispatchAudienceOption{
				Value:       "specified_users",
				Label:       "指定成员",
				Description: "给当前协作空间内指定成员发送消息。",
			},
			dispatchAudienceOption{
				Value:       "recipient_group",
				Label:       "接收组",
				Description: "按协作空间下预设的接收组展开成员并发送。",
			},
			dispatchAudienceOption{
				Value:       "role",
				Label:       "接收组中的角色规则",
				Description: "仅展开已选接收组中的角色规则匹配成员。",
			},
			dispatchAudienceOption{
				Value:       "feature_package",
				Label:       "接收组中的功能包规则",
				Description: "仅展开已选接收组中的功能包规则匹配成员。",
			},
		)
	} else {
		result.SenderScope = "personal"
		result.AudienceOptions = append(result.AudienceOptions,
			dispatchAudienceOption{
				Value:       "all_users",
				Label:       "所有用户",
				Description: "当前个人空间向全部有效用户发送。",
			},
			dispatchAudienceOption{
				Value:       "collaboration_workspace_admins",
				Label:       "协作空间管理员",
				Description: "当前个人空间给选定协作空间的管理员发送。",
			},
			dispatchAudienceOption{
				Value:       "collaboration_workspace_users",
				Label:       "指定协作空间成员",
				Description: "当前个人空间给选定协作空间的全部有效成员发送。",
			},
			dispatchAudienceOption{
				Value:       "specified_users",
				Label:       "指定用户",
				Description: "当前个人空间直接给一个或多个指定用户发送。",
			},
			dispatchAudienceOption{
				Value:       "recipient_group",
				Label:       "接收组",
				Description: "按个人空间或协作空间预设的接收组展开成员并发送。",
			},
			dispatchAudienceOption{
				Value:       "role",
				Label:       "接收组中的角色规则",
				Description: "仅展开已选接收组中的角色规则匹配用户。",
			},
			dispatchAudienceOption{
				Value:       "feature_package",
				Label:       "接收组中的功能包规则",
				Description: "仅展开已选接收组中的功能包规则匹配用户。",
			},
		)
		var collaborationWorkspaces []models.CollaborationWorkspace
		if err := s.db.
			Select("id", "name").
			Where("status = ?", "active").
			Order("created_at ASC").
			Find(&collaborationWorkspaces).Error; err != nil {
			return dispatchOptions{}, err
		}
		for _, item := range collaborationWorkspaces {
			result.CollaborationWorkspaces = append(result.CollaborationWorkspaces, dispatchCollaborationWorkspaceOption{ID: item.ID, Name: item.Name})
		}
	}

	senders, err := s.listSenderOptions(collaborationWorkspaceID)
	if err != nil {
		return dispatchOptions{}, err
	}
	result.SenderOptions = senders
	for _, item := range senders {
		if item.IsDefault {
			result.DefaultSenderID = item.ID.String()
			break
		}
	}
	if result.DefaultSenderID == "" && len(senders) > 0 {
		result.DefaultSenderID = senders[0].ID.String()
	}

	templateQuery := s.db.Model(&models.MessageTemplate{}).
		Select("id", "template_key", "name", "description", "message_type", "owner_scope", "audience_type", "title_template", "summary_template", "content_template").
		Where("status = ?", "normal")
	if collaborationWorkspaceID != nil {
		templateQuery = templateQuery.Where("owner_scope = ? AND owner_collaboration_workspace_id = ?", "collaboration", *collaborationWorkspaceID)
	} else {
		templateQuery = templateQuery.Where("owner_scope = ?", "personal")
	}

	var templates []dispatchTemplateOption
	if err := templateQuery.Order("created_at ASC").Scan(&templates).Error; err != nil {
		return dispatchOptions{}, err
	}
	result.TemplateOptions = templates

	users, err := s.listDispatchUsers(collaborationWorkspaceID)
	if err != nil {
		return dispatchOptions{}, err
	}
	result.Users = users

	groups, err := s.listDispatchRecipientGroups(collaborationWorkspaceID)
	if err != nil {
		return dispatchOptions{}, err
	}
	result.RecipientGroups = groups

	roles, err := s.listDispatchRoles(collaborationWorkspaceID)
	if err != nil {
		return dispatchOptions{}, err
	}
	result.Roles = roles

	packages, err := s.listDispatchFeaturePackages(collaborationWorkspaceID)
	if err != nil {
		return dispatchOptions{}, err
	}
	result.FeaturePackages = packages

	return result, nil
}

func (s *messageService) DispatchMessage(userID uuid.UUID, collaborationWorkspaceID *uuid.UUID, req dispatchRequest) (dispatchResult, error) {
	sender, err := s.resolveDispatchSender(collaborationWorkspaceID, strings.TrimSpace(req.SenderID))
	if err != nil {
		return dispatchResult{}, err
	}

	template, err := s.resolveTemplate(strings.TrimSpace(req.TemplateID), strings.TrimSpace(req.TemplateKey), collaborationWorkspaceID)
	if err != nil {
		return dispatchResult{}, err
	}

	messageType := normalizeMessageType(strings.TrimSpace(req.MessageType))
	if messageType == "" && template != nil {
		messageType = normalizeMessageType(template.MessageType)
	}
	if messageType == "" {
		messageType = "notice"
	}

	audienceType := normalizeAudienceType(strings.TrimSpace(req.AudienceType))
	if audienceType == "" && template != nil {
		audienceType = normalizeAudienceType(template.AudienceType)
	}
	if audienceType == "" {
		if collaborationWorkspaceID != nil {
			audienceType = "collaboration_workspace_users"
		} else {
			audienceType = "all_users"
		}
	}

	title := strings.TrimSpace(req.Title)
	summary := strings.TrimSpace(req.Summary)
	content := strings.TrimSpace(req.Content)
	actionType := normalizeActionType(strings.TrimSpace(req.ActionType))
	actionTarget := strings.TrimSpace(req.ActionTarget)
	if template != nil {
		if title == "" {
			title = strings.TrimSpace(template.TitleTemplate)
		}
		if summary == "" {
			summary = strings.TrimSpace(template.SummaryTemplate)
		}
		if content == "" {
			content = strings.TrimSpace(template.ContentTemplate)
		}
		if actionType == "" {
			actionType = normalizeActionType(template.ActionType)
		}
		if actionTarget == "" {
			actionTarget = strings.TrimSpace(template.ActionTargetTemplate)
		}
	}
	if title == "" {
		return dispatchResult{}, errors.New("消息标题不能为空")
	}
	if actionType == "" {
		actionType = "none"
	}

	targetCollaborationWorkspaceIDs, err := parseTargetCollaborationWorkspaceIDs(req.TargetCollaborationWorkspaceIDs)
	if err != nil {
		return dispatchResult{}, err
	}
	targetCollaborationWorkspaceIDs, err = s.resolveLegacyCollaborationWorkspaceIDs(targetCollaborationWorkspaceIDs)
	if err != nil {
		return dispatchResult{}, err
	}
	targetUserIDs, err := parseUUIDStrings(req.TargetUserIDs, "目标用户标识无效")
	if err != nil {
		return dispatchResult{}, err
	}
	targetGroupIDs, err := parseUUIDStrings(req.TargetGroupIDs, "接收组标识无效")
	if err != nil {
		return dispatchResult{}, err
	}
	targetCollaborationWorkspaceIDs, err = s.normalizeTargetCollaborationWorkspaces(audienceType, collaborationWorkspaceID, targetCollaborationWorkspaceIDs)
	if err != nil {
		return dispatchResult{}, err
	}
	targetUserIDs, targetGroupIDs, err = s.normalizeAudienceTargets(audienceType, collaborationWorkspaceID, targetUserIDs, targetGroupIDs)
	if err != nil {
		return dispatchResult{}, err
	}

	priority := normalizePriority(req.Priority)
	if priority == "" {
		priority = "normal"
	}

	var expiredAt *time.Time
	if target := strings.TrimSpace(req.ExpiredAt); target != "" {
		parsed, parseErr := time.Parse(time.RFC3339, target)
		if parseErr != nil {
			return dispatchResult{}, fmt.Errorf("失效时间格式错误")
		}
		expiredAt = &parsed
	}

	scopeType := "personal"
	var scopeID *uuid.UUID
	audienceScope := "personal"
	targetCollaborationWorkspaceID := singleCollaborationWorkspaceID(targetCollaborationWorkspaceIDs)
	senderType := "personal_workspace_sender"
	senderName := strings.TrimSpace(sender.Name)
	if collaborationWorkspaceID != nil {
		scopeType = "collaboration"
		scopeID = collaborationWorkspaceID
		audienceScope = "collaboration"
		senderType = "collaboration_workspace_sender"
		targetCollaborationWorkspaceID = collaborationWorkspaceID
	}

	now := time.Now()
	meta := models.MetaJSON{
		"target_collaboration_workspace_ids": uuidListToStringList(targetCollaborationWorkspaceIDs),
		"target_user_ids":                    uuidListToStringList(targetUserIDs),
		"target_group_ids":                   uuidListToStringList(targetGroupIDs),
		"dispatch_status":                    "queued",
		"dispatch_error":                     "",
		"recipient_count":                    0,
		"queued_at":                          now.Format(time.RFC3339),
	}

	message := models.Message{
		MessageType:                    messageType,
		BizType:                        strings.TrimSpace(req.BizType),
		ScopeType:                      scopeType,
		ScopeID:                        scopeID,
		SenderID:                       &sender.ID,
		SenderType:                     senderType,
		SenderUserID:                   &userID,
		SenderNameSnapshot:             senderName,
		SenderAvatarSnapshot:           strings.TrimSpace(sender.AvatarURL),
		AudienceType:                   audienceType,
		AudienceScope:                  audienceScope,
		TargetCollaborationWorkspaceID: targetCollaborationWorkspaceID,
		TargetUserIDs:                  uuidListToStringList(targetUserIDs),
		TargetGroupIDs:                 uuidListToStringList(targetGroupIDs),
		TemplateID:                     uuidPtrFromTemplate(template),
		Title:                          title,
		Summary:                        summary,
		Content:                        content,
		Priority:                       priority,
		ActionType:                     actionType,
		ActionTarget:                   actionTarget,
		Status:                         "queued",
		PublishedAt:                    nil,
		ExpiredAt:                      expiredAt,
		Meta:                           meta,
	}

	// DryRun：校验已全部通过，但不落库、不入队、不触发投递。
	// 用于 /system/message 与 /collaboration-workspace/message 的沙箱预览。
	// 注意：此分支必须放在 db.Create 之前，确保零副作用。
	if req.DryRun {
		return dispatchResult{
			MessageID:      uuid.Nil,
			DeliveryCount:  0,
			DispatchStatus: "preview",
		}, nil
	}

	if err := s.db.Create(&message).Error; err != nil {
		return dispatchResult{}, err
	}
	s.enqueueDispatch(message.ID)

	return dispatchResult{
		MessageID:      message.ID,
		DeliveryCount:  0,
		DispatchStatus: "queued",
	}, nil
}

func (s *messageService) ListTemplates(collaborationWorkspaceID *uuid.UUID, query messageTemplateQuery) (messageTemplateListResult, error) {
	current := query.Current
	size := query.Size
	if current <= 0 {
		current = 1
	}
	if size <= 0 || size > 100 {
		size = 20
	}

	base := s.templateScopeQuery(collaborationWorkspaceID)
	if keyword := strings.TrimSpace(query.Keyword); keyword != "" {
		like := "%" + keyword + "%"
		base = base.Where(
			"message_templates.template_key ILIKE ? OR message_templates.name ILIKE ? OR message_templates.description ILIKE ?",
			like,
			like,
			like,
		)
	}

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return messageTemplateListResult{}, err
	}

	var rows []struct {
		models.MessageTemplate
		OwnerCollaborationWorkspaceName string `gorm:"column:owner_collaboration_workspace_name"`
	}
	err := base.
		Select("message_templates.*", "COALESCE(owner_collaboration_workspaces.name, '') AS owner_collaboration_workspace_name").
		Joins("LEFT JOIN collaboration_workspaces AS owner_collaboration_workspaces ON owner_collaboration_workspaces.id = message_templates.owner_collaboration_workspace_id").
		Order("CASE WHEN message_templates.owner_scope = 'collaboration' THEN 0 ELSE 1 END").
		Order("message_templates.updated_at DESC").
		Offset((current - 1) * size).
		Limit(size).
		Scan(&rows).Error
	if err != nil {
		return messageTemplateListResult{}, err
	}

	records := make([]messageTemplateListItem, 0, len(rows))
	for _, row := range rows {
		records = append(records, s.buildTemplateListItem(row.MessageTemplate, row.OwnerCollaborationWorkspaceName, collaborationWorkspaceID))
	}

	return messageTemplateListResult{
		Records: records,
		Current: current,
		Size:    size,
		Total:   total,
	}, nil
}

func (s *messageService) SaveTemplate(templateID string, collaborationWorkspaceID *uuid.UUID, req messageTemplateUpsertRequest) (messageTemplateListItem, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return messageTemplateListItem{}, errors.New("模板名称不能为空")
	}

	messageType := normalizeMessageType(req.MessageType)
	if messageType == "" {
		return messageTemplateListItem{}, errors.New("消息类型无效")
	}

	audienceType := normalizeAudienceType(req.AudienceType)
	if audienceType == "" {
		if collaborationWorkspaceID != nil {
			audienceType = "collaboration_workspace_users"
		} else {
			return messageTemplateListItem{}, errors.New("发送对象无效")
		}
	}
	if collaborationWorkspaceID != nil && audienceType != "collaboration_workspace_users" {
		return messageTemplateListItem{}, errors.New("协作空间模板只能面向当前协作空间成员")
	}

	status := normalizeTemplateStatus(req.Status)
	if status == "" {
		status = "normal"
	}

	now := time.Now().Unix()
	templateKey := buildTemplateKey(req.TemplateKey, collaborationWorkspaceID, now)

	var saved models.MessageTemplate
	err := s.db.Transaction(func(tx *gorm.DB) error {
		var target models.MessageTemplate
		if trimmedID := strings.TrimSpace(templateID); trimmedID != "" {
			id, parseErr := uuid.Parse(trimmedID)
			if parseErr != nil {
				return errors.New("模板标识无效")
			}
			existing, loadErr := s.loadEditableTemplate(tx, id, collaborationWorkspaceID)
			if loadErr != nil {
				return loadErr
			}
			target = existing
		} else {
			target = models.MessageTemplate{}
			if collaborationWorkspaceID != nil {
				target.OwnerScope = "collaboration"
				target.OwnerCollaborationWorkspaceID = collaborationWorkspaceID
			} else {
				target.OwnerScope = "personal"
			}
		}

		target.TemplateKey = templateKey
		target.Name = name
		target.Description = strings.TrimSpace(req.Description)
		target.MessageType = messageType
		target.AudienceType = audienceType
		target.TitleTemplate = strings.TrimSpace(req.TitleTemplate)
		target.SummaryTemplate = strings.TrimSpace(req.SummaryTemplate)
		target.ContentTemplate = strings.TrimSpace(req.ContentTemplate)
		target.ActionType = "none"
		target.ActionTargetTemplate = ""
		target.Status = status

		if target.ID == uuid.Nil {
			if err := tx.Create(&target).Error; err != nil {
				return convertTemplatePersistenceError(err)
			}
		} else {
			if err := tx.Save(&target).Error; err != nil {
				return convertTemplatePersistenceError(err)
			}
		}
		saved = target
		return nil
	})
	if err != nil {
		return messageTemplateListItem{}, err
	}

	ownerCollaborationWorkspaceName := ""
	if saved.OwnerCollaborationWorkspaceID != nil {
		var collaborationWorkspace models.CollaborationWorkspace
		if err := s.db.Select("name").Where("id = ?", *saved.OwnerCollaborationWorkspaceID).First(&collaborationWorkspace).Error; err == nil {
			ownerCollaborationWorkspaceName = collaborationWorkspace.Name
		}
	}
	return s.buildTemplateListItem(saved, ownerCollaborationWorkspaceName, collaborationWorkspaceID), nil
}

func (s *messageService) ListSenders(collaborationWorkspaceID *uuid.UUID) ([]messageSenderListItem, error) {
	senders, err := s.ensureSenderOptions(collaborationWorkspaceID)
	if err != nil {
		return nil, err
	}
	result := make([]messageSenderListItem, 0, len(senders))
	for _, item := range senders {
		result = append(result, messageSenderListItem{
			ID:          item.ID,
			ScopeType:   item.ScopeType,
			ScopeID:     item.ScopeID,
			Name:        item.Name,
			Description: item.Description,
			AvatarURL:   item.AvatarURL,
			IsDefault:   item.IsDefault,
			Status:      item.Status,
			Editable:    true,
			Meta:        item.Meta,
			CreatedAt:   item.CreatedAt,
			UpdatedAt:   item.UpdatedAt,
		})
	}
	return result, nil
}

func (s *messageService) SaveSender(senderID string, collaborationWorkspaceID *uuid.UUID, req messageSenderSaveRequest) (messageSenderListItem, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return messageSenderListItem{}, errors.New("发送人名称不能为空")
	}
	status := strings.TrimSpace(req.Status)
	if status == "" {
		status = "normal"
	}
	if status != "normal" && status != "disabled" {
		return messageSenderListItem{}, errors.New("发送人状态无效")
	}
	meta := req.Meta
	if meta == nil {
		meta = models.MetaJSON{}
	}

	var saved models.MessageSender
	err := s.db.Transaction(func(tx *gorm.DB) error {
		scopeType := "personal"
		var scopeID *uuid.UUID
		if collaborationWorkspaceID != nil {
			scopeType = "collaboration"
			scopeID = collaborationWorkspaceID
		}

		var target models.MessageSender
		if trimmedID := strings.TrimSpace(senderID); trimmedID != "" {
			id, parseErr := uuid.Parse(trimmedID)
			if parseErr != nil {
				return errors.New("发送人标识无效")
			}
			query := tx.Model(&models.MessageSender{}).Where("id = ? AND deleted_at IS NULL", id)
			if collaborationWorkspaceID != nil {
				query = query.Where("scope_type = ? AND scope_id = ?", "collaboration", *collaborationWorkspaceID)
			} else {
				query = query.Where("scope_type = ? AND scope_id IS NULL", "personal")
			}
			if err := query.First(&target).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return errors.New("发送人不存在或当前上下文不可编辑")
				}
				return err
			}
		} else {
			target = models.MessageSender{
				ScopeType: scopeType,
				ScopeID:   scopeID,
			}
		}

		target.Name = name
		target.Description = strings.TrimSpace(req.Description)
		target.AvatarURL = strings.TrimSpace(req.AvatarURL)
		target.Status = status
		target.Meta = meta
		target.IsDefault = req.IsDefault && status == "normal"

		if target.ID == uuid.Nil {
			if err := tx.Create(&target).Error; err != nil {
				return err
			}
		} else {
			if err := tx.Save(&target).Error; err != nil {
				return err
			}
		}

		if target.IsDefault {
			scopeQuery := tx.Model(&models.MessageSender{}).Where("id <> ? AND deleted_at IS NULL AND scope_type = ?", target.ID, target.ScopeType)
			if target.ScopeID != nil {
				scopeQuery = scopeQuery.Where("scope_id = ?", *target.ScopeID)
			} else {
				scopeQuery = scopeQuery.Where("scope_id IS NULL")
			}
			if err := scopeQuery.Update("is_default", false).Error; err != nil {
				return err
			}
		}

		saved = target
		return nil
	})
	if err != nil {
		return messageSenderListItem{}, err
	}

	if _, err := s.ensureSenderOptions(collaborationWorkspaceID); err != nil {
		return messageSenderListItem{}, err
	}

	return messageSenderListItem{
		ID:          saved.ID,
		ScopeType:   saved.ScopeType,
		ScopeID:     saved.ScopeID,
		Name:        saved.Name,
		Description: saved.Description,
		AvatarURL:   saved.AvatarURL,
		IsDefault:   saved.IsDefault,
		Status:      saved.Status,
		Editable:    true,
		Meta:        saved.Meta,
		CreatedAt:   saved.CreatedAt,
		UpdatedAt:   saved.UpdatedAt,
	}, nil
}

func (s *messageService) ListDispatchRecords(collaborationWorkspaceID *uuid.UUID, query dispatchRecordQuery) (dispatchRecordListResult, error) {
	current := query.Current
	size := query.Size
	if current <= 0 {
		current = 1
	}
	if size <= 0 || size > 100 {
		size = 20
	}

	base := s.dispatchRecordBaseQuery(collaborationWorkspaceID)
	if messageType := normalizeMessageType(query.MessageType); messageType != "" {
		base = base.Where("messages.message_type = ?", messageType)
	}
	if audienceType := normalizeAudienceType(query.AudienceType); audienceType != "" {
		base = base.Where("messages.audience_type = ?", audienceType)
	}
	if keyword := strings.TrimSpace(query.Keyword); keyword != "" {
		like := "%" + keyword + "%"
		base = base.Where(
			"messages.title ILIKE ? OR messages.summary ILIKE ? OR messages.content ILIKE ? OR messages.sender_name_snapshot ILIKE ?",
			like,
			like,
			like,
			like,
		)
	}

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return dispatchRecordListResult{}, err
	}

	summary, err := s.loadDispatchRecordSummary(collaborationWorkspaceID)
	if err != nil {
		return dispatchRecordListResult{}, err
	}

	var records []dispatchRecordListItem
	err = base.
		Select(strings.Join([]string{
			"messages.id AS id",
			"messages.title AS title",
			"messages.summary AS summary",
			"messages.content AS content",
			"messages.message_type AS message_type",
			"messages.audience_type AS audience_type",
			"messages.scope_type AS scope_type",
			"messages.scope_id AS scope_id",
			"messages.target_collaboration_workspace_id AS target_collaboration_workspace_id",
			"COALESCE(target_collaboration_workspaces.name, '') AS target_collaboration_workspace_name",
			"messages.sender_name_snapshot AS sender_name",
			"COALESCE(message_templates.name, '') AS template_name",
			"messages.priority AS priority",
			"messages.status AS status",
			"messages.published_at AS published_at",
			"messages.created_at AS created_at",
			"COUNT(message_deliveries.id) AS delivery_count",
			"SUM(CASE WHEN message_deliveries.delivery_status = 'read' THEN 1 ELSE 0 END) AS read_count",
			"SUM(CASE WHEN message_deliveries.delivery_status = 'unread' THEN 1 ELSE 0 END) AS unread_count",
			"SUM(CASE WHEN message_deliveries.todo_status = 'pending' THEN 1 ELSE 0 END) AS pending_todo_count",
		}, ", ")).
		Joins("LEFT JOIN message_templates ON message_templates.id = messages.template_id").
		Joins("LEFT JOIN collaboration_workspaces AS target_collaboration_workspaces ON target_collaboration_workspaces.id = messages.target_collaboration_workspace_id").
		Joins("LEFT JOIN message_deliveries ON message_deliveries.message_id = messages.id AND message_deliveries.deleted_at IS NULL").
		Group("messages.id, target_collaboration_workspaces.name, message_templates.name").
		Order("COALESCE(messages.published_at, messages.created_at) DESC").
		Order("messages.created_at DESC").
		Offset((current - 1) * size).
		Limit(size).
		Scan(&records).Error
	if err != nil {
		return dispatchRecordListResult{}, err
	}

	return dispatchRecordListResult{
		Records: records,
		Current: current,
		Size:    size,
		Total:   total,
		Summary: summary,
	}, nil
}

func (s *messageService) GetDispatchRecordDetail(collaborationWorkspaceID *uuid.UUID, recordID string) (dispatchRecordDetail, error) {
	id, err := uuid.Parse(strings.TrimSpace(recordID))
	if err != nil {
		return dispatchRecordDetail{}, errors.New("发送记录标识无效")
	}

	type dispatchRecordDetailRow struct {
		ID                               uuid.UUID  `gorm:"column:id"`
		Title                            string     `gorm:"column:title"`
		Summary                          string     `gorm:"column:summary"`
		Content                          string     `gorm:"column:content"`
		MessageType                      string     `gorm:"column:message_type"`
		AudienceType                     string     `gorm:"column:audience_type"`
		ScopeType                        string     `gorm:"column:scope_type"`
		ScopeID                          *uuid.UUID `gorm:"column:scope_id"`
		TargetCollaborationWorkspaceID   *uuid.UUID `gorm:"column:target_collaboration_workspace_id"`
		TargetCollaborationWorkspaceName string     `gorm:"column:target_collaboration_workspace_name"`
		SenderName                       string     `gorm:"column:sender_name"`
		TemplateName                     string     `gorm:"column:template_name"`
		Priority                         string     `gorm:"column:priority"`
		Status                           string     `gorm:"column:status"`
		PublishedAt                      *time.Time `gorm:"column:published_at"`
		CreatedAt                        time.Time  `gorm:"column:created_at"`
		DeliveryCount                    int64      `gorm:"column:delivery_count"`
		ReadCount                        int64      `gorm:"column:read_count"`
		UnreadCount                      int64      `gorm:"column:unread_count"`
		PendingTodoCount                 int64      `gorm:"column:pending_todo_count"`
	}

	var row dispatchRecordDetailRow
	err = s.dispatchRecordBaseQuery(collaborationWorkspaceID).
		Where("messages.id = ?", id).
		Select(strings.Join([]string{
			"messages.id AS id",
			"messages.title AS title",
			"messages.summary AS summary",
			"messages.content AS content",
			"messages.message_type AS message_type",
			"messages.audience_type AS audience_type",
			"messages.scope_type AS scope_type",
			"messages.scope_id AS scope_id",
			"messages.target_collaboration_workspace_id AS target_collaboration_workspace_id",
			"COALESCE(target_collaboration_workspaces.name, '') AS target_collaboration_workspace_name",
			"messages.sender_name_snapshot AS sender_name",
			"COALESCE(message_templates.name, '') AS template_name",
			"messages.priority AS priority",
			"messages.status AS status",
			"messages.published_at AS published_at",
			"messages.created_at AS created_at",
			"COUNT(message_deliveries.id) AS delivery_count",
			"SUM(CASE WHEN message_deliveries.delivery_status = 'read' THEN 1 ELSE 0 END) AS read_count",
			"SUM(CASE WHEN message_deliveries.delivery_status = 'unread' THEN 1 ELSE 0 END) AS unread_count",
			"SUM(CASE WHEN message_deliveries.todo_status = 'pending' THEN 1 ELSE 0 END) AS pending_todo_count",
		}, ", ")).
		Joins("LEFT JOIN message_templates ON message_templates.id = messages.template_id").
		Joins("LEFT JOIN collaboration_workspaces AS target_collaboration_workspaces ON target_collaboration_workspaces.id = messages.target_collaboration_workspace_id").
		Joins("LEFT JOIN message_deliveries ON message_deliveries.message_id = messages.id AND message_deliveries.deleted_at IS NULL").
		Group("messages.id, target_collaboration_workspaces.name, message_templates.name").
		Scan(&row).Error
	if err != nil {
		return dispatchRecordDetail{}, err
	}
	if row.ID == uuid.Nil {
		return dispatchRecordDetail{}, gorm.ErrRecordNotFound
	}

	detail := dispatchRecordDetail{
		dispatchRecordListItem: dispatchRecordListItem{
			ID:                               row.ID,
			Title:                            row.Title,
			Summary:                          row.Summary,
			Content:                          row.Content,
			MessageType:                      row.MessageType,
			AudienceType:                     row.AudienceType,
			ScopeType:                        row.ScopeType,
			ScopeID:                          row.ScopeID,
			TargetCollaborationWorkspaceID:   row.TargetCollaborationWorkspaceID,
			TargetCollaborationWorkspaceName: row.TargetCollaborationWorkspaceName,
			SenderName:                       row.SenderName,
			TemplateName:                     row.TemplateName,
			Priority:                         row.Priority,
			Status:                           row.Status,
			PublishedAt:                      row.PublishedAt,
			CreatedAt:                        row.CreatedAt,
			DeliveryCount:                    row.DeliveryCount,
			ReadCount:                        row.ReadCount,
			UnreadCount:                      row.UnreadCount,
			PendingTodoCount:                 row.PendingTodoCount,
		},
		Deliveries: make([]dispatchRecordDeliveryItem, 0),
	}

	type deliveryRow struct {
		ID                                uuid.UUID  `gorm:"column:id"`
		RecipientUserID                   uuid.UUID  `gorm:"column:recipient_user_id"`
		RecipientName                     string     `gorm:"column:recipient_name"`
		RecipientCollaborationWorkspaceID *uuid.UUID `gorm:"column:recipient_collaboration_workspace_id"`
		RecipientCollaborationWorkspace   string     `gorm:"column:recipient_collaboration_workspace_name"`
		DeliveryStatus                    string     `gorm:"column:delivery_status"`
		TodoStatus                        string     `gorm:"column:todo_status"`
		ReadAt                            *time.Time `gorm:"column:read_at"`
		DoneAt                            *time.Time `gorm:"column:done_at"`
		LastActionAt                      *time.Time `gorm:"column:last_action_at"`
		SourceGroupID                     string     `gorm:"column:source_group_id"`
		SourceGroupName                   string     `gorm:"column:source_group_name"`
		SourceRuleType                    string     `gorm:"column:source_rule_type"`
		SourceRuleLabel                   string     `gorm:"column:source_rule_label"`
		SourceTargetID                    string     `gorm:"column:source_target_id"`
		SourceTargetType                  string     `gorm:"column:source_target_type"`
		SourceTargetValue                 string     `gorm:"column:source_target_value"`
	}

	var rows []deliveryRow
	err = s.db.Table("message_deliveries").
		Select(strings.Join([]string{
			"message_deliveries.id AS id",
			"message_deliveries.recipient_user_id AS recipient_user_id",
			"COALESCE(users.nickname, users.username, '') AS recipient_name",
			"message_deliveries.recipient_collaboration_workspace_id AS recipient_collaboration_workspace_id",
			"COALESCE(collaboration_workspaces.name, '') AS recipient_collaboration_workspace_name",
			"message_deliveries.delivery_status AS delivery_status",
			"message_deliveries.todo_status AS todo_status",
			"message_deliveries.read_at AS read_at",
			"message_deliveries.done_at AS done_at",
			"message_deliveries.last_action_at AS last_action_at",
			"COALESCE(message_deliveries.meta ->> 'source_group_id', '') AS source_group_id",
			"COALESCE(message_deliveries.meta ->> 'source_group_name', '') AS source_group_name",
			"COALESCE(message_deliveries.meta ->> 'source_rule_type', '') AS source_rule_type",
			"COALESCE(message_deliveries.meta ->> 'source_rule_label', '') AS source_rule_label",
			"COALESCE(message_deliveries.meta ->> 'source_target_id', '') AS source_target_id",
			"COALESCE(message_deliveries.meta ->> 'source_target_type', '') AS source_target_type",
			"COALESCE(message_deliveries.meta ->> 'source_target_value', '') AS source_target_value",
		}, ", ")).
		Joins("LEFT JOIN users ON users.id = message_deliveries.recipient_user_id").
		Joins("LEFT JOIN collaboration_workspaces ON collaboration_workspaces.id = message_deliveries.recipient_collaboration_workspace_id").
		Where("message_deliveries.message_id = ? AND message_deliveries.deleted_at IS NULL", detail.ID).
		Order("message_deliveries.created_at ASC").
		Scan(&rows).Error
	if err != nil {
		return dispatchRecordDetail{}, err
	}

	detail.Deliveries = make([]dispatchRecordDeliveryItem, 0, len(rows))
	for _, row := range rows {
		item := dispatchRecordDeliveryItem{
			ID:                                row.ID,
			RecipientUserID:                   row.RecipientUserID,
			RecipientName:                     row.RecipientName,
			RecipientCollaborationWorkspaceID: row.RecipientCollaborationWorkspaceID,
			RecipientCollaborationWorkspace:   row.RecipientCollaborationWorkspace,
			DeliveryStatus:                    row.DeliveryStatus,
			TodoStatus:                        row.TodoStatus,
			ReadAt:                            row.ReadAt,
			DoneAt:                            row.DoneAt,
			LastActionAt:                      row.LastActionAt,
			SourceGroupName:                   row.SourceGroupName,
			SourceRuleType:                    row.SourceRuleType,
			SourceRuleLabel:                   row.SourceRuleLabel,
			SourceTargetType:                  row.SourceTargetType,
			SourceTargetValue:                 row.SourceTargetValue,
		}
		if sourceGroupID := strings.TrimSpace(row.SourceGroupID); sourceGroupID != "" {
			if parsed, parseErr := uuid.Parse(sourceGroupID); parseErr == nil {
				item.SourceGroupID = &parsed
			}
		}
		if sourceTargetID := strings.TrimSpace(row.SourceTargetID); sourceTargetID != "" {
			if parsed, parseErr := uuid.Parse(sourceTargetID); parseErr == nil {
				item.SourceTargetID = &parsed
			}
		}
		detail.Deliveries = append(detail.Deliveries, item)
	}

	return detail, nil
}

func (s *messageService) baseInboxQuery(userID uuid.UUID) *gorm.DB {
	return s.db.Table("message_deliveries").
		Joins("JOIN messages ON messages.id = message_deliveries.message_id").
		Where("message_deliveries.recipient_user_id = ?", userID).
		Where("message_deliveries.deleted_at IS NULL").
		Where("messages.deleted_at IS NULL")
}

func (s *messageService) templateScopeQuery(collaborationWorkspaceID *uuid.UUID) *gorm.DB {
	tx := s.db.Model(&models.MessageTemplate{}).Where("message_templates.deleted_at IS NULL")
	if collaborationWorkspaceID != nil {
		return tx.Where(
			"message_templates.owner_scope = ? OR (message_templates.owner_scope = ? AND message_templates.owner_collaboration_workspace_id = ?)",
			"personal",
			"collaboration",
			*collaborationWorkspaceID,
		)
	}
	return tx.Where("message_templates.owner_scope = ?", "personal")
}

func (s *messageService) dispatchRecordBaseQuery(collaborationWorkspaceID *uuid.UUID) *gorm.DB {
	tx := s.db.Model(&models.Message{}).Where("messages.deleted_at IS NULL")
	if collaborationWorkspaceID != nil {
		return tx.Where("messages.scope_type = ? AND messages.scope_id = ?", "collaboration", *collaborationWorkspaceID)
	}
	return tx.Where("messages.scope_type = ?", "personal")
}

func (s *messageService) loadDispatchRecordSummary(collaborationWorkspaceID *uuid.UUID) (dispatchRecordSummary, error) {
	summary := dispatchRecordSummary{}
	if err := s.dispatchRecordBaseQuery(collaborationWorkspaceID).Count(&summary.TotalMessages).Error; err != nil {
		return dispatchRecordSummary{}, err
	}
	if err := s.dispatchRecordBaseQuery(collaborationWorkspaceID).Where("messages.message_type = ?", "todo").Count(&summary.TodoMessages).Error; err != nil {
		return dispatchRecordSummary{}, err
	}

	var deliveryAgg struct {
		Total int64 `gorm:"column:total"`
		Read  int64 `gorm:"column:read"`
	}
	err := s.db.Table("message_deliveries").
		Select(
			"COUNT(message_deliveries.id) AS total",
			"SUM(CASE WHEN message_deliveries.delivery_status = 'read' THEN 1 ELSE 0 END) AS read",
		).
		Joins("JOIN messages ON messages.id = message_deliveries.message_id").
		Where("message_deliveries.deleted_at IS NULL").
		Where("messages.deleted_at IS NULL").
		Scopes(func(tx *gorm.DB) *gorm.DB {
			if collaborationWorkspaceID != nil {
				return tx.Where("messages.scope_type = ? AND messages.scope_id = ?", "collaboration", *collaborationWorkspaceID)
			}
			return tx.Where("messages.scope_type = ?", "personal")
		}).
		Scan(&deliveryAgg).Error
	if err != nil {
		return dispatchRecordSummary{}, err
	}
	summary.TotalDeliveries = deliveryAgg.Total
	summary.ReadDeliveries = deliveryAgg.Read
	return summary, nil
}

func (s *messageService) filteredInboxQuery(userID uuid.UUID, query inboxQuery) *gorm.DB {
	tx := s.baseInboxQuery(userID).
		Where("messages.status = ?", "published")
	now := time.Now()
	tx = tx.Where("messages.expired_at IS NULL OR messages.expired_at > ?", now)
	if normalized := normalizeBoxType(query.BoxType); normalized != "" {
		tx = tx.Where("message_deliveries.box_type = ?", normalized)
	}
	if query.UnreadOnly {
		tx = tx.Where("message_deliveries.delivery_status = ?", "unread")
	}
	return tx
}

func normalizeBoxType(value string) string {
	switch strings.TrimSpace(value) {
	case "notice", "message", "todo":
		return strings.TrimSpace(value)
	default:
		return ""
	}
}

func inboxSelectColumns() string {
	return strings.Join([]string{
		"message_deliveries.id AS id",
		"message_deliveries.message_id AS message_id",
		"message_deliveries.box_type AS box_type",
		"message_deliveries.delivery_status AS delivery_status",
		"message_deliveries.todo_status AS todo_status",
		"message_deliveries.read_at AS read_at",
		"message_deliveries.done_at AS done_at",
		"message_deliveries.last_action_at AS last_action_at",
		"message_deliveries.recipient_collaboration_workspace_id AS recipient_collaboration_workspace_id",
		"messages.title AS title",
		"messages.summary AS summary",
		"messages.content AS content",
		"messages.priority AS priority",
		"messages.action_type AS action_type",
		"messages.action_target AS action_target",
		"messages.message_type AS message_type",
		"messages.biz_type AS biz_type",
		"messages.scope_type AS scope_type",
		"messages.scope_id AS scope_id",
		"messages.sender_type AS sender_type",
		"messages.sender_name_snapshot AS sender_name_snapshot",
		"messages.sender_avatar_snapshot AS sender_avatar_snapshot",
		"messages.sender_service_key AS sender_service_key",
		"messages.audience_type AS audience_type",
		"messages.audience_scope AS audience_scope",
		"messages.target_collaboration_workspace_id AS target_collaboration_workspace_id",
		"messages.published_at AS published_at",
		"messages.expired_at AS expired_at",
		"messages.created_at AS created_at",
		"messages.meta AS meta",
	}, ", ")
}

func (s *messageService) listSenderOptions(collaborationWorkspaceID *uuid.UUID) ([]dispatchSenderOption, error) {
	items, err := s.ensureSenderOptions(collaborationWorkspaceID)
	if err != nil {
		return nil, err
	}
	result := make([]dispatchSenderOption, 0, len(items))
	for _, item := range items {
		result = append(result, dispatchSenderOption{
			ID:          item.ID,
			Name:        item.Name,
			Description: item.Description,
			AvatarURL:   item.AvatarURL,
			IsDefault:   item.IsDefault,
		})
	}
	return result, nil
}

func (s *messageService) ensureSenderOptions(collaborationWorkspaceID *uuid.UUID) ([]models.MessageSender, error) {
	scopeType := "personal"
	defaultName := "个人空间"
	defaultDescription := "个人空间默认发送人"
	var scopeID *uuid.UUID
	if collaborationWorkspaceID != nil {
		scopeType = "collaboration"
		scopeID = collaborationWorkspaceID
		defaultName = "协作空间"
		defaultDescription = "协作空间默认发送人"
	}

	if err := s.db.Transaction(func(tx *gorm.DB) error {
		query := tx.Model(&models.MessageSender{}).Where("deleted_at IS NULL AND scope_type = ?", scopeType)
		if scopeID != nil {
			query = query.Where("scope_id = ?", *scopeID)
		} else {
			query = query.Where("scope_id IS NULL")
		}
		var count int64
		if err := query.Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			item := models.MessageSender{
				ScopeType:   scopeType,
				ScopeID:     scopeID,
				Name:        defaultName,
				Description: defaultDescription,
				IsDefault:   true,
				Status:      "normal",
				Meta:        models.MetaJSON{},
			}
			query := tx.Where("scope_type = ? AND name = ? AND deleted_at IS NULL", scopeType, defaultName)
			if scopeID != nil {
				query = query.Where("scope_id = ?", *scopeID)
			} else {
				query = query.Where("scope_id IS NULL")
			}
			if err := query.FirstOrCreate(&item).Error; err != nil {
				return err
			}
			if item.Status != "normal" || !item.IsDefault {
				if err := tx.Model(&models.MessageSender{}).Where("id = ?", item.ID).Updates(map[string]interface{}{
					"description": defaultDescription,
					"status":      "normal",
					"is_default":  true,
				}).Error; err != nil {
					return err
				}
			}
			return nil
		}

		var activeDefaultCount int64
		defaultQuery := tx.Model(&models.MessageSender{}).
			Where("deleted_at IS NULL AND scope_type = ? AND status = ? AND is_default = ?", scopeType, "normal", true)
		if scopeID != nil {
			defaultQuery = defaultQuery.Where("scope_id = ?", *scopeID)
		} else {
			defaultQuery = defaultQuery.Where("scope_id IS NULL")
		}
		if err := defaultQuery.Count(&activeDefaultCount).Error; err != nil {
			return err
		}
		if activeDefaultCount > 0 {
			return nil
		}

		var fallback models.MessageSender
		fallbackQuery := tx.Model(&models.MessageSender{}).
			Where("deleted_at IS NULL AND scope_type = ? AND status = ?", scopeType, "normal")
		if scopeID != nil {
			fallbackQuery = fallbackQuery.Where("scope_id = ?", *scopeID)
		} else {
			fallbackQuery = fallbackQuery.Where("scope_id IS NULL")
		}
		fallbackErr := fallbackQuery.Order("created_at ASC").First(&fallback).Error
		if fallbackErr == nil {
			return tx.Model(&models.MessageSender{}).Where("id = ?", fallback.ID).Update("is_default", true).Error
		}
		if errors.Is(fallbackErr, gorm.ErrRecordNotFound) {
			item := models.MessageSender{
				ScopeType:   scopeType,
				ScopeID:     scopeID,
				Name:        defaultName,
				Description: defaultDescription,
				IsDefault:   true,
				Status:      "normal",
				Meta:        models.MetaJSON{},
			}
			query := tx.Where("scope_type = ? AND name = ? AND deleted_at IS NULL", scopeType, defaultName)
			if scopeID != nil {
				query = query.Where("scope_id = ?", *scopeID)
			} else {
				query = query.Where("scope_id IS NULL")
			}
			if err := query.FirstOrCreate(&item).Error; err != nil {
				return err
			}
			return tx.Model(&models.MessageSender{}).Where("id = ?", item.ID).Updates(map[string]interface{}{
				"description": defaultDescription,
				"status":      "normal",
				"is_default":  true,
			}).Error
		}
		return fallbackErr
	}); err != nil {
		return nil, err
	}

	query := s.db.Model(&models.MessageSender{}).Where("deleted_at IS NULL AND scope_type = ?", scopeType)
	if scopeID != nil {
		query = query.Where("scope_id = ?", *scopeID)
	} else {
		query = query.Where("scope_id IS NULL")
	}

	var items []models.MessageSender
	if err := query.Order("is_default DESC").Order("created_at ASC").Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (s *messageService) resolveDispatchSender(collaborationWorkspaceID *uuid.UUID, senderID string) (models.MessageSender, error) {
	items, err := s.ensureSenderOptions(collaborationWorkspaceID)
	if err != nil {
		return models.MessageSender{}, err
	}
	if trimmedID := strings.TrimSpace(senderID); trimmedID != "" {
		id, parseErr := uuid.Parse(trimmedID)
		if parseErr != nil {
			return models.MessageSender{}, errors.New("发送人标识无效")
		}
		for _, item := range items {
			if item.ID == id && item.Status == "normal" {
				return item, nil
			}
		}
		return models.MessageSender{}, errors.New("发送人不存在或当前上下文不可用")
	}
	for _, item := range items {
		if item.IsDefault && item.Status == "normal" {
			return item, nil
		}
	}
	for _, item := range items {
		if item.Status == "normal" {
			return item, nil
		}
	}
	return models.MessageSender{}, errors.New("当前作用域没有可用发送人")
}

func (s *messageService) loadEditableTemplate(tx *gorm.DB, templateID uuid.UUID, collaborationWorkspaceID *uuid.UUID) (models.MessageTemplate, error) {
	query := tx.Model(&models.MessageTemplate{}).Where("id = ? AND deleted_at IS NULL", templateID)
	if collaborationWorkspaceID != nil {
		query = query.Where("owner_scope = ? AND owner_collaboration_workspace_id = ?", "collaboration", *collaborationWorkspaceID)
	} else {
		query = query.Where("owner_scope = ?", "personal")
	}

	var template models.MessageTemplate
	if err := query.First(&template).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.MessageTemplate{}, errors.New("消息模板不存在或当前上下文不可编辑")
		}
		return models.MessageTemplate{}, err
	}
	return template, nil
}

func (s *messageService) resolveTemplate(templateID, templateKey string, collaborationWorkspaceID *uuid.UUID) (*models.MessageTemplate, error) {
	if templateID == "" && templateKey == "" {
		return nil, nil
	}
	query := s.db.Model(&models.MessageTemplate{}).Where("status = ?", "normal")
	if collaborationWorkspaceID != nil {
		query = query.Where(
			"owner_scope = ? OR (owner_scope = ? AND owner_collaboration_workspace_id = ?)",
			"personal",
			"collaboration",
			*collaborationWorkspaceID,
		)
	} else {
		query = query.Where("owner_scope = ?", "personal")
	}

	var template models.MessageTemplate
	var err error
	switch {
	case templateID != "":
		id, parseErr := uuid.Parse(templateID)
		if parseErr != nil {
			return nil, errors.New("模板标识无效")
		}
		err = query.Where("id = ?", id).First(&template).Error
	default:
		err = query.Where("template_key = ?", templateKey).First(&template).Error
	}
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("消息模板不存在或当前上下文不可用")
		}
		return nil, err
	}
	return &template, nil
}

func (s *messageService) normalizeTargetCollaborationWorkspaces(audienceType string, collaborationWorkspaceID *uuid.UUID, targets []uuid.UUID) ([]uuid.UUID, error) {
	switch audienceType {
	case "all_users":
		if collaborationWorkspaceID != nil {
			return nil, errors.New("协作空间上下文不支持给所有用户发送")
		}
		return nil, nil
	case "collaboration_workspace_admins", "collaboration_workspace_users":
		if collaborationWorkspaceID != nil {
			if len(targets) == 0 {
				return []uuid.UUID{*collaborationWorkspaceID}, nil
			}
			if len(targets) != 1 || targets[0] != *collaborationWorkspaceID {
				return nil, errors.New("协作空间上下文只能给当前协作空间发送")
			}
			return targets, nil
		}
		if len(targets) == 0 {
			return nil, errors.New("请选择目标协作空间")
		}
		return targets, nil
	case "specified_users", "recipient_group", "role", "feature_package":
		return nil, nil
	default:
		return nil, errors.New("不支持的发送对象")
	}
}

func (s *messageService) resolveRecipients(
	audienceType string,
	collaborationWorkspaceID *uuid.UUID,
	targetCollaborationWorkspaceIDs []uuid.UUID,
	targetUserIDs []uuid.UUID,
	targetGroupIDs []uuid.UUID,
) ([]dispatchRecipient, error) {
	switch audienceType {
	case "all_users":
		return s.loadAllUsers()
	case "collaboration_workspace_admins":
		return s.loadCollaborationWorkspaceRecipients(targetCollaborationWorkspaceIDs, true)
	case "collaboration_workspace_users":
		return s.loadCollaborationWorkspaceRecipients(targetCollaborationWorkspaceIDs, false)
	case "specified_users":
		return s.loadSpecifiedUsers(targetUserIDs, collaborationWorkspaceID)
	case "recipient_group":
		return s.loadGroupRecipients(targetGroupIDs, collaborationWorkspaceID)
	case "role":
		return s.loadRoleRecipients(targetGroupIDs, collaborationWorkspaceID)
	case "feature_package":
		return s.loadFeaturePackageRecipients(targetGroupIDs, collaborationWorkspaceID)
	default:
		return nil, errors.New("不支持的发送对象")
	}
}

func (s *messageService) normalizeAudienceTargets(
	audienceType string,
	collaborationWorkspaceID *uuid.UUID,
	targetUserIDs []uuid.UUID,
	targetGroupIDs []uuid.UUID,
) ([]uuid.UUID, []uuid.UUID, error) {
	switch audienceType {
	case "specified_users":
		if len(targetUserIDs) == 0 {
			return nil, nil, errors.New("请选择目标用户")
		}
		return targetUserIDs, nil, nil
	case "recipient_group":
		if len(targetGroupIDs) == 0 {
			return nil, nil, errors.New("请选择接收组")
		}
		return nil, targetGroupIDs, nil
	case "role", "feature_package":
		if len(targetGroupIDs) == 0 {
			return nil, nil, errors.New("请至少选择一个包含规则的接收组")
		}
		return nil, targetGroupIDs, nil
	case "all_users", "collaboration_workspace_admins", "collaboration_workspace_users":
		return nil, nil, nil
	default:
		if collaborationWorkspaceID != nil {
			return nil, nil, errors.New("协作空间上下文不支持当前发送对象")
		}
		return nil, nil, errors.New("不支持的发送对象")
	}
}

func (s *messageService) loadAllUsers() ([]dispatchRecipient, error) {
	var users []models.User
	if err := s.db.Select("id", "username").Where("status = ?", "active").Find(&users).Error; err != nil {
		return nil, err
	}
	result := make([]dispatchRecipient, 0, len(users))
	for _, item := range users {
		result = append(result, dispatchRecipient{
			UserID:            item.ID,
			Username:          item.Username,
			SourceRuleType:    "all_users",
			SourceRuleLabel:   "所有用户",
			SourceTargetType:  "all_users",
			SourceTargetValue: "all_users",
		})
	}
	return result, nil
}

func (s *messageService) loadCollaborationWorkspaceRecipients(targetCollaborationWorkspaceIDs []uuid.UUID, adminOnly bool) ([]dispatchRecipient, error) {
	if len(targetCollaborationWorkspaceIDs) == 0 {
		return nil, errors.New("请选择目标协作空间")
	}
	type collaborationWorkspaceRecipientRow struct {
		UserID                   uuid.UUID `gorm:"column:user_id"`
		CollaborationWorkspaceID uuid.UUID `gorm:"column:collaboration_workspace_id"`
		Username                 string    `gorm:"column:username"`
		Nickname                 string    `gorm:"column:nickname"`
		RoleCode                 string    `gorm:"column:role_code"`
		Status                   string    `gorm:"column:status"`
	}
	query := s.db.Table("collaboration_workspace_members").
		Select("collaboration_workspace_members.user_id AS user_id", "collaboration_workspace_members.collaboration_workspace_id AS collaboration_workspace_id", "users.username AS username", "users.nickname AS nickname", "collaboration_workspace_members.role_code AS role_code", "collaboration_workspace_members.status AS status").
		Joins("JOIN users ON users.id = collaboration_workspace_members.user_id").
		Where("collaboration_workspace_members.collaboration_workspace_id IN ?", targetCollaborationWorkspaceIDs).
		Where("collaboration_workspace_members.status = ?", "active").
		Where("users.status = ?", "active")
	if adminOnly {
		query = query.Where("collaboration_workspace_members.role_code = ?", "collaboration_workspace_admin")
	}

	var rows []collaborationWorkspaceRecipientRow
	if err := query.Order("collaboration_workspace_members.created_at ASC").Scan(&rows).Error; err != nil {
		return nil, err
	}

	result := make([]dispatchRecipient, 0, len(rows))
	seen := make(map[uuid.UUID]struct{}, len(rows))
	for _, row := range rows {
		if _, ok := seen[row.UserID]; ok {
			continue
		}
		seen[row.UserID] = struct{}{}
		collaborationWorkspaceID := row.CollaborationWorkspaceID
		username := strings.TrimSpace(row.Nickname)
		if username == "" {
			username = strings.TrimSpace(row.Username)
		}
		result = append(result, dispatchRecipient{
			UserID:                   row.UserID,
			CollaborationWorkspaceID: &collaborationWorkspaceID,
			Username:                 username,
			SourceRuleType:           map[bool]string{true: "collaboration_workspace_admins", false: "collaboration_workspace_users"}[adminOnly],
			SourceRuleLabel:          membershipRecipientRuleLabel(map[bool]string{true: "协作空间管理员", false: "协作空间成员"}[adminOnly]),
			SourceTargetType:         map[bool]string{true: "collaboration_workspace_admins", false: "collaboration_workspace_users"}[adminOnly],
			SourceTargetValue:        collaborationWorkspaceID.String(),
		})
	}
	return result, nil
}

func (s *messageService) loadSpecifiedUsers(targetUserIDs []uuid.UUID, collaborationWorkspaceID *uuid.UUID) ([]dispatchRecipient, error) {
	if len(targetUserIDs) == 0 {
		return nil, errors.New("请选择目标用户")
	}
	if collaborationWorkspaceID != nil {
		return s.loadSpecifiedCollaborationWorkspaceUsers(*collaborationWorkspaceID, targetUserIDs)
	}
	var rows []struct {
		ID       uuid.UUID `gorm:"column:id"`
		Username string    `gorm:"column:username"`
		Nickname string    `gorm:"column:nickname"`
	}
	if err := s.db.Model(&models.User{}).
		Select("id", "username", "nickname").
		Where("id IN ?", targetUserIDs).
		Where("status = ?", "active").
		Order("created_at ASC").
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, errors.New("当前没有可投递的目标用户")
	}
	result := make([]dispatchRecipient, 0, len(rows))
	for _, row := range rows {
		name := strings.TrimSpace(row.Nickname)
		if name == "" {
			name = strings.TrimSpace(row.Username)
		}
		result = append(result, dispatchRecipient{
			UserID:            row.ID,
			Username:          name,
			SourceRuleType:    "specified_users",
			SourceRuleLabel:   "指定用户",
			SourceTargetID:    &row.ID,
			SourceTargetType:  "user",
			SourceTargetValue: row.ID.String(),
		})
	}
	return result, nil
}

func (s *messageService) loadSpecifiedCollaborationWorkspaceUsers(collaborationWorkspaceID uuid.UUID, targetUserIDs []uuid.UUID) ([]dispatchRecipient, error) {
	type row struct {
		UserID                   uuid.UUID `gorm:"column:user_id"`
		CollaborationWorkspaceID uuid.UUID `gorm:"column:collaboration_workspace_id"`
		Username                 string    `gorm:"column:username"`
		Nickname                 string    `gorm:"column:nickname"`
	}
	var rows []row
	if err := s.db.Table("collaboration_workspace_members").
		Select("collaboration_workspace_members.user_id AS user_id", "collaboration_workspace_members.collaboration_workspace_id AS collaboration_workspace_id", "users.username AS username", "users.nickname AS nickname").
		Joins("JOIN users ON users.id = collaboration_workspace_members.user_id").
		Where("collaboration_workspace_members.collaboration_workspace_id = ?", collaborationWorkspaceID).
		Where("collaboration_workspace_members.user_id IN ?", targetUserIDs).
		Where("collaboration_workspace_members.status = ?", "active").
		Where("users.status = ?", "active").
		Order("collaboration_workspace_members.created_at ASC").
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, errors.New("当前没有可投递的目标用户")
	}
	result := make([]dispatchRecipient, 0, len(rows))
	for _, row := range rows {
		name := strings.TrimSpace(row.Nickname)
		if name == "" {
			name = strings.TrimSpace(row.Username)
		}
		collaborationWorkspaceID := row.CollaborationWorkspaceID
		result = append(result, dispatchRecipient{
			UserID:                   row.UserID,
			CollaborationWorkspaceID: &collaborationWorkspaceID,
			Username:                 name,
			SourceRuleType:           "specified_users",
			SourceRuleLabel:          "指定用户",
			SourceTargetID:           &row.UserID,
			SourceTargetType:         "user",
			SourceTargetValue:        row.UserID.String(),
		})
	}
	return result, nil
}

func (s *messageService) loadRecipientsByRoleCode(roleCode string, collaborationWorkspaceID *uuid.UUID) ([]dispatchRecipient, error) {
	roleCode = strings.TrimSpace(roleCode)
	if roleCode == "" {
		return nil, errors.New("角色规则不能为空")
	}
	if collaborationWorkspaceID != nil {
		return s.loadCollaborationWorkspaceRecipientsByRoleCode(*collaborationWorkspaceID, roleCode)
	}
	return s.loadPlatformRecipientsByRoleCode(roleCode)
}

func (s *messageService) loadPlatformRecipientsByRoleCode(roleCode string) ([]dispatchRecipient, error) {
	seen := make(map[uuid.UUID]dispatchRecipient)
	var legacyUserIDs []uuid.UUID
	if err := s.db.Table("user_roles").
		Select("user_roles.user_id").
		Joins("JOIN roles ON roles.id = user_roles.role_id").
		Where("user_roles.collaboration_workspace_id IS NULL").
		Where("roles.code = ? AND roles.collaboration_workspace_id IS NULL AND roles.status = ? AND roles.deleted_at IS NULL", roleCode, "normal").
		Distinct("user_roles.user_id").
		Pluck("user_roles.user_id", &legacyUserIDs).Error; err != nil {
		return nil, err
	}
	legacyRows, err := s.loadActiveRecipientUserRows(legacyUserIDs)
	if err != nil {
		return nil, err
	}
	for _, row := range legacyRows {
		seen[row.UserID] = dispatchRecipient{
			UserID:            row.UserID,
			Username:          displayRecipientName(row.Username, row.Nickname),
			SourceRuleType:    "role",
			SourceRuleLabel:   roleRecipientRuleLabel(roleCode, "legacy_user_role"),
			SourceTargetType:  "role",
			SourceTargetValue: roleCode,
		}
	}
	workspaceUserIDs, err := workspacerolebinding.ListPlatformUserIDsByRoleCodes(s.db, []string{roleCode}, true)
	if err != nil {
		return nil, err
	}
	workspaceRows, err := s.loadActiveRecipientUserRows(workspaceUserIDs)
	if err != nil {
		return nil, err
	}
	for _, row := range workspaceRows {
		seen[row.UserID] = dispatchRecipient{
			UserID:            row.UserID,
			Username:          displayRecipientName(row.Username, row.Nickname),
			SourceRuleType:    "role",
			SourceRuleLabel:   roleRecipientRuleLabel(roleCode, "workspace_role_binding"),
			SourceTargetType:  "role",
			SourceTargetValue: roleCode,
		}
	}
	result := make([]dispatchRecipient, 0, len(seen))
	for _, item := range seen {
		result = append(result, item)
	}
	return result, nil
}

func (s *messageService) loadCollaborationWorkspaceRecipientsByRoleCode(collaborationWorkspaceID uuid.UUID, roleCode string) ([]dispatchRecipient, error) {
	seen := make(map[uuid.UUID]dispatchRecipient)
	var memberRows []dispatchRecipientUserRow
	if err := s.db.Table("collaboration_workspace_members").
		Select("users.id AS user_id", "users.username AS username", "users.nickname AS nickname").
		Joins("JOIN users ON users.id = collaboration_workspace_members.user_id").
		Where("collaboration_workspace_members.collaboration_workspace_id = ? AND collaboration_workspace_members.status = ? AND collaboration_workspace_members.role_code = ?", collaborationWorkspaceID, "active", roleCode).
		Where("users.status = ? AND users.deleted_at IS NULL", "active").
		Scan(&memberRows).Error; err != nil {
		return nil, err
	}
	for _, row := range memberRows {
		collaborationWorkspaceRef := collaborationWorkspaceID
		seen[row.UserID] = dispatchRecipient{
			UserID:                   row.UserID,
			CollaborationWorkspaceID: &collaborationWorkspaceRef,
			Username:                 displayRecipientName(row.Username, row.Nickname),
			SourceRuleType:           "role",
			SourceRuleLabel:          roleRecipientRuleLabel(roleCode, "membership_identity"),
			SourceTargetType:         "role",
			SourceTargetValue:        roleCode,
		}
	}

	var roleIDs []uuid.UUID
	if err := s.db.Model(&models.Role{}).
		Where("(collaboration_workspace_id = ? OR collaboration_workspace_id IS NULL) AND code = ? AND status = ? AND deleted_at IS NULL", collaborationWorkspaceID, roleCode, "normal").
		Pluck("id", &roleIDs).Error; err != nil {
		return nil, err
	}
	if len(roleIDs) > 0 {
		var customRows []dispatchRecipientUserRow
		if err := s.db.Table("user_roles").
			Select("users.id AS user_id", "users.username AS username", "users.nickname AS nickname").
			Joins("JOIN users ON users.id = user_roles.user_id").
			Where("user_roles.collaboration_workspace_id = ? AND user_roles.role_id IN ?", collaborationWorkspaceID, roleIDs).
			Where("users.status = ? AND users.deleted_at IS NULL", "active").
			Scan(&customRows).Error; err != nil {
			return nil, err
		}
		for _, row := range customRows {
			collaborationWorkspaceRef := collaborationWorkspaceID
			seen[row.UserID] = dispatchRecipient{
				UserID:                   row.UserID,
				CollaborationWorkspaceID: &collaborationWorkspaceRef,
				Username:                 displayRecipientName(row.Username, row.Nickname),
				SourceRuleType:           "role",
				SourceRuleLabel:          roleRecipientRuleLabel(roleCode, "legacy_user_role"),
				SourceTargetType:         "role",
				SourceTargetValue:        roleCode,
			}
		}
	}
	workspaceUserIDs, err := workspacerolebinding.ListUserIDsByCollaborationWorkspaceRoleCodes(s.db, collaborationWorkspaceID, []string{roleCode}, true)
	if err != nil {
		return nil, err
	}
	workspaceRows, err := s.loadActiveRecipientUserRows(workspaceUserIDs)
	if err != nil {
		return nil, err
	}
	for _, row := range workspaceRows {
		collaborationWorkspaceRef := collaborationWorkspaceID
		seen[row.UserID] = dispatchRecipient{
			UserID:                   row.UserID,
			CollaborationWorkspaceID: &collaborationWorkspaceRef,
			Username:                 displayRecipientName(row.Username, row.Nickname),
			SourceRuleType:           "role",
			SourceRuleLabel:          roleRecipientRuleLabel(roleCode, "workspace_role_binding"),
			SourceTargetType:         "role",
			SourceTargetValue:        roleCode,
		}
	}

	result := make([]dispatchRecipient, 0, len(seen))
	for _, item := range seen {
		result = append(result, item)
	}
	return result, nil
}

func (s *messageService) loadRecipientsByPackageKey(packageKey string, collaborationWorkspaceID *uuid.UUID) ([]dispatchRecipient, error) {
	packageKey = strings.TrimSpace(packageKey)
	if packageKey == "" {
		return nil, errors.New("功能包规则不能为空")
	}
	if collaborationWorkspaceID != nil {
		return s.loadCollaborationWorkspaceRecipientsByPackageKey(*collaborationWorkspaceID, packageKey)
	}
	return s.loadPlatformRecipientsByPackageKey(packageKey)
}

func (s *messageService) loadPlatformRecipientsByPackageKey(packageKey string) ([]dispatchRecipient, error) {
	var pkg models.FeaturePackage
	if err := s.db.Model(&models.FeaturePackage{}).
		Where("package_key = ? AND context_type = ? AND status = ? AND deleted_at IS NULL", packageKey, "personal", "normal").
		First(&pkg).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	type row struct {
		UserID   uuid.UUID `gorm:"column:user_id"`
		Username string    `gorm:"column:username"`
		Nickname string    `gorm:"column:nickname"`
	}
	var rows []row
	if err := s.db.Table("personal_workspace_access_snapshots").
		Select("users.id AS user_id", "users.username AS username", "users.nickname AS nickname").
		Joins("JOIN users ON users.id = personal_workspace_access_snapshots.user_id").
		Where("personal_workspace_access_snapshots.expanded_package_ids @> ?", fmt.Sprintf("[\"%s\"]", pkg.ID.String())).
		Where("users.status = ? AND users.deleted_at IS NULL", "active").
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	result := make([]dispatchRecipient, 0, len(rows))
	for _, row := range rows {
		name := strings.TrimSpace(row.Nickname)
		if name == "" {
			name = strings.TrimSpace(row.Username)
		}
		result = append(result, dispatchRecipient{
			UserID:            row.UserID,
			Username:          name,
			SourceRuleType:    "feature_package",
			SourceRuleLabel:   packageKey,
			SourceTargetType:  "feature_package",
			SourceTargetValue: packageKey,
		})
	}
	return result, nil
}

func (s *messageService) loadCollaborationWorkspaceRecipientsByPackageKey(collaborationWorkspaceID uuid.UUID, packageKey string) ([]dispatchRecipient, error) {
	var pkg models.FeaturePackage
	if err := s.db.Model(&models.FeaturePackage{}).
		Where("package_key = ? AND context_type IN ? AND status = ? AND deleted_at IS NULL", packageKey, []string{"collaboration"}, "normal").
		First(&pkg).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	var roleIDs []uuid.UUID
	if err := s.db.Table("collaboration_workspace_role_access_snapshots").
		Where("collaboration_workspace_id = ? AND expanded_package_ids @> ?", collaborationWorkspaceID, fmt.Sprintf("[\"%s\"]", pkg.ID.String())).
		Pluck("role_id", &roleIDs).Error; err != nil {
		return nil, err
	}
	if len(roleIDs) == 0 {
		return nil, nil
	}
	var roles []models.Role
	if err := s.db.Model(&models.Role{}).Where("id IN ?", roleIDs).Find(&roles).Error; err != nil {
		return nil, err
	}
	roleCodeSet := make(map[string]struct{})
	for _, role := range roles {
		roleCodeSet[role.Code] = struct{}{}
	}
	seen := make(map[uuid.UUID]dispatchRecipient)
	if len(roleCodeSet) > 0 {
		roleCodes := make([]string, 0, len(roleCodeSet))
		for code := range roleCodeSet {
			roleCodes = append(roleCodes, code)
		}
		var identityRows []dispatchRecipientUserRow
		if err := s.db.Table("collaboration_workspace_members").
			Select("users.id AS user_id", "users.username AS username", "users.nickname AS nickname").
			Joins("JOIN users ON users.id = collaboration_workspace_members.user_id").
			Where("collaboration_workspace_members.collaboration_workspace_id = ? AND collaboration_workspace_members.status = ? AND collaboration_workspace_members.role_code IN ?", collaborationWorkspaceID, "active", roleCodes).
			Where("users.status = ? AND users.deleted_at IS NULL", "active").
			Scan(&identityRows).Error; err != nil {
			return nil, err
		}
		for _, row := range identityRows {
			collaborationWorkspaceRef := collaborationWorkspaceID
			seen[row.UserID] = dispatchRecipient{
				UserID:                   row.UserID,
				CollaborationWorkspaceID: &collaborationWorkspaceRef,
				Username:                 displayRecipientName(row.Username, row.Nickname),
				SourceRuleType:           "feature_package",
				SourceRuleLabel:          packageRecipientRuleLabel(packageKey, "membership_identity"),
				SourceTargetType:         "feature_package",
				SourceTargetValue:        packageKey,
			}
		}
	}
	var customRows []dispatchRecipientUserRow
	if err := s.db.Table("user_roles").
		Select("users.id AS user_id", "users.username AS username", "users.nickname AS nickname").
		Joins("JOIN users ON users.id = user_roles.user_id").
		Where("user_roles.collaboration_workspace_id = ? AND user_roles.role_id IN ?", collaborationWorkspaceID, roleIDs).
		Where("users.status = ? AND users.deleted_at IS NULL", "active").
		Scan(&customRows).Error; err != nil {
		return nil, err
	}
	for _, row := range customRows {
		collaborationWorkspaceRef := collaborationWorkspaceID
		seen[row.UserID] = dispatchRecipient{
			UserID:                   row.UserID,
			CollaborationWorkspaceID: &collaborationWorkspaceRef,
			Username:                 displayRecipientName(row.Username, row.Nickname),
			SourceRuleType:           "feature_package",
			SourceRuleLabel:          packageRecipientRuleLabel(packageKey, "legacy_user_role"),
			SourceTargetType:         "feature_package",
			SourceTargetValue:        packageKey,
		}
	}
	workspaceUserIDs, err := workspacerolebinding.ListUserIDsByCollaborationWorkspaceRoleIDs(s.db, collaborationWorkspaceID, roleIDs, true)
	if err != nil {
		return nil, err
	}
	workspaceRows, err := s.loadActiveRecipientUserRows(workspaceUserIDs)
	if err != nil {
		return nil, err
	}
	for _, row := range workspaceRows {
		collaborationWorkspaceRef := collaborationWorkspaceID
		seen[row.UserID] = dispatchRecipient{
			UserID:                   row.UserID,
			CollaborationWorkspaceID: &collaborationWorkspaceRef,
			Username:                 displayRecipientName(row.Username, row.Nickname),
			SourceRuleType:           "feature_package",
			SourceRuleLabel:          packageRecipientRuleLabel(packageKey, "workspace_role_binding"),
			SourceTargetType:         "feature_package",
			SourceTargetValue:        packageKey,
		}
	}
	result := make([]dispatchRecipient, 0, len(seen))
	for _, item := range seen {
		result = append(result, item)
	}
	return result, nil
}

func (s *messageService) loadGroupRecipients(groupIDs []uuid.UUID, collaborationWorkspaceID *uuid.UUID) ([]dispatchRecipient, error) {
	groups, err := s.loadEditableRecipientGroups(groupIDs, collaborationWorkspaceID)
	if err != nil {
		return nil, err
	}
	recipients := make([]dispatchRecipient, 0)
	seen := make(map[uuid.UUID]dispatchRecipient)
	for _, group := range groups {
		targets, targetErr := s.loadRecipientGroupTargets(group.ID)
		if targetErr != nil {
			return nil, targetErr
		}
		for _, target := range targets {
			var groupRecipients []dispatchRecipient
			switch target.TargetType {
			case "user":
				if target.UserID == nil {
					continue
				}
				groupRecipients, err = s.loadSpecifiedUsers([]uuid.UUID{*target.UserID}, collaborationWorkspaceID)
			case "collaboration_workspace_users":
				if target.CollaborationWorkspaceID == nil {
					continue
				}
				groupRecipients, err = s.loadCollaborationWorkspaceRecipients([]uuid.UUID{*target.CollaborationWorkspaceID}, false)
			case "collaboration_workspace_admins":
				if target.CollaborationWorkspaceID == nil {
					continue
				}
				groupRecipients, err = s.loadCollaborationWorkspaceRecipients([]uuid.UUID{*target.CollaborationWorkspaceID}, true)
			case "role":
				groupRecipients, err = s.loadRecipientsByRoleCode(strings.TrimSpace(target.RoleCode), collaborationWorkspaceID)
			case "feature_package":
				groupRecipients, err = s.loadRecipientsByPackageKey(strings.TrimSpace(target.PackageKey), collaborationWorkspaceID)
			default:
				continue
			}
			if err != nil {
				return nil, err
			}
			for _, recipient := range groupRecipients {
				recipient.SourceGroupID = &group.ID
				recipient.SourceGroupName = group.Name
				if recipient.SourceRuleType == "" {
					recipient.SourceRuleType = target.TargetType
				}
				if recipient.SourceRuleLabel == "" {
					recipient.SourceRuleLabel = s.resolveRecipientRuleLabel(target, collaborationWorkspaceID)
				}
				if recipient.SourceTargetID == nil {
					targetID := target.ID
					recipient.SourceTargetID = &targetID
				}
				if recipient.SourceTargetType == "" {
					recipient.SourceTargetType = target.TargetType
				}
				if recipient.SourceTargetValue == "" {
					recipient.SourceTargetValue = s.resolveRecipientTargetValue(target)
				}
				seen[recipient.UserID] = recipient
			}
		}
	}
	for _, item := range seen {
		recipients = append(recipients, item)
	}
	if len(recipients) == 0 {
		return nil, errors.New("当前接收组没有可投递的成员")
	}
	return recipients, nil
}

func (s *messageService) loadRoleRecipients(groupIDs []uuid.UUID, collaborationWorkspaceID *uuid.UUID) ([]dispatchRecipient, error) {
	return s.loadGroupRecipientsByRuleType(groupIDs, collaborationWorkspaceID, "role")
}

func (s *messageService) loadFeaturePackageRecipients(groupIDs []uuid.UUID, collaborationWorkspaceID *uuid.UUID) ([]dispatchRecipient, error) {
	return s.loadGroupRecipientsByRuleType(groupIDs, collaborationWorkspaceID, "feature_package")
}

func (s *messageService) loadGroupRecipientsByRuleType(groupIDs []uuid.UUID, collaborationWorkspaceID *uuid.UUID, ruleType string) ([]dispatchRecipient, error) {
	groups, err := s.loadEditableRecipientGroups(groupIDs, collaborationWorkspaceID)
	if err != nil {
		return nil, err
	}
	recipients := make([]dispatchRecipient, 0)
	seen := make(map[uuid.UUID]dispatchRecipient)
	for _, group := range groups {
		targets, targetErr := s.loadRecipientGroupTargets(group.ID)
		if targetErr != nil {
			return nil, targetErr
		}
		for _, target := range targets {
			if target.TargetType != ruleType {
				continue
			}
			var groupRecipients []dispatchRecipient
			switch ruleType {
			case "role":
				groupRecipients, err = s.loadRecipientsByRoleCode(strings.TrimSpace(target.RoleCode), collaborationWorkspaceID)
			case "feature_package":
				groupRecipients, err = s.loadRecipientsByPackageKey(strings.TrimSpace(target.PackageKey), collaborationWorkspaceID)
			}
			if err != nil {
				return nil, err
			}
			for _, recipient := range groupRecipients {
				recipient.SourceGroupID = &group.ID
				recipient.SourceGroupName = group.Name
				if recipient.SourceRuleType == "" {
					recipient.SourceRuleType = target.TargetType
				}
				if recipient.SourceRuleLabel == "" {
					recipient.SourceRuleLabel = s.resolveRecipientRuleLabel(target, collaborationWorkspaceID)
				}
				if recipient.SourceTargetID == nil {
					targetID := target.ID
					recipient.SourceTargetID = &targetID
				}
				if recipient.SourceTargetType == "" {
					recipient.SourceTargetType = target.TargetType
				}
				if recipient.SourceTargetValue == "" {
					recipient.SourceTargetValue = s.resolveRecipientTargetValue(target)
				}
				seen[recipient.UserID] = recipient
			}
		}
	}
	for _, item := range seen {
		recipients = append(recipients, item)
	}
	if len(recipients) == 0 {
		return nil, errors.New("当前接收组没有命中可投递成员")
	}
	return recipients, nil
}

func (s *messageService) listDispatchUsers(collaborationWorkspaceID *uuid.UUID) ([]dispatchUserOption, error) {
	if collaborationWorkspaceID != nil {
		type row struct {
			UserID                     uuid.UUID `gorm:"column:user_id"`
			Username                   string    `gorm:"column:username"`
			Nickname                   string    `gorm:"column:nickname"`
			CollaborationWorkspaceID   uuid.UUID `gorm:"column:collaboration_workspace_id"`
			CollaborationWorkspaceName string    `gorm:"column:collaboration_workspace_name"`
		}
		var rows []row
		if err := s.db.Table("collaboration_workspace_members").
			Select("collaboration_workspace_members.user_id AS user_id", "users.username AS username", "users.nickname AS nickname", "collaboration_workspace_members.collaboration_workspace_id AS collaboration_workspace_id", "collaboration_workspaces.name AS collaboration_workspace_name").
			Joins("JOIN users ON users.id = collaboration_workspace_members.user_id").
			Joins("JOIN collaboration_workspaces ON collaboration_workspaces.id = collaboration_workspace_members.collaboration_workspace_id").
			Where("collaboration_workspace_members.collaboration_workspace_id = ?", *collaborationWorkspaceID).
			Where("collaboration_workspace_members.status = ?", "active").
			Where("users.status = ?", "active").
			Order("collaboration_workspace_members.created_at ASC").
			Scan(&rows).Error; err != nil {
			return nil, err
		}
		result := make([]dispatchUserOption, 0, len(rows))
		for _, row := range rows {
			name := strings.TrimSpace(row.Nickname)
			if name == "" {
				name = strings.TrimSpace(row.Username)
			}
			collaborationWorkspaceID := row.CollaborationWorkspaceID
			result = append(result, dispatchUserOption{
				ID:                         row.UserID,
				Name:                       row.Username,
				DisplayName:                name,
				Description:                row.CollaborationWorkspaceName,
				CollaborationWorkspaceID:   &collaborationWorkspaceID,
				CollaborationWorkspaceName: row.CollaborationWorkspaceName,
			})
		}
		return result, nil
	}
	var users []models.User
	if err := s.db.Select("id", "username", "nickname").Where("status = ?", "active").Order("created_at ASC").Find(&users).Error; err != nil {
		return nil, err
	}
	result := make([]dispatchUserOption, 0, len(users))
	for _, user := range users {
		displayName := strings.TrimSpace(user.Nickname)
		if displayName == "" {
			displayName = strings.TrimSpace(user.Username)
		}
		result = append(result, dispatchUserOption{
			ID:          user.ID,
			Name:        user.Username,
			DisplayName: displayName,
			Description: "个人空间用户",
		})
	}
	return result, nil
}

func (s *messageService) listDispatchRecipientGroups(collaborationWorkspaceID *uuid.UUID) ([]dispatchRecipientGroupOption, error) {
	items, err := s.ListRecipientGroups(collaborationWorkspaceID)
	if err != nil {
		return nil, err
	}
	result := make([]dispatchRecipientGroupOption, 0, len(items))
	for _, item := range items {
		if item.Status != "normal" {
			continue
		}
		result = append(result, dispatchRecipientGroupOption{
			ID:             item.ID,
			Name:           item.Name,
			Description:    item.Description,
			MatchMode:      item.MatchMode,
			EstimatedCount: item.EstimatedCount,
		})
	}
	return result, nil
}

func (s *messageService) listDispatchRoles(collaborationWorkspaceID *uuid.UUID) ([]dispatchRoleOption, error) {
	var rows []models.Role
	query := s.db.Model(&models.Role{}).
		Select("id", "code", "name", "description", "collaboration_workspace_id", "status").
		Where("status = ? AND deleted_at IS NULL", "normal")
	if collaborationWorkspaceID != nil {
		query = query.Where("(collaboration_workspace_id = ? OR (collaboration_workspace_id IS NULL AND code IN ?))", *collaborationWorkspaceID, []string{"collaboration_workspace_admin", "member"})
	} else {
		query = query.Where("collaboration_workspace_id IS NULL")
	}
	if err := query.Order("sort_order ASC, created_at ASC").Find(&rows).Error; err != nil {
		return nil, err
	}
	result := make([]dispatchRoleOption, 0, len(rows))
	for _, row := range rows {
		result = append(result, dispatchRoleOption{
			ID:          row.ID,
			Code:        row.Code,
			Name:        row.Name,
			Description: row.Description,
		})
	}
	return result, nil
}

func (s *messageService) listDispatchFeaturePackages(collaborationWorkspaceID *uuid.UUID) ([]dispatchFeaturePackageOption, error) {
	contextType := "personal"
	if collaborationWorkspaceID != nil {
		contextType = "collaboration"
	}
	var rows []models.FeaturePackage
	if err := s.db.Model(&models.FeaturePackage{}).
		Select("id", "package_key", "name", "description").
		Where("context_type = ? AND status = ? AND deleted_at IS NULL", contextType, "normal").
		Order("sort_order ASC, created_at ASC").
		Find(&rows).Error; err != nil {
		return nil, err
	}
	result := make([]dispatchFeaturePackageOption, 0, len(rows))
	for _, row := range rows {
		result = append(result, dispatchFeaturePackageOption{
			ID:          row.ID,
			PackageKey:  row.PackageKey,
			Name:        row.Name,
			Description: row.Description,
		})
	}
	return result, nil
}

func (s *messageService) ListRecipientGroups(collaborationWorkspaceID *uuid.UUID) ([]messageRecipientGroupListItem, error) {
	var groups []models.MessageRecipientGroup
	query := s.db.Model(&models.MessageRecipientGroup{}).Where("deleted_at IS NULL")
	if collaborationWorkspaceID != nil {
		query = query.Where("scope_type = ? AND scope_id = ?", "collaboration", *collaborationWorkspaceID)
	} else {
		query = query.Where("scope_type = ? AND scope_id IS NULL", "personal")
	}
	if err := query.Order("updated_at DESC").Find(&groups).Error; err != nil {
		return nil, err
	}
	result := make([]messageRecipientGroupListItem, 0, len(groups))
	for _, group := range groups {
		targets, err := s.loadRecipientGroupTargetItems(group.ID)
		if err != nil {
			return nil, err
		}
		estimated, err := s.estimateRecipientGroup(group.ID, collaborationWorkspaceID)
		if err != nil {
			return nil, err
		}
		result = append(result, messageRecipientGroupListItem{
			ID:             group.ID,
			ScopeType:      group.ScopeType,
			ScopeID:        group.ScopeID,
			Name:           group.Name,
			Description:    group.Description,
			MatchMode:      group.MatchMode,
			Status:         group.Status,
			Editable:       true,
			EstimatedCount: estimated,
			Meta:           group.Meta,
			Targets:        targets,
			CreatedAt:      group.CreatedAt,
			UpdatedAt:      group.UpdatedAt,
		})
	}
	return result, nil
}

func (s *messageService) SaveRecipientGroup(groupID string, collaborationWorkspaceID *uuid.UUID, req messageRecipientGroupSaveRequest) (messageRecipientGroupListItem, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return messageRecipientGroupListItem{}, errors.New("接收组名称不能为空")
	}
	matchMode := strings.TrimSpace(req.MatchMode)
	if matchMode == "" {
		matchMode = "manual"
	}
	if matchMode != "manual" {
		return messageRecipientGroupListItem{}, errors.New("当前仅支持手动接收组")
	}
	status := strings.TrimSpace(req.Status)
	if status == "" {
		status = "normal"
	}
	if status != "normal" && status != "disabled" {
		return messageRecipientGroupListItem{}, errors.New("接收组状态无效")
	}
	meta := req.Meta
	if meta == nil {
		meta = models.MetaJSON{}
	}
	var saved models.MessageRecipientGroup
	err := s.db.Transaction(func(tx *gorm.DB) error {
		scopeType := "personal"
		var scopeID *uuid.UUID
		if collaborationWorkspaceID != nil {
			scopeType = "collaboration"
			scopeID = collaborationWorkspaceID
		}
		var target models.MessageRecipientGroup
		if trimmedID := strings.TrimSpace(groupID); trimmedID != "" {
			id, parseErr := uuid.Parse(trimmedID)
			if parseErr != nil {
				return errors.New("接收组标识无效")
			}
			query := tx.Model(&models.MessageRecipientGroup{}).Where("id = ? AND deleted_at IS NULL", id)
			if collaborationWorkspaceID != nil {
				query = query.Where("scope_type = ? AND scope_id = ?", "collaboration", *collaborationWorkspaceID)
			} else {
				query = query.Where("scope_type = ? AND scope_id IS NULL", "personal")
			}
			if err := query.First(&target).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return errors.New("接收组不存在或当前上下文不可编辑")
				}
				return err
			}
		} else {
			target = models.MessageRecipientGroup{ScopeType: scopeType, ScopeID: scopeID}
		}
		target.Name = name
		target.Description = strings.TrimSpace(req.Description)
		target.MatchMode = matchMode
		target.Status = status
		target.Meta = meta
		if target.ID == uuid.Nil {
			if err := tx.Create(&target).Error; err != nil {
				return err
			}
		} else {
			if err := tx.Save(&target).Error; err != nil {
				return err
			}
			if err := tx.Where("group_id = ?", target.ID).Delete(&models.MessageRecipientGroupTarget{}).Error; err != nil {
				return err
			}
		}
		targets, err := buildRecipientGroupTargets(target.ID, req.Targets, collaborationWorkspaceID)
		if err != nil {
			return err
		}
		if len(targets) > 0 {
			if err := tx.Create(&targets).Error; err != nil {
				return err
			}
		}
		saved = target
		return nil
	})
	if err != nil {
		return messageRecipientGroupListItem{}, err
	}
	items, err := s.ListRecipientGroups(collaborationWorkspaceID)
	if err != nil {
		return messageRecipientGroupListItem{}, err
	}
	for _, item := range items {
		if item.ID == saved.ID {
			return item, nil
		}
	}
	return messageRecipientGroupListItem{}, errors.New("接收组保存成功，但结果读取失败")
}

func (s *messageService) loadEditableRecipientGroups(groupIDs []uuid.UUID, collaborationWorkspaceID *uuid.UUID) ([]models.MessageRecipientGroup, error) {
	if len(groupIDs) == 0 {
		return nil, errors.New("请选择接收组")
	}
	query := s.db.Model(&models.MessageRecipientGroup{}).
		Where("id IN ? AND deleted_at IS NULL AND status = ?", groupIDs, "normal")
	if collaborationWorkspaceID != nil {
		query = query.Where("scope_type = ? AND scope_id = ?", "collaboration", *collaborationWorkspaceID)
	} else {
		query = query.Where("scope_type = ? AND scope_id IS NULL", "personal")
	}
	var groups []models.MessageRecipientGroup
	if err := query.Find(&groups).Error; err != nil {
		return nil, err
	}
	if len(groups) == 0 {
		return nil, errors.New("接收组不存在或当前上下文不可用")
	}
	return groups, nil
}

func (s *messageService) loadRecipientGroupTargets(groupID uuid.UUID) ([]models.MessageRecipientGroupTarget, error) {
	var items []models.MessageRecipientGroupTarget
	if err := s.db.Model(&models.MessageRecipientGroupTarget{}).
		Where("group_id = ? AND deleted_at IS NULL", groupID).
		Order("sort_order ASC").
		Order("created_at ASC").
		Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (s *messageService) loadRecipientGroupTargetItems(groupID uuid.UUID) ([]messageRecipientGroupTargetItem, error) {
	type row struct {
		models.MessageRecipientGroupTarget
		UserName                   string `gorm:"column:user_name"`
		CollaborationWorkspaceName string `gorm:"column:collaboration_workspace_name"`
	}
	var rows []row
	if err := s.db.Model(&models.MessageRecipientGroupTarget{}).
		Select("message_recipient_group_targets.*", "COALESCE(users.nickname, users.username, '') AS user_name", "COALESCE(collaboration_workspaces.name, '') AS collaboration_workspace_name").
		Joins("LEFT JOIN users ON users.id = message_recipient_group_targets.user_id").
		Joins("LEFT JOIN collaboration_workspaces ON collaboration_workspaces.id = message_recipient_group_targets.collaboration_workspace_id").
		Where("message_recipient_group_targets.group_id = ? AND message_recipient_group_targets.deleted_at IS NULL", groupID).
		Order("message_recipient_group_targets.sort_order ASC").
		Order("message_recipient_group_targets.created_at ASC").
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	items := make([]messageRecipientGroupTargetItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, messageRecipientGroupTargetItem{
			ID:                         row.ID,
			TargetType:                 row.TargetType,
			UserID:                     row.UserID,
			UserName:                   row.UserName,
			CollaborationWorkspaceID:   row.CollaborationWorkspaceID,
			CollaborationWorkspaceName: row.CollaborationWorkspaceName,
			RoleCode:                   row.RoleCode,
			RoleName:                   s.lookupRoleName(row.RoleCode, row.CollaborationWorkspaceID),
			PackageKey:                 row.PackageKey,
			PackageName:                s.lookupPackageName(row.PackageKey, row.CollaborationWorkspaceID),
			SortOrder:                  row.SortOrder,
			Meta:                       row.Meta,
		})
	}
	return items, nil
}

func (s *messageService) lookupRoleName(roleCode string, collaborationWorkspaceID *uuid.UUID) string {
	roleCode = strings.TrimSpace(roleCode)
	if roleCode == "" {
		return ""
	}
	var role models.Role
	query := s.db.Model(&models.Role{}).
		Select("name").
		Where("code = ? AND status = ? AND deleted_at IS NULL", roleCode, "normal")
	if collaborationWorkspaceID != nil {
		query = query.Where("(collaboration_workspace_id = ? OR collaboration_workspace_id IS NULL)", *collaborationWorkspaceID).Order("collaboration_workspace_id DESC NULLS LAST")
	} else {
		query = query.Where("collaboration_workspace_id IS NULL")
	}
	if err := query.First(&role).Error; err != nil {
		return ""
	}
	return role.Name
}

func (s *messageService) lookupPackageName(packageKey string, collaborationWorkspaceID *uuid.UUID) string {
	packageKey = strings.TrimSpace(packageKey)
	if packageKey == "" {
		return ""
	}
	contextType := "personal"
	if collaborationWorkspaceID != nil {
		contextType = "collaboration"
	}
	var pkg models.FeaturePackage
	if err := s.db.Model(&models.FeaturePackage{}).
		Select("name").
		Where("package_key = ? AND context_type = ? AND status = ? AND deleted_at IS NULL", packageKey, contextType, "normal").
		First(&pkg).Error; err != nil {
		return ""
	}
	return pkg.Name
}

func (s *messageService) resolveRecipientRuleLabel(target models.MessageRecipientGroupTarget, collaborationWorkspaceID *uuid.UUID) string {
	switch target.TargetType {
	case "user":
		return "指定用户"
	case "collaboration_workspace_users":
		return "协作空间成员"
	case "collaboration_workspace_admins":
		return "协作空间管理员"
	case "role":
		name := s.lookupRoleName(target.RoleCode, collaborationWorkspaceID)
		if name != "" {
			return name
		}
		return target.RoleCode
	case "feature_package":
		name := s.lookupPackageName(target.PackageKey, collaborationWorkspaceID)
		if name != "" {
			return name
		}
		return target.PackageKey
	default:
		return target.TargetType
	}
}

func (s *messageService) resolveRecipientTargetValue(target models.MessageRecipientGroupTarget) string {
	switch target.TargetType {
	case "user":
		return uuidString(target.UserID)
	case "collaboration_workspace_users", "collaboration_workspace_admins":
		return uuidString(target.CollaborationWorkspaceID)
	case "role":
		return strings.TrimSpace(target.RoleCode)
	case "feature_package":
		return strings.TrimSpace(target.PackageKey)
	default:
		return target.TargetType
	}
}

func (s *messageService) estimateRecipientGroup(groupID uuid.UUID, collaborationWorkspaceID *uuid.UUID) (int, error) {
	targets, err := s.loadRecipientGroupTargets(groupID)
	if err != nil {
		return 0, err
	}
	seen := make(map[uuid.UUID]struct{})
	for _, target := range targets {
		var recipients []dispatchRecipient
		switch target.TargetType {
		case "user":
			if target.UserID == nil {
				continue
			}
			recipients, err = s.loadSpecifiedUsers([]uuid.UUID{*target.UserID}, collaborationWorkspaceID)
		case "collaboration_workspace_users":
			if target.CollaborationWorkspaceID == nil {
				continue
			}
			recipients, err = s.loadCollaborationWorkspaceRecipients([]uuid.UUID{*target.CollaborationWorkspaceID}, false)
		case "collaboration_workspace_admins":
			if target.CollaborationWorkspaceID == nil {
				continue
			}
			recipients, err = s.loadCollaborationWorkspaceRecipients([]uuid.UUID{*target.CollaborationWorkspaceID}, true)
		case "role":
			recipients, err = s.loadRecipientsByRoleCode(strings.TrimSpace(target.RoleCode), collaborationWorkspaceID)
		case "feature_package":
			recipients, err = s.loadRecipientsByPackageKey(strings.TrimSpace(target.PackageKey), collaborationWorkspaceID)
		default:
			continue
		}
		if err != nil {
			return 0, err
		}
		for _, recipient := range recipients {
			seen[recipient.UserID] = struct{}{}
		}
	}
	return len(seen), nil
}

func buildRecipientGroupTargets(
	groupID uuid.UUID,
	inputs []messageRecipientGroupTargetSaveRequest,
	collaborationWorkspaceID *uuid.UUID,
) ([]models.MessageRecipientGroupTarget, error) {
	result := make([]models.MessageRecipientGroupTarget, 0, len(inputs))
	for index, item := range inputs {
		targetType := strings.TrimSpace(item.TargetType)
		target := models.MessageRecipientGroupTarget{
			GroupID:    groupID,
			TargetType: targetType,
			RoleCode:   strings.TrimSpace(item.RoleCode),
			PackageKey: strings.TrimSpace(item.PackageKey),
			SortOrder:  item.SortOrder,
			Meta:       item.Meta,
		}
		if target.Meta == nil {
			target.Meta = models.MetaJSON{}
		}
		if target.SortOrder == 0 {
			target.SortOrder = index + 1
		}
		switch targetType {
		case "user":
			if strings.TrimSpace(item.UserID) == "" {
				return nil, errors.New("接收组成员缺少用户")
			}
			id, err := uuid.Parse(strings.TrimSpace(item.UserID))
			if err != nil {
				return nil, errors.New("接收组用户标识无效")
			}
			target.UserID = &id
		case "collaboration_workspace_users", "collaboration_workspace_admins":
			if collaborationWorkspaceID != nil {
				target.CollaborationWorkspaceID = collaborationWorkspaceID
			} else {
				if strings.TrimSpace(item.CollaborationWorkspaceID) == "" {
					return nil, errors.New("接收组协作空间标识不能为空")
				}
				id, err := uuid.Parse(strings.TrimSpace(item.CollaborationWorkspaceID))
				if err != nil {
					return nil, errors.New("接收组协作空间标识无效")
				}
				target.CollaborationWorkspaceID = &id
			}
		case "role":
			if target.RoleCode == "" {
				return nil, errors.New("接收组角色规则不能为空")
			}
		case "feature_package":
			if target.PackageKey == "" {
				return nil, errors.New("接收组功能包规则不能为空")
			}
		default:
			return nil, errors.New("接收组目标类型无效")
		}
		result = append(result, target)
	}
	return result, nil
}

func normalizeMessageType(value string) string {
	switch strings.TrimSpace(value) {
	case "notice", "message", "todo":
		return strings.TrimSpace(value)
	default:
		return ""
	}
}

func normalizeAudienceType(value string) string {
	switch strings.TrimSpace(value) {
	case "all_users", "collaboration_workspace_admins", "collaboration_workspace_users", "specified_users", "recipient_group", "role", "feature_package":
		return strings.TrimSpace(value)
	default:
		return ""
	}
}

func normalizeActionType(value string) string {
	switch strings.TrimSpace(value) {
	case "route", "external_link", "api", "none":
		return strings.TrimSpace(value)
	default:
		return ""
	}
}

func normalizePriority(value string) string {
	switch strings.TrimSpace(value) {
	case "low", "normal", "high", "urgent":
		return strings.TrimSpace(value)
	default:
		return ""
	}
}

func normalizeTemplateStatus(value string) string {
	switch strings.TrimSpace(value) {
	case "normal", "disabled":
		return strings.TrimSpace(value)
	default:
		return ""
	}
}

func (s *messageService) buildTemplateListItem(template models.MessageTemplate, ownerCollaborationWorkspaceName string, collaborationWorkspaceID *uuid.UUID) messageTemplateListItem {
	editable := template.OwnerScope == "personal" && collaborationWorkspaceID == nil
	if collaborationWorkspaceID != nil && template.OwnerScope == "collaboration" && template.OwnerCollaborationWorkspaceID != nil && *template.OwnerCollaborationWorkspaceID == *collaborationWorkspaceID {
		editable = true
	}
	return messageTemplateListItem{
		ID:                              template.ID,
		TemplateKey:                     template.TemplateKey,
		Name:                            template.Name,
		Description:                     template.Description,
		MessageType:                     template.MessageType,
		OwnerScope:                      template.OwnerScope,
		OwnerCollaborationWorkspaceID:   template.OwnerCollaborationWorkspaceID,
		OwnerCollaborationWorkspaceName: ownerCollaborationWorkspaceName,
		AudienceType:                    template.AudienceType,
		TitleTemplate:                   template.TitleTemplate,
		SummaryTemplate:                 template.SummaryTemplate,
		ContentTemplate:                 template.ContentTemplate,
		Status:                          template.Status,
		Editable:                        editable,
		Meta:                            template.Meta,
		CreatedAt:                       template.CreatedAt,
		UpdatedAt:                       template.UpdatedAt,
	}
}

func buildTemplateKey(raw string, collaborationWorkspaceID *uuid.UUID, nowUnix int64) string {
	normalized := strings.ToLower(strings.TrimSpace(raw))
	normalized = strings.ReplaceAll(normalized, " ", "-")
	normalized = strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z':
			return r
		case r >= '0' && r <= '9':
			return r
		case r == '-', r == '_', r == '.':
			return r
		default:
			return -1
		}
	}, normalized)
	normalized = strings.Trim(normalized, "-._")
	if normalized == "" {
		normalized = fmt.Sprintf("template-%d", nowUnix)
	}
	if collaborationWorkspaceID != nil {
		prefix := "collaboration_workspace." + collaborationWorkspaceID.String() + "."
		if strings.HasPrefix(normalized, prefix) {
			return normalized
		}
		return prefix + normalized
	}
	if strings.HasPrefix(normalized, "personal.") {
		return normalized
	}
	return "personal." + normalized
}

func (s *messageService) resolveLegacyCollaborationWorkspaceIDs(values []uuid.UUID) ([]uuid.UUID, error) {
	if len(values) == 0 {
		return []uuid.UUID{}, nil
	}
	result := make([]uuid.UUID, 0, len(values))
	seen := make(map[uuid.UUID]struct{}, len(values))
	for _, item := range values {
		resolved, err := s.resolveLegacyCollaborationWorkspaceID(item)
		if err != nil {
			return nil, err
		}
		if _, ok := seen[resolved]; ok {
			continue
		}
		seen[resolved] = struct{}{}
		result = append(result, resolved)
	}
	return result, nil
}

func (s *messageService) resolveLegacyCollaborationWorkspaceIDString(value string) (string, error) {
	target := strings.TrimSpace(value)
	if target == "" {
		return "", nil
	}
	parsed, err := uuid.Parse(target)
	if err != nil {
		return "", errors.New("协作空间标识无效")
	}
	resolved, err := s.resolveLegacyCollaborationWorkspaceID(parsed)
	if err != nil {
		return "", err
	}
	return resolved.String(), nil
}

func (s *messageService) resolveLegacyCollaborationWorkspaceID(reference uuid.UUID) (uuid.UUID, error) {
	if reference == uuid.Nil {
		return uuid.Nil, nil
	}
	var workspace models.Workspace
	err := s.db.Select("id", "workspace_type", "collaboration_workspace_id").
		Where("id = ? AND deleted_at IS NULL", reference).
		First(&workspace).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return reference, nil
		}
		return uuid.Nil, err
	}
	if workspace.WorkspaceType != models.WorkspaceTypeCollaboration || workspace.CollaborationWorkspaceID == nil || *workspace.CollaborationWorkspaceID == uuid.Nil {
		return reference, nil
	}
	return *workspace.CollaborationWorkspaceID, nil
}

func convertTemplatePersistenceError(err error) error {
	if err == nil {
		return nil
	}
	message := strings.ToLower(strings.TrimSpace(err.Error()))
	if strings.Contains(message, "duplicate key") || strings.Contains(message, "unique constraint") {
		return errors.New("模板标识已存在，请更换后重试")
	}
	return err
}

func parseTargetCollaborationWorkspaceIDs(values []string) ([]uuid.UUID, error) {
	return parseUUIDStrings(values, "目标协作空间标识无效")
}

func parseUUIDStrings(values []string, errorMessage string) ([]uuid.UUID, error) {
	result := make([]uuid.UUID, 0, len(values))
	seen := make(map[uuid.UUID]struct{}, len(values))
	for _, raw := range values {
		target := strings.TrimSpace(raw)
		if target == "" {
			continue
		}
		id, err := uuid.Parse(target)
		if err != nil {
			return nil, errors.New(errorMessage)
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	return result, nil
}

func parseMetaUUIDList(raw interface{}, errorMessage string) ([]uuid.UUID, error) {
	switch typed := raw.(type) {
	case []string:
		return parseUUIDStrings(typed, errorMessage)
	case []interface{}:
		values := make([]string, 0, len(typed))
		for _, item := range typed {
			if target, ok := item.(string); ok {
				values = append(values, target)
			}
		}
		return parseUUIDStrings(values, errorMessage)
	default:
		return []uuid.UUID{}, nil
	}
}

func cloneMetaJSON(value models.MetaJSON) models.MetaJSON {
	if value == nil {
		return models.MetaJSON{}
	}
	result := make(models.MetaJSON, len(value))
	for key, item := range value {
		result[key] = item
	}
	return result
}

func uuidPtrFromTemplate(template *models.MessageTemplate) *uuid.UUID {
	if template == nil {
		return nil
	}
	id := template.ID
	return &id
}

func uuidString(target *uuid.UUID) string {
	if target == nil {
		return ""
	}
	return target.String()
}

func singleCollaborationWorkspaceID(values []uuid.UUID) *uuid.UUID {
	if len(values) != 1 {
		return nil
	}
	target := values[0]
	return &target
}

func uuidListToStringList(values []uuid.UUID) []string {
	result := make([]string, 0, len(values))
	for _, item := range values {
		result = append(result, item.String())
	}
	return result
}

func (s *messageService) loadActiveRecipientUserRows(userIDs []uuid.UUID) ([]dispatchRecipientUserRow, error) {
	if len(userIDs) == 0 {
		return []dispatchRecipientUserRow{}, nil
	}
	rows := make([]dispatchRecipientUserRow, 0, len(userIDs))
	if err := s.db.Model(&models.User{}).
		Select("id AS user_id", "username", "nickname").
		Where("id IN ? AND status = ? AND deleted_at IS NULL", mergeDispatchUserIDs(userIDs), "active").
		Order("created_at ASC").
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func mergeDispatchUserIDs(groups ...[]uuid.UUID) []uuid.UUID {
	result := make([]uuid.UUID, 0)
	seen := make(map[uuid.UUID]struct{})
	for _, group := range groups {
		for _, id := range group {
			if id == uuid.Nil {
				continue
			}
			if _, ok := seen[id]; ok {
				continue
			}
			seen[id] = struct{}{}
			result = append(result, id)
		}
	}
	return result
}

func displayRecipientName(username, nickname string) string {
	if value := strings.TrimSpace(nickname); value != "" {
		return value
	}
	return strings.TrimSpace(username)
}

func roleRecipientRuleLabel(roleCode, source string) string {
	return strings.TrimSpace(roleCode) + " · " + strings.TrimSpace(source)
}

func packageRecipientRuleLabel(packageKey, source string) string {
	return strings.TrimSpace(packageKey) + " · " + strings.TrimSpace(source)
}

func membershipRecipientRuleLabel(label string) string {
	return strings.TrimSpace(label) + " · membership_identity"
}


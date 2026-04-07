// message_facade.go — Phase 4: exposes the unexported messageService through
// the Facade so internal/api/handlers can call message operations without
// re-implementing the logic.
package system

import (
	"github.com/google/uuid"
)

// ── public DTOs ────────────────────────────────────────────────────────────

// MessageInboxSummary mirrors inboxSummary for external callers.
type MessageInboxSummary = inboxSummary

// MessageInboxListResult mirrors inboxListResult for external callers.
type MessageInboxListResult = inboxListResult

// MessageInboxListItem mirrors inboxListItem for external callers.
type MessageInboxListItem = inboxListItem

// MessageInboxDetail mirrors inboxDetail for external callers.
type MessageInboxDetail = inboxDetail

// MessageDispatchOptions mirrors dispatchOptions for external callers.
type MessageDispatchOptions = dispatchOptions

// MessageDispatchResult mirrors dispatchResult for external callers.
type MessageDispatchResult = dispatchResult

// MessageTemplateListResult mirrors messageTemplateListResult for external callers.
type MessageTemplateListResult = messageTemplateListResult

// MessageTemplateListItem mirrors messageTemplateListItem for external callers.
type MessageTemplateListItem = messageTemplateListItem

// MessageSenderListItem mirrors messageSenderListItem for external callers.
type MessageSenderListItem = messageSenderListItem

// MessageRecipientGroupListItem mirrors messageRecipientGroupListItem for external callers.
type MessageRecipientGroupListItem = messageRecipientGroupListItem

// DispatchRecordListResult mirrors dispatchRecordListResult for external callers.
type DispatchRecordListResult = dispatchRecordListResult

// DispatchRecordDetail mirrors dispatchRecordDetail for external callers.
type DispatchRecordDetail = dispatchRecordDetail

// ── public request types ────────────────────────────────────────────────────

// MessageInboxQuery is the public version of inboxQuery.
type MessageInboxQuery = inboxQuery

// MessageDispatchRequest is the public version of dispatchRequest.
type MessageDispatchRequest = dispatchRequest

// MessageTemplateUpsertRequest is the public version of messageTemplateUpsertRequest.
type MessageTemplateUpsertRequest = messageTemplateUpsertRequest

// MessageSenderSaveRequest is the public version of messageSenderSaveRequest.
type MessageSenderSaveRequest = messageSenderSaveRequest

// MessageRecipientGroupSaveRequest is the public version of messageRecipientGroupSaveRequest.
type MessageRecipientGroupSaveRequest = messageRecipientGroupSaveRequest

// MessageRecipientGroupTargetSaveRequest is the public version of messageRecipientGroupTargetSaveRequest.
type MessageRecipientGroupTargetSaveRequest = messageRecipientGroupTargetSaveRequest

// MessageTemplateQuery is the public version of messageTemplateQuery.
type MessageTemplateQuery = messageTemplateQuery

// DispatchRecordQuery is the public version of dispatchRecordQuery.
type DispatchRecordQuery = dispatchRecordQuery

// ── Facade message methods ──────────────────────────────────────────────────

// GetInboxSummary delegates to messageService.
func (f *Facade) GetInboxSummary(userID uuid.UUID) (MessageInboxSummary, error) {
	return f.messageSvc.GetInboxSummary(userID)
}

// ListInbox delegates to messageService.
func (f *Facade) ListInbox(userID uuid.UUID, q MessageInboxQuery) (MessageInboxListResult, error) {
	return f.messageSvc.ListInbox(userID, q)
}

// GetInboxDetail delegates to messageService.
func (f *Facade) GetInboxDetail(userID, deliveryID uuid.UUID) (MessageInboxDetail, error) {
	return f.messageSvc.GetInboxDetail(userID, deliveryID)
}

// MarkRead delegates to messageService.
func (f *Facade) MarkRead(userID, deliveryID uuid.UUID) error {
	return f.messageSvc.MarkRead(userID, deliveryID)
}

// MarkAllRead delegates to messageService.
func (f *Facade) MarkAllRead(userID uuid.UUID, boxType string) error {
	return f.messageSvc.MarkAllRead(userID, boxType)
}

// UpdateTodoStatus delegates to messageService.
func (f *Facade) UpdateTodoStatus(userID, deliveryID uuid.UUID, action string) error {
	return f.messageSvc.UpdateTodoStatus(userID, deliveryID, action)
}

// GetDispatchOptions delegates to messageService.
func (f *Facade) GetDispatchOptions(userID uuid.UUID, cwID *uuid.UUID) (MessageDispatchOptions, error) {
	return f.messageSvc.GetDispatchOptions(userID, cwID)
}

// DispatchMessage delegates to messageService.
func (f *Facade) DispatchMessage(userID uuid.UUID, cwID *uuid.UUID, req MessageDispatchRequest) (MessageDispatchResult, error) {
	return f.messageSvc.DispatchMessage(userID, cwID, req)
}

// ListTemplates delegates to messageService.
func (f *Facade) ListTemplates(cwID *uuid.UUID, q MessageTemplateQuery) (MessageTemplateListResult, error) {
	return f.messageSvc.ListTemplates(cwID, q)
}

// SaveTemplate delegates to messageService (create or update based on templateID).
func (f *Facade) SaveTemplate(templateID string, cwID *uuid.UUID, req MessageTemplateUpsertRequest) (MessageTemplateListItem, error) {
	return f.messageSvc.SaveTemplate(templateID, cwID, req)
}

// ListSenders delegates to messageService.
func (f *Facade) ListSenders(cwID *uuid.UUID) ([]MessageSenderListItem, error) {
	return f.messageSvc.ListSenders(cwID)
}

// SaveSender delegates to messageService (create or update based on senderID).
func (f *Facade) SaveSender(senderID string, cwID *uuid.UUID, req MessageSenderSaveRequest) (MessageSenderListItem, error) {
	return f.messageSvc.SaveSender(senderID, cwID, req)
}

// ListRecipientGroups delegates to messageService.
func (f *Facade) ListRecipientGroups(cwID *uuid.UUID) ([]MessageRecipientGroupListItem, error) {
	return f.messageSvc.ListRecipientGroups(cwID)
}

// SaveRecipientGroup delegates to messageService (create or update based on groupID).
func (f *Facade) SaveRecipientGroup(groupID string, cwID *uuid.UUID, req MessageRecipientGroupSaveRequest) (MessageRecipientGroupListItem, error) {
	return f.messageSvc.SaveRecipientGroup(groupID, cwID, req)
}

// ListDispatchRecords delegates to messageService.
func (f *Facade) ListDispatchRecords(cwID *uuid.UUID, q DispatchRecordQuery) (DispatchRecordListResult, error) {
	return f.messageSvc.ListDispatchRecords(cwID, q)
}

// GetDispatchRecordDetail delegates to messageService.
func (f *Facade) GetDispatchRecordDetail(cwID *uuid.UUID, recordID string) (DispatchRecordDetail, error) {
	return f.messageSvc.GetDispatchRecordDetail(cwID, recordID)
}

// ResolveLegacyCollaborationWorkspaceIDString delegates to messageService.
func (f *Facade) ResolveLegacyCollaborationWorkspaceIDString(value string) (string, error) {
	return f.messageSvc.resolveLegacyCollaborationWorkspaceIDString(value)
}

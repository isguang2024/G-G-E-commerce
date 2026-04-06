package workspace

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/gg-ecommerce/backend/internal/modules/system/models"
)

type Summary struct {
	ID                       uuid.UUID  `json:"id"`
	WorkspaceType            string     `json:"workspace_type"`
	Name                     string     `json:"name"`
	Code                     string     `json:"code"`
	OwnerUserID              *uuid.UUID `json:"owner_user_id,omitempty"`
	CollaborationWorkspaceID *uuid.UUID `json:"collaboration_workspace_id,omitempty"`
	Status                   string     `json:"status"`
}

type Service interface {
	EnsureWorkspaceBackfill() error
	EnsurePersonalWorkspaceForUser(userID uuid.UUID) (*models.Workspace, error)
	GetByID(id uuid.UUID) (*models.Workspace, error)
	GetMember(workspaceID, userID uuid.UUID) (*models.WorkspaceMember, error)
	GetPersonalWorkspaceByUserID(userID uuid.UUID) (*models.Workspace, error)
	GetCollaborationWorkspaceByCollaborationWorkspaceID(collaborationWorkspaceID uuid.UUID) (*models.Workspace, error)
	GetWorkspaceByCollaborationWorkspaceID(collaborationWorkspaceID uuid.UUID) (*models.Workspace, error)
	ListByUserID(userID uuid.UUID) ([]Summary, error)
}

type service struct {
	db     *gorm.DB
	logger *zap.Logger
}

func NewService(db *gorm.DB, logger *zap.Logger) Service {
	return &service{db: db, logger: logger}
}

func (s *service) EnsureWorkspaceBackfill() error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := ensurePersonalWorkspacesTx(tx); err != nil {
			return err
		}
		if err := ensureCollaborationWorkspacesTx(tx); err != nil {
			return err
		}
		if err := ensureWorkspaceMembersTx(tx); err != nil {
			return err
		}
		return nil
	})
}

func (s *service) EnsurePersonalWorkspaceForUser(userID uuid.UUID) (*models.Workspace, error) {
	if userID == uuid.Nil {
		return nil, fmt.Errorf("invalid user id")
	}
	if err := s.EnsureWorkspaceBackfill(); err != nil {
		return nil, err
	}
	return s.GetPersonalWorkspaceByUserID(userID)
}

func (s *service) GetByID(id uuid.UUID) (*models.Workspace, error) {
	var workspace models.Workspace
	err := s.db.Where("id = ? AND deleted_at IS NULL", id).First(&workspace).Error
	if err != nil {
		return nil, err
	}
	return &workspace, nil
}

func (s *service) GetMember(workspaceID, userID uuid.UUID) (*models.WorkspaceMember, error) {
	var member models.WorkspaceMember
	err := s.db.Where("workspace_id = ? AND user_id = ? AND deleted_at IS NULL", workspaceID, userID).First(&member).Error
	if err != nil {
		return nil, err
	}
	return &member, nil
}

func (s *service) GetPersonalWorkspaceByUserID(userID uuid.UUID) (*models.Workspace, error) {
	var workspace models.Workspace
	err := s.db.Where("workspace_type = ? AND owner_user_id = ? AND deleted_at IS NULL", models.WorkspaceTypePersonal, userID).First(&workspace).Error
	if err != nil {
		return nil, err
	}
	return &workspace, nil
}

func (s *service) GetCollaborationWorkspaceByCollaborationWorkspaceID(collaborationWorkspaceID uuid.UUID) (*models.Workspace, error) {
	var workspace models.Workspace
	err := s.db.Where("workspace_type = ? AND collaboration_workspace_id = ? AND deleted_at IS NULL", models.WorkspaceTypeCollaboration, collaborationWorkspaceID).First(&workspace).Error
	if err != nil {
		return nil, err
	}
	return &workspace, nil
}

func (s *service) GetWorkspaceByCollaborationWorkspaceID(collaborationWorkspaceID uuid.UUID) (*models.Workspace, error) {
	return s.GetCollaborationWorkspaceByCollaborationWorkspaceID(collaborationWorkspaceID)
}

func (s *service) ListByUserID(userID uuid.UUID) ([]Summary, error) {
	var memberships []models.WorkspaceMember
	if err := s.db.Where("user_id = ? AND deleted_at IS NULL", userID).Find(&memberships).Error; err != nil {
		return nil, err
	}
	if len(memberships) == 0 {
		return []Summary{}, nil
	}

	ids := make([]uuid.UUID, 0, len(memberships))
	seen := make(map[uuid.UUID]struct{}, len(memberships))
	for _, item := range memberships {
		if _, ok := seen[item.WorkspaceID]; ok {
			continue
		}
		seen[item.WorkspaceID] = struct{}{}
		ids = append(ids, item.WorkspaceID)
	}

	var workspaces []models.Workspace
	if err := s.db.Where("id IN ? AND deleted_at IS NULL", ids).Order("workspace_type ASC, created_at ASC").Find(&workspaces).Error; err != nil {
		return nil, err
	}

	result := make([]Summary, 0, len(workspaces))
	for _, item := range workspaces {
		result = append(result, workspaceToSummary(&item))
	}
	return result, nil
}

func workspaceToSummary(item *models.Workspace) Summary {
	if item == nil {
		return Summary{}
	}
	return Summary{
		ID:                       item.ID,
		WorkspaceType:            item.WorkspaceType,
		Name:                     item.Name,
		Code:                     item.Code,
		OwnerUserID:              item.OwnerUserID,
		CollaborationWorkspaceID: item.CollaborationWorkspaceID,
		Status:                   item.Status,
	}
}

func ensurePersonalWorkspacesTx(tx *gorm.DB) error {
	var users []models.User
	if err := tx.Where("deleted_at IS NULL").Find(&users).Error; err != nil {
		return err
	}
	for _, item := range users {
		if item.ID == uuid.Nil {
			continue
		}
		var count int64
		if err := tx.Model(&models.Workspace{}).
			Where("workspace_type = ? AND owner_user_id = ? AND deleted_at IS NULL", models.WorkspaceTypePersonal, item.ID).
			Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			continue
		}
		workspace := models.Workspace{
			WorkspaceType: models.WorkspaceTypePersonal,
			Name:          buildPersonalWorkspaceName(item),
			Code:          buildPersonalWorkspaceCode(item),
			OwnerUserID:   uuidPtr(item.ID),
			Status:        models.WorkspaceStatusActive,
			Meta: models.MetaJSON{
				"legacy_source": "user",
			},
		}
		if err := tx.Create(&workspace).Error; err != nil {
			return err
		}
	}
	return nil
}

func ensureCollaborationWorkspacesTx(tx *gorm.DB) error {
	var collaborationWorkspaces []models.CollaborationWorkspace
	if err := tx.Where("deleted_at IS NULL").Find(&collaborationWorkspaces).Error; err != nil {
		return err
	}
	for _, item := range collaborationWorkspaces {
		if item.ID == uuid.Nil {
			continue
		}
		var count int64
		if err := tx.Model(&models.Workspace{}).
			Where("workspace_type = ? AND collaboration_workspace_id = ? AND deleted_at IS NULL", models.WorkspaceTypeCollaboration, item.ID).
			Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			continue
		}
		workspace := models.Workspace{
			WorkspaceType:            models.WorkspaceTypeCollaboration,
			Name:                     strings.TrimSpace(item.Name),
			Code:                     buildCollaborationWorkspaceCode(item),
			OwnerUserID:              uuidPtr(item.OwnerID),
			CollaborationWorkspaceID: uuidPtr(item.ID),
			Status:                   normalizeWorkspaceStatus(item.Status),
			Meta: models.MetaJSON{
				"legacy_source":                "collaboration_workspace",
				"collaboration_workspace_plan": item.Plan,
			},
		}
		if err := tx.Create(&workspace).Error; err != nil {
			return err
		}
	}
	return nil
}

func ensureWorkspaceMembersTx(tx *gorm.DB) error {
	var personalWorkspaces []models.Workspace
	if err := tx.Where("workspace_type = ? AND deleted_at IS NULL", models.WorkspaceTypePersonal).Find(&personalWorkspaces).Error; err != nil {
		return err
	}
	for _, item := range personalWorkspaces {
		if item.OwnerUserID == nil || *item.OwnerUserID == uuid.Nil {
			continue
		}
		if err := ensureWorkspaceMemberTx(tx, item.ID, *item.OwnerUserID, models.WorkspaceMemberOwner, models.WorkspaceStatusActive, nil); err != nil {
			return err
		}
	}

	var collaborationWorkspaceMembers []models.CollaborationWorkspaceMember
	if err := tx.Where("deleted_at IS NULL").Find(&collaborationWorkspaceMembers).Error; err != nil {
		return err
	}
	for _, item := range collaborationWorkspaceMembers {
		var workspace models.Workspace
		if err := tx.Where("workspace_type = ? AND collaboration_workspace_id = ? AND deleted_at IS NULL", models.WorkspaceTypeCollaboration, item.CollaborationWorkspaceID).First(&workspace).Error; err != nil {
			return err
		}
		if err := ensureWorkspaceMemberTx(tx, workspace.ID, item.UserID, mapCollaborationWorkspaceRoleToWorkspaceMemberType(item.RoleCode), normalizeWorkspaceStatus(item.Status), uuidPtr(item.ID)); err != nil {
			return err
		}
	}
	return nil
}

func ensureWorkspaceMemberTx(tx *gorm.DB, workspaceID, userID uuid.UUID, memberType, status string, collaborationWorkspaceMemberID *uuid.UUID) error {
	var existing models.WorkspaceMember
	err := tx.Where("workspace_id = ? AND user_id = ? AND deleted_at IS NULL", workspaceID, userID).First(&existing).Error
	if err == nil {
		updates := map[string]interface{}{
			"member_type": memberType,
			"status":      status,
			"updated_at":  tx.NowFunc(),
		}
		if collaborationWorkspaceMemberID != nil {
			updates["collaboration_workspace_member_id"] = *collaborationWorkspaceMemberID
		}
		return tx.Model(&existing).Updates(updates).Error
	}
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}
	member := models.WorkspaceMember{
		WorkspaceID:                    workspaceID,
		UserID:                         userID,
		MemberType:                     memberType,
		Status:                         status,
		CollaborationWorkspaceMemberID: collaborationWorkspaceMemberID,
	}
	return tx.Create(&member).Error
}

func buildPersonalWorkspaceName(user models.User) string {
	for _, candidate := range []string{
		strings.TrimSpace(user.Nickname),
		strings.TrimSpace(user.Username),
		strings.TrimSpace(user.Email),
	} {
		if candidate != "" {
			return candidate + " Personal Workspace"
		}
	}
	return "Personal Workspace"
}

func buildPersonalWorkspaceCode(user models.User) string {
	base := normalizeWorkspaceCodeComponent(firstNonEmpty(
		strings.TrimSpace(user.Username),
		strings.TrimSpace(user.Email),
		user.ID.String(),
	))
	if base == "" {
		base = user.ID.String()
	}
	return "personal-" + base
}

func buildCollaborationWorkspaceCode(collaborationWorkspace models.CollaborationWorkspace) string {
	base := normalizeWorkspaceCodeComponent(strings.TrimSpace(collaborationWorkspace.Name))
	if base == "" {
		base = collaborationWorkspace.ID.String()
	}
	return "collaboration-" + base
}

func normalizeWorkspaceCodeComponent(value string) string {
	target := strings.ToLower(strings.TrimSpace(value))
	if target == "" {
		return ""
	}
	reg := regexp.MustCompile(`[^a-z0-9]+`)
	target = reg.ReplaceAllString(target, "-")
	return strings.Trim(target, "-")
}

func normalizeWorkspaceStatus(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "disabled", "inactive", "suspended":
		return "disabled"
	default:
		return models.WorkspaceStatusActive
	}
}

func mapCollaborationWorkspaceRoleToWorkspaceMemberType(roleCode string) string {
	switch strings.ToLower(strings.TrimSpace(roleCode)) {
	case "owner":
		return models.WorkspaceMemberOwner
	case "collaboration_workspace_admin", "admin":
		return models.WorkspaceMemberAdmin
	case "viewer":
		return models.WorkspaceMemberViewer
	default:
		return models.WorkspaceMemberMember
	}
}

func firstNonEmpty(values ...string) string {
	for _, item := range values {
		if strings.TrimSpace(item) != "" {
			return strings.TrimSpace(item)
		}
	}
	return ""
}

func uuidPtr(value uuid.UUID) *uuid.UUID {
	if value == uuid.Nil {
		return nil
	}
	target := value
	return &target
}

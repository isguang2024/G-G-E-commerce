package collaborationworkspace

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/gg-ecommerce/backend/internal/api/dto"
	"github.com/gg-ecommerce/backend/internal/modules/system/models"
	"github.com/gg-ecommerce/backend/internal/modules/system/user"
	"github.com/gg-ecommerce/backend/internal/pkg/workspacerolebinding"
)

var ErrCollaborationWorkspaceNotFound = errors.New("collaboration workspace not found")
var ErrCollaborationWorkspaceMemberExists = errors.New("user already in collaboration workspace")
var ErrCollaborationWorkspaceMemberNotFound = errors.New("member not in collaboration workspace")

type CollaborationWorkspaceService interface {
	List(req *dto.CollaborationWorkspaceListRequest) ([]user.CollaborationWorkspace, int64, error)
	ListOptions(req *dto.CollaborationWorkspaceListRequest) ([]user.CollaborationWorkspace, error)
	Get(id uuid.UUID) (*user.CollaborationWorkspace, error)
	Create(req *dto.CollaborationWorkspaceCreateRequest, ownerID *uuid.UUID) (*user.CollaborationWorkspace, error)
	Update(id uuid.UUID, req *dto.CollaborationWorkspaceUpdateRequest) error
	Delete(id uuid.UUID) error
	ListMembers(collaborationWorkspaceID uuid.UUID, searchParams *user.MemberSearchParams) ([]user.CollaborationWorkspaceMember, error)
	AddMember(collaborationWorkspaceID uuid.UUID, req *dto.CollaborationWorkspaceAddMemberRequest, invitedBy *uuid.UUID) error
	RemoveMember(collaborationWorkspaceID, userID uuid.UUID) error
	UpdateMemberRole(collaborationWorkspaceID, userID uuid.UUID, roleCode string) error
}

const defaultCollaborationWorkspaceRoleAdminCode = "collaboration_workspace_admin"
const defaultCollaborationWorkspaceRoleMemberCode = "collaboration_workspace_member"

type collaborationWorkspaceService struct {
	db                               *gorm.DB
	collaborationWorkspaceRepo       user.CollaborationWorkspaceRepository
	collaborationWorkspaceMemberRepo user.CollaborationWorkspaceMemberRepository
	userRepo                         user.UserRepository
	roleRepo                         user.RoleRepository
	userRoleRepo                     user.UserRoleRepository
	refresher                        interface {
		RefreshCollaborationWorkspace(collaborationWorkspaceID uuid.UUID) error
	}
	logger *zap.Logger
}

func NewCollaborationWorkspaceService(
	db *gorm.DB,
	collaborationWorkspaceRepo user.CollaborationWorkspaceRepository,
	collaborationWorkspaceMemberRepo user.CollaborationWorkspaceMemberRepository,
	userRepo user.UserRepository,
	roleRepo user.RoleRepository,
	userRoleRepo user.UserRoleRepository,
	refresher interface {
		RefreshCollaborationWorkspace(collaborationWorkspaceID uuid.UUID) error
	},
	logger *zap.Logger,
) CollaborationWorkspaceService {
	return &collaborationWorkspaceService{
		db:                               db,
		collaborationWorkspaceRepo:       collaborationWorkspaceRepo,
		collaborationWorkspaceMemberRepo: collaborationWorkspaceMemberRepo,
		userRepo:                         userRepo,
		roleRepo:                         roleRepo,
		userRoleRepo:                     userRoleRepo,
		refresher:                        refresher,
		logger:                           logger,
	}
}

func (s *collaborationWorkspaceService) List(req *dto.CollaborationWorkspaceListRequest) ([]user.CollaborationWorkspace, int64, error) {
	if req.Current <= 0 {
		req.Current = 1
	}
	if req.Size <= 0 {
		req.Size = 20
	}
	offset := (req.Current - 1) * req.Size
	return s.collaborationWorkspaceRepo.List(offset, req.Size, req.Name, req.Status)
}

func (s *collaborationWorkspaceService) ListOptions(req *dto.CollaborationWorkspaceListRequest) ([]user.CollaborationWorkspace, error) {
	query := s.db.Model(&user.CollaborationWorkspace{})
	if req != nil {
		if name := strings.TrimSpace(req.Name); name != "" {
			query = query.Where("name LIKE ?", "%"+name+"%")
		}
		if status := strings.TrimSpace(req.Status); status != "" {
			query = query.Where("status = ?", status)
		}
	}

	items := make([]user.CollaborationWorkspace, 0)
	err := query.
		Select("id", "name", "remark", "logo_url", "plan", "owner_id", "max_members", "status", "created_at", "updated_at").
		Order("created_at DESC").
		Find(&items).Error
	return items, err
}

func (s *collaborationWorkspaceService) Get(id uuid.UUID) (*user.CollaborationWorkspace, error) {
	t, err := s.collaborationWorkspaceRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCollaborationWorkspaceNotFound
		}
		return nil, err
	}
	return t, nil
}

func (s *collaborationWorkspaceService) Create(req *dto.CollaborationWorkspaceCreateRequest, ownerID *uuid.UUID) (*user.CollaborationWorkspace, error) {
	if ownerID == nil || *ownerID == uuid.Nil {
		return nil, errors.New("invalid owner id")
	}

	plan := req.Plan
	if plan == "" {
		plan = "free"
	}
	status := req.Status
	if status == "" {
		status = "active"
	}
	maxMembers := req.MaxMembers
	if maxMembers <= 0 {
		maxMembers = 5
	}
	t := &user.CollaborationWorkspace{
		Name:       req.Name,
		Remark:     req.Remark,
		LogoURL:    req.LogoURL,
		Plan:       plan,
		OwnerID:    *ownerID,
		MaxMembers: maxMembers,
		Status:     status,
	}
	adminIDs, err := parseAdminUserIDs(req.AdminUserIDs)
	if err != nil {
		return nil, err
	}
	if err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(t).Error; err != nil {
			return err
		}
		if err := ensureCollaborationWorkspaceWorkspaceTx(tx, t); err != nil {
			return err
		}
		return s.syncCollaborationWorkspaceAdminsTx(tx, t.ID, adminIDs, ownerID)
	}); err != nil {
		return nil, err
	}
	if s.refresher != nil {
		if err := s.refresher.RefreshCollaborationWorkspace(t.ID); err != nil {
			return nil, err
		}
	}

	return t, nil
}

func (s *collaborationWorkspaceService) Update(id uuid.UUID, req *dto.CollaborationWorkspaceUpdateRequest) error {
	updates := map[string]interface{}{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Remark != "" {
		updates["remark"] = req.Remark
	}
	if req.LogoURL != "" {
		updates["logo_url"] = req.LogoURL
	}
	if req.Plan != "" {
		updates["plan"] = req.Plan
	}
	if req.Status != "" {
		updates["status"] = req.Status
	}
	if req.MaxMembers > 0 {
		updates["max_members"] = req.MaxMembers
	}
	var adminIDs []uuid.UUID
	if req.AdminUserIDs != nil {
		var err error
		adminIDs, err = parseAdminUserIDs(req.AdminUserIDs)
		if err != nil {
			return err
		}
	}
	if len(updates) == 0 && req.AdminUserIDs == nil {
		return nil
	}
	if err := s.db.Transaction(func(tx *gorm.DB) error {
		var collaborationWorkspace user.CollaborationWorkspace
		if err := tx.Where("id = ?", id).First(&collaborationWorkspace).Error; err != nil {
			return err
		}
		if len(updates) > 0 {
			if err := tx.Model(&collaborationWorkspace).Updates(updates).Error; err != nil {
				return err
			}
		}
		if req.AdminUserIDs != nil {
			return s.syncCollaborationWorkspaceAdminsTx(tx, id, adminIDs, nil)
		}
		return nil
	}); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrCollaborationWorkspaceNotFound
		}
		return err
	}
	if req.AdminUserIDs != nil && s.refresher != nil {
		if err := s.refresher.RefreshCollaborationWorkspace(id); err != nil {
			return err
		}
	}
	return nil
}

func (s *collaborationWorkspaceService) Delete(id uuid.UUID) error {
	if _, err := s.collaborationWorkspaceRepo.GetByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrCollaborationWorkspaceNotFound
		}
		return err
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		var roleIDs []uuid.UUID
		if err := tx.Model(&user.Role{}).Where("collaboration_workspace_id = ?", id).Pluck("id", &roleIDs).Error; err != nil {
			return err
		}

		if err := tx.Where("collaboration_workspace_id = ?", id).Delete(&user.UserActionPermission{}).Error; err != nil {
			return err
		}

		if err := tx.Where("collaboration_workspace_id = ?", id).Delete(&user.CollaborationWorkspaceFeaturePackage{}).Error; err != nil {
			return err
		}

		if err := tx.Where("collaboration_workspace_id = ?", id).Delete(&user.CollaborationWorkspaceBlockedMenu{}).Error; err != nil {
			return err
		}

		if err := tx.Where("collaboration_workspace_id = ?", id).Delete(&user.CollaborationWorkspaceBlockedAction{}).Error; err != nil {
			return err
		}

		if err := tx.Where("collaboration_workspace_id = ?", id).Delete(&user.APIKey{}).Error; err != nil {
			return err
		}

		if err := tx.Where("collaboration_workspace_id = ?", id).Delete(&user.MediaAsset{}).Error; err != nil {
			return err
		}

		if err := tx.Where("collaboration_workspace_id = ?", id).Delete(&user.UserRole{}).Error; err != nil {
			return err
		}

		workspace, err := workspacerolebinding.GetCollaborationWorkspaceByCollaborationWorkspaceID(tx, id)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if err == nil && workspace != nil {
			if err := tx.Where("workspace_id = ?", workspace.ID).Delete(&models.WorkspaceRoleBinding{}).Error; err != nil {
				return err
			}
			if err := tx.Where("workspace_id = ?", workspace.ID).Delete(&models.WorkspaceFeaturePackage{}).Error; err != nil {
				return err
			}
			if err := tx.Where("workspace_id = ?", workspace.ID).Delete(&models.WorkspaceMember{}).Error; err != nil {
				return err
			}
			if err := tx.Delete(&models.Workspace{}, "id = ?", workspace.ID).Error; err != nil {
				return err
			}
		}

		if err := tx.Where("collaboration_workspace_id = ?", id).Delete(&user.CollaborationWorkspaceMember{}).Error; err != nil {
			return err
		}

		if len(roleIDs) > 0 {
			if err := tx.Where("role_id IN ?", roleIDs).Delete(&user.RoleFeaturePackage{}).Error; err != nil {
				return err
			}
			if err := tx.Where("role_id IN ?", roleIDs).Delete(&user.RoleHiddenMenu{}).Error; err != nil {
				return err
			}
			if err := tx.Where("role_id IN ?", roleIDs).Delete(&user.RoleDisabledAction{}).Error; err != nil {
				return err
			}
			if err := tx.Where("role_id IN ?", roleIDs).Delete(&user.RoleDataPermission{}).Error; err != nil {
				return err
			}
			if err := tx.Where("role_id IN ?", roleIDs).Delete(&models.CollaborationWorkspaceRoleAccessSnapshot{}).Error; err != nil {
				return err
			}
			if err := tx.Where("id IN ?", roleIDs).Delete(&user.Role{}).Error; err != nil {
				return err
			}
		}

		if err := tx.Where("collaboration_workspace_id = ?", id).Delete(&models.CollaborationWorkspaceAccessSnapshot{}).Error; err != nil {
			return err
		}

		if err := tx.Delete(&user.CollaborationWorkspace{}, "id = ?", id).Error; err != nil {
			return err
		}

		return nil
	})
}

func (s *collaborationWorkspaceService) ListMembers(collaborationWorkspaceID uuid.UUID, searchParams *user.MemberSearchParams) ([]user.CollaborationWorkspaceMember, error) {
	return s.collaborationWorkspaceMemberRepo.List(collaborationWorkspaceID, searchParams)
}

func (s *collaborationWorkspaceService) AddMember(collaborationWorkspaceID uuid.UUID, req *dto.CollaborationWorkspaceAddMemberRequest, invitedBy *uuid.UUID) error {
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return errors.New("invalid user id")
	}

	existing, err := s.collaborationWorkspaceMemberRepo.GetByUserAndCollaborationWorkspace(userID, collaborationWorkspaceID)
	if err == nil && existing != nil {
		return ErrCollaborationWorkspaceMemberExists
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	roleCode := normalizeCollaborationWorkspaceRoleCode(req.RoleCode)

	member := &user.CollaborationWorkspaceMember{
		CollaborationWorkspaceID: collaborationWorkspaceID,
		UserID:                   userID,
		RoleCode:                 roleCode,
		JoinedAt:                 time.Now(),
	}
	if invitedBy != nil {
		member.InvitedBy = invitedBy
	}

	if err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(member).Error; err != nil {
			return err
		}
		if err := s.syncCollaborationWorkspaceIdentityRoleTx(tx, userID, collaborationWorkspaceID, roleCode); err != nil {
			return err
		}
		return s.syncCollaborationWorkspaceCanonicalAccessTx(tx, userID, collaborationWorkspaceID, roleCode, member.ID, member.Status)
	}); err != nil {
		return err
	}
	if s.refresher != nil {
		if err := s.refresher.RefreshCollaborationWorkspace(collaborationWorkspaceID); err != nil {
			return err
		}
	}

	return nil
}

func (s *collaborationWorkspaceService) RemoveMember(collaborationWorkspaceID, userID uuid.UUID) error {
	member, err := s.collaborationWorkspaceMemberRepo.GetByUserAndCollaborationWorkspace(userID, collaborationWorkspaceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrCollaborationWorkspaceMemberNotFound
		}
		return err
	}

	if err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ? AND collaboration_workspace_id = ?", userID, collaborationWorkspaceID).Delete(&user.UserActionPermission{}).Error; err != nil {
			return err
		}

		if err := tx.Where("user_id = ? AND collaboration_workspace_id = ?", userID, collaborationWorkspaceID).Delete(&user.UserRole{}).Error; err != nil {
			return err
		}

		if err := s.clearCollaborationWorkspaceCanonicalAccessTx(tx, userID, collaborationWorkspaceID); err != nil {
			return err
		}

		if err := tx.Delete(&user.CollaborationWorkspaceMember{}, "id = ?", member.ID).Error; err != nil {
			return err
		}

		return nil
	}); err != nil {
		return err
	}
	if s.refresher != nil {
		if err := s.refresher.RefreshCollaborationWorkspace(collaborationWorkspaceID); err != nil {
			return err
		}
	}
	return nil
}

func (s *collaborationWorkspaceService) UpdateMemberRole(collaborationWorkspaceID, userID uuid.UUID, roleCode string) error {
	member, err := s.collaborationWorkspaceMemberRepo.GetByUserAndCollaborationWorkspace(userID, collaborationWorkspaceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrCollaborationWorkspaceMemberNotFound
		}
		return err
	}

	if err := s.db.Transaction(func(tx *gorm.DB) error {
		normalizedRoleCode := normalizeCollaborationWorkspaceRoleCode(roleCode)
		if err := tx.Model(&user.CollaborationWorkspaceMember{}).Where("id = ?", member.ID).Update("role_code", normalizedRoleCode).Error; err != nil {
			return err
		}
		if err := s.syncCollaborationWorkspaceIdentityRoleTx(tx, userID, collaborationWorkspaceID, normalizedRoleCode); err != nil {
			return err
		}
		return s.syncCollaborationWorkspaceCanonicalAccessTx(tx, userID, collaborationWorkspaceID, normalizedRoleCode, member.ID, member.Status)
	}); err != nil {
		return err
	}
	if s.refresher != nil {
		if err := s.refresher.RefreshCollaborationWorkspace(collaborationWorkspaceID); err != nil {
			return err
		}
	}

	return nil
}

func (s *collaborationWorkspaceService) syncCollaborationWorkspaceAdmins(collaborationWorkspaceID uuid.UUID, adminIDs []uuid.UUID, invitedBy *uuid.UUID) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		return s.syncCollaborationWorkspaceAdminsTx(tx, collaborationWorkspaceID, adminIDs, invitedBy)
	})
}

func (s *collaborationWorkspaceService) syncCollaborationWorkspaceAdminsTx(tx *gorm.DB, collaborationWorkspaceID uuid.UUID, adminIDs []uuid.UUID, invitedBy *uuid.UUID) error {
	var collaborationWorkspace user.CollaborationWorkspace
	if err := tx.Select("id", "owner_id").Where("id = ?", collaborationWorkspaceID).First(&collaborationWorkspace).Error; err != nil {
		return err
	}
	adminIDs = appendIfMissingUUID(adminIDs, collaborationWorkspace.OwnerID)
	var currentMembers []user.CollaborationWorkspaceMember
	if err := tx.Where("collaboration_workspace_id = ?", collaborationWorkspaceID).Find(&currentMembers).Error; err != nil {
		return err
	}

	existingByUser := make(map[uuid.UUID]*user.CollaborationWorkspaceMember, len(currentMembers))
	for i := range currentMembers {
		existingByUser[currentMembers[i].UserID] = &currentMembers[i]
	}

	adminSet := make(map[uuid.UUID]struct{}, len(adminIDs))
	for _, adminID := range adminIDs {
		adminSet[adminID] = struct{}{}
	}

	for _, adminID := range adminIDs {
		member, exists := existingByUser[adminID]
		if exists {
			if member.RoleCode != defaultCollaborationWorkspaceRoleAdminCode {
				if err := tx.Model(&user.CollaborationWorkspaceMember{}).Where("id = ?", member.ID).Update("role_code", defaultCollaborationWorkspaceRoleAdminCode).Error; err != nil {
					return err
				}
				if err := s.syncCollaborationWorkspaceIdentityRoleTx(tx, adminID, collaborationWorkspaceID, defaultCollaborationWorkspaceRoleAdminCode); err != nil {
					return err
				}
				if err := s.syncCollaborationWorkspaceCanonicalAccessTx(tx, adminID, collaborationWorkspaceID, defaultCollaborationWorkspaceRoleAdminCode, member.ID, member.Status); err != nil {
					return err
				}
			}
			continue
		}

		var adminUser user.User
		if err := tx.Where("id = ?", adminID).First(&adminUser).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("admin user not found")
			}
			return err
		}

		member = &user.CollaborationWorkspaceMember{
			CollaborationWorkspaceID: collaborationWorkspaceID,
			UserID:                   adminID,
			RoleCode:                 defaultCollaborationWorkspaceRoleAdminCode,
			JoinedAt:                 time.Now(),
		}
		if invitedBy != nil {
			member.InvitedBy = invitedBy
		}
		if err := tx.Create(member).Error; err != nil {
			return err
		}
		if err := s.syncCollaborationWorkspaceIdentityRoleTx(tx, adminID, collaborationWorkspaceID, defaultCollaborationWorkspaceRoleAdminCode); err != nil {
			return err
		}
		if err := s.syncCollaborationWorkspaceCanonicalAccessTx(tx, adminID, collaborationWorkspaceID, defaultCollaborationWorkspaceRoleAdminCode, member.ID, member.Status); err != nil {
			return err
		}
	}

	for _, member := range currentMembers {
		if member.RoleCode != defaultCollaborationWorkspaceRoleAdminCode {
			continue
		}
		if _, keep := adminSet[member.UserID]; !keep {
			if err := tx.Model(&user.CollaborationWorkspaceMember{}).Where("id = ?", member.ID).Update("role_code", defaultCollaborationWorkspaceRoleMemberCode).Error; err != nil {
				return err
			}
			if err := s.syncCollaborationWorkspaceIdentityRoleTx(tx, member.UserID, collaborationWorkspaceID, defaultCollaborationWorkspaceRoleMemberCode); err != nil {
				return err
			}
			if err := s.syncCollaborationWorkspaceCanonicalAccessTx(tx, member.UserID, collaborationWorkspaceID, defaultCollaborationWorkspaceRoleMemberCode, member.ID, member.Status); err != nil {
				return err
			}
		}
	}

	return nil
}

func parseAdminUserIDs(rawIDs []string) ([]uuid.UUID, error) {
	adminIDs := make([]uuid.UUID, 0, len(rawIDs))
	for _, rawID := range rawIDs {
		rawID = strings.TrimSpace(rawID)
		if rawID == "" {
			continue
		}
		id, err := uuid.Parse(rawID)
		if err != nil {
			return nil, errors.New("invalid admin user id")
		}
		adminIDs = appendIfMissingUUID(adminIDs, id)
	}
	return adminIDs, nil
}

func appendIfMissingUUID(items []uuid.UUID, id uuid.UUID) []uuid.UUID {
	for _, item := range items {
		if item == id {
			return items
		}
	}
	return append(items, id)
}

func normalizeCollaborationWorkspaceRoleCode(roleCode string) string {
	if strings.TrimSpace(roleCode) != "" {
		return roleCode
	}
	return defaultCollaborationWorkspaceRoleMemberCode
}

func (s *collaborationWorkspaceService) syncCollaborationWorkspaceIdentityRole(userID, collaborationWorkspaceID uuid.UUID, roleCode string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		return s.syncCollaborationWorkspaceIdentityRoleTx(tx, userID, collaborationWorkspaceID, roleCode)
	})
}

func (s *collaborationWorkspaceService) syncCollaborationWorkspaceIdentityRoleTx(tx *gorm.DB, userID, collaborationWorkspaceID uuid.UUID, roleCode string) error {
	roleCode = normalizeCollaborationWorkspaceRoleCode(roleCode)
	var roleIDs []uuid.UUID
	if err := tx.Model(&user.Role{}).
		Where("collaboration_workspace_id IS NULL AND code IN ?", []string{defaultCollaborationWorkspaceRoleAdminCode, defaultCollaborationWorkspaceRoleMemberCode}).
		Pluck("id", &roleIDs).Error; err != nil {
		return err
	}
	if len(roleIDs) > 0 {
		if err := tx.Where("user_id = ? AND collaboration_workspace_id = ? AND role_id IN ?", userID, collaborationWorkspaceID, roleIDs).Delete(&user.UserRole{}).Error; err != nil {
			return err
		}
	}

	var roles []user.Role
	if err := tx.Where("code = ? AND collaboration_workspace_id IS NULL", roleCode).Find(&roles).Error; err != nil {
		return err
	}
	if len(roles) == 0 {
		return nil
	}

	return tx.Create(&user.UserRole{UserID: userID, RoleID: roles[0].ID, CollaborationWorkspaceID: &collaborationWorkspaceID}).Error
}

func (s *collaborationWorkspaceService) syncCollaborationWorkspaceCanonicalAccessTx(tx *gorm.DB, userID, collaborationWorkspaceID uuid.UUID, roleCode string, memberID uuid.UUID, memberStatus string) error {
	var collaborationWorkspace user.CollaborationWorkspace
	if err := tx.Select("id", "name", "plan", "owner_id", "status").Where("id = ?", collaborationWorkspaceID).First(&collaborationWorkspace).Error; err != nil {
		return err
	}
	if err := ensureCollaborationWorkspaceWorkspaceTx(tx, &collaborationWorkspace); err != nil {
		return err
	}
	workspace, err := workspacerolebinding.GetCollaborationWorkspaceByCollaborationWorkspaceID(tx, collaborationWorkspaceID)
	if err != nil {
		return err
	}
	if err := ensureCollaborationWorkspaceWorkspaceMemberTx(tx, workspace.ID, userID, memberID, memberStatus, roleCode); err != nil {
		return err
	}

	var roleIDs []uuid.UUID
	if roleCode != "" {
		if err := tx.Model(&user.Role{}).
			Where("code = ? AND collaboration_workspace_id IS NULL AND deleted_at IS NULL", roleCode).
			Pluck("id", &roleIDs).Error; err != nil {
			return err
		}
	}
	return workspacerolebinding.ReplaceCollaborationWorkspaceRoleBindings(tx, collaborationWorkspaceID, userID, roleIDs)
}

func (s *collaborationWorkspaceService) clearCollaborationWorkspaceCanonicalAccessTx(tx *gorm.DB, userID, collaborationWorkspaceID uuid.UUID) error {
	workspace, err := workspacerolebinding.GetCollaborationWorkspaceByCollaborationWorkspaceID(tx, collaborationWorkspaceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}
	if err := workspacerolebinding.ReplaceCollaborationWorkspaceRoleBindings(tx, collaborationWorkspaceID, userID, nil); err != nil {
		return err
	}
	return tx.Where("workspace_id = ? AND user_id = ?", workspace.ID, userID).Delete(&models.WorkspaceMember{}).Error
}

func ensureCollaborationWorkspaceWorkspaceTx(tx *gorm.DB, item *user.CollaborationWorkspace) error {
	if tx == nil || item == nil || item.ID == uuid.Nil {
		return nil
	}
	var existing models.Workspace
	err := tx.Where("workspace_type = ? AND collaboration_workspace_id = ? AND deleted_at IS NULL", models.WorkspaceTypeCollaboration, item.ID).First(&existing).Error
	if err == nil {
		return tx.Model(&existing).Updates(map[string]interface{}{
			"name":        strings.TrimSpace(item.Name),
			"owner_user_id": item.OwnerID,
			"status":      normalizeWorkspaceStatus(item.Status),
			"updated_at":  tx.NowFunc(),
			"meta": models.MetaJSON{
				"legacy_source":                "collaboration_workspace",
				"collaboration_workspace_plan": item.Plan,
			},
		}).Error
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	ownerID := item.OwnerID
	workspace := models.Workspace{
		WorkspaceType:            models.WorkspaceTypeCollaboration,
		Name:                     strings.TrimSpace(item.Name),
		Code:                     buildCollaborationWorkspaceCode(*item),
		OwnerUserID:              &ownerID,
		CollaborationWorkspaceID: &item.ID,
		Status:                   normalizeWorkspaceStatus(item.Status),
		Meta: models.MetaJSON{
			"legacy_source":                "collaboration_workspace",
			"collaboration_workspace_plan": item.Plan,
		},
	}
	return tx.Create(&workspace).Error
}

func ensureCollaborationWorkspaceWorkspaceMemberTx(tx *gorm.DB, workspaceID, userID, collaborationWorkspaceMemberID uuid.UUID, memberStatus, roleCode string) error {
	if tx == nil || workspaceID == uuid.Nil || userID == uuid.Nil {
		return nil
	}
	memberType := mapCollaborationWorkspaceRoleToWorkspaceMemberType(roleCode)
	status := normalizeWorkspaceStatus(memberStatus)
	var existing models.WorkspaceMember
	err := tx.Where("workspace_id = ? AND user_id = ? AND deleted_at IS NULL", workspaceID, userID).First(&existing).Error
	if err == nil {
		return tx.Model(&existing).Updates(map[string]interface{}{
			"member_type":                       memberType,
			"status":                            status,
			"collaboration_workspace_member_id": collaborationWorkspaceMemberID,
			"updated_at":                        tx.NowFunc(),
		}).Error
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	return tx.Create(&models.WorkspaceMember{
		WorkspaceID:                    workspaceID,
		UserID:                         userID,
		MemberType:                     memberType,
		Status:                         status,
		CollaborationWorkspaceMemberID: &collaborationWorkspaceMemberID,
	}).Error
}

func mapCollaborationWorkspaceRoleToWorkspaceMemberType(roleCode string) string {
	switch normalizeCollaborationWorkspaceRoleCode(roleCode) {
	case defaultCollaborationWorkspaceRoleAdminCode:
		return models.WorkspaceMemberAdmin
	default:
		return models.WorkspaceMemberMember
	}
}

func normalizeWorkspaceStatus(status string) string {
	if strings.EqualFold(strings.TrimSpace(status), models.WorkspaceStatusActive) {
		return models.WorkspaceStatusActive
	}
	return models.WorkspaceStatusActive
}

func buildCollaborationWorkspaceCode(collaborationWorkspace user.CollaborationWorkspace) string {
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
	var builder strings.Builder
	lastDash := false
	for _, r := range target {
		isAlphaNum := (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9')
		if isAlphaNum {
			builder.WriteRune(r)
			lastDash = false
			continue
		}
		if !lastDash {
			builder.WriteByte('-')
			lastDash = true
		}
	}
	return strings.Trim(builder.String(), "-")
}

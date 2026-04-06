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
)

var ErrCollaborationWorkspaceNotFound = errors.New("collaboration workspace not found")
var ErrCollaborationWorkspaceMemberExists = errors.New("user already in collaboration workspace")
var ErrCollaborationWorkspaceMemberNotFound = errors.New("member not in collaboration workspace")

type CollaborationWorkspaceService interface {
	List(req *dto.CollaborationWorkspaceListRequest) ([]user.Tenant, int64, error)
	ListOptions(req *dto.CollaborationWorkspaceListRequest) ([]user.Tenant, error)
	Get(id uuid.UUID) (*user.Tenant, error)
	Create(req *dto.CollaborationWorkspaceCreateRequest, ownerID *uuid.UUID) (*user.Tenant, error)
	Update(id uuid.UUID, req *dto.CollaborationWorkspaceUpdateRequest) error
	Delete(id uuid.UUID) error
	ListMembers(tenantID uuid.UUID, searchParams *user.MemberSearchParams) ([]user.TenantMember, error)
	AddMember(tenantID uuid.UUID, req *dto.CollaborationWorkspaceAddMemberRequest, invitedBy *uuid.UUID) error
	RemoveMember(tenantID, userID uuid.UUID) error
	UpdateMemberRole(tenantID, userID uuid.UUID, roleCode string) error
}

const defaultTeamRoleAdminCode = "collaboration_workspace_admin"
const defaultTeamRoleMemberCode = "collaboration_workspace_member"

type collaborationWorkspaceService struct {
	db                               *gorm.DB
	collaborationWorkspaceRepo       user.TenantRepository
	collaborationWorkspaceMemberRepo user.TenantMemberRepository
	userRepo                         user.UserRepository
	roleRepo                         user.RoleRepository
	userRoleRepo                     user.UserRoleRepository
	refresher                        interface {
		RefreshTeam(teamID uuid.UUID) error
	}
	logger *zap.Logger
}

func NewCollaborationWorkspaceService(
	db *gorm.DB,
	collaborationWorkspaceRepo user.TenantRepository,
	collaborationWorkspaceMemberRepo user.TenantMemberRepository,
	userRepo user.UserRepository,
	roleRepo user.RoleRepository,
	userRoleRepo user.UserRoleRepository,
	refresher interface {
		RefreshTeam(teamID uuid.UUID) error
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

func (s *collaborationWorkspaceService) List(req *dto.CollaborationWorkspaceListRequest) ([]user.Tenant, int64, error) {
	if req.Current <= 0 {
		req.Current = 1
	}
	if req.Size <= 0 {
		req.Size = 20
	}
	offset := (req.Current - 1) * req.Size
	return s.collaborationWorkspaceRepo.List(offset, req.Size, req.Name, req.Status)
}

func (s *collaborationWorkspaceService) ListOptions(req *dto.CollaborationWorkspaceListRequest) ([]user.Tenant, error) {
	query := s.db.Model(&user.Tenant{})
	if req != nil {
		if name := strings.TrimSpace(req.Name); name != "" {
			query = query.Where("name LIKE ?", "%"+name+"%")
		}
		if status := strings.TrimSpace(req.Status); status != "" {
			query = query.Where("status = ?", status)
		}
	}

	items := make([]user.Tenant, 0)
	err := query.
		Select("id", "name", "remark", "logo_url", "plan", "owner_id", "max_members", "status", "created_at", "updated_at").
		Order("created_at DESC").
		Find(&items).Error
	return items, err
}

func (s *collaborationWorkspaceService) Get(id uuid.UUID) (*user.Tenant, error) {
	t, err := s.collaborationWorkspaceRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCollaborationWorkspaceNotFound
		}
		return nil, err
	}
	return t, nil
}

func (s *collaborationWorkspaceService) Create(req *dto.CollaborationWorkspaceCreateRequest, ownerID *uuid.UUID) (*user.Tenant, error) {
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
	t := &user.Tenant{
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
		return s.syncTenantAdminsTx(tx, t.ID, adminIDs, ownerID)
	}); err != nil {
		return nil, err
	}
	if s.refresher != nil {
		if err := s.refresher.RefreshTeam(t.ID); err != nil {
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
		var tenant user.Tenant
		if err := tx.Where("id = ?", id).First(&tenant).Error; err != nil {
			return err
		}
		if len(updates) > 0 {
			if err := tx.Model(&tenant).Updates(updates).Error; err != nil {
				return err
			}
		}
		if req.AdminUserIDs != nil {
			return s.syncTenantAdminsTx(tx, id, adminIDs, nil)
		}
		return nil
	}); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrCollaborationWorkspaceNotFound
		}
		return err
	}
	if req.AdminUserIDs != nil && s.refresher != nil {
		if err := s.refresher.RefreshTeam(id); err != nil {
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

		if err := tx.Where("collaboration_workspace_id = ?", id).Delete(&user.TenantMember{}).Error; err != nil {
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

		if err := tx.Delete(&user.Tenant{}, "id = ?", id).Error; err != nil {
			return err
		}

		return nil
	})
}

func (s *collaborationWorkspaceService) ListMembers(tenantID uuid.UUID, searchParams *user.MemberSearchParams) ([]user.TenantMember, error) {
	return s.collaborationWorkspaceMemberRepo.List(tenantID, searchParams)
}

func (s *collaborationWorkspaceService) AddMember(tenantID uuid.UUID, req *dto.CollaborationWorkspaceAddMemberRequest, invitedBy *uuid.UUID) error {
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return errors.New("invalid user id")
	}

	existing, err := s.collaborationWorkspaceMemberRepo.GetByUserAndTenant(userID, tenantID)
	if err == nil && existing != nil {
		return ErrCollaborationWorkspaceMemberExists
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	roleCode := normalizeTenantRoleCode(req.RoleCode)

	member := &user.TenantMember{
		CollaborationWorkspaceID: tenantID,
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
		return s.syncTenantIdentityRoleTx(tx, userID, tenantID, roleCode)
	}); err != nil {
		return err
	}
	if s.refresher != nil {
		if err := s.refresher.RefreshTeam(tenantID); err != nil {
			return err
		}
	}

	return nil
}

func (s *collaborationWorkspaceService) RemoveMember(tenantID, userID uuid.UUID) error {
	member, err := s.collaborationWorkspaceMemberRepo.GetByUserAndTenant(userID, tenantID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrCollaborationWorkspaceMemberNotFound
		}
		return err
	}

	if err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ? AND collaboration_workspace_id = ?", userID, tenantID).Delete(&user.UserActionPermission{}).Error; err != nil {
			return err
		}

		if err := tx.Where("user_id = ? AND collaboration_workspace_id = ?", userID, tenantID).Delete(&user.UserRole{}).Error; err != nil {
			return err
		}

		if err := tx.Delete(&user.TenantMember{}, "id = ?", member.ID).Error; err != nil {
			return err
		}

		return nil
	}); err != nil {
		return err
	}
	if s.refresher != nil {
		if err := s.refresher.RefreshTeam(tenantID); err != nil {
			return err
		}
	}
	return nil
}

func (s *collaborationWorkspaceService) UpdateMemberRole(tenantID, userID uuid.UUID, roleCode string) error {
	member, err := s.collaborationWorkspaceMemberRepo.GetByUserAndTenant(userID, tenantID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrCollaborationWorkspaceMemberNotFound
		}
		return err
	}

	if err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&user.TenantMember{}).Where("id = ?", member.ID).Update("role_code", normalizeTenantRoleCode(roleCode)).Error; err != nil {
			return err
		}
		return s.syncTenantIdentityRoleTx(tx, userID, tenantID, roleCode)
	}); err != nil {
		return err
	}
	if s.refresher != nil {
		if err := s.refresher.RefreshTeam(tenantID); err != nil {
			return err
		}
	}

	return nil
}

func (s *collaborationWorkspaceService) syncTenantAdmins(tenantID uuid.UUID, adminIDs []uuid.UUID, invitedBy *uuid.UUID) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		return s.syncTenantAdminsTx(tx, tenantID, adminIDs, invitedBy)
	})
}

func (s *collaborationWorkspaceService) syncTenantAdminsTx(tx *gorm.DB, tenantID uuid.UUID, adminIDs []uuid.UUID, invitedBy *uuid.UUID) error {
	var tenant user.Tenant
	if err := tx.Select("id", "owner_id").Where("id = ?", tenantID).First(&tenant).Error; err != nil {
		return err
	}
	adminIDs = appendIfMissingUUID(adminIDs, tenant.OwnerID)
	var currentMembers []user.TenantMember
	if err := tx.Where("collaboration_workspace_id = ?", tenantID).Find(&currentMembers).Error; err != nil {
		return err
	}

	existingByUser := make(map[uuid.UUID]*user.TenantMember, len(currentMembers))
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
			if member.RoleCode != defaultTeamRoleAdminCode {
				if err := tx.Model(&user.TenantMember{}).Where("id = ?", member.ID).Update("role_code", defaultTeamRoleAdminCode).Error; err != nil {
					return err
				}
				if err := s.syncTenantIdentityRoleTx(tx, adminID, tenantID, defaultTeamRoleAdminCode); err != nil {
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

		member = &user.TenantMember{
			CollaborationWorkspaceID: tenantID,
			UserID:                   adminID,
			RoleCode:                 defaultTeamRoleAdminCode,
			JoinedAt:                 time.Now(),
		}
		if invitedBy != nil {
			member.InvitedBy = invitedBy
		}
		if err := tx.Create(member).Error; err != nil {
			return err
		}
		if err := s.syncTenantIdentityRoleTx(tx, adminID, tenantID, defaultTeamRoleAdminCode); err != nil {
			return err
		}
	}

	for _, member := range currentMembers {
		if member.RoleCode != defaultTeamRoleAdminCode {
			continue
		}
		if _, keep := adminSet[member.UserID]; !keep {
			if err := tx.Model(&user.TenantMember{}).Where("id = ?", member.ID).Update("role_code", defaultTeamRoleMemberCode).Error; err != nil {
				return err
			}
			if err := s.syncTenantIdentityRoleTx(tx, member.UserID, tenantID, defaultTeamRoleMemberCode); err != nil {
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

func normalizeTenantRoleCode(roleCode string) string {
	if strings.TrimSpace(roleCode) != "" {
		return roleCode
	}
	return defaultTeamRoleMemberCode
}

func (s *collaborationWorkspaceService) syncTenantIdentityRole(userID, tenantID uuid.UUID, roleCode string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		return s.syncTenantIdentityRoleTx(tx, userID, tenantID, roleCode)
	})
}

func (s *collaborationWorkspaceService) syncTenantIdentityRoleTx(tx *gorm.DB, userID, tenantID uuid.UUID, roleCode string) error {
	roleCode = normalizeTenantRoleCode(roleCode)
	var roleIDs []uuid.UUID
	if err := tx.Model(&user.Role{}).
		Where("collaboration_workspace_id IS NULL AND code IN ?", []string{defaultTeamRoleAdminCode, defaultTeamRoleMemberCode}).
		Pluck("id", &roleIDs).Error; err != nil {
		return err
	}
	if len(roleIDs) > 0 {
		if err := tx.Where("user_id = ? AND collaboration_workspace_id = ? AND role_id IN ?", userID, tenantID, roleIDs).Delete(&user.UserRole{}).Error; err != nil {
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

	return tx.Create(&user.UserRole{UserID: userID, RoleID: roles[0].ID, CollaborationWorkspaceID: &tenantID}).Error
}

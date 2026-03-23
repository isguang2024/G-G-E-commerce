package tenant

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/gg-ecommerce/backend/internal/api/dto"
	"github.com/gg-ecommerce/backend/internal/modules/system/user"
)

var ErrTenantNotFound = errors.New("tenant not found")
var ErrTenantMemberExists = errors.New("user already in team")
var ErrTenantMemberNotFound = errors.New("member not in team")

type TenantService interface {
	List(req *dto.TenantListRequest) ([]user.Tenant, int64, error)
	Get(id uuid.UUID) (*user.Tenant, error)
	Create(req *dto.TenantCreateRequest, ownerID *uuid.UUID) (*user.Tenant, error)
	Update(id uuid.UUID, req *dto.TenantUpdateRequest) error
	Delete(id uuid.UUID) error
	ListMembers(tenantID uuid.UUID, searchParams *user.MemberSearchParams) ([]user.TenantMember, error)
	AddMember(tenantID uuid.UUID, req *dto.TenantAddMemberRequest, invitedBy *uuid.UUID) error
	RemoveMember(tenantID, userID uuid.UUID) error
	UpdateMemberRole(tenantID, userID uuid.UUID, roleCode string) error
}

const defaultTeamRoleAdminCode = "team_admin"
const defaultTeamRoleMemberCode = "team_member"

type tenantService struct {
	db               *gorm.DB
	tenantRepo       user.TenantRepository
	tenantMemberRepo user.TenantMemberRepository
	userRepo         user.UserRepository
	roleRepo         user.RoleRepository
	userRoleRepo     user.UserRoleRepository
	refresher        interface {
		RefreshTeam(teamID uuid.UUID) error
	}
	logger *zap.Logger
}

func NewTenantService(
	db *gorm.DB,
	tenantRepo user.TenantRepository,
	tenantMemberRepo user.TenantMemberRepository,
	userRepo user.UserRepository,
	roleRepo user.RoleRepository,
	userRoleRepo user.UserRoleRepository,
	refresher interface {
		RefreshTeam(teamID uuid.UUID) error
	},
	logger *zap.Logger,
) TenantService {
	return &tenantService{
		db:               db,
		tenantRepo:       tenantRepo,
		tenantMemberRepo: tenantMemberRepo,
		userRepo:         userRepo,
		roleRepo:         roleRepo,
		userRoleRepo:     userRoleRepo,
		refresher:        refresher,
		logger:           logger,
	}
}

func (s *tenantService) List(req *dto.TenantListRequest) ([]user.Tenant, int64, error) {
	if req.Current <= 0 {
		req.Current = 1
	}
	if req.Size <= 0 {
		req.Size = 20
	}
	offset := (req.Current - 1) * req.Size
	return s.tenantRepo.List(offset, req.Size, req.Name, req.Status)
}

func (s *tenantService) Get(id uuid.UUID) (*user.Tenant, error) {
	t, err := s.tenantRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTenantNotFound
		}
		return nil, err
	}
	return t, nil
}

func (s *tenantService) Create(req *dto.TenantCreateRequest, ownerID *uuid.UUID) (*user.Tenant, error) {
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
	if err := s.tenantRepo.Create(t); err != nil {
		return nil, err
	}

	adminIDs, err := parseAdminUserIDs(req.AdminUserIDs)
	if err != nil {
		return nil, err
	}
	if err := s.syncTenantAdmins(t.ID, adminIDs, ownerID); err != nil {
		return nil, err
	}
	if s.refresher != nil {
		if err := s.refresher.RefreshTeam(t.ID); err != nil {
			return nil, err
		}
	}

	return t, nil
}

func (s *tenantService) Update(id uuid.UUID, req *dto.TenantUpdateRequest) error {
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
	if len(updates) == 0 {
		return nil
	}
	if err := s.tenantRepo.UpdateWithMap(id, updates); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrTenantNotFound
		}
		return err
	}

	if req.AdminUserIDs != nil {
		adminIDs, err := parseAdminUserIDs(req.AdminUserIDs)
		if err != nil {
			return err
		}
		if err := s.syncTenantAdmins(id, adminIDs, nil); err != nil {
			return err
		}
		if s.refresher != nil {
			if err := s.refresher.RefreshTeam(id); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *tenantService) Delete(id uuid.UUID) error {
	if _, err := s.tenantRepo.GetByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrTenantNotFound
		}
		return err
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("tenant_id = ?", id).Delete(&user.UserActionPermission{}).Error; err != nil {
			return err
		}

		if err := tx.Where("tenant_id = ?", id).Delete(&user.TenantActionPermission{}).Error; err != nil {
			return err
		}

		if err := tx.Where("tenant_id = ?", id).Delete(&user.TeamManualActionPermission{}).Error; err != nil {
			return err
		}

		if err := tx.Where("team_id = ?", id).Delete(&user.TeamFeaturePackage{}).Error; err != nil {
			return err
		}

		if err := tx.Where("tenant_id = ?", id).Delete(&user.APIKey{}).Error; err != nil {
			return err
		}

		if err := tx.Where("tenant_id = ?", id).Delete(&user.MediaAsset{}).Error; err != nil {
			return err
		}

		if err := tx.Where("tenant_id = ?", id).Delete(&user.UserRole{}).Error; err != nil {
			return err
		}

		if err := tx.Where("tenant_id = ?", id).Delete(&user.TenantMember{}).Error; err != nil {
			return err
		}

		if err := tx.Delete(&user.Tenant{}, "id = ?", id).Error; err != nil {
			return err
		}

		return nil
	})
}

func (s *tenantService) ListMembers(tenantID uuid.UUID, searchParams *user.MemberSearchParams) ([]user.TenantMember, error) {
	return s.tenantMemberRepo.List(tenantID, searchParams)
}

func (s *tenantService) AddMember(tenantID uuid.UUID, req *dto.TenantAddMemberRequest, invitedBy *uuid.UUID) error {
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return errors.New("invalid user id")
	}

	existing, err := s.tenantMemberRepo.GetByUserAndTenant(userID, tenantID)
	if err == nil && existing != nil {
		return ErrTenantMemberExists
	}

	roleCode := normalizeTenantRoleCode(req.RoleCode, req.Role)

	member := &user.TenantMember{
		TenantID: tenantID,
		UserID:   userID,
		RoleCode: roleCode,
		JoinedAt: time.Now(),
	}
	if invitedBy != nil {
		member.InvitedBy = invitedBy
	}

	if err := s.tenantMemberRepo.Create(member); err != nil {
		return err
	}

	if err := s.syncTenantIdentityRole(userID, tenantID, roleCode); err != nil {
		return err
	}
	if s.refresher != nil {
		if err := s.refresher.RefreshTeam(tenantID); err != nil {
			return err
		}
	}

	return nil
}

func (s *tenantService) RemoveMember(tenantID, userID uuid.UUID) error {
	member, err := s.tenantMemberRepo.GetByUserAndTenant(userID, tenantID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrTenantMemberNotFound
		}
		return err
	}

	if err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ? AND tenant_id = ?", userID, tenantID).Delete(&user.UserActionPermission{}).Error; err != nil {
			return err
		}

		if err := tx.Where("user_id = ? AND tenant_id = ?", userID, tenantID).Delete(&user.UserRole{}).Error; err != nil {
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

func (s *tenantService) UpdateMemberRole(tenantID, userID uuid.UUID, roleCode string) error {
	member, err := s.tenantMemberRepo.GetByUserAndTenant(userID, tenantID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrTenantMemberNotFound
		}
		return err
	}

	if err := s.tenantMemberRepo.UpdateRole(member.ID, roleCode); err != nil {
		return err
	}

	if err := s.syncTenantIdentityRole(userID, tenantID, roleCode); err != nil {
		return err
	}
	if s.refresher != nil {
		if err := s.refresher.RefreshTeam(tenantID); err != nil {
			return err
		}
	}

	return nil
}

func (s *tenantService) syncTenantAdmins(tenantID uuid.UUID, adminIDs []uuid.UUID, invitedBy *uuid.UUID) error {
	currentMembers, err := s.tenantMemberRepo.List(tenantID, nil)
	if err != nil {
		return err
	}

	existingByUser := make(map[uuid.UUID]*user.TenantMember, len(currentMembers))
	for i := range currentMembers {
		member := currentMembers[i]
		existingByUser[member.UserID] = &member
	}

	adminSet := make(map[uuid.UUID]struct{}, len(adminIDs))
	for _, adminID := range adminIDs {
		adminSet[adminID] = struct{}{}
	}

	for _, adminID := range adminIDs {
		member, exists := existingByUser[adminID]
		if exists {
			if member.RoleCode != defaultTeamRoleAdminCode {
				if err := s.tenantMemberRepo.UpdateRole(member.ID, defaultTeamRoleAdminCode); err != nil {
					return err
				}
				if err := s.syncTenantIdentityRole(adminID, tenantID, defaultTeamRoleAdminCode); err != nil {
					return err
				}
			}
			continue
		}

		if _, err := s.userRepo.GetByID(adminID); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("admin user not found")
			}
			return err
		}

		member = &user.TenantMember{
			TenantID: tenantID,
			UserID:   adminID,
			RoleCode: defaultTeamRoleAdminCode,
			JoinedAt: time.Now(),
		}
		if invitedBy != nil {
			member.InvitedBy = invitedBy
		}
		if err := s.tenantMemberRepo.Create(member); err != nil {
			return err
		}
		if err := s.syncTenantIdentityRole(adminID, tenantID, defaultTeamRoleAdminCode); err != nil {
			return err
		}
	}

	for _, member := range currentMembers {
		if member.RoleCode != defaultTeamRoleAdminCode {
			continue
		}
		if _, keep := adminSet[member.UserID]; !keep {
			if err := s.tenantMemberRepo.UpdateRole(member.ID, defaultTeamRoleMemberCode); err != nil {
				return err
			}
			if err := s.syncTenantIdentityRole(member.UserID, tenantID, defaultTeamRoleMemberCode); err != nil {
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

func normalizeTenantRoleCode(roleCode string, role string) string {
	if strings.TrimSpace(roleCode) != "" {
		return roleCode
	}
	if strings.TrimSpace(role) != "" {
		return role
	}
	return defaultTeamRoleMemberCode
}

func (s *tenantService) syncTenantIdentityRole(userID, tenantID uuid.UUID, roleCode string) error {
	roleCode = normalizeTenantRoleCode(roleCode, "")
	if err := s.userRoleRepo.RemoveRolesByCodes(
		userID,
		&tenantID,
		[]string{defaultTeamRoleAdminCode, defaultTeamRoleMemberCode},
	); err != nil {
		return err
	}

	roles, err := s.roleRepo.FindByCode(roleCode)
	if err != nil {
		return err
	}
	if len(roles) == 0 {
		return nil
	}

	return s.userRoleRepo.AssignRole(userID, roles[0].ID, &tenantID)
}

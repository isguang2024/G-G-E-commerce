package tenant

import (
	"errors"
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
	tenantRepo       user.TenantRepository
	tenantMemberRepo user.TenantMemberRepository
	userRepo         user.UserRepository
	roleRepo         user.RoleRepository
	userRoleRepo     user.UserRoleRepository
	logger           *zap.Logger
}

func NewTenantService(tenantRepo user.TenantRepository, tenantMemberRepo user.TenantMemberRepository, userRepo user.UserRepository, roleRepo user.RoleRepository, userRoleRepo user.UserRoleRepository, logger *zap.Logger) TenantService {
	return &tenantService{
		tenantRepo:       tenantRepo,
		tenantMemberRepo: tenantMemberRepo,
		userRepo:         userRepo,
		roleRepo:         roleRepo,
		userRoleRepo:     userRoleRepo,
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
	return nil
}

func (s *tenantService) Delete(id uuid.UUID) error {
	if err := s.tenantRepo.Delete(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrTenantNotFound
		}
		return err
	}
	return nil
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

	member := &user.TenantMember{
		TenantID: tenantID,
		UserID:   userID,
		RoleCode: req.RoleCode,
		JoinedAt: time.Now(),
	}
	if invitedBy != nil {
		member.InvitedBy = invitedBy
	}

	if err := s.tenantMemberRepo.Create(member); err != nil {
		return err
	}

	if req.RoleCode != "" {
		if err := s.assignTeamRole(userID, tenantID, req.RoleCode); err != nil {
			s.logger.Error("failed to assign team role", zap.Error(err))
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

	if err := s.tenantMemberRepo.Delete(member.ID); err != nil {
		return err
	}

	if err := s.userRoleRepo.RemoveUserRole(userID, &tenantID); err != nil {
		s.logger.Error("failed to remove user team role", zap.Error(err))
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

	if err := s.assignTeamRole(userID, tenantID, roleCode); err != nil {
		return err
	}

	return nil
}

func (s *tenantService) assignTeamRole(userID, tenantID uuid.UUID, roleCode string) error {
	roles, err := s.roleRepo.FindByCode(roleCode)
	if err != nil {
		return err
	}
	if len(roles) == 0 {
		return errors.New("role not found")
	}
	roleID := roles[0].ID
	return s.userRoleRepo.AssignRole(userID, roleID, &tenantID)
}

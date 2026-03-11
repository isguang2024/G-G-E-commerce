package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/gg-ecommerce/backend/internal/api/dto"
	"github.com/gg-ecommerce/backend/internal/model"
	"github.com/gg-ecommerce/backend/internal/repository"
)

var ErrTenantNotFound = errors.New("tenant not found")
var ErrTenantMemberExists = errors.New("user already in team")
var ErrTenantMemberNotFound = errors.New("member not in team")

type TenantService interface {
	List(req *dto.TenantListRequest) ([]model.Tenant, int64, error)
	Get(id uuid.UUID) (*model.Tenant, error)
	Create(req *dto.TenantCreateRequest, ownerID *uuid.UUID) (*model.Tenant, error)
	Update(id uuid.UUID, req *dto.TenantUpdateRequest) error
	Delete(id uuid.UUID) error
	ListMembers(tenantID uuid.UUID, searchParams *repository.MemberSearchParams) ([]model.TenantMember, error)
	AddMember(tenantID uuid.UUID, req *dto.TenantAddMemberRequest, invitedBy *uuid.UUID) error
	RemoveMember(tenantID, userID uuid.UUID) error
	UpdateMemberRole(tenantID, userID uuid.UUID, roleCode string) error
}

// 团队内身份使用的默认角色编码（全局 scope=team 角色，通过 user_roles 关联）
const defaultTeamRoleAdminCode = "team_admin"
const defaultTeamRoleMemberCode = "team_member"

type tenantService struct {
	tenantRepo       repository.TenantRepository
	tenantMemberRepo repository.TenantMemberRepository
	userRepo         repository.UserRepository
	roleRepo         repository.RoleRepository
	userRoleRepo     repository.UserRoleRepository
	logger           *zap.Logger
}

func NewTenantService(tenantRepo repository.TenantRepository, tenantMemberRepo repository.TenantMemberRepository, userRepo repository.UserRepository, roleRepo repository.RoleRepository, userRoleRepo repository.UserRoleRepository, logger *zap.Logger) TenantService {
	return &tenantService{tenantRepo: tenantRepo, tenantMemberRepo: tenantMemberRepo, userRepo: userRepo, roleRepo: roleRepo, userRoleRepo: userRoleRepo, logger: logger}
}

func (s *tenantService) List(req *dto.TenantListRequest) ([]model.Tenant, int64, error) {
	if req.Current <= 0 {
		req.Current = 1
	}
	if req.Size <= 0 {
		req.Size = 20
	}
	offset := (req.Current - 1) * req.Size
	return s.tenantRepo.List(offset, req.Size, req.Name, req.Status)
}

func (s *tenantService) Get(id uuid.UUID) (*model.Tenant, error) {
	t, err := s.tenantRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTenantNotFound
		}
		return nil, err
	}
	return t, nil
}

func (s *tenantService) Create(req *dto.TenantCreateRequest, ownerID *uuid.UUID) (*model.Tenant, error) {
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
	t := &model.Tenant{
		Name:       req.Name,
		Remark:     req.Remark,
		LogoURL:    req.LogoURL,
		Plan:       plan,
		OwnerID:    ownerID,
		MaxMembers: maxMembers,
		Status:     status,
	}
	if err := s.tenantRepo.Create(t); err != nil {
		return nil, err
	}

	// 注意：不再自动将创建者添加为团队成员
	// 团队创建者只是 owner（拥有者），需要手动在团队中添加成员
	// 如果需要指定管理员，通过 req.AdminUserIDs 添加

	// 添加指定的管理员
	adminRole, _ := s.roleRepo.GetByCode(defaultTeamRoleAdminCode)
	if len(req.AdminUserIDs) > 0 && adminRole != nil {
		for _, adminUserIDStr := range req.AdminUserIDs {
			adminUID, err := uuid.Parse(adminUserIDStr)
			if err != nil {
				continue // Skip invalid UUIDs
			}
			if ownerID != nil && adminUID == *ownerID {
				continue // Owner is already added as admin
			}
			now := t.CreatedAt
			_ = s.tenantMemberRepo.Upsert(&model.TenantMember{
				TenantID: t.ID,
				UserID:   adminUID,
				RoleID:   &adminRole.ID,
				Status:   "active",
				JoinedAt: &now,
			})
		}
	}
	return s.tenantRepo.GetByID(t.ID)
}

func (s *tenantService) Update(id uuid.UUID, req *dto.TenantUpdateRequest) error {
	t, err := s.tenantRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrTenantNotFound
		}
		return err
	}
	if req.Name != "" {
		t.Name = req.Name
	}
	t.Remark = req.Remark
	if req.LogoURL != "" {
		t.LogoURL = req.LogoURL
	}
	if req.Plan != "" {
		t.Plan = req.Plan
	}
	if req.MaxMembers > 0 {
		t.MaxMembers = req.MaxMembers
	}
	if req.Status != "" {
		t.Status = req.Status
	}

	if err := s.tenantRepo.Update(t); err != nil {
		return err
	}

	if req.AdminUserIDs != nil {
		adminRole, _ := s.roleRepo.GetByCode(defaultTeamRoleAdminCode)
		if adminRole != nil {
			for _, adminUserIDStr := range req.AdminUserIDs {
				adminUID, err := uuid.Parse(adminUserIDStr)
				if err != nil {
					continue
				}
				now := time.Now()
				_ = s.tenantMemberRepo.Upsert(&model.TenantMember{
					TenantID: t.ID,
					UserID:   adminUID,
					RoleID:   &adminRole.ID,
					Status:   "active",
					JoinedAt: &now,
				})
			}
		}
	}
	return nil
}

func (s *tenantService) Delete(id uuid.UUID) error {
	_, err := s.tenantRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrTenantNotFound
		}
		return err
	}
	return s.tenantRepo.Delete(id)
}

func (s *tenantService) ListMembers(tenantID uuid.UUID, searchParams *repository.MemberSearchParams) ([]model.TenantMember, error) {
	_, err := s.tenantRepo.GetByID(tenantID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTenantNotFound
		}
		return nil, err
	}

	// 如果有搜索参数，使用搜索查询
	if searchParams != nil && (searchParams.UserID != "" || searchParams.UserName != "" || searchParams.Role != "") {
		searchParams.TenantID = tenantID
		return s.tenantMemberRepo.ListByTenantIDWithSearch(*searchParams)
	}

	return s.tenantMemberRepo.ListByTenantID(tenantID)
}

func (s *tenantService) AddMember(tenantID uuid.UUID, req *dto.TenantAddMemberRequest, invitedBy *uuid.UUID) error {
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return fmt.Errorf("invalid user_id")
	}
	_, err = s.tenantRepo.GetByID(tenantID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrTenantNotFound
		}
		return err
	}
	_, err = s.tenantMemberRepo.Get(tenantID, userID)
	if err == nil {
		return ErrTenantMemberExists
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	roleCode := req.Role
	if roleCode == "" {
		roleCode = defaultTeamRoleMemberCode
	}
	role, err := s.roleRepo.GetByCode(roleCode)
	if err != nil || role == nil {
		return fmt.Errorf("role not found: %s", roleCode)
	}
	return s.tenantMemberRepo.Add(&model.TenantMember{
		TenantID:  tenantID,
		UserID:    userID,
		RoleID:    &role.ID,
		Status:    "active",
		InvitedBy: invitedBy,
	})
}

func (s *tenantService) RemoveMember(tenantID, userID uuid.UUID) error {
	_, err := s.tenantMemberRepo.Get(tenantID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrTenantMemberNotFound
		}
		return err
	}
	return s.tenantMemberRepo.Remove(tenantID, userID)
}

func (s *tenantService) UpdateMemberRole(tenantID, userID uuid.UUID, roleCode string) error {
	_, err := s.tenantMemberRepo.Get(tenantID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrTenantMemberNotFound
		}
		return err
	}
	role, err := s.roleRepo.GetByCode(roleCode)
	if err != nil || role == nil {
		return fmt.Errorf("role not found: %s", roleCode)
	}
	return s.tenantMemberRepo.UpdateRole(tenantID, userID, &role.ID)
}

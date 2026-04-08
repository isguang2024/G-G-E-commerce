package user

import (
	"errors"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/gg-ecommerce/backend/internal/api/dto"
	"github.com/gg-ecommerce/backend/internal/modules/system/models"
	appctx "github.com/gg-ecommerce/backend/internal/pkg/appctx"
	"github.com/gg-ecommerce/backend/internal/pkg/collaborationworkspaceboundary"
	"github.com/gg-ecommerce/backend/internal/pkg/password"
	"github.com/gg-ecommerce/backend/internal/pkg/permissionrefresh"
	"github.com/gg-ecommerce/backend/internal/pkg/platformaccess"
	"github.com/gg-ecommerce/backend/internal/pkg/workspacerolebinding"
)

var (
	ErrUserNotFound = errors.New("用户不存在")
	ErrUserExists   = errors.New("用户名已存在")
	ErrEmailExists  = errors.New("邮箱已存在")
)

type UserService interface {
	List(req *dto.UserListRequest) ([]User, int64, error)
	Get(id uuid.UUID) (*User, error)
	GetByIDs(ids []uuid.UUID) ([]User, error)
	Create(req *dto.UserCreateRequest) (*User, error)
	Update(id uuid.UUID, req *dto.UserUpdateRequest) error
	Delete(id uuid.UUID) error
	AssignRoles(id uuid.UUID, roleIDs []string) error
}

type userService struct {
	userRepo UserRepository
	db       *gorm.DB
	roleRepo interface {
		GetByID(id uuid.UUID) (*Role, error)
	}
	refresher permissionrefresh.Service
	logger    *zap.Logger
}

func NewUserService(db *gorm.DB, userRepo UserRepository, roleRepo interface {
	GetByID(id uuid.UUID) (*Role, error)
}, refresher permissionrefresh.Service, logger *zap.Logger) UserService {
	return &userService{db: db, userRepo: userRepo, roleRepo: roleRepo, refresher: refresher, logger: logger}
}

func (s *userService) List(req *dto.UserListRequest) ([]User, int64, error) {
	if req.Current <= 0 {
		req.Current = 1
	}
	if req.Size <= 0 {
		req.Size = 20
	}
	offset := (req.Current - 1) * req.Size
	return s.userRepo.List(offset, req.Size, req.UserName, req.UserPhone, req.UserEmail, req.Status, req.RoleID, req.ID, req.RegisterSource, req.InvitedBy)
}

func (s *userService) Get(id uuid.UUID) (*User, error) {
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

func (s *userService) GetByIDs(ids []uuid.UUID) ([]User, error) {
	return s.userRepo.GetByIDs(ids)
}

func (s *userService) Create(req *dto.UserCreateRequest) (*User, error) {
	exists, err := s.userRepo.ExistsByUsername(req.Username)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrUserExists
	}
	email := strings.TrimSpace(req.Email)
	if email != "" {
		exists, err := s.userRepo.ExistsByEmail(email)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, ErrEmailExists
		}
	}
	hash, err := password.Hash(req.Password)
	if err != nil {
		return nil, err
	}
	status := req.Status
	if status == "" {
		status = "active"
	}
	user := &User{
		Username:       req.Username,
		PasswordHash:   hash,
		Email:          email,
		Nickname:       req.Nickname,
		Phone:          req.Phone,
		SystemRemark:   req.SystemRemark,
		RegisterSource: "admin",
		Status:         status,
	}
	roleUUIDs, _ := parseUUIDs(req.RoleIDs)
	if err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(user).Error; err != nil {
			return err
		}
		if len(roleUUIDs) == 0 {
			return nil
		}
		return replaceGlobalUserRoles(tx, user.ID, roleUUIDs)
	}); err != nil {
		return nil, err
	}
	if s.refresher != nil {
		if err := s.refresher.RefreshPersonalWorkspaceUser(user.ID); err != nil {
			return nil, err
		}
	}
	return s.userRepo.GetByID(user.ID)
}

func (s *userService) Update(id uuid.UUID, req *dto.UserUpdateRequest) error {
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrUserNotFound
		}
		return err
	}
	email := strings.TrimSpace(req.Email)
	if email != "" {
		exists, err := s.userRepo.GetByEmail(email)
		if err != nil && err != gorm.ErrRecordNotFound {
			return err
		}
		if err == nil && exists.ID != id {
			return ErrEmailExists
		}
	}
	user.Email = email
	user.Nickname = req.Nickname
	user.Phone = req.Phone
	user.SystemRemark = req.SystemRemark
	if req.Status != "" {
		user.Status = req.Status
	}
	roleUUIDs := []uuid.UUID(nil)
	if req.RoleIDs != nil {
		roleUUIDs, err = parseUUIDs(req.RoleIDs)
		if err != nil {
			return err
		}
	}
	if err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.User{}).Where("id = ?", id).Updates(map[string]interface{}{
			"email":         user.Email,
			"nickname":      user.Nickname,
			"phone":         user.Phone,
			"system_remark": user.SystemRemark,
			"status":        user.Status,
		}).Error; err != nil {
			return err
		}
		if req.RoleIDs == nil {
			return nil
		}
		return replaceGlobalUserRoles(tx, id, roleUUIDs)
	}); err != nil {
		return err
	}
	if s.refresher != nil {
		return s.refresher.RefreshPersonalWorkspaceUser(id)
	}
	return nil
}

func (s *userService) Delete(id uuid.UUID) error {
	_, err := s.userRepo.GetByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrUserNotFound
		}
		return err
	}
	return s.userRepo.Delete(id)
}

func (s *userService) AssignRoles(id uuid.UUID, roleIDs []string) error {
	_, err := s.userRepo.GetByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrUserNotFound
		}
		return err
	}
	roleUUIDs, err := parseUUIDs(roleIDs)
	if err != nil {
		return err
	}
	return s.userRepo.ReplaceRoles(id, roleUUIDs)
}

func parseUUIDs(ids []string) ([]uuid.UUID, error) {
	var result []uuid.UUID
	for _, s := range ids {
		if s == "" {
			continue
		}
		id, err := uuid.Parse(s)
		if err != nil {
			return nil, err
		}
		result = append(result, id)
	}
	return result, nil
}

func replaceGlobalUserRoles(tx *gorm.DB, userID uuid.UUID, roleIDs []uuid.UUID) error {
	if err := workspacerolebinding.ReplacePersonalRoleBindings(tx, userID, roleIDs); err != nil {
		return err
	}
	if err := tx.Where("user_id = ? AND collaboration_workspace_id IS NULL", userID).Delete(&models.UserRole{}).Error; err != nil {
		return err
	}
	if len(roleIDs) == 0 {
		return nil
	}
	items := make([]models.UserRole, 0, len(roleIDs))
	seen := make(map[uuid.UUID]struct{}, len(roleIDs))
	for _, roleID := range roleIDs {
		if roleID == uuid.Nil {
			continue
		}
		if _, ok := seen[roleID]; ok {
			continue
		}
		seen[roleID] = struct{}{}
		items = append(items, models.UserRole{UserID: userID, RoleID: roleID})
	}
	if len(items) == 0 {
		return nil
	}
	return tx.Create(&items).Error
}

type PermissionService interface {
	GetUserMenuIDs(userID uuid.UUID, collaborationWorkspaceID *uuid.UUID) ([]uuid.UUID, error)
	GetUserMenuIDsInApp(userID uuid.UUID, collaborationWorkspaceID *uuid.UUID, appKey string) ([]uuid.UUID, error)
}

type permissionService struct {
	userRepo           UserRepository
	roleRepo           RoleRepository
	userRoleRepo       UserRoleRepository
	rolePackageRepo    RoleFeaturePackageRepository
	userPackageRepo    UserFeaturePackageRepository
	packageRepo        FeaturePackageRepository
	packageMenuRepo    FeaturePackageMenuRepository
	packageBundleRepo  FeaturePackageBundleRepository
	roleHiddenMenuRepo RoleHiddenMenuRepository
	userHiddenMenuRepo UserHiddenMenuRepository
	menuRepo           MenuRepository
	boundaryService    collaborationworkspaceboundary.Service
	personalWorkspaceAccessService platformaccess.Service
}

func NewPermissionService(
	userRepo UserRepository,
	roleRepo RoleRepository,
	userRoleRepo UserRoleRepository,
	rolePackageRepo RoleFeaturePackageRepository,
	userPackageRepo UserFeaturePackageRepository,
	packageRepo FeaturePackageRepository,
	packageMenuRepo FeaturePackageMenuRepository,
	packageBundleRepo FeaturePackageBundleRepository,
	roleHiddenMenuRepo RoleHiddenMenuRepository,
	userHiddenMenuRepo UserHiddenMenuRepository,
	menuRepo MenuRepository,
	boundaryService collaborationworkspaceboundary.Service,
	personalWorkspaceAccessService platformaccess.Service,
) PermissionService {
	return &permissionService{
		userRepo:           userRepo,
		roleRepo:           roleRepo,
		userRoleRepo:       userRoleRepo,
		rolePackageRepo:    rolePackageRepo,
		userPackageRepo:    userPackageRepo,
		packageRepo:        packageRepo,
		packageMenuRepo:    packageMenuRepo,
		packageBundleRepo:  packageBundleRepo,
		roleHiddenMenuRepo: roleHiddenMenuRepo,
		userHiddenMenuRepo: userHiddenMenuRepo,
		menuRepo:                       menuRepo,
		boundaryService:                boundaryService,
		personalWorkspaceAccessService: personalWorkspaceAccessService,
	}
}

func (s *permissionService) GetUserMenuIDs(userID uuid.UUID, collaborationWorkspaceID *uuid.UUID) ([]uuid.UUID, error) {
	return s.GetUserMenuIDsInApp(userID, collaborationWorkspaceID, models.DefaultAppKey)
}

func (s *permissionService) GetUserMenuIDsInApp(userID uuid.UUID, collaborationWorkspaceID *uuid.UUID, appKey string) ([]uuid.UUID, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, err
	}

	if user.Status != "active" {
		return []uuid.UUID{}, nil
	}
	if collaborationWorkspaceID != nil {
		return s.getCollaborationWorkspaceUserMenuIDs(userID, *collaborationWorkspaceID, appKey)
	}
	return s.getPersonalWorkspaceMenuIDs(userID, appKey)
}

func (s *permissionService) getPersonalWorkspaceMenuIDs(userID uuid.UUID, appKey string) ([]uuid.UUID, error) {
	if s.personalWorkspaceAccessService != nil {
		snapshot, err := s.personalWorkspaceAccessService.GetSnapshot(userID, appctx.NormalizeAppKey(appKey))
		if err != nil {
			return nil, err
		}
		return s.finalizeMenuIDs(snapshot.MenuIDs, appKey)
	}
	return s.finalizeMenuIDs(nil, appKey)
}

func (s *permissionService) getCollaborationWorkspaceUserMenuIDs(userID, collaborationWorkspaceID uuid.UUID, appKey string) ([]uuid.UUID, error) {
	roleIDs, err := s.userRoleRepo.GetEffectiveActiveRoleIDsByUserAndCollaborationWorkspace(userID, &collaborationWorkspaceID)
	if err != nil {
		return nil, err
	}
	if len(roleIDs) == 0 {
		return s.finalizeMenuIDs(nil, appKey)
	}
	roles, err := s.roleRepo.GetByIDs(roleIDs)
	if err != nil {
		return nil, err
	}
	roleMap := make(map[uuid.UUID]Role, len(roles))
	for _, role := range roles {
		roleMap[role.ID] = role
	}
	menuSet := make(map[uuid.UUID]struct{})
	for _, roleID := range roleIDs {
		role, ok := roleMap[roleID]
		if !ok {
			continue
		}
		snapshot, snapshotErr := s.boundaryService.GetRoleSnapshot(collaborationWorkspaceID, roleID, role.CollaborationWorkspaceID == nil, appctx.NormalizeAppKey(appKey))
		if snapshotErr != nil {
			return nil, snapshotErr
		}
		for _, menuID := range snapshot.MenuIDs {
			menuSet[menuID] = struct{}{}
		}
	}
	menuIDs := make([]uuid.UUID, 0, len(menuSet))
	for menuID := range menuSet {
		menuIDs = append(menuIDs, menuID)
	}
	return s.finalizeMenuIDs(menuIDs, appKey)
}

func (s *permissionService) getPersonalWorkspacePackageMenuIDs(packageIDs []uuid.UUID) ([]uuid.UUID, error) {
	expandedIDs, err := s.expandPackageIDs(packageIDs, "personal")
	if err != nil {
		return nil, err
	}
	return s.packageMenuRepo.GetMenuIDsByPackageIDs(expandedIDs)
}

func (s *permissionService) getRolePackageIDs(roleIDs []uuid.UUID) ([]uuid.UUID, error) {
	result := make([]uuid.UUID, 0)
	seen := make(map[uuid.UUID]struct{})
	for _, roleID := range roleIDs {
		packageIDs, err := s.rolePackageRepo.GetPackageIDsByRoleID(roleID)
		if err != nil {
			return nil, err
		}
		for _, packageID := range packageIDs {
			if _, ok := seen[packageID]; ok {
				continue
			}
			seen[packageID] = struct{}{}
			result = append(result, packageID)
		}
	}
	return result, nil
}

func (s *permissionService) getHiddenRoleMenuIDs(roleIDs []uuid.UUID) ([]uuid.UUID, error) {
	result := make([]uuid.UUID, 0)
	seen := make(map[uuid.UUID]struct{})
	for _, roleID := range roleIDs {
		menuIDs, err := s.roleHiddenMenuRepo.GetMenuIDsByRoleID(roleID)
		if err != nil {
			return nil, err
		}
		for _, menuID := range menuIDs {
			if _, ok := seen[menuID]; ok {
				continue
			}
			seen[menuID] = struct{}{}
			result = append(result, menuID)
		}
	}
	return result, nil
}

func (s *permissionService) expandPackageIDs(packageIDs []uuid.UUID, context string) ([]uuid.UUID, error) {
	result := make([]uuid.UUID, 0, len(packageIDs))
	seen := make(map[uuid.UUID]struct{}, len(packageIDs))
	visited := make(map[uuid.UUID]struct{}, len(packageIDs))
	for _, packageID := range packageIDs {
		if err := s.expandPackageID(packageID, context, visited, seen, &result); err != nil {
			return nil, err
		}
	}
	return result, nil
}

func (s *permissionService) expandPackageID(packageID uuid.UUID, context string, visited map[uuid.UUID]struct{}, seen map[uuid.UUID]struct{}, result *[]uuid.UUID) error {
	if _, ok := visited[packageID]; ok {
		return nil
	}
	visited[packageID] = struct{}{}
	pkg, err := s.packageRepo.GetByID(packageID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}
	if pkg.Status != "normal" || !packageMatchesWorkspaceScope(pkg.WorkspaceScope, pkg.ContextType, context) {
		return nil
	}
	if pkg.PackageType == "bundle" {
		childPackageIDs, err := s.packageBundleRepo.GetChildPackageIDs(packageID)
		if err != nil {
			return err
		}
		for _, childPackageID := range childPackageIDs {
			if err := s.expandPackageID(childPackageID, context, visited, seen, result); err != nil {
				return err
			}
		}
		return nil
	}
	if _, ok := seen[packageID]; ok {
		return nil
	}
	seen[packageID] = struct{}{}
	*result = append(*result, packageID)
	return nil
}

func (s *permissionService) finalizeMenuIDs(menuIDs []uuid.UUID, appKey string) ([]uuid.UUID, error) {
	allMenus, err := s.menuRepo.ListAll()
	if err != nil {
		return nil, err
	}
	enabledSet := make(map[uuid.UUID]struct{}, len(allMenus))
	publicIDs := make([]uuid.UUID, 0)
	for _, menu := range allMenus {
		if appctx.NormalizeAppKey(menu.AppKey) != appctx.NormalizeAppKey(appKey) {
			continue
		}
		if !isMenuEnabled(menu) {
			continue
		}
		enabledSet[menu.ID] = struct{}{}
		if isPublicMenu(menu) {
			publicIDs = append(publicIDs, menu.ID)
		}
	}
	result := make([]uuid.UUID, 0, len(menuIDs)+len(publicIDs))
	seen := make(map[uuid.UUID]struct{}, len(menuIDs)+len(publicIDs))
	for _, menuID := range mergeUUIDLists(menuIDs, publicIDs) {
		if _, ok := enabledSet[menuID]; !ok {
			continue
		}
		if _, ok := seen[menuID]; ok {
			continue
		}
		seen[menuID] = struct{}{}
		result = append(result, menuID)
	}
	return result, nil
}

func mergeUUIDLists(groups ...[]uuid.UUID) []uuid.UUID {
	result := make([]uuid.UUID, 0)
	seen := make(map[uuid.UUID]struct{})
	for _, group := range groups {
		for _, id := range group {
			if _, ok := seen[id]; ok {
				continue
			}
			seen[id] = struct{}{}
			result = append(result, id)
		}
	}
	return result
}

func subtractUUIDs(source []uuid.UUID, blocked []uuid.UUID) []uuid.UUID {
	if len(source) == 0 {
		return []uuid.UUID{}
	}
	if len(blocked) == 0 {
		return append([]uuid.UUID{}, source...)
	}
	blockedSet := make(map[uuid.UUID]struct{}, len(blocked))
	for _, id := range blocked {
		blockedSet[id] = struct{}{}
	}
	result := make([]uuid.UUID, 0, len(source))
	seen := make(map[uuid.UUID]struct{}, len(source))
	for _, id := range source {
		if _, ok := blockedSet[id]; ok {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	return result
}

func intersectUUIDs(left []uuid.UUID, right []uuid.UUID) []uuid.UUID {
	if len(left) == 0 || len(right) == 0 {
		return []uuid.UUID{}
	}
	rightSet := make(map[uuid.UUID]struct{}, len(right))
	for _, id := range right {
		rightSet[id] = struct{}{}
	}
	result := make([]uuid.UUID, 0, len(left))
	seen := make(map[uuid.UUID]struct{}, len(left))
	for _, id := range left {
		if _, ok := rightSet[id]; !ok {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	return result
}

func packageMatchesWorkspaceScope(packageWorkspaceScope, packageContext, currentContext string) bool {
	scope := strings.TrimSpace(packageWorkspaceScope)
	if scope == "" {
		scope = strings.TrimSpace(packageContext)
	}
	if scope == "" || scope == "all" || scope == "common" {
		return true
	}
	if scope == currentContext {
		return true
	}
	return currentContext == "common"
}

func isMenuEnabled(menu Menu) bool {
	if menu.Meta == nil {
		return true
	}
	if enabled, ok := menu.Meta["isEnable"].(bool); ok {
		return enabled
	}
	return true
}

func isPublicMenu(menu Menu) bool {
	if menu.Meta == nil {
		return false
	}
	if accessMode := menuAccessMode(menu.Meta); accessMode == "public" || accessMode == "jwt" {
		return true
	}
	for _, key := range []string{"isPublic", "public", "globalVisible", "publicMenu", "public_menu"} {
		if value, ok := menu.Meta[key].(bool); ok && value {
			return true
		}
	}
	return false
}

func menuAccessMode(meta map[string]interface{}) string {
	if meta == nil {
		return "permission"
	}
	value, ok := meta["accessMode"].(string)
	if !ok {
		return "permission"
	}
	switch strings.TrimSpace(strings.ToLower(value)) {
	case "public", "jwt":
		return strings.TrimSpace(strings.ToLower(value))
	default:
		return "permission"
	}
}

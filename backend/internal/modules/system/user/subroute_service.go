// subroute_service.go — UserSubrouteService wraps the complex per-user
// sub-route operations (menus, packages, permissions, diagnosis) so they can
// be called from the ogen handler layer without importing gin.
package user

import (
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/gg-ecommerce/backend/internal/modules/system/models"
	appctx "github.com/gg-ecommerce/backend/internal/pkg/appctx"
	"github.com/gg-ecommerce/backend/internal/pkg/appscope"
	"github.com/gg-ecommerce/backend/internal/pkg/collaborationworkspaceboundary"
	"github.com/gg-ecommerce/backend/internal/pkg/platformaccess"
	"github.com/gg-ecommerce/backend/internal/pkg/workspacerolebinding"
)

// UserMenusResult is returned by GetMenus.
type UserMenusResult struct {
	MenuIDs            []string
	AvailableMenuIDs   []string
	HiddenMenuIDs      []string
	ExpandedPackageIDs []string
	DerivedSources     []map[string]interface{}
	HasPackageConfig   bool
}

// UserPackagesResult is returned by GetPackages.
type UserPackagesResult struct {
	PackageIDs           []string
	Packages             []FeaturePackage
	BindingWorkspaceID   string
	BindingWorkspaceType string
}

// SubrouteService exposes the user sub-route business logic to the ogen
// handler layer without requiring a gin.Context.
type SubrouteService interface {
	// GetMenus returns the menu snapshot for userID under appKey.
	GetMenus(userID uuid.UUID, appKey string) (*UserMenusResult, error)
	// SetMenus replaces the hidden menus for userID under appKey.
	// selectedMenuIDs must be a subset of the user's available menus.
	SetMenus(userID uuid.UUID, appKey string, selectedMenuIDs []uuid.UUID) error
	// GetPackages returns the assigned packages for userID under appKey.
	GetPackages(userID uuid.UUID, appKey string) (*UserPackagesResult, error)
	// SetPackages replaces the assigned packages for userID under appKey.
	// operatorID may be nil.
	SetPackages(userID uuid.UUID, appKey string, packageIDs []uuid.UUID, operatorID *uuid.UUID) error
	// GetPermissions returns the menu-permission tree for userID.
	GetPermissions(userID uuid.UUID, appKey string, collaborationWorkspaceID *uuid.UUID) (interface{}, error)
	// GetPermissionDiagnosis returns the full permission diagnosis payload.
	GetPermissionDiagnosis(userID uuid.UUID, appKey string, permissionKey string, collaborationWorkspaceID *uuid.UUID) (interface{}, error)
}

// subrouteService implements SubrouteService.
type subrouteService struct {
	// Wraps the existing UserHandler so we reuse all its private helpers.
	handler *UserHandler
}

// Sentinel errors returned by subrouteService.
var (
	ErrMenuOutOfScope   = &subrouteErr{"存在超出当前用户已生效功能包范围的菜单"}
	ErrNoPackageConfig  = &subrouteErr{"当前用户尚未绑定功能包，不能配置菜单裁剪"}
	ErrPackageNotFound  = &subrouteErr{"包含不存在的功能包"}
	ErrWrongApp         = &subrouteErr{"仅支持绑定当前应用内的功能包"}
	ErrWrongContextType = &subrouteErr{"仅支持绑定个人空间功能包"}
)

type subrouteErr struct{ msg string }

func (e *subrouteErr) Error() string { return e.msg }

// NewSubrouteService constructs a SubrouteService from a fully-wired
// UserHandler. Use NewSubrouteServiceFromDeps if you need to construct from
// raw dependencies.
func NewSubrouteService(h *UserHandler) SubrouteService {
	return &subrouteService{handler: h}
}

// NewSubrouteServiceFromDeps constructs a SubrouteService directly from the
// same set of repositories and services used by NewAPIHandler.
func NewSubrouteServiceFromDeps(
	db *gorm.DB,
	userSvc UserService,
	featurePkgRepo interface {
		GetByIDs(ids []uuid.UUID) ([]FeaturePackage, error)
	},
	keyRepo interface {
		GetByPermissionKey(permissionKey string) (*PermissionKey, error)
	},
	personalAccess platformaccess.Service,
	boundarySvc collaborationworkspaceboundary.Service,
	roleRepo interface {
		GetByIDs(ids []uuid.UUID) ([]Role, error)
	},
	authzSvc interface {
		Authorize(userID uuid.UUID, collaborationWorkspaceID *uuid.UUID, permissionKey string, legacy ...string) (bool, *models.PermissionKey, error)
		AuthorizeInApp(userID uuid.UUID, collaborationWorkspaceID *uuid.UUID, appKey string, permissionKey string, legacy ...string) (bool, *models.PermissionKey, error)
	},
	userRoleRepo interface {
		GetEffectiveActiveRoleIDsByUserAndCollaborationWorkspace(userID uuid.UUID, collaborationWorkspaceID *uuid.UUID) ([]uuid.UUID, error)
	},
	cwMemberRepo interface {
		GetByUserAndCollaborationWorkspace(userID, collaborationWorkspaceID uuid.UUID) (*CollaborationWorkspaceMember, error)
		GetCollaborationWorkspacesByUserID(userID uuid.UUID) ([]CollaborationWorkspace, error)
	},
	userPackageRepo interface {
		GetPackageIDsByUserID(userID uuid.UUID) ([]uuid.UUID, error)
		ReplaceUserPackages(userID uuid.UUID, packageIDs []uuid.UUID, grantedBy *uuid.UUID) error
	},
	userHiddenMenuRepo interface {
		GetMenuIDsByUserID(userID uuid.UUID) ([]uuid.UUID, error)
		ReplaceUserHiddenMenus(userID uuid.UUID, menuIDs []uuid.UUID) error
	},
	menuRepo interface {
		ListAll() ([]Menu, error)
	},
	refresher interface {
		RefreshPersonalWorkspaceUser(userID uuid.UUID) error
		RefreshCollaborationWorkspace(collaborationWorkspaceID uuid.UUID) error
	},
	logger *zap.Logger,
) SubrouteService {
	h := NewUserHandler(
		db, userSvc, featurePkgRepo, keyRepo,
		personalAccess, boundarySvc,
		roleRepo, authzSvc, userRoleRepo, cwMemberRepo,
		userPackageRepo, userHiddenMenuRepo, menuRepo, refresher,
		logger,
	)
	return &subrouteService{handler: h}
}

// ── GetMenus ──────────────────────────────────────────────────────────────────

func (s *subrouteService) GetMenus(userID uuid.UUID, appKey string) (*UserMenusResult, error) {
	snapshot, err := s.handler.getPersonalWorkspaceSnapshot(userID, appKey)
	if err != nil {
		return nil, err
	}
	menuIDs := snapshot.MenuIDs
	if menuIDs == nil {
		menuIDs = []uuid.UUID{}
	}
	availableMenuIDs := snapshot.AvailableMenuIDs
	if availableMenuIDs == nil {
		availableMenuIDs = []uuid.UUID{}
	}
	hiddenMenuIDs := snapshot.HiddenMenuIDs
	if hiddenMenuIDs == nil {
		hiddenMenuIDs = []uuid.UUID{}
	}

	ginSources := buildUserMenuSourceMaps(snapshot.AvailableMenuMap)
	sourcesOut := make([]map[string]interface{}, 0, len(ginSources))
	for _, gh := range ginSources {
		m := make(map[string]interface{}, len(gh))
		for k, v := range gh {
			m[k] = v
		}
		sourcesOut = append(sourcesOut, m)
	}

	return &UserMenusResult{
		MenuIDs:            packageIDsToStrings(menuIDs),
		AvailableMenuIDs:   packageIDsToStrings(availableMenuIDs),
		HiddenMenuIDs:      packageIDsToStrings(hiddenMenuIDs),
		ExpandedPackageIDs: packageIDsToStrings(snapshot.ExpandedPackageIDs),
		DerivedSources:     sourcesOut,
		HasPackageConfig:   snapshot.HasPackageConfig,
	}, nil
}

// ── SetMenus ──────────────────────────────────────────────────────────────────

func (s *subrouteService) SetMenus(userID uuid.UUID, appKey string, selectedMenuIDs []uuid.UUID) error {
	snapshot, err := s.handler.getPersonalWorkspaceSnapshot(userID, appKey)
	if err != nil {
		return err
	}
	if !snapshot.HasPackageConfig {
		return ErrNoPackageConfig
	}
	availableMenuSet := uuidSliceToSet(snapshot.AvailableMenuIDs)
	for _, menuID := range selectedMenuIDs {
		if !availableMenuSet[menuID] {
			return ErrMenuOutOfScope
		}
	}
	blockedMenuIDs := excludeUUIDs(snapshot.AvailableMenuIDs, selectedMenuIDs)
	if err := appscope.ReplaceUserHiddenMenusInApp(s.handler.db, userID, appKey, blockedMenuIDs); err != nil {
		return err
	}
	if s.handler.refresher != nil {
		return s.handler.refresher.RefreshPersonalWorkspaceUser(userID)
	}
	return nil
}

// ── GetPackages ───────────────────────────────────────────────────────────────

func (s *subrouteService) GetPackages(userID uuid.UUID, appKey string) (*UserPackagesResult, error) {
	packageIDs, err := appscope.PackageIDsByUser(s.handler.db, userID, appKey)
	if err != nil {
		return nil, err
	}
	packages, err := s.handler.featurePkgRepo.GetByIDs(packageIDs)
	if err != nil {
		return nil, err
	}
	bindingWorkspaceID := ""
	bindingWorkspaceType := models.WorkspaceTypePersonal
	if workspace, workspaceErr := workspacerolebinding.GetPersonalWorkspaceByUserID(s.handler.db, userID); workspaceErr == nil && workspace != nil {
		bindingWorkspaceID = workspace.ID.String()
		bindingWorkspaceType = workspace.WorkspaceType
	}
	return &UserPackagesResult{
		PackageIDs:           packageIDsToStrings(packageIDs),
		Packages:             packages,
		BindingWorkspaceID:   bindingWorkspaceID,
		BindingWorkspaceType: bindingWorkspaceType,
	}, nil
}

// ── SetPackages ───────────────────────────────────────────────────────────────

func (s *subrouteService) SetPackages(userID uuid.UUID, appKey string, packageIDs []uuid.UUID, operatorID *uuid.UUID) error {
	if len(packageIDs) > 0 {
		packages, err := s.handler.featurePkgRepo.GetByIDs(packageIDs)
		if err != nil {
			return err
		}
		if len(packages) != len(packageIDs) {
			return ErrPackageNotFound
		}
		for _, pkg := range packages {
			if appctx.NormalizeAppKey(pkg.AppKey) != appctx.NormalizeAppKey(appKey) {
				return ErrWrongApp
			}
			if !supportsPersonalWorkspaceContext(pkg.ContextType) {
				return ErrWrongContextType
			}
		}
	}
	if err := appscope.ReplaceUserPackagesInApp(s.handler.db, userID, appKey, packageIDs, operatorID); err != nil {
		return err
	}
	if s.handler.refresher != nil {
		return s.handler.refresher.RefreshPersonalWorkspaceUser(userID)
	}
	return nil
}

// ── GetPermissions ────────────────────────────────────────────────────────────

func (s *subrouteService) GetPermissions(userID uuid.UUID, appKey string, collaborationWorkspaceID *uuid.UUID) (interface{}, error) {
	menuIDs, err := s.handler.getPermissionMenuIDs(userID, collaborationWorkspaceID, appKey)
	if err != nil {
		return nil, err
	}
	allMenus, err := s.handler.menuRepo.ListAll()
	if err != nil {
		return nil, err
	}
	menuIDSet := make(map[uuid.UUID]bool)
	for _, mid := range menuIDs {
		menuIDSet[mid] = true
	}
	return buildMenuTree(filterMenusByApp(allMenus, appKey), menuIDSet), nil
}

// ── GetPermissionDiagnosis ────────────────────────────────────────────────────

func (s *subrouteService) GetPermissionDiagnosis(userID uuid.UUID, appKey string, permissionKey string, collaborationWorkspaceID *uuid.UUID) (interface{}, error) {
	return s.handler.buildPermissionDiagnosis(userID, collaborationWorkspaceID, permissionKey, appKey)
}


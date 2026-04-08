package navigation

import (
	"fmt"
	"testing"

	"github.com/google/uuid"

	"github.com/gg-ecommerce/backend/internal/api/dto"
	apppkg "github.com/gg-ecommerce/backend/internal/modules/system/app"
	menupkg "github.com/gg-ecommerce/backend/internal/modules/system/menu"
	"github.com/gg-ecommerce/backend/internal/modules/system/models"
	pagepkg "github.com/gg-ecommerce/backend/internal/modules/system/page"
	spacepkg "github.com/gg-ecommerce/backend/internal/modules/system/space"
	"github.com/gg-ecommerce/backend/internal/modules/system/user"
)

func TestCompileUsesResolvedSpaceAndCompiledAccessGraph(t *testing.T) {
	rootID := uuid.New()
	entryID := uuid.New()
	userID := uuid.New()
	collaborationWorkspaceID := uuid.New()

	accessCtx := &pagepkg.CompiledAccessContext{
		SpaceKey:      "ops",
		Authenticated: true,
		ActionKeys: map[string]struct{}{
			"collaboration_workspace.read": {},
		},
		VisibleMenuIDs: map[uuid.UUID]struct{}{
			rootID:  {},
			entryID: {},
		},
	}

	var receivedSpaceKey string
	var receivedMenuIDs []uuid.UUID
	var listRuntimeSpaceKey string
	compiler := NewService(
		nil,
		&stubAppService{},
		&stubMenuService{
			getTreeFn: func(all bool, allowedMenuIDs []uuid.UUID, appKey, spaceKey string) ([]*user.Menu, error) {
				if all {
					t.Fatalf("GetTree all = true, want false")
				}
				if appKey != models.DefaultAppKey {
					t.Fatalf("GetTree appKey = %q, want %q", appKey, models.DefaultAppKey)
				}
				receivedMenuIDs = append([]uuid.UUID(nil), allowedMenuIDs...)
				receivedSpaceKey = spaceKey
				return []*user.Menu{
					{
						ID:       rootID,
						SpaceKey: "ops",
						Kind:     models.MenuKindDirectory,
						Path:     "collaboration",
						Name:     "TeamRoot",
						Title:    "协作空间管理",
						Meta:     models.MetaJSON{"isEnable": true},
						Children: []*user.Menu{
							{
								ID:        entryID,
								ParentID:  &rootID,
								SpaceKey:  "ops",
								Kind:      models.MenuKindEntry,
								Path:      "all",
								Name:      "TeamAll",
								Title:     "所有协作空间",
								Component: "/collaboration-workspace/workspaces",
								Meta:      models.MetaJSON{"accessMode": "permission", "isEnable": true},
							},
						},
					},
				}, nil
			},
		},
		&stubPageService{
			resolveCompiledAccessContextFn: func(appKey string, spaceKey string, gotUserID *uuid.UUID, gotCollaborationWorkspaceID *uuid.UUID) (*pagepkg.CompiledAccessContext, error) {
				if appKey != models.DefaultAppKey {
					t.Fatalf("ResolveCompiledAccessContext appKey = %q, want %q", appKey, models.DefaultAppKey)
				}
				if spaceKey != "ops" {
					t.Fatalf("ResolveCompiledAccessContext spaceKey = %q, want ops", spaceKey)
				}
				if gotUserID == nil || *gotUserID != userID {
					t.Fatalf("ResolveCompiledAccessContext userID = %v, want %s", gotUserID, userID)
				}
				if gotCollaborationWorkspaceID == nil || *gotCollaborationWorkspaceID != collaborationWorkspaceID {
					t.Fatalf("ResolveCompiledAccessContext collaborationWorkspaceID = %v, want %s", gotCollaborationWorkspaceID, collaborationWorkspaceID)
				}
				return accessCtx, nil
			},
			listRuntimeWithAccessFn: func(appKey string, spaceKey string, ctx *pagepkg.CompiledAccessContext) ([]pagepkg.Record, error) {
				if appKey != models.DefaultAppKey {
					t.Fatalf("ListRuntimeWithAccess appKey = %q, want %q", appKey, models.DefaultAppKey)
				}
				listRuntimeSpaceKey = spaceKey
				if ctx != accessCtx {
					t.Fatalf("ListRuntimeWithAccess received unexpected access context pointer")
				}
				return []pagepkg.Record{
					{
						UIPage: models.UIPage{
							PageKey:        "collaboration_workspace.detail",
							Name:           "协作空间详情",
							RouteName:      "TeamDetail",
							RoutePath:      "/detail/:id",
							Component:      "/collaboration-workspace/detail",
							PageType:       models.PageTypeInner,
							ParentMenuID:   &entryID,
							ActiveMenuPath: "/collaboration-workspace/all",
							Status:         "normal",
						},
					},
				}, nil
			},
		},
		&stubSpaceService{
			getCurrentFn: func(appKey string, host string, requestedSpaceKey string, gotUserID *uuid.UUID, gotCollaborationWorkspaceID *uuid.UUID) (*spacepkg.CurrentResponse, error) {
				if appKey != models.DefaultAppKey {
					t.Fatalf("GetCurrent appKey = %q, want %q", appKey, models.DefaultAppKey)
				}
				if host != " ops.example.com " {
					t.Fatalf("GetCurrent host = %q, want original request host", host)
				}
				if requestedSpaceKey != " ops " {
					t.Fatalf("GetCurrent requestedSpaceKey = %q, want original request space", requestedSpaceKey)
				}
				if gotUserID == nil || *gotUserID != userID {
					t.Fatalf("GetCurrent userID = %v, want %s", gotUserID, userID)
				}
				if gotCollaborationWorkspaceID == nil || *gotCollaborationWorkspaceID != collaborationWorkspaceID {
					t.Fatalf("GetCurrent collaborationWorkspaceID = %v, want %s", gotCollaborationWorkspaceID, collaborationWorkspaceID)
				}
				return &spacepkg.CurrentResponse{
					Space: spacepkg.SpaceRecord{
						MenuSpace: models.MenuSpace{
							SpaceKey: "ops",
							Name:     "运营空间",
						},
					},
					ResolvedBy:    "explicit",
					RequestHost:   "ops.example.com",
					AccessGranted: true,
				}, nil
			},
		},
	)

	manifest, err := compiler.Compile(models.DefaultAppKey, " ops.example.com ", " ops ", &userID, &collaborationWorkspaceID)
	if err != nil {
		t.Fatalf("Compile() error = %v", err)
	}

	if receivedSpaceKey != "ops" {
		t.Fatalf("GetTree spaceKey = %q, want ops", receivedSpaceKey)
	}
	if listRuntimeSpaceKey != "ops" {
		t.Fatalf("ListRuntimeWithAccess spaceKey = %q, want ops", listRuntimeSpaceKey)
	}
	if !sameUUIDSet(receivedMenuIDs, accessCtx.VisibleMenuIDList()) {
		t.Fatalf("GetTree allowedMenuIDs = %v, want %v", receivedMenuIDs, accessCtx.VisibleMenuIDList())
	}

	if manifest.CurrentSpace == nil || manifest.CurrentSpace.Space.SpaceKey != "ops" {
		t.Fatalf("manifest.CurrentSpace = %#v, want ops space", manifest.CurrentSpace)
	}
	if got := manifest.Context["space_key"]; got != "ops" {
		t.Fatalf("context.space_key = %#v, want ops", got)
	}
	if got := manifest.Context["request_host"]; got != "ops.example.com" {
		t.Fatalf("context.request_host = %#v, want trimmed host", got)
	}
	if got := manifest.Context["requested_space_key"]; got != "ops" {
		t.Fatalf("context.requested_space_key = %#v, want trimmed space key", got)
	}
	if got := manifest.Context["visible_menu_count"]; got != 2 {
		t.Fatalf("context.visible_menu_count = %#v, want 2", got)
	}
	if got := manifest.Context["managed_page_count"]; got != 1 {
		t.Fatalf("context.managed_page_count = %#v, want 1", got)
	}
	if got := manifest.Context["action_key_count"]; got != 1 {
		t.Fatalf("context.action_key_count = %#v, want 1", got)
	}
	if got := manifest.Context["authenticated"]; got != true {
		t.Fatalf("context.authenticated = %#v, want true", got)
	}
	if got := manifest.Context["user_id"]; got != userID.String() {
		t.Fatalf("context.user_id = %#v, want %s", got, userID)
	}
	if got := manifest.Context["collaboration_workspace_id"]; got != collaborationWorkspaceID.String() {
		t.Fatalf("context.collaboration_workspace_id = %#v, want %s", got, collaborationWorkspaceID)
	}
	if len(manifest.MenuTree) != 1 {
		t.Fatalf("len(manifest.MenuTree) = %d, want 1", len(manifest.MenuTree))
	}
	if len(manifest.EntryRoutes) != 1 {
		t.Fatalf("len(manifest.EntryRoutes) = %d, want 1", len(manifest.EntryRoutes))
	}
	if len(manifest.ManagedPages) != 1 {
		t.Fatalf("len(manifest.ManagedPages) = %d, want 1", len(manifest.ManagedPages))
	}
	if got := manifest.EntryRoutes[0]["name"]; got != "TeamAll" {
		t.Fatalf("entry route name = %#v, want TeamAll", got)
	}
	if got := manifest.ManagedPages[0]["page_key"]; got != "collaboration_workspace.detail" {
		t.Fatalf("managed page key = %#v, want collaboration_workspace.detail", got)
	}
	if manifest.VersionStamp != "platform-admin:ops:1:1" {
		t.Fatalf("manifest.VersionStamp = %q, want platform-admin:ops:1:1", manifest.VersionStamp)
	}
}

func sameUUIDSet(left, right []uuid.UUID) bool {
	if len(left) != len(right) {
		return false
	}
	set := make(map[uuid.UUID]int, len(left))
	for _, id := range left {
		set[id]++
	}
	for _, id := range right {
		if set[id] == 0 {
			return false
		}
		set[id]--
	}
	for _, count := range set {
		if count != 0 {
			return false
		}
	}
	return true
}

type stubMenuService struct {
	getTreeFn func(all bool, allowedMenuIDs []uuid.UUID, appKey, spaceKey string) ([]*user.Menu, error)
}

func (s *stubMenuService) GetTree(all bool, allowedMenuIDs []uuid.UUID, appKey, spaceKey string) ([]*user.Menu, error) {
	if s.getTreeFn == nil {
		return nil, fmt.Errorf("unexpected GetTree call")
	}
	return s.getTreeFn(all, allowedMenuIDs, appKey, spaceKey)
}

func (s *stubMenuService) Create(req *dto.MenuCreateRequest) (*user.Menu, error) {
	return nil, fmt.Errorf("unexpected Create call")
}

func (s *stubMenuService) Update(id uuid.UUID, req *dto.MenuUpdateRequest) error {
	return fmt.Errorf("unexpected Update call")
}

func (s *stubMenuService) Delete(id uuid.UUID, mode string, targetParentID *uuid.UUID) error {
	return fmt.Errorf("unexpected Delete call")
}

func (s *stubMenuService) DeletePreview(id uuid.UUID, mode string, targetParentID *uuid.UUID) (*menupkg.MenuDeletePreview, error) {
	return nil, fmt.Errorf("unexpected DeletePreview call")
}

type stubPageService struct {
	resolveCompiledAccessContextFn func(appKey, spaceKey string, userID *uuid.UUID, collaborationWorkspaceID *uuid.UUID) (*pagepkg.CompiledAccessContext, error)
	listRuntimeWithAccessFn        func(appKey, spaceKey string, accessCtx *pagepkg.CompiledAccessContext) ([]pagepkg.Record, error)
}

func (s *stubPageService) List(req *pagepkg.ListRequest) ([]pagepkg.Record, int64, error) {
	return nil, 0, fmt.Errorf("unexpected List call")
}

func (s *stubPageService) ListOptions(appKey, spaceKey string) ([]models.UIPage, error) {
	return nil, fmt.Errorf("unexpected ListOptions call")
}

func (s *stubPageService) ListRuntime(appKey, host, requestedSpaceKey string, userID *uuid.UUID, collaborationWorkspaceID *uuid.UUID) ([]pagepkg.Record, error) {
	return nil, fmt.Errorf("unexpected ListRuntime call")
}

func (s *stubPageService) ListRuntimePublic(appKey, host, requestedSpaceKey string, userID *uuid.UUID, collaborationWorkspaceID *uuid.UUID) ([]pagepkg.Record, error) {
	return nil, fmt.Errorf("unexpected ListRuntimePublic call")
}

func (s *stubPageService) ResolveCompiledAccessContext(appKey, spaceKey string, userID *uuid.UUID, collaborationWorkspaceID *uuid.UUID) (*pagepkg.CompiledAccessContext, error) {
	if s.resolveCompiledAccessContextFn == nil {
		return nil, fmt.Errorf("unexpected ResolveCompiledAccessContext call")
	}
	return s.resolveCompiledAccessContextFn(appKey, spaceKey, userID, collaborationWorkspaceID)
}

func (s *stubPageService) GetAccessTrace(appKey string, req *pagepkg.AccessTraceRequest) (*pagepkg.AccessTraceResult, error) {
	return nil, fmt.Errorf("unexpected GetAccessTrace call")
}

func (s *stubPageService) ListRuntimeWithAccess(appKey, spaceKey string, accessCtx *pagepkg.CompiledAccessContext) ([]pagepkg.Record, error) {
	if s.listRuntimeWithAccessFn == nil {
		return nil, fmt.Errorf("unexpected ListRuntimeWithAccess call")
	}
	return s.listRuntimeWithAccessFn(appKey, spaceKey, accessCtx)
}

func (s *stubPageService) ListUnregistered(appKey string) ([]pagepkg.UnregisteredRecord, error) {
	return nil, fmt.Errorf("unexpected ListUnregistered call")
}

func (s *stubPageService) Sync(appKey string) (*pagepkg.SyncResult, error) {
	return nil, fmt.Errorf("unexpected Sync call")
}

func (s *stubPageService) PreviewBreadcrumb(id uuid.UUID, appKey string) ([]pagepkg.BreadcrumbPreviewItem, error) {
	return nil, fmt.Errorf("unexpected PreviewBreadcrumb call")
}

func (s *stubPageService) Get(id uuid.UUID, appKey string) (*pagepkg.Record, error) {
	return nil, fmt.Errorf("unexpected Get call")
}

func (s *stubPageService) Create(req *pagepkg.SaveRequest) (*pagepkg.Record, error) {
	return nil, fmt.Errorf("unexpected Create call")
}

func (s *stubPageService) Update(id uuid.UUID, req *pagepkg.SaveRequest) (*pagepkg.Record, error) {
	return nil, fmt.Errorf("unexpected Update call")
}

func (s *stubPageService) Delete(id uuid.UUID, appKey string) error {
	return fmt.Errorf("unexpected Delete call")
}

func (s *stubPageService) ListMenuOptions(appKey, spaceKey string) ([]pagepkg.MenuOption, error) {
	return nil, fmt.Errorf("unexpected ListMenuOptions call")
}

type stubSpaceService struct {
	getCurrentFn func(appKey string, host string, requestedSpaceKey string, userID *uuid.UUID, collaborationWorkspaceID *uuid.UUID) (*spacepkg.CurrentResponse, error)
}

type stubAppService struct{}

func (s *stubAppService) ListApps() ([]apppkg.AppRecord, error) {
	return []apppkg.AppRecord{{
		App: models.App{AppKey: models.DefaultAppKey, Name: models.DefaultAppName, DefaultSpaceKey: models.DefaultMenuSpaceKey},
	}}, nil
}

func (s *stubAppService) GetCurrent(host, requestedAppKey string) (*apppkg.CurrentResponse, error) {
	return &apppkg.CurrentResponse{
		App: apppkg.AppRecord{
			App: models.App{AppKey: apppkg.NormalizeAppKey(requestedAppKey), Name: models.DefaultAppName, DefaultSpaceKey: models.DefaultMenuSpaceKey},
		},
		ResolvedBy:  "explicit",
		RequestHost: host,
	}, nil
}

func (s *stubAppService) SaveApp(req *apppkg.SaveAppRequest) (*apppkg.AppRecord, error) {
	return nil, fmt.Errorf("unexpected SaveApp call")
}

func (s *stubAppService) ListHostBindings(appKey string) ([]apppkg.HostBindingRecord, error) {
	return nil, fmt.Errorf("unexpected ListHostBindings call")
}

func (s *stubAppService) SaveHostBinding(appKey string, req *apppkg.SaveHostBindingRequest) (*apppkg.HostBindingRecord, error) {
	return nil, fmt.Errorf("unexpected SaveHostBinding call")
}

func (s *stubSpaceService) ListSpaces(appKey string) ([]spacepkg.SpaceRecord, error) {
	return nil, fmt.Errorf("unexpected ListSpaces call")
}

func (s *stubSpaceService) GetCurrent(appKey string, host string, requestedSpaceKey string, userID *uuid.UUID, collaborationWorkspaceID *uuid.UUID) (*spacepkg.CurrentResponse, error) {
	if s.getCurrentFn == nil {
		return nil, fmt.Errorf("unexpected GetCurrent call")
	}
	return s.getCurrentFn(appKey, host, requestedSpaceKey, userID, collaborationWorkspaceID)
}

func (s *stubSpaceService) ListHostBindings(appKey string) ([]spacepkg.HostBindingRecord, error) {
	return nil, fmt.Errorf("unexpected ListHostBindings call")
}

func (s *stubSpaceService) GetMode(appKey string) (string, error) {
	return "", fmt.Errorf("unexpected GetMode call")
}

func (s *stubSpaceService) SaveMode(appKey, mode string) (string, error) {
	return "", fmt.Errorf("unexpected SaveMode call")
}

func (s *stubSpaceService) SaveSpace(appKey string, req *spacepkg.SaveSpaceRequest) (*spacepkg.SpaceRecord, error) {
	return nil, fmt.Errorf("unexpected SaveSpace call")
}

func (s *stubSpaceService) SaveHostBinding(appKey string, req *spacepkg.SaveHostBindingRequest) (*spacepkg.HostBindingRecord, error) {
	return nil, fmt.Errorf("unexpected SaveHostBinding call")
}

func (s *stubSpaceService) InitializeFromDefault(appKey string, targetSpaceKey string, force bool, actorUserID *uuid.UUID) (*spacepkg.InitializeResult, error) {
	return nil, fmt.Errorf("unexpected InitializeFromDefault call")
}

var (
	_ menupkg.MenuService = (*stubMenuService)(nil)
	_ pagepkg.Service     = (*stubPageService)(nil)
	_ spacepkg.Service    = (*stubSpaceService)(nil)
)

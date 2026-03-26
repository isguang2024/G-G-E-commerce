package apiendpoint

import (
	"crypto/sha1"
	"errors"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/gg-ecommerce/backend/internal/modules/system/user"
	"github.com/gg-ecommerce/backend/internal/pkg/apiendpointaccess"
	"github.com/gg-ecommerce/backend/internal/pkg/apiregistry"
)

type ListRequest struct {
	Current       int
	Size          int
	PermissionKey string
	Keyword       string
	Method        string
	Path          string
	CategoryID    string
	ContextScope  string
	Source        string
	FeatureKind   string
	Status        string
	HasPermission *bool
	HasCategory   *bool
}

type UnregisteredRouteListRequest struct {
	Current    int
	Size       int
	Method     string
	Path       string
	Keyword    string
	OnlyNoMeta bool
}

type UnregisteredRouteItem struct {
	Method  string                 `json:"method"`
	Path    string                 `json:"path"`
	Spec    string                 `json:"spec"`
	Handler string                 `json:"handler"`
	HasMeta bool                   `json:"has_meta"`
	Meta    map[string]interface{} `json:"meta"`
}

type Service interface {
	List(req *ListRequest) ([]user.APIEndpoint, int64, error)
	ListUnregisteredRoutes(req *UnregisteredRouteListRequest) ([]UnregisteredRouteItem, int64, error)
	ListBindingsByEndpointIDs(endpointIDs []uuid.UUID) ([]user.APIEndpointPermissionBinding, error)
	ListBindings(endpointID uuid.UUID) ([]user.APIEndpointPermissionBinding, error)
	ListCategories() ([]user.APIEndpointCategory, error)
	Save(endpoint *user.APIEndpoint, permissionKeys []string) (*user.APIEndpoint, error)
	SaveCategory(item *user.APIEndpointCategory) (*user.APIEndpointCategory, error)
	UpdateContextScope(endpointID uuid.UUID, contextScope string) (*user.APIEndpoint, error)
	Sync() error
}

type service struct {
	db             *gorm.DB
	repo           user.APIEndpointRepository
	categoryRepo   user.APIEndpointCategoryRepository
	bindingRepo    user.APIEndpointPermissionBindingRepository
	router         *gin.Engine
	logger         *zap.Logger
	env            string
	endpointAccess apiendpointaccess.Service
}

func NewService(db *gorm.DB, repo user.APIEndpointRepository, categoryRepo user.APIEndpointCategoryRepository, bindingRepo user.APIEndpointPermissionBindingRepository, router *gin.Engine, logger *zap.Logger, env string, endpointAccess apiendpointaccess.Service) Service {
	return &service{
		db:             db,
		repo:           repo,
		categoryRepo:   categoryRepo,
		bindingRepo:    bindingRepo,
		router:         router,
		logger:         logger,
		env:            strings.TrimSpace(env),
		endpointAccess: endpointAccess,
	}
}

func (s *service) List(req *ListRequest) ([]user.APIEndpoint, int64, error) {
	if req.Current <= 0 {
		req.Current = 1
	}
	if req.Size <= 0 {
		req.Size = 20
	}
	params := &user.APIEndpointListParams{
		Method:        strings.ToUpper(strings.TrimSpace(req.Method)),
		PermissionKey: strings.TrimSpace(req.PermissionKey),
		Keyword:       strings.TrimSpace(req.Keyword),
		Path:          strings.TrimSpace(req.Path),
		CategoryID:    strings.TrimSpace(req.CategoryID),
		ContextScope:  strings.TrimSpace(req.ContextScope),
		Source:        strings.TrimSpace(req.Source),
		FeatureKind:   strings.TrimSpace(req.FeatureKind),
		Status:        strings.TrimSpace(req.Status),
		HasPermission: req.HasPermission,
		HasCategory:   req.HasCategory,
	}
	if params.PermissionKey != "" {
		endpointIDs, err := s.bindingRepo.ListEndpointIDsByPermissionKey(params.PermissionKey)
		if err != nil {
			return nil, 0, err
		}
		if len(endpointIDs) == 0 {
			return []user.APIEndpoint{}, 0, nil
		}
		params.EndpointIDs = endpointIDs
	}
	return s.repo.List((req.Current-1)*req.Size, req.Size, params)
}

func (s *service) ListUnregisteredRoutes(req *UnregisteredRouteListRequest) ([]UnregisteredRouteItem, int64, error) {
	if req == nil {
		req = &UnregisteredRouteListRequest{}
	}
	if req.Current <= 0 {
		req.Current = 1
	}
	if req.Size <= 0 {
		req.Size = 20
	}
	if s.router == nil {
		return []UnregisteredRouteItem{}, 0, nil
	}

	runtimeRoutes := apiregistry.CollectRuntimeRoutes(s.router.Routes())
	registered, _, err := s.repo.List(0, 5000, nil)
	if err != nil {
		return nil, 0, err
	}
	registeredSet := make(map[string]struct{}, len(registered))
	for _, item := range registered {
		registeredSet[routeSpec(item.Method, item.Path)] = struct{}{}
	}

	methodFilter := strings.ToUpper(strings.TrimSpace(req.Method))
	pathFilter := strings.TrimSpace(req.Path)
	keyword := strings.TrimSpace(req.Keyword)

	items := make([]UnregisteredRouteItem, 0, len(runtimeRoutes))
	for _, route := range runtimeRoutes {
		if _, ok := registeredSet[routeSpec(route.Method, route.Path)]; ok {
			continue
		}
		if req.OnlyNoMeta && route.HasMeta {
			continue
		}
		if methodFilter != "" && route.Method != methodFilter {
			continue
		}
		if pathFilter != "" && !strings.Contains(route.Path, pathFilter) {
			continue
		}
		if keyword != "" {
			target := strings.ToLower(route.Method + " " + route.Path + " " + route.Handler + " " + route.RouteMeta.Summary + " " + route.RouteMeta.CategoryCode)
			if !strings.Contains(target, strings.ToLower(keyword)) {
				continue
			}
		}
		meta := map[string]interface{}{}
		if route.HasMeta {
			meta["summary"] = route.RouteMeta.Summary
			meta["category_code"] = route.RouteMeta.CategoryCode
			meta["context_scope"] = route.RouteMeta.ContextScope
			meta["source"] = route.RouteMeta.Source
			meta["feature_kind"] = route.RouteMeta.FeatureKind
			meta["permission_keys"] = route.RouteMeta.PermissionKeys
		}
		items = append(items, UnregisteredRouteItem{
			Method:  route.Method,
			Path:    route.Path,
			Spec:    routeSpec(route.Method, route.Path),
			Handler: route.Handler,
			HasMeta: route.HasMeta,
			Meta:    meta,
		})
	}

	total := int64(len(items))
	start := (req.Current - 1) * req.Size
	if start >= len(items) {
		return []UnregisteredRouteItem{}, total, nil
	}
	end := start + req.Size
	if end > len(items) {
		end = len(items)
	}
	return items[start:end], total, nil
}

func (s *service) Sync() error {
	if s.router == nil {
		return nil
	}
	if err := apiregistry.SyncRoutes(s.db, s.logger, s.router.Routes()); err != nil {
		return err
	}
	return s.refreshRuntimeCache()
}

func (s *service) ListBindings(endpointID uuid.UUID) ([]user.APIEndpointPermissionBinding, error) {
	return s.bindingRepo.ListByEndpointID(endpointID)
}

func (s *service) ListBindingsByEndpointIDs(endpointIDs []uuid.UUID) ([]user.APIEndpointPermissionBinding, error) {
	return s.bindingRepo.ListByEndpointIDs(endpointIDs)
}

func (s *service) ListCategories() ([]user.APIEndpointCategory, error) {
	return s.categoryRepo.List()
}

func (s *service) Save(endpoint *user.APIEndpoint, permissionKeys []string) (*user.APIEndpoint, error) {
	if endpoint == nil {
		return nil, errors.New("endpoint is nil")
	}
	endpoint.Method = strings.ToUpper(strings.TrimSpace(endpoint.Method))
	endpoint.Path = strings.TrimSpace(endpoint.Path)
	if endpoint.Method == "" || endpoint.Path == "" {
		return nil, errors.New("method 和 path 不能为空")
	}
	if endpoint.Code == "" {
		endpoint.Code = deriveStableEndpointCode(endpoint.Method, endpoint.Path)
	}
	if endpoint.ContextScope == "" {
		endpoint.ContextScope = "optional"
	}
	switch endpoint.ContextScope {
	case "required", "forbidden", "optional":
	default:
		return nil, errors.New("context_scope 仅支持 required|forbidden|optional")
	}
	if endpoint.Source == "" {
		endpoint.Source = "manual"
	}
	switch endpoint.Source {
	case "sync", "seed", "manual":
	default:
		return nil, errors.New("source 仅支持 sync|seed|manual")
	}
	if endpoint.FeatureKind == "" {
		endpoint.FeatureKind = "system"
	}
	switch endpoint.FeatureKind {
	case "system", "business":
	default:
		return nil, errors.New("feature_kind 仅支持 system|business")
	}
	if endpoint.Status == "" {
		endpoint.Status = "normal"
	}
	if endpoint.CategoryID != nil && *endpoint.CategoryID != uuid.Nil {
		if _, err := s.categoryRepo.GetByID(*endpoint.CategoryID); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errors.New("分类不存在")
			}
			return nil, err
		}
	} else {
		endpoint.CategoryID = nil
	}

	var oldMethod string
	var oldPath string
	if endpoint.ID == uuid.Nil {
		exists, err := s.repo.GetByMethodAndPath(endpoint.Method, endpoint.Path)
		if err == nil && exists != nil {
			return nil, errors.New("method+path 已存在")
		}
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		if err := s.repo.Create(endpoint); err != nil {
			return nil, err
		}
	} else {
		current, err := s.repo.GetByID(endpoint.ID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errors.New("API 不存在")
			}
			return nil, err
		}
		oldMethod = current.Method
		oldPath = current.Path
		exists, err := s.repo.GetByMethodAndPath(endpoint.Method, endpoint.Path)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		if exists != nil && exists.ID != endpoint.ID {
			return nil, errors.New("method+path 已存在")
		}
		updates := map[string]interface{}{
			"code":          endpoint.Code,
			"method":        endpoint.Method,
			"path":          endpoint.Path,
			"feature_kind":  endpoint.FeatureKind,
			"handler":       endpoint.Handler,
			"summary":       endpoint.Summary,
			"category_id":   endpoint.CategoryID,
			"context_scope": endpoint.ContextScope,
			"source":        endpoint.Source,
			"status":        endpoint.Status,
		}
		if err := s.repo.UpdateWithMap(endpoint.ID, updates); err != nil {
			return nil, err
		}
	}

	keys := make([]string, 0, len(permissionKeys))
	seen := make(map[string]struct{}, len(permissionKeys))
	actionRepo := user.NewPermissionKeyRepository(s.db)
	for _, key := range permissionKeys {
		target := strings.TrimSpace(key)
		if target == "" {
			continue
		}
		if _, ok := seen[target]; ok {
			continue
		}
		seen[target] = struct{}{}
		if _, err := actionRepo.GetByPermissionKey(target); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errors.New("存在未注册的功能权限键")
			}
			return nil, err
		}
		keys = append(keys, target)
	}
	items := make([]user.APIEndpointPermissionBinding, 0, len(keys))
	for idx, key := range keys {
		items = append(items, user.APIEndpointPermissionBinding{
			EndpointID:    endpoint.ID,
			PermissionKey: key,
			MatchMode:     "ANY",
			SortOrder:     idx,
		})
	}
	if err := s.bindingRepo.ReplaceByEndpointID(endpoint.ID, items); err != nil {
		return nil, err
	}
	if err := s.refreshRuntimeCacheForSave(oldMethod, oldPath, endpoint.Method, endpoint.Path); err != nil {
		return nil, err
	}

	return s.repo.GetByID(endpoint.ID)
}

func (s *service) SaveCategory(item *user.APIEndpointCategory) (*user.APIEndpointCategory, error) {
	if item == nil {
		return nil, errors.New("category is nil")
	}
	item.Code = strings.TrimSpace(item.Code)
	item.Name = strings.TrimSpace(item.Name)
	item.NameEn = strings.TrimSpace(item.NameEn)
	if item.Code == "" || item.Name == "" || item.NameEn == "" {
		return nil, errors.New("分类编码、中文名、英文名不能为空")
	}
	if item.Status == "" {
		item.Status = "normal"
	}
	if item.ID == uuid.Nil {
		exists, err := s.categoryRepo.GetByCode(item.Code)
		if err == nil && exists != nil {
			return nil, errors.New("分类编码已存在")
		}
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		if err := s.categoryRepo.Create(item); err != nil {
			return nil, err
		}
		return s.categoryRepo.GetByID(item.ID)
	}
	exists, err := s.categoryRepo.GetByCode(item.Code)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if exists != nil && exists.ID != item.ID {
		return nil, errors.New("分类编码已存在")
	}
	if err := s.categoryRepo.UpdateWithMap(item.ID, map[string]interface{}{
		"code":       item.Code,
		"name":       item.Name,
		"name_en":    item.NameEn,
		"sort_order": item.SortOrder,
		"status":     item.Status,
	}); err != nil {
		return nil, err
	}
	return s.categoryRepo.GetByID(item.ID)
}

func (s *service) UpdateContextScope(endpointID uuid.UUID, contextScope string) (*user.APIEndpoint, error) {
	switch strings.TrimSpace(contextScope) {
	case "required", "forbidden", "optional":
	default:
		return nil, errors.New("context_scope 仅支持 required|forbidden|optional")
	}
	if err := s.repo.UpdateWithMap(endpointID, map[string]interface{}{
		"context_scope": strings.TrimSpace(contextScope),
	}); err != nil {
		return nil, err
	}
	return s.repo.GetByID(endpointID)
}

func deriveStableEndpointCode(method, path string) string {
	targetMethod := strings.ToUpper(strings.TrimSpace(method))
	targetPath := strings.TrimSpace(path)
	return uuid.NewHash(sha1.New(), uuid.NameSpaceURL, []byte("api-endpoint:"+targetMethod+" "+targetPath), 5).String()
}

func routeSpec(method, path string) string {
	return strings.ToUpper(strings.TrimSpace(method)) + " " + strings.TrimSpace(path)
}

func (s *service) refreshRuntimeCache() error {
	if s.endpointAccess == nil {
		return nil
	}
	return s.endpointAccess.Refresh()
}

func (s *service) refreshRuntimeCacheForSave(oldMethod, oldPath, newMethod, newPath string) error {
	if s.endpointAccess == nil {
		return nil
	}

	oldKeyMethod := strings.ToUpper(strings.TrimSpace(oldMethod))
	oldKeyPath := strings.TrimSpace(oldPath)
	newKeyMethod := strings.ToUpper(strings.TrimSpace(newMethod))
	newKeyPath := strings.TrimSpace(newPath)

	if oldKeyMethod != "" && oldKeyPath != "" && (oldKeyMethod != newKeyMethod || oldKeyPath != newKeyPath) {
		s.endpointAccess.RemoveByRoute(oldKeyMethod, oldKeyPath)
	}

	return s.endpointAccess.RefreshByRoute(newKeyMethod, newKeyPath)
}

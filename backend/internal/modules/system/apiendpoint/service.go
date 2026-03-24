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
	"github.com/gg-ecommerce/backend/internal/pkg/apiregistry"
)

type ListRequest struct {
	Current       int
	Size          int
	PermissionKey string
	Keyword       string
	Method        string
	Path          string
	Module        string
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
	Module     string
	Keyword    string
	OnlyNoMeta bool
}

type UnregisteredRouteItem struct {
	Method  string                 `json:"method"`
	Path    string                 `json:"path"`
	Spec    string                 `json:"spec"`
	Handler string                 `json:"handler"`
	Module  string                 `json:"module"`
	HasMeta bool                   `json:"has_meta"`
	Meta    map[string]interface{} `json:"meta"`
}

type Service interface {
	List(req *ListRequest) ([]user.APIEndpoint, int64, error)
	ListUnregisteredRoutes(req *UnregisteredRouteListRequest) ([]UnregisteredRouteItem, int64, error)
	ListBindings(endpointID uuid.UUID) ([]user.APIEndpointPermissionBinding, error)
	ListCategories() ([]user.APIEndpointCategory, error)
	Save(endpoint *user.APIEndpoint, permissionKeys []string) (*user.APIEndpoint, error)
	SaveCategory(item *user.APIEndpointCategory) (*user.APIEndpointCategory, error)
	UpdateContextScope(endpointID uuid.UUID, contextScope string) (*user.APIEndpoint, error)
	Sync() error
}

type service struct {
	db           *gorm.DB
	repo         user.APIEndpointRepository
	categoryRepo user.APIEndpointCategoryRepository
	bindingRepo  user.APIEndpointPermissionBindingRepository
	router       *gin.Engine
	logger       *zap.Logger
	env          string
}

func NewService(db *gorm.DB, repo user.APIEndpointRepository, categoryRepo user.APIEndpointCategoryRepository, bindingRepo user.APIEndpointPermissionBindingRepository, router *gin.Engine, logger *zap.Logger, env string) Service {
	return &service{
		db:           db,
		repo:         repo,
		categoryRepo: categoryRepo,
		bindingRepo:  bindingRepo,
		router:       router,
		logger:       logger,
		env:          strings.TrimSpace(env),
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
		Module:        strings.TrimSpace(req.Module),
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
		all, _, err := s.repo.List(0, 5000, params)
		if err != nil {
			return nil, 0, err
		}
		idSet := make(map[uuid.UUID]struct{}, len(endpointIDs))
		for _, endpointID := range endpointIDs {
			idSet[endpointID] = struct{}{}
		}
		filtered := make([]user.APIEndpoint, 0, len(all))
		for _, item := range all {
			if _, ok := idSet[item.ID]; ok {
				filtered = append(filtered, item)
			}
		}
		total := int64(len(filtered))
		start := (req.Current - 1) * req.Size
		if start >= len(filtered) {
			return []user.APIEndpoint{}, total, nil
		}
		end := start + req.Size
		if end > len(filtered) {
			end = len(filtered)
		}
		return filtered[start:end], total, nil
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
	moduleFilter := strings.TrimSpace(req.Module)
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
		if moduleFilter != "" && !strings.Contains(route.Module, moduleFilter) {
			continue
		}
		if keyword != "" {
			target := strings.ToLower(route.Method + " " + route.Path + " " + route.Module + " " + route.Handler + " " + route.RouteMeta.Summary)
			if !strings.Contains(target, strings.ToLower(keyword)) {
				continue
			}
		}
		meta := map[string]interface{}{}
		if route.HasMeta {
			meta["summary"] = route.RouteMeta.Summary
			meta["module"] = route.RouteMeta.Module
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
			Module:  route.Module,
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
	return apiregistry.SyncRoutes(s.db, s.logger, s.router.Routes())
}

func (s *service) ListBindings(endpointID uuid.UUID) ([]user.APIEndpointPermissionBinding, error) {
	return s.bindingRepo.ListByEndpointID(endpointID)
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
	if endpoint.Module == "" {
		endpoint.Module = "manual"
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
			"module":        endpoint.Module,
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
	for _, key := range permissionKeys {
		target := strings.TrimSpace(key)
		if target == "" {
			continue
		}
		if _, ok := seen[target]; ok {
			continue
		}
		seen[target] = struct{}{}
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

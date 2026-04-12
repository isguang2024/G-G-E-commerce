package apiendpointaccess

import (
	"errors"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/gg-ecommerce/backend/internal/api/legacyresp"
	"github.com/gg-ecommerce/backend/internal/modules/system/models"
)

type EndpointMeta struct {
	ID     uuid.UUID
	Method string
	Path   string
	Status string
}

type Service interface {
	Refresh() error
	RefreshByRoute(method, path string) error
	RemoveByRoute(method, path string)
	Get(method, path string) (EndpointMeta, bool)
	RequireActiveEndpoint() gin.HandlerFunc
}

type service struct {
	db       *gorm.DB
	logger   *zap.Logger
	mu       sync.RWMutex
	routeMap map[string]EndpointMeta
}

func NewService(db *gorm.DB, logger *zap.Logger) Service {
	return &service{
		db:       db,
		logger:   logger,
		routeMap: make(map[string]EndpointMeta),
	}
}

func (s *service) Refresh() error {
	var endpoints []models.APIEndpoint
	if err := s.db.Select("id", "method", "path", "status").Find(&endpoints).Error; err != nil {
		return err
	}

	next := make(map[string]EndpointMeta, len(endpoints))
	for _, endpoint := range endpoints {
		item := EndpointMeta{
			ID:     endpoint.ID,
			Method: strings.ToUpper(strings.TrimSpace(endpoint.Method)),
			Path:   strings.TrimSpace(endpoint.Path),
			Status: strings.TrimSpace(endpoint.Status),
		}
		if endpointStatusCode(item.Status) != 1 {
			continue
		}
		next[routeKey(item.Method, item.Path)] = item
	}

	s.mu.Lock()
	s.routeMap = next
	s.mu.Unlock()
	return nil
}

func (s *service) Get(method, path string) (EndpointMeta, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	item, ok := s.routeMap[routeKey(method, path)]
	return item, ok
}

func (s *service) RefreshByRoute(method, path string) error {
	targetMethod := strings.ToUpper(strings.TrimSpace(method))
	targetPath := strings.TrimSpace(path)
	if targetMethod == "" || targetPath == "" {
		return nil
	}

	var endpoint models.APIEndpoint
	err := s.db.Select("id", "method", "path", "status").
		Where("method = ? AND path = ?", targetMethod, targetPath).
		First(&endpoint).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.RemoveByRoute(targetMethod, targetPath)
			return nil
		}
		return err
	}

	item := EndpointMeta{
		ID:     endpoint.ID,
		Method: strings.ToUpper(strings.TrimSpace(endpoint.Method)),
		Path:   strings.TrimSpace(endpoint.Path),
		Status: strings.TrimSpace(endpoint.Status),
	}

	s.mu.Lock()
	key := routeKey(item.Method, item.Path)
	if endpointStatusCode(item.Status) == 1 {
		s.routeMap[key] = item
	} else {
		delete(s.routeMap, key)
	}
	s.mu.Unlock()
	return nil
}

func (s *service) RemoveByRoute(method, path string) {
	targetMethod := strings.ToUpper(strings.TrimSpace(method))
	targetPath := strings.TrimSpace(path)
	if targetMethod == "" || targetPath == "" {
		return
	}

	s.mu.Lock()
	delete(s.routeMap, routeKey(targetMethod, targetPath))
	s.mu.Unlock()
}

func (s *service) RequireActiveEndpoint() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := strings.TrimSpace(c.FullPath())
		if path == "" {
			c.Next()
			return
		}
		if shouldBypassEndpointStatusCheck(path) {
			c.Next()
			return
		}

		meta, ok := s.Get(c.Request.Method, path)
		if !ok {
			c.Next()
			return
		}

		c.Set("api_endpoint_id", meta.ID.String())
		c.Set("api_endpoint_status", meta.Status)
		if endpointStatusCode(meta.Status) == 1 {
			legacyresp.Forbidden(c, "当前 API 已停用")
			return
		}

		c.Next()
	}
}

func shouldBypassEndpointStatusCheck(path string) bool {
	switch {
	case path == "/api/v1/api-endpoints":
		return true
	case strings.HasPrefix(path, "/api/v1/api-endpoints/"):
		return true
	default:
		return false
	}
}

func routeKey(method, path string) string {
	return strings.ToUpper(strings.TrimSpace(method)) + " " + strings.TrimSpace(path)
}

func endpointStatusCode(status string) int {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "suspended", "disabled", "stopped", "inactive", "off", "1":
		return 1
	default:
		return 0
	}
}

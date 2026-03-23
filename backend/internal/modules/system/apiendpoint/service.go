package apiendpoint

import (
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/gg-ecommerce/backend/internal/modules/system/user"
	"github.com/gg-ecommerce/backend/internal/pkg/apiregistry"
)

type ListRequest struct {
	Current      int
	Size         int
	Method       string
	Path         string
	Module       string
	FeatureKind  string
	ResourceCode string
	ActionCode   string
	Status       string
}

type Service interface {
	List(req *ListRequest) ([]user.APIEndpoint, int64, error)
	Sync() error
}

type service struct {
	db     *gorm.DB
	repo   user.APIEndpointRepository
	router *gin.Engine
	logger *zap.Logger
}

func NewService(db *gorm.DB, repo user.APIEndpointRepository, router *gin.Engine, logger *zap.Logger) Service {
	return &service{db: db, repo: repo, router: router, logger: logger}
}

func (s *service) List(req *ListRequest) ([]user.APIEndpoint, int64, error) {
	if req.Current <= 0 {
		req.Current = 1
	}
	if req.Size <= 0 {
		req.Size = 20
	}
	return s.repo.List((req.Current-1)*req.Size, req.Size, &user.APIEndpointListParams{
		Method:       strings.ToUpper(strings.TrimSpace(req.Method)),
		Path:         strings.TrimSpace(req.Path),
		Module:       strings.TrimSpace(req.Module),
		FeatureKind:  strings.TrimSpace(req.FeatureKind),
		ResourceCode: strings.TrimSpace(req.ResourceCode),
		ActionCode:   strings.TrimSpace(req.ActionCode),
		Status:       strings.TrimSpace(req.Status),
	})
}

func (s *service) Sync() error {
	if s.router == nil {
		return nil
	}
	return apiregistry.SyncRoutes(s.db, s.logger, s.router.Routes())
}

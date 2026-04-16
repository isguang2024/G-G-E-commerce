package upload

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"

	"go.uber.org/zap"

	"github.com/gg-ecommerce/backend/internal/config"
	"github.com/gg-ecommerce/backend/internal/modules/system/models"
)

const (
	DriverErrorCodeInvalidInput          = "invalid_input"
	DriverErrorCodeCapabilityUnsupported = "capability_unsupported"
	DriverErrorCodeUnavailable           = "unavailable"
	DriverErrorCodePathViolation         = "path_violation"
	DriverErrorCodeIO                    = "io"
)

type DriverCapabilities struct {
	Relay     bool `json:"relay"`
	Direct    bool `json:"direct"`
	STS       bool `json:"sts"`
	Multipart bool `json:"multipart"`
	Delete    bool `json:"delete"`
}

type UploadRequest struct {
	TenantID         string
	StorageKey       string
	PublicPath       string
	OriginalFilename string
	ContentType      string
	Size             int64
	File             io.Reader
	PublicBaseURL    string
}

type UploadResult struct {
	StorageKey  string
	URL         string
	ContentType string
	Size        int64
	Checksum    string
}

type DeleteRequest struct {
	TenantID   string
	StorageKey string
}

type DirectUploadRequest struct {
	TenantID    string
	StorageKey  string
	ContentType string
	Size        int64
}

type DirectUploadResult struct {
	Method  string
	URL     string
	Headers map[string]string
	Form    map[string]string
}

type ObjectStat struct {
	StorageKey  string
	ContentType string
	Size        int64
	Checksum    string
}

type Driver interface {
	Name() string
	Capabilities() DriverCapabilities
	HealthCheck(ctx context.Context) error
	Upload(ctx context.Context, req UploadRequest) (*UploadResult, error)
	Delete(ctx context.Context, req DeleteRequest) error
	PrepareDirectUpload(ctx context.Context, req DirectUploadRequest) (*DirectUploadResult, error)
}

type ObjectStatProvider interface {
	StatObject(ctx context.Context, storageKey string) (*ObjectStat, error)
}

type DriverFactoryInput struct {
	Config   *config.Config
	Provider models.StorageProvider
	Bucket   models.StorageBucket
	Logger   *zap.Logger
}

type DriverFactory func(input DriverFactoryInput) (Driver, error)

type DriverRegistry struct {
	mu        sync.RWMutex
	factories map[string]DriverFactory
}

type DriverError struct {
	Code      string
	Driver    string
	Operation string
	Message   string
	Err       error
	Retryable bool
}

func (e *DriverError) Error() string {
	if e == nil {
		return ""
	}
	parts := make([]string, 0, 3)
	if e.Driver != "" {
		parts = append(parts, e.Driver)
	}
	if e.Operation != "" {
		parts = append(parts, e.Operation)
	}
	if e.Message != "" {
		parts = append(parts, e.Message)
	}
	if len(parts) == 0 && e.Err != nil {
		return e.Err.Error()
	}
	base := strings.Join(parts, ": ")
	if e.Err == nil {
		return base
	}
	if base == "" {
		return e.Err.Error()
	}
	return base + ": " + e.Err.Error()
}

func (e *DriverError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

func NewDriverRegistry() *DriverRegistry {
	return &DriverRegistry{
		factories: make(map[string]DriverFactory),
	}
}

func (r *DriverRegistry) Register(driverName string, factory DriverFactory) error {
	name := strings.TrimSpace(driverName)
	if name == "" {
		return errors.New("driver name is empty")
	}
	if factory == nil {
		return errors.New("driver factory is nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.factories[name]; exists {
		return fmt.Errorf("driver %s already registered", name)
	}
	r.factories[name] = factory
	return nil
}

func (r *DriverRegistry) MustRegister(driverName string, factory DriverFactory) {
	if err := r.Register(driverName, factory); err != nil {
		panic(err)
	}
}

func (r *DriverRegistry) Open(input DriverFactoryInput) (Driver, error) {
	driverName := strings.TrimSpace(input.Provider.Driver)
	if driverName == "" {
		return nil, &DriverError{
			Code:      DriverErrorCodeInvalidInput,
			Operation: "open",
			Message:   "provider driver is empty",
		}
	}
	if status := strings.TrimSpace(input.Provider.Status); status != "" && status != models.UploadProviderStatusReady {
		return nil, &DriverError{
			Code:      DriverErrorCodeUnavailable,
			Driver:    driverName,
			Operation: "open",
			Message:   "provider is not ready",
		}
	}

	r.mu.RLock()
	factory, exists := r.factories[driverName]
	r.mu.RUnlock()
	if !exists {
		return nil, &DriverError{
			Code:      DriverErrorCodeUnavailable,
			Driver:    driverName,
			Operation: "open",
			Message:   "driver is not registered",
		}
	}
	return factory(input)
}

func (r *DriverRegistry) Probe(driverName string) (DriverCapabilities, bool) {
	name := strings.TrimSpace(driverName)
	if name == "" {
		return DriverCapabilities{}, false
	}
	r.mu.RLock()
	factory, exists := r.factories[name]
	r.mu.RUnlock()
	if !exists {
		return DriverCapabilities{}, false
	}
	driver, err := factory(DriverFactoryInput{
		Provider: models.StorageProvider{Driver: name, Status: models.UploadProviderStatusReady},
		Bucket:   models.StorageBucket{},
	})
	if err != nil {
		return DriverCapabilities{}, false
	}
	return driver.Capabilities(), true
}

func (r *DriverRegistry) HealthCheck(ctx context.Context, input DriverFactoryInput) error {
	driver, err := r.Open(input)
	if err != nil {
		return err
	}
	return driver.HealthCheck(ctx)
}

func newCapabilityUnsupported(driverName, capability string) error {
	return &DriverError{
		Code:      DriverErrorCodeCapabilityUnsupported,
		Driver:    driverName,
		Operation: capability,
		Message:   "capability is not supported by this driver",
	}
}

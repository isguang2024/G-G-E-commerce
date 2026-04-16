package upload

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/maben/backend/internal/modules/system/models"
)

type stubDriver struct {
	capabilities DriverCapabilities
	healthErr    error
}

func (d *stubDriver) Name() string { return "stub" }

func (d *stubDriver) Capabilities() DriverCapabilities { return d.capabilities }

func (d *stubDriver) HealthCheck(context.Context) error { return d.healthErr }

func (d *stubDriver) Upload(context.Context, UploadRequest) (*UploadResult, error) {
	return &UploadResult{}, nil
}

func (d *stubDriver) Delete(context.Context, DeleteRequest) error { return nil }

func (d *stubDriver) PrepareDirectUpload(context.Context, DirectUploadRequest) (*DirectUploadResult, error) {
	return nil, newCapabilityUnsupported("stub", "direct")
}

func TestDriverRegistryLifecycle(t *testing.T) {
	registry := NewDriverRegistry()
	driver := &stubDriver{capabilities: DriverCapabilities{Relay: true, Delete: true}}
	factory := func(DriverFactoryInput) (Driver, error) { return driver, nil }

	if err := registry.Register("stub", factory); err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if err := registry.Register("stub", factory); err == nil {
		t.Fatalf("Register() should reject duplicate driver")
	}

	opened, err := registry.Open(DriverFactoryInput{
		Provider: models.StorageProvider{Driver: "stub", Status: models.UploadProviderStatusReady},
	})
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	if opened.Name() != "stub" {
		t.Fatalf("Open() name = %q, want %q", opened.Name(), "stub")
	}

	capabilities, ok := registry.Probe("stub")
	if !ok || !capabilities.Relay || !capabilities.Delete {
		t.Fatalf("Probe() = (%+v,%v), want relay+delete true", capabilities, ok)
	}
	if err := registry.HealthCheck(context.Background(), DriverFactoryInput{
		Provider: models.StorageProvider{Driver: "stub", Status: models.UploadProviderStatusReady},
	}); err != nil {
		t.Fatalf("HealthCheck() error = %v", err)
	}
}

func TestDriverRegistryErrorsAndDriverErrorFormatting(t *testing.T) {
	registry := NewDriverRegistry()

	if _, err := registry.Open(DriverFactoryInput{}); err == nil {
		t.Fatalf("Open() should reject empty driver")
	}
	if _, err := registry.Open(DriverFactoryInput{
		Provider: models.StorageProvider{Driver: "stub", Status: "disabled"},
	}); err == nil {
		t.Fatalf("Open() should reject non-ready provider")
	}
	if _, err := registry.Open(DriverFactoryInput{
		Provider: models.StorageProvider{Driver: "missing", Status: models.UploadProviderStatusReady},
	}); err == nil {
		t.Fatalf("Open() should reject missing driver")
	}

	inner := errors.New("disk failed")
	driverErr := &DriverError{
		Code:      DriverErrorCodeIO,
		Driver:    "local",
		Operation: "upload",
		Message:   "write file",
		Err:       inner,
	}
	if !strings.Contains(driverErr.Error(), "local: upload: write file: disk failed") {
		t.Fatalf("DriverError.Error() = %q", driverErr.Error())
	}
	if !errors.Is(driverErr, inner) {
		t.Fatalf("DriverError should unwrap inner error")
	}
}

func TestDriverRegistryMustRegisterPanicsOnDuplicate(t *testing.T) {
	registry := NewDriverRegistry()
	factory := func(DriverFactoryInput) (Driver, error) { return &stubDriver{}, nil }
	registry.MustRegister("stub", factory)

	defer func() {
		if recover() == nil {
			t.Fatalf("MustRegister() should panic on duplicate driver")
		}
	}()
	registry.MustRegister("stub", factory)
}


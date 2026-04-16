package uploadtest

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/gg-ecommerce/backend/internal/modules/system/upload"
)

type DriverFactory func(t *testing.T) upload.Driver

type ContractCase struct {
	Name               string
	Factory            DriverFactory
	WantCapabilities   upload.DriverCapabilities
	ExpectDirectError  bool
	ExpectDeleteError  bool
	SupportsObjectStat bool
}

func RunDriverContractSuite(t *testing.T, tc ContractCase) {
	t.Helper()
	if tc.Factory == nil {
		t.Fatal("contract case factory is nil")
	}

	t.Run(tc.Name+"/capabilities", func(t *testing.T) {
		driver := tc.Factory(t)
		if got := driver.Capabilities(); got != tc.WantCapabilities {
			t.Fatalf("Capabilities() = %+v, want %+v", got, tc.WantCapabilities)
		}
	})

	t.Run(tc.Name+"/health", func(t *testing.T) {
		driver := tc.Factory(t)
		if err := driver.HealthCheck(context.Background()); err != nil {
			t.Fatalf("HealthCheck() error = %v", err)
		}
	})

	t.Run(tc.Name+"/relay", func(t *testing.T) {
		driver := tc.Factory(t)
		result, err := driver.Upload(context.Background(), upload.UploadRequest{
			TenantID:         "tenant-test",
			StorageKey:       "tenant-test/contracts/demo.txt",
			PublicPath:       "contracts/demo.txt",
			OriginalFilename: "demo.txt",
			ContentType:      "text/plain",
			Size:             4,
			File:             strings.NewReader("demo"),
			PublicBaseURL:    "/uploads",
		})
		if tc.WantCapabilities.Relay {
			if err != nil {
				t.Fatalf("Upload() error = %v", err)
			}
			if result == nil || strings.TrimSpace(result.StorageKey) == "" {
				t.Fatalf("Upload() result = %#v, want non-empty storage key", result)
			}
			return
		}
		assertCapabilityUnsupported(t, err, "upload")
	})

	t.Run(tc.Name+"/direct", func(t *testing.T) {
		driver := tc.Factory(t)
		result, err := driver.PrepareDirectUpload(context.Background(), upload.DirectUploadRequest{
			TenantID:    "tenant-test",
			StorageKey:  "tenant-test/contracts/direct.txt",
			ContentType: "text/plain",
			Size:        4,
		})
		if tc.WantCapabilities.Direct {
			if err != nil {
				t.Fatalf("PrepareDirectUpload() error = %v", err)
			}
			if result == nil || strings.TrimSpace(result.Method) == "" {
				t.Fatalf("PrepareDirectUpload() result = %#v, want non-empty method", result)
			}
			return
		}
		if tc.ExpectDirectError {
			if err == nil {
				t.Fatal("PrepareDirectUpload() error = nil, want error")
			}
			return
		}
		assertCapabilityUnsupported(t, err, "prepare_direct_upload")
	})

	t.Run(tc.Name+"/delete", func(t *testing.T) {
		driver := tc.Factory(t)
		err := driver.Delete(context.Background(), upload.DeleteRequest{
			TenantID:   "tenant-test",
			StorageKey: "tenant-test/contracts/demo.txt",
		})
		if tc.WantCapabilities.Delete {
			if err != nil && !tc.ExpectDeleteError {
				t.Fatalf("Delete() error = %v", err)
			}
			return
		}
		assertCapabilityUnsupported(t, err, "delete")
	})

	t.Run(tc.Name+"/object-stat", func(t *testing.T) {
		driver := tc.Factory(t)
		statProvider, ok := driver.(upload.ObjectStatProvider)
		if tc.SupportsObjectStat {
			if !ok {
				t.Fatal("driver does not implement ObjectStatProvider")
			}
			_, _ = statProvider.StatObject(context.Background(), "tenant-test/contracts/demo.txt")
			return
		}
		if ok {
			t.Fatal("driver unexpectedly implements ObjectStatProvider")
		}
	})
}

func assertCapabilityUnsupported(t *testing.T, err error, operation string) {
	t.Helper()
	if err == nil {
		t.Fatalf("%s error = nil, want capability unsupported", operation)
	}
	var driverErr *upload.DriverError
	if !errors.As(err, &driverErr) {
		t.Fatalf("%s error type = %T, want *upload.DriverError", operation, err)
	}
	if driverErr.Code != upload.DriverErrorCodeCapabilityUnsupported {
		t.Fatalf("%s error code = %q, want %q", operation, driverErr.Code, upload.DriverErrorCodeCapabilityUnsupported)
	}
}

type MemoryDriver struct {
	DriverName string
	Caps       upload.DriverCapabilities
	Objects    map[string][]byte
}

func NewMemoryDriver(name string, caps upload.DriverCapabilities) *MemoryDriver {
	return &MemoryDriver{
		DriverName: name,
		Caps:       caps,
		Objects:    map[string][]byte{},
	}
}

func (d *MemoryDriver) Name() string { return d.DriverName }

func (d *MemoryDriver) Capabilities() upload.DriverCapabilities { return d.Caps }

func (d *MemoryDriver) HealthCheck(context.Context) error { return nil }

func (d *MemoryDriver) Upload(_ context.Context, req upload.UploadRequest) (*upload.UploadResult, error) {
	if !d.Caps.Relay {
		return nil, unsupported(d.Name(), "upload")
	}
	body, err := io.ReadAll(req.File)
	if err != nil {
		return nil, err
	}
	d.Objects[req.StorageKey] = body
	return &upload.UploadResult{
		StorageKey:  req.StorageKey,
		URL:         strings.TrimRight(req.PublicBaseURL, "/") + "/" + strings.TrimLeft(req.PublicPath, "/"),
		ContentType: req.ContentType,
		Size:        int64(len(body)),
	}, nil
}

func (d *MemoryDriver) Delete(_ context.Context, req upload.DeleteRequest) error {
	if !d.Caps.Delete {
		return unsupported(d.Name(), "delete")
	}
	delete(d.Objects, req.StorageKey)
	return nil
}

func (d *MemoryDriver) PrepareDirectUpload(_ context.Context, req upload.DirectUploadRequest) (*upload.DirectUploadResult, error) {
	if !d.Caps.Direct {
		return nil, unsupported(d.Name(), "prepare_direct_upload")
	}
	return &upload.DirectUploadResult{
		Method: "POST",
		URL:    "https://example.invalid/upload",
		Form: map[string]string{
			"key": req.StorageKey,
		},
	}, nil
}

func unsupported(driverName, operation string) error {
	return &upload.DriverError{
		Code:      upload.DriverErrorCodeCapabilityUnsupported,
		Driver:    driverName,
		Operation: operation,
		Message:   "capability is not supported by this driver",
	}
}

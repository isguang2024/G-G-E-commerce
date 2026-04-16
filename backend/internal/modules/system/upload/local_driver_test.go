package upload

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLocalDriverUploadWritesFileAndBuildsURL(t *testing.T) {
	rootDir := t.TempDir()
	driver := NewLocalDriver(LocalDriverConfig{
		RootPath:      rootDir,
		TempDirectory: filepath.Join(rootDir, ".tmp"),
	})

	result, err := driver.Upload(context.Background(), UploadRequest{
		StorageKey:       "public-media/2026/04/16/test.txt",
		PublicPath:       "2026/04/16/test.txt",
		OriginalFilename: "test.txt",
		ContentType:      "text/plain",
		File:             strings.NewReader("hello"),
		PublicBaseURL:    "/uploads/public-media",
	})
	if err != nil {
		t.Fatalf("Upload() error = %v", err)
	}
	if result.URL != "/uploads/public-media/2026/04/16/test.txt" {
		t.Fatalf("Upload() URL = %q, want %q", result.URL, "/uploads/public-media/2026/04/16/test.txt")
	}
	content, readErr := os.ReadFile(filepath.Join(rootDir, "public-media", "2026", "04", "16", "test.txt"))
	if readErr != nil {
		t.Fatalf("ReadFile() error = %v", readErr)
	}
	if string(content) != "hello" {
		t.Fatalf("stored content = %q, want %q", string(content), "hello")
	}
}

func TestLocalDriverRejectsPathTraversal(t *testing.T) {
	driver := NewLocalDriver(LocalDriverConfig{
		RootPath: t.TempDir(),
	})

	_, err := driver.Upload(context.Background(), UploadRequest{
		StorageKey:       "../escape.txt",
		OriginalFilename: "escape.txt",
		ContentType:      "text/plain",
		File:             strings.NewReader("blocked"),
		PublicBaseURL:    "/uploads",
	})
	var driverErr *DriverError
	if !errors.As(err, &driverErr) {
		t.Fatalf("Upload() error = %v, want DriverError", err)
	}
	if driverErr.Code != DriverErrorCodePathViolation {
		t.Fatalf("DriverError.Code = %q, want %q", driverErr.Code, DriverErrorCodePathViolation)
	}
}

func TestLocalDriverPrepareDirectUploadIsUnsupported(t *testing.T) {
	driver := NewLocalDriver(LocalDriverConfig{
		RootPath: t.TempDir(),
	})
	_, err := driver.PrepareDirectUpload(context.Background(), DirectUploadRequest{})
	var driverErr *DriverError
	if !errors.As(err, &driverErr) {
		t.Fatalf("PrepareDirectUpload() error = %v, want DriverError", err)
	}
	if driverErr.Code != DriverErrorCodeCapabilityUnsupported {
		t.Fatalf("DriverError.Code = %q, want %q", driverErr.Code, DriverErrorCodeCapabilityUnsupported)
	}
}

func TestLocalDriverCapabilitiesHealthCheckAndDelete(t *testing.T) {
	rootDir := t.TempDir()
	driver := NewLocalDriver(LocalDriverConfig{
		RootPath:      rootDir,
		TempDirectory: filepath.Join(rootDir, ".tmp"),
	})

	capabilities := driver.Capabilities()
	if !capabilities.Relay || !capabilities.Delete || capabilities.Direct {
		t.Fatalf("Capabilities() = %+v", capabilities)
	}
	if err := driver.HealthCheck(context.Background()); err != nil {
		t.Fatalf("HealthCheck() error = %v", err)
	}

	target := filepath.Join(rootDir, "public-media", "demo.txt")
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(target, []byte("demo"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	if err := driver.Delete(context.Background(), DeleteRequest{StorageKey: "public-media/demo.txt"}); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if _, err := os.Stat(target); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("Delete() stat error = %v, want not exist", err)
	}
}

func TestLocalDriverHealthCheckFailure(t *testing.T) {
	rootDir := t.TempDir()
	blockingFile := filepath.Join(rootDir, "blocked")
	if err := os.WriteFile(blockingFile, []byte("x"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	driver := NewLocalDriver(LocalDriverConfig{
		RootPath:      blockingFile,
		TempDirectory: filepath.Join(blockingFile, ".tmp"),
	})
	if err := driver.HealthCheck(context.Background()); err == nil {
		t.Fatalf("HealthCheck() should fail when root path is a file")
	}
}

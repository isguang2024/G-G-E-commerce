package upload

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"

	"go.uber.org/zap"

	"github.com/gg-ecommerce/backend/internal/modules/system/models"
)

type LocalDriverConfig struct {
	RootPath      string
	TempDirectory string
	Logger        *zap.Logger
}

type LocalDriver struct {
	rootPath      string
	tempDirectory string
	logger        *zap.Logger
}

func NewLocalDriver(cfg LocalDriverConfig) *LocalDriver {
	rootPath := strings.TrimSpace(cfg.RootPath)
	if rootPath == "" {
		rootPath = filepath.Clean(filepath.Join(".", "data", "uploads"))
	}
	tempDirectory := strings.TrimSpace(cfg.TempDirectory)
	if tempDirectory == "" {
		tempDirectory = filepath.Join(rootPath, ".tmp")
	}
	return &LocalDriver{
		rootPath:      filepath.Clean(rootPath),
		tempDirectory: filepath.Clean(tempDirectory),
		logger:        cfg.Logger,
	}
}

func (d *LocalDriver) Name() string {
	return models.UploadProviderDriverLocal
}

func (d *LocalDriver) Capabilities() DriverCapabilities {
	return DriverCapabilities{
		Relay:  true,
		Delete: true,
	}
}

func (d *LocalDriver) HealthCheck(_ context.Context) error {
	if err := os.MkdirAll(d.rootPath, 0o755); err != nil {
		return &DriverError{
			Code:      DriverErrorCodeUnavailable,
			Driver:    d.Name(),
			Operation: "health_check",
			Message:   "create root directory failed",
			Err:       err,
		}
	}
	if err := os.MkdirAll(d.tempDirectory, 0o755); err != nil {
		return &DriverError{
			Code:      DriverErrorCodeUnavailable,
			Driver:    d.Name(),
			Operation: "health_check",
			Message:   "create temp directory failed",
			Err:       err,
		}
	}
	return nil
}

func (d *LocalDriver) Upload(_ context.Context, req UploadRequest) (*UploadResult, error) {
	if req.File == nil {
		return nil, &DriverError{
			Code:      DriverErrorCodeInvalidInput,
			Driver:    d.Name(),
			Operation: "upload",
			Message:   "file reader is nil",
		}
	}

	targetPath, err := d.resolveStoragePath(req.StorageKey)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		return nil, &DriverError{
			Code:      DriverErrorCodeIO,
			Driver:    d.Name(),
			Operation: "upload",
			Message:   "create target directory failed",
			Err:       err,
		}
	}
	if err := os.MkdirAll(d.tempDirectory, 0o755); err != nil {
		return nil, &DriverError{
			Code:      DriverErrorCodeIO,
			Driver:    d.Name(),
			Operation: "upload",
			Message:   "create temp directory failed",
			Err:       err,
		}
	}

	tempFile, err := os.CreateTemp(d.tempDirectory, "upload-*.part")
	if err != nil {
		return nil, &DriverError{
			Code:      DriverErrorCodeIO,
			Driver:    d.Name(),
			Operation: "upload",
			Message:   "create temp file failed",
			Err:       err,
		}
	}
	tempPath := tempFile.Name()
	cleanupTemp := true
	defer func() {
		if cleanupTemp {
			_ = os.Remove(tempPath)
		}
	}()

	hasher := sha256.New()
	size, copyErr := io.Copy(io.MultiWriter(tempFile, hasher), req.File)
	closeErr := tempFile.Close()
	if copyErr != nil {
		return nil, &DriverError{
			Code:      DriverErrorCodeIO,
			Driver:    d.Name(),
			Operation: "upload",
			Message:   "stream file content failed",
			Err:       copyErr,
		}
	}
	if closeErr != nil {
		return nil, &DriverError{
			Code:      DriverErrorCodeIO,
			Driver:    d.Name(),
			Operation: "upload",
			Message:   "close temp file failed",
			Err:       closeErr,
		}
	}
	if err := os.Rename(tempPath, targetPath); err != nil {
		return nil, &DriverError{
			Code:      DriverErrorCodeIO,
			Driver:    d.Name(),
			Operation: "upload",
			Message:   "atomic rename failed",
			Err:       err,
		}
	}
	cleanupTemp = false

	return &UploadResult{
		StorageKey:  filepath.ToSlash(strings.TrimSpace(req.StorageKey)),
		URL:         buildPublicURL(req.PublicBaseURL, req.PublicPath),
		ContentType: normalizeContentType(req.ContentType, req.OriginalFilename),
		Size:        size,
		Checksum:    hex.EncodeToString(hasher.Sum(nil)),
	}, nil
}

func (d *LocalDriver) Delete(_ context.Context, req DeleteRequest) error {
	targetPath, err := d.resolveStoragePath(req.StorageKey)
	if err != nil {
		return err
	}
	removeErr := os.Remove(targetPath)
	if removeErr != nil && !errors.Is(removeErr, os.ErrNotExist) {
		return &DriverError{
			Code:      DriverErrorCodeIO,
			Driver:    d.Name(),
			Operation: "delete",
			Message:   "remove file failed",
			Err:       removeErr,
		}
	}
	return nil
}

func (d *LocalDriver) PrepareDirectUpload(_ context.Context, _ DirectUploadRequest) (*DirectUploadResult, error) {
	return nil, newCapabilityUnsupported(d.Name(), "prepare_direct_upload")
}

func (d *LocalDriver) resolveStoragePath(storageKey string) (string, error) {
	normalizedKey := filepath.Clean(filepath.FromSlash(strings.TrimSpace(storageKey)))
	if normalizedKey == "." || normalizedKey == "" {
		return "", &DriverError{
			Code:      DriverErrorCodeInvalidInput,
			Driver:    d.Name(),
			Operation: "path",
			Message:   "storage key is empty",
		}
	}
	if filepath.IsAbs(normalizedKey) || normalizedKey == ".." || strings.HasPrefix(normalizedKey, ".."+string(os.PathSeparator)) {
		return "", &DriverError{
			Code:      DriverErrorCodePathViolation,
			Driver:    d.Name(),
			Operation: "path",
			Message:   "storage key escapes root path",
		}
	}
	targetPath := filepath.Join(d.rootPath, normalizedKey)
	relativeToRoot, err := filepath.Rel(d.rootPath, targetPath)
	if err != nil {
		return "", &DriverError{
			Code:      DriverErrorCodePathViolation,
			Driver:    d.Name(),
			Operation: "path",
			Message:   "resolve storage path failed",
			Err:       err,
		}
	}
	if relativeToRoot == ".." || strings.HasPrefix(relativeToRoot, ".."+string(os.PathSeparator)) {
		return "", &DriverError{
			Code:      DriverErrorCodePathViolation,
			Driver:    d.Name(),
			Operation: "path",
			Message:   "storage key escapes root path",
		}
	}
	return targetPath, nil
}

package upload

import (
	"errors"
	"fmt"
	"strings"
)

const (
	UploadProviderDriverAliyunOSS   = "aliyun_oss"
	DriverErrorCodeObjectNotFound   = "object_not_found"
	DriverErrorCodePermissionDenied = "permission_denied"
	DriverErrorCodeRateLimited      = "rate_limited"
	DriverErrorCodeProviderError    = "provider_error"
)

type OSSAPIError struct {
	Code       string
	Message    string
	StatusCode int
}

func (e *OSSAPIError) Error() string {
	if e == nil {
		return ""
	}
	if strings.TrimSpace(e.Message) != "" {
		return e.Message
	}
	return e.Code
}

func mapAliyunOSSDriverError(operation string, err error) error {
	if err == nil {
		return nil
	}
	var apiErr *OSSAPIError
	if errors.As(err, &apiErr) {
		switch strings.TrimSpace(apiErr.Code) {
		case "NoSuchKey":
			return &DriverError{
				Code:      DriverErrorCodeObjectNotFound,
				Driver:    UploadProviderDriverAliyunOSS,
				Operation: operation,
				Message:   "oss object not found",
				Err:       err,
			}
		case "AccessDenied":
			return &DriverError{
				Code:      DriverErrorCodePermissionDenied,
				Driver:    UploadProviderDriverAliyunOSS,
				Operation: operation,
				Message:   "oss access denied",
				Err:       err,
			}
		case "RequestTimeout":
			return &DriverError{
				Code:      DriverErrorCodeRateLimited,
				Driver:    UploadProviderDriverAliyunOSS,
				Operation: operation,
				Message:   "oss request timeout",
				Err:       err,
				Retryable: true,
			}
		}
	}
	return &DriverError{
		Code:      DriverErrorCodeProviderError,
		Driver:    UploadProviderDriverAliyunOSS,
		Operation: operation,
		Message:   "oss provider request failed",
		Err:       err,
	}
}

func newOSSAPIError(code, message string, statusCode int) *OSSAPIError {
	return &OSSAPIError{
		Code:       strings.TrimSpace(code),
		Message:    strings.TrimSpace(message),
		StatusCode: statusCode,
	}
}

func (e *OSSAPIError) String() string {
	if e == nil {
		return ""
	}
	return fmt.Sprintf("%s (%d): %s", e.Code, e.StatusCode, e.Message)
}

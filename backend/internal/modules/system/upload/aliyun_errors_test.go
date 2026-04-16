package upload

import "testing"

func TestMapAliyunOSSDriverError(t *testing.T) {
	cases := []struct {
		name      string
		err       error
		wantCode  string
		retryable bool
	}{
		{
			name:      "NoSuchKey",
			err:       newOSSAPIError("NoSuchKey", "missing object", 404),
			wantCode:  DriverErrorCodeObjectNotFound,
			retryable: false,
		},
		{
			name:      "AccessDenied",
			err:       newOSSAPIError("AccessDenied", "denied", 403),
			wantCode:  DriverErrorCodePermissionDenied,
			retryable: false,
		},
		{
			name:      "RequestTimeout",
			err:       newOSSAPIError("RequestTimeout", "timeout", 408),
			wantCode:  DriverErrorCodeRateLimited,
			retryable: true,
		},
		{
			name:      "Other",
			err:       newOSSAPIError("InternalError", "boom", 500),
			wantCode:  DriverErrorCodeProviderError,
			retryable: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mapped := mapAliyunOSSDriverError("upload", tc.err)
			driverErr, ok := mapped.(*DriverError)
			if !ok {
				t.Fatalf("mapped error type = %T, want *DriverError", mapped)
			}
			if driverErr.Code != tc.wantCode {
				t.Fatalf("DriverError.Code = %q, want %q", driverErr.Code, tc.wantCode)
			}
			if driverErr.Retryable != tc.retryable {
				t.Fatalf("DriverError.Retryable = %v, want %v", driverErr.Retryable, tc.retryable)
			}
		})
	}
}

func TestOSSAPIErrorString(t *testing.T) {
	err := newOSSAPIError("NoSuchKey", "missing", 404)
	if err.String() != "NoSuchKey (404): missing" {
		t.Fatalf("OSSAPIError.String() = %q", err.String())
	}
}

package upload_test

import (
	"testing"

	"github.com/maben/backend/internal/modules/system/upload"
	"github.com/maben/backend/internal/modules/system/upload/uploadtest"
)

func TestMemoryDriverContractHarness(t *testing.T) {
	uploadtest.RunDriverContractSuite(t, uploadtest.ContractCase{
		Name: "memory-driver",
		Factory: func(t *testing.T) upload.Driver {
			t.Helper()
			return uploadtest.NewMemoryDriver("memory", upload.DriverCapabilities{
				Relay:     true,
				Direct:    true,
				Delete:    true,
				Multipart: false,
				STS:       false,
			})
		},
		WantCapabilities: upload.DriverCapabilities{
			Relay:     true,
			Direct:    true,
			Delete:    true,
			Multipart: false,
			STS:       false,
		},
		SupportsObjectStat: false,
	})
}


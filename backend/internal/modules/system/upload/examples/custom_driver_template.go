package examples

import (
	"context"

	"github.com/maben/backend/internal/modules/system/upload"
)

// CustomDriverTemplate 展示一个最小自定义 driver 骨架。
// 生产实现应当在独立包中保留自己的配置解析、超时和鉴权策略。
type CustomDriverTemplate struct {
	NameValue string
	Caps      upload.DriverCapabilities
}

func NewCustomDriverTemplate(name string, caps upload.DriverCapabilities) *CustomDriverTemplate {
	return &CustomDriverTemplate{
		NameValue: name,
		Caps:      caps,
	}
}

func (d *CustomDriverTemplate) Name() string {
	return d.NameValue
}

func (d *CustomDriverTemplate) Capabilities() upload.DriverCapabilities {
	return d.Caps
}

func (d *CustomDriverTemplate) HealthCheck(context.Context) error {
	return nil
}

func (d *CustomDriverTemplate) Upload(context.Context, upload.UploadRequest) (*upload.UploadResult, error) {
	return nil, unsupported(d.Name(), "upload")
}

func (d *CustomDriverTemplate) Delete(context.Context, upload.DeleteRequest) error {
	return unsupported(d.Name(), "delete")
}

func (d *CustomDriverTemplate) PrepareDirectUpload(context.Context, upload.DirectUploadRequest) (*upload.DirectUploadResult, error) {
	return nil, unsupported(d.Name(), "prepare_direct_upload")
}

func unsupported(driverName, operation string) error {
	return &upload.DriverError{
		Code:      upload.DriverErrorCodeCapabilityUnsupported,
		Driver:    driverName,
		Operation: operation,
		Message:   "capability is not supported by this driver",
	}
}


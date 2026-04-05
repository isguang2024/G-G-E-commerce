package app

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type appHandlerServiceStub struct {
	listHostBindingsAppKey string
	saveHostBindingAppKey  string
	saveHostBindingReq     *SaveHostBindingRequest
}

func (s *appHandlerServiceStub) ListApps() ([]AppRecord, error) { return nil, nil }

func (s *appHandlerServiceStub) GetCurrent(host, requestedAppKey string) (*CurrentResponse, error) {
	return nil, nil
}

func (s *appHandlerServiceStub) SaveApp(req *SaveAppRequest) (*AppRecord, error) { return nil, nil }

func (s *appHandlerServiceStub) ListHostBindings(appKey string) ([]HostBindingRecord, error) {
	s.listHostBindingsAppKey = appKey
	return nil, nil
}

func (s *appHandlerServiceStub) SaveHostBinding(appKey string, req *SaveHostBindingRequest) (*HostBindingRecord, error) {
	s.saveHostBindingAppKey = appKey
	s.saveHostBindingReq = req
	return &HostBindingRecord{}, nil
}

func TestHandlerListHostBindingsRequiresExplicitAppKey(t *testing.T) {
	gin.SetMode(gin.TestMode)
	stub := &appHandlerServiceStub{}
	handler := NewHandler(zap.NewNop(), stub)

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/system/app-host-bindings", nil)
	ctx.Request = req

	handler.ListHostBindings(ctx)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
	if stub.listHostBindingsAppKey != "" {
		t.Fatalf("service should not be called without app_key, got %q", stub.listHostBindingsAppKey)
	}
}

func TestHandlerSaveHostBindingUsesQueryAppKey(t *testing.T) {
	gin.SetMode(gin.TestMode)
	stub := &appHandlerServiceStub{}
	handler := NewHandler(zap.NewNop(), stub)

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	body := bytes.NewBufferString(`{"app_key":"merchant-console","host":"admin.example.com"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/system/app-host-bindings?app_key=platform-admin", body)
	req.Header.Set("Content-Type", "application/json")
	ctx.Request = req

	handler.SaveHostBinding(ctx)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if stub.saveHostBindingAppKey != "platform-admin" {
		t.Fatalf("service appKey = %q, want %q", stub.saveHostBindingAppKey, "platform-admin")
	}
	if stub.saveHostBindingReq == nil {
		t.Fatal("service request should not be nil")
	}
	if stub.saveHostBindingReq.AppKey != "merchant-console" {
		t.Fatalf("request body app_key = %q, want %q", stub.saveHostBindingReq.AppKey, "merchant-console")
	}
}

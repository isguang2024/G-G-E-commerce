package handlers

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/gg-ecommerce/backend/api/gen"
	"github.com/gg-ecommerce/backend/internal/modules/observability/audit"
	"github.com/gg-ecommerce/backend/internal/modules/observability/logpolicy"
)

type fakePolicyRepo struct {
	listFn        func(ctx context.Context, tenantID, pipeline string, enabled *bool) ([]logpolicy.LogPolicy, error)
	getFn         func(ctx context.Context, tenantID string, id uuid.UUID) (*logpolicy.LogPolicy, error)
	createFn      func(ctx context.Context, policy *logpolicy.LogPolicy) error
	updateFn      func(ctx context.Context, policy *logpolicy.LogPolicy) error
	deleteFn      func(ctx context.Context, tenantID string, id uuid.UUID) error
	listEnabledFn func(ctx context.Context, tenantID, pipeline string) ([]logpolicy.LogPolicy, error)
}

func (f *fakePolicyRepo) List(ctx context.Context, tenantID, pipeline string, enabled *bool) ([]logpolicy.LogPolicy, error) {
	if f.listFn != nil {
		return f.listFn(ctx, tenantID, pipeline, enabled)
	}
	return nil, nil
}

func (f *fakePolicyRepo) Get(ctx context.Context, tenantID string, id uuid.UUID) (*logpolicy.LogPolicy, error) {
	if f.getFn != nil {
		return f.getFn(ctx, tenantID, id)
	}
	return nil, nil
}

func (f *fakePolicyRepo) Create(ctx context.Context, policy *logpolicy.LogPolicy) error {
	if f.createFn != nil {
		return f.createFn(ctx, policy)
	}
	return nil
}

func (f *fakePolicyRepo) Update(ctx context.Context, policy *logpolicy.LogPolicy) error {
	if f.updateFn != nil {
		return f.updateFn(ctx, policy)
	}
	return nil
}

func (f *fakePolicyRepo) Delete(ctx context.Context, tenantID string, id uuid.UUID) error {
	if f.deleteFn != nil {
		return f.deleteFn(ctx, tenantID, id)
	}
	return nil
}

func (f *fakePolicyRepo) ListEnabled(ctx context.Context, tenantID, pipeline string) ([]logpolicy.LogPolicy, error) {
	if f.listEnabledFn != nil {
		return f.listEnabledFn(ctx, tenantID, pipeline)
	}
	return nil, nil
}

type fakePolicyEngine struct {
	decideFn     func(pipeline string, fields map[string]string) logpolicy.Decision
	refreshCalls int
}

func (f *fakePolicyEngine) Decide(pipeline string, fields map[string]string) logpolicy.Decision {
	if f.decideFn != nil {
		return f.decideFn(pipeline, fields)
	}
	return logpolicy.Decision{Decision: logpolicy.DecisionAllow}
}

func (f *fakePolicyEngine) Refresh(_ context.Context) error {
	f.refreshCalls++
	return nil
}

func (f *fakePolicyEngine) Start(_ context.Context) {}

type recordingAudit struct {
	events []audit.Event
}

func (r *recordingAudit) Record(_ context.Context, e audit.Event) {
	r.events = append(r.events, e)
}

func (r *recordingAudit) Stats() audit.Stats { return audit.Stats{} }

func (r *recordingAudit) Shutdown(_ context.Context) error { return nil }

func authCtx() context.Context {
	return context.WithValue(context.Background(), CtxUserID, uuid.NewString())
}

func TestCreateLogPolicy_SuccessRefreshAndAudit(t *testing.T) {
	repo := &fakePolicyRepo{
		createFn: func(_ context.Context, policy *logpolicy.LogPolicy) error {
			policy.ID = uuid.New()
			policy.CreatedAt = time.Now().UTC()
			policy.UpdatedAt = policy.CreatedAt
			return nil
		},
	}
	engine := &fakePolicyEngine{}
	aud := &recordingAudit{}
	h := &APIHandler{
		policyRepo:   repo,
		policyEngine: engine,
		audit:        aud,
		logger:       zap.NewNop(),
	}

	req := &gen.LogPolicyCreateRequest{
		Pipeline:   gen.LogPolicyCreateRequestPipeline(logpolicy.PipelineAudit),
		MatchField: gen.LogPolicyCreateRequestMatchField(logpolicy.MatchFieldAction),
		Pattern:    "custom.action",
		Decision:   gen.LogPolicyCreateRequestDecision(logpolicy.DecisionAllow),
		Enabled:    gen.NewOptBool(true),
	}
	res, err := h.CreateLogPolicy(authCtx(), req)
	if err != nil {
		t.Fatalf("CreateLogPolicy error: %v", err)
	}

	item, ok := res.(*gen.LogPolicyItem)
	if !ok {
		t.Fatalf("want *gen.LogPolicyItem, got %T", res)
	}
	if item.Pattern != "custom.action" {
		t.Fatalf("pattern = %q, want custom.action", item.Pattern)
	}
	if engine.refreshCalls != 1 {
		t.Fatalf("refresh calls = %d, want 1", engine.refreshCalls)
	}
	if len(aud.events) != 1 || aud.events[0].Action != "observability.policy.create" {
		t.Fatalf("audit events = %+v, want one create event", aud.events)
	}
}

func TestUpdateLogPolicy_ComplianceLockedReturns409(t *testing.T) {
	id := uuid.New()
	repo := &fakePolicyRepo{
		getFn: func(_ context.Context, _ string, _ uuid.UUID) (*logpolicy.LogPolicy, error) {
			return &logpolicy.LogPolicy{
				ID:         id,
				TenantID:   logpolicy.DefaultTenantID,
				Pipeline:   logpolicy.PipelineAudit,
				MatchField: logpolicy.MatchFieldAction,
				Pattern:    "system.auth.login",
				Decision:   logpolicy.DecisionAllow,
				Enabled:    true,
			}, nil
		},
		updateFn: func(_ context.Context, _ *logpolicy.LogPolicy) error {
			t.Fatal("update should not be called for compliance lock policy")
			return nil
		},
	}
	engine := &fakePolicyEngine{}
	h := &APIHandler{
		policyRepo:   repo,
		policyEngine: engine,
		audit:        &recordingAudit{},
		logger:       zap.NewNop(),
	}

	res, err := h.UpdateLogPolicy(authCtx(), &gen.LogPolicyUpdateRequest{}, gen.UpdateLogPolicyParams{ID: id})
	if err != nil {
		t.Fatalf("UpdateLogPolicy error: %v", err)
	}
	conflict, ok := res.(*gen.UpdateLogPolicyConflict)
	if !ok {
		t.Fatalf("want *gen.UpdateLogPolicyConflict, got %T", res)
	}
	if conflict.Code != 409 {
		t.Fatalf("conflict code = %d, want 409", conflict.Code)
	}
	if engine.refreshCalls != 0 {
		t.Fatalf("refresh calls = %d, want 0", engine.refreshCalls)
	}
}

func TestPreviewLogPolicy_ReturnsDecisionAndMatchedPolicy(t *testing.T) {
	matched := &logpolicy.LogPolicy{
		ID:         uuid.New(),
		TenantID:   logpolicy.DefaultTenantID,
		Pipeline:   logpolicy.PipelineTelemetry,
		MatchField: logpolicy.MatchFieldEvent,
		Pattern:    "checkout.submit",
		Decision:   logpolicy.DecisionSample,
		Enabled:    true,
		Priority:   10,
	}
	engine := &fakePolicyEngine{
		decideFn: func(pipeline string, fields map[string]string) logpolicy.Decision {
			if pipeline != logpolicy.PipelineTelemetry || fields[logpolicy.MatchFieldEvent] != "checkout.submit" {
				t.Fatalf("unexpected decide input: pipeline=%s fields=%v", pipeline, fields)
			}
			return logpolicy.Decision{
				Decision:   logpolicy.DecisionSample,
				SampleRate: 25,
				Matched:    matched,
			}
		},
	}
	h := &APIHandler{
		policyEngine: engine,
		policyRepo:   &fakePolicyRepo{},
		audit:        &recordingAudit{},
		logger:       zap.NewNop(),
	}

	req := &gen.LogPolicyPreviewRequest{
		Pipeline: gen.LogPolicyPreviewRequestPipeline(logpolicy.PipelineTelemetry),
		Fields: gen.LogPolicyPreviewRequestFields{
			logpolicy.MatchFieldEvent: "checkout.submit",
		},
	}
	res, err := h.PreviewLogPolicy(authCtx(), req)
	if err != nil {
		t.Fatalf("PreviewLogPolicy error: %v", err)
	}

	okResp, ok := res.(*gen.LogPolicyPreviewResponse)
	if !ok {
		t.Fatalf("want *gen.LogPolicyPreviewResponse, got %T", res)
	}
	if okResp.Decision != gen.LogPolicyPreviewResponseDecisionSample {
		t.Fatalf("decision = %q, want sample", okResp.Decision)
	}
	if !okResp.Matched {
		t.Fatal("matched = false, want true")
	}
	if !okResp.SampleRate.Set || okResp.SampleRate.Value != 25 {
		t.Fatalf("sample_rate = %+v, want set to 25", okResp.SampleRate)
	}
	if !okResp.Policy.Set || okResp.Policy.Value.Pattern != "checkout.submit" {
		t.Fatalf("policy = %+v, want matched policy", okResp.Policy)
	}
}


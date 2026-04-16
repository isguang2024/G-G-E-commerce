package apperr

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-faster/jx"
	"github.com/ogen-go/ogen/ogenerrors"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/maben/backend/api/gen"
	"github.com/maben/backend/internal/modules/system/auth"
)

// ErrPermissionDenied is returned by the OpenAPIPermission middleware when
// the evaluator rejects an operation. Defined here (in apperr) to avoid an
// import cycle: apperr ← middleware ← handlers ← apperr.
var ErrPermissionDenied = errors.New("openapi permission denied")

// ── Handler 层专用哨兵（临时，待 ogen spec 校验覆盖后可删）────────────────

// ParamError 供 handler 在参数前置校验时 return nil, &apperr.ParamError{"..."}。
// mapper 将其翻译为 400 / CodeParamInvalid。
type ParamError struct{ Msg string }

func (e *ParamError) Error() string { return e.Msg }

// UnauthError 供 handler 在确认身份失败时 return nil, &apperr.UnauthError{"..."}。
// mapper 将其翻译为 401 / CodeUnauthorized。
type UnauthError struct{ Msg string }

func (e *UnauthError) Error() string { return e.Msg }

// FieldError 表达"单个字段级"校验失败。
//
// mapper 翻译为：HTTP 400 + Error{Code: CodeParamInvalid, Message: Msg|Reason,
// Details: {Field: Reason}}。Frontend 读取 `error.details.field` 并映射到
// `el-form-item` 的 `error` 属性实现定位回显。
//
// 规范文档：`docs/guides/frontend-observability-spec.md` §2.3。
//
// 用法：
//
//	return nil, &apperr.FieldError{Field: "code", Reason: "已存在", Msg: "入口 Code 已存在"}
type FieldError struct {
	Field  string
	Reason string
	Msg    string // 面向用户的整体提示；留空则使用 Reason
	// Code 可选：业务场景需要特定业务码（如 3013 Conflict）时覆写；默认 CodeParamInvalid。
	Code int
}

func (e *FieldError) Error() string {
	if e.Msg != "" {
		return e.Msg
	}
	return e.Reason
}

// ── 内部 ──────────────────────────────────────────────────────────────────

// mapped 是 Map 的中间结果。
type mapped struct {
	status int
	body   *gen.Error
}

// Map 将任意 error 翻译成 (HTTP status, gen.Error)。
// 覆盖两类：
//   - 协议错误：ogen decode/validation/security 失败
//   - 业务错误：service 层 sentinel
//
// 新增映射只在这里加 case，不需要改任何其他文件。
func Map(err error) (status int, body *gen.Error) {
	m := doMap(err)
	return m.status, m.body
}

func doMap(err error) mapped {
	// ── handler 层哨兵 ────────────────────────────────────────────────────

	var pe *ParamError
	if errors.As(err, &pe) {
		return mapped{http.StatusBadRequest, &gen.Error{
			Code:    CodeParamInvalid,
			Message: pe.Msg,
		}}
	}

	var ue *UnauthError
	if errors.As(err, &ue) {
		return mapped{http.StatusUnauthorized, &gen.Error{
			Code:    CodeUnauthorized,
			Message: ue.Msg,
		}}
	}

	// 字段级校验失败（哨兵位置必须在 ParamError 之后，CodeConflict 冲突之前）
	var fe *FieldError
	if errors.As(err, &fe) {
		code := fe.Code
		if code == 0 {
			code = CodeParamInvalid
		}
		msg := fe.Msg
		if msg == "" {
			msg = fe.Reason
		}
		return mapped{http.StatusBadRequest, &gen.Error{
			Code:    code,
			Message: msg,
			Details: gen.NewOptNilErrorDetails(gen.ErrorDetails{
				fe.Field: jx.Raw(strconv.Quote(fe.Reason)),
			}),
		}}
	}

	// ── 协议错误（ogen 框架产生）──────────────────────────────────────────

	// 权限拒绝（来自 OpenAPIPermission 中间件）
	if errors.Is(err, ErrPermissionDenied) {
		return mapped{http.StatusForbidden, &gen.Error{
			Code:    CodeForbidden,
			Message: "无权访问",
		}}
	}

	// 请求体 / query / path 参数解码失败
	var decodeReq *ogenerrors.DecodeRequestError
	if errors.As(err, &decodeReq) {
		return mapped{http.StatusBadRequest, &gen.Error{
			Code:    CodeParamFormat,
			Message: "请求体格式错误",
		}}
	}

	var decodeParams *ogenerrors.DecodeParamsError
	if errors.As(err, &decodeParams) {
		return mapped{http.StatusBadRequest, &gen.Error{
			Code:    CodeParamFormat,
			Message: "请求参数格式错误",
		}}
	}

	// security scheme 未满足（token 缺失 / 格式错误）
	var secErr *ogenerrors.SecurityError
	if errors.As(err, &secErr) {
		return mapped{http.StatusUnauthorized, &gen.Error{
			Code:    CodeUnauthorized,
			Message: "未认证，请先登录",
		}}
	}

	// context 取消 / 超时
	if errors.Is(err, context.Canceled) {
		return mapped{499, &gen.Error{
			Code:    CodeInternal,
			Message: "请求已取消",
		}}
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return mapped{http.StatusGatewayTimeout, &gen.Error{
			Code:    CodeInternal,
			Message: "请求超时，请稍后重试",
		}}
	}

	// ── 业务错误（service 层 sentinel）───────────────────────────────────

	// auth 域
	if errors.Is(err, auth.ErrInvalidCredentials) {
		return mapped{http.StatusUnauthorized, &gen.Error{
			Code:    CodeInvalidCredentials,
			Message: "邮箱或密码错误",
		}}
	}
	if errors.Is(err, auth.ErrUserInactive) {
		return mapped{http.StatusUnauthorized, &gen.Error{
			Code:    CodeUserInactive,
			Message: "账号已被禁用",
		}}
	}
	if errors.Is(err, auth.ErrUserExists) {
		return mapped{http.StatusConflict, &gen.Error{
			Code:    CodeUserExists,
			Message: "用户名已存在",
		}}
	}
	if errors.Is(err, auth.ErrEmailExists) {
		return mapped{http.StatusConflict, &gen.Error{
			Code:    CodeEmailExists,
			Message: "邮箱已被注册",
		}}
	}
	if errors.Is(err, auth.ErrUserNotFound) {
		return mapped{http.StatusNotFound, &gen.Error{
			Code:    CodeUserNotFound,
			Message: "用户不存在",
		}}
	}

	// 通用 GORM not found
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return mapped{http.StatusNotFound, &gen.Error{
			Code:    CodeNotFound,
			Message: "资源不存在",
		}}
	}

	// ── 兜底：500，不暴露内部细节 ─────────────────────────────────────────
	return mapped{http.StatusInternalServerError, &gen.Error{
		Code:    CodeInternal,
		Message: "服务器内部错误，请稍后重试",
	}}
}

// ErrorHandler 实现 ogenerrors.ErrorHandler，挂到 ogen NewServer 的 WithErrorHandler。
// 它是 ogen 框架与业务错误的唯一出口，handler 只需 return nil, err。
func ErrorHandler(logger *zap.Logger) func(context.Context, http.ResponseWriter, *http.Request, error) {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request, err error) {
		status, body := Map(err)

		// 5xx 记详细日志；4xx 只记 debug（避免日志噪音）
		if status >= 500 {
			logger.Error("internal error", zap.Error(err),
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
			)
		} else {
			logger.Debug("client error",
				zap.Int("status", status),
				zap.Error(err),
				zap.String("path", r.URL.Path),
			)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(body)
	}
}


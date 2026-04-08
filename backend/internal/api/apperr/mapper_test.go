package apperr_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/ogen-go/ogen/ogenerrors"
	"gorm.io/gorm"

	"github.com/gg-ecommerce/backend/internal/api/apperr"
	"github.com/gg-ecommerce/backend/internal/modules/system/auth"
)

type mapCase struct {
	name       string
	err        error
	wantStatus int
	wantCode   int
}

func TestMap(t *testing.T) {
	cases := []mapCase{
		// handler 哨兵
		{"ParamError", &apperr.ParamError{Msg: "x"}, http.StatusBadRequest, apperr.CodeParamInvalid},
		{"UnauthError", &apperr.UnauthError{Msg: "x"}, http.StatusUnauthorized, apperr.CodeUnauthorized},

		// 权限中间件
		{"ErrPermissionDenied", apperr.ErrPermissionDenied, http.StatusForbidden, apperr.CodeForbidden},

		// ogen 协议错误
		{"DecodeRequestError", &ogenerrors.DecodeRequestError{}, http.StatusBadRequest, apperr.CodeParamFormat},
		{"DecodeParamsError", &ogenerrors.DecodeParamsError{}, http.StatusBadRequest, apperr.CodeParamFormat},
		{"SecurityError", &ogenerrors.SecurityError{}, http.StatusUnauthorized, apperr.CodeUnauthorized},

		// context 错误
		{"Canceled", context.Canceled, 499, apperr.CodeInternal},
		{"DeadlineExceeded", context.DeadlineExceeded, http.StatusGatewayTimeout, apperr.CodeInternal},

		// auth sentinel
		{"ErrInvalidCredentials", auth.ErrInvalidCredentials, http.StatusUnauthorized, apperr.CodeInvalidCredentials},
		{"ErrUserInactive", auth.ErrUserInactive, http.StatusUnauthorized, apperr.CodeUserInactive},
		{"ErrUserExists", auth.ErrUserExists, http.StatusConflict, apperr.CodeUserExists},
		{"ErrEmailExists", auth.ErrEmailExists, http.StatusConflict, apperr.CodeEmailExists},
		{"ErrUserNotFound", auth.ErrUserNotFound, http.StatusNotFound, apperr.CodeUserNotFound},

		// gorm
		{"gorm.ErrRecordNotFound", gorm.ErrRecordNotFound, http.StatusNotFound, apperr.CodeNotFound},

		// 兜底
		{"unknown", errors.New("boom"), http.StatusInternalServerError, apperr.CodeInternal},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			status, body := apperr.Map(c.err)
			if status != c.wantStatus {
				t.Errorf("status: got %d, want %d", status, c.wantStatus)
			}
			if body.Code != c.wantCode {
				t.Errorf("code: got %d, want %d", body.Code, c.wantCode)
			}
		})
	}
}

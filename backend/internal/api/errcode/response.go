// Deprecated: 见 codes.go 的说明。
package errcode

import (
	"net/http"

	"github.com/gg-ecommerce/backend/internal/api/dto"
)

// Response 根据错误码返回 (HTTP 状态码, 统一响应体)。用于 Handler：c.JSON(errcode.Response(errcode.ErrParamInvalid))
func Response(code int) (int, *dto.Response) {
	return HTTPStatus(code), &dto.Response{
		Code:    code,
		Message: Message(code),
		Data:    nil,
	}
}

// ResponseWithMsg 根据错误码返回响应，但使用自定义 message（覆盖默认说明）
func ResponseWithMsg(code int, message string) (int, *dto.Response) {
	if message == "" {
		message = Message(code)
	}
	return HTTPStatus(code), &dto.Response{
		Code:    code,
		Message: message,
		Data:    nil,
	}
}

// ResponseData 仅构造响应体，HTTP 状态码由调用方指定（如 200 成功但业务 code 非 0 的场景较少用）
func ResponseData(code int, message string, data interface{}) *dto.Response {
	if message == "" {
		message = Message(code)
	}
	return &dto.Response{Code: code, Message: message, Data: data}
}

// OK 成功响应，HTTP 200 + body.code=0
func OK(data interface{}) (int, *dto.Response) {
	return http.StatusOK, dto.SuccessResponse(data)
}

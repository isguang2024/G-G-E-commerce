package social

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type HTTPHandler struct {
	service Service
	logger  *zap.Logger
}

func NewHTTPHandler(service Service, logger *zap.Logger) *HTTPHandler {
	return &HTTPHandler{service: service, logger: logger}
}

func (h *HTTPHandler) Authorize(c *gin.Context) {
	providerKey := strings.TrimSpace(c.Param("provider"))
	if providerKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "provider 不能为空"})
		return
	}

	authorizeURL, err := h.service.BuildAuthorizeURL(c.Request.Context(), providerKey, AuthorizeInput{
		RequestHost:  resolveRequestHost(c),
		RequestPath:  c.Query("redirect_path"),
		LoginPageKey: c.Query("login_page_key"),
		PageScene:    c.DefaultQuery("page_scene", "login"),
		TargetAppKey: c.Query("target_app_key"),
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	c.Redirect(http.StatusFound, authorizeURL)
}

func (h *HTTPHandler) Callback(c *gin.Context) {
	providerKey := strings.TrimSpace(c.Param("provider"))
	if providerKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "provider 不能为空"})
		return
	}
	result, err := h.service.HandleCallback(c.Request.Context(), providerKey, CallbackInput{
		Code:  c.Query("code"),
		State: c.Query("state"),
	})
	if err != nil {
		h.logger.Warn("social oauth callback failed", zap.Error(err), zap.String("provider", providerKey))
		c.Redirect(http.StatusFound, "/account/auth/login?social_error=oauth_callback_failed")
		return
	}
	c.Redirect(http.StatusFound, result.RedirectPath)
}

func (h *HTTPHandler) Exchange(c *gin.Context) {
	var req struct {
		SocialToken string `json:"social_token"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "请求体格式错误"})
		return
	}
	result, err := h.service.ExchangeSocialToken(c.Request.Context(), strings.TrimSpace(req.SocialToken))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	payload := gin.H{
		"intent":          result.Intent,
		"provider_key":    result.ProviderKey,
		"provider_name":   result.ProviderName,
		"provider_uid":    result.ProviderUID,
		"provider_user":   result.ProviderUser,
		"email":           result.Email,
		"avatar_url":      result.AvatarURL,
		"matched_user_id": result.MatchedUserID,
		"need_register":   result.NeedRegister,
	}
	if result.LoginResponse != nil {
		payload["access_token"] = result.LoginResponse.AccessToken
		payload["refresh_token"] = result.LoginResponse.RefreshToken
		payload["expires_in"] = result.LoginResponse.ExpiresIn
		payload["user"] = result.LoginResponse.User
	}
	c.JSON(http.StatusOK, payload)
}

func resolveRequestHost(c *gin.Context) string {
	host := strings.TrimSpace(c.GetString("request_host"))
	if host != "" {
		return host
	}
	if c.Request != nil {
		return strings.TrimSpace(c.Request.Host)
	}
	return ""
}

// Package middleware 提供 HTTP 中间件。
package middleware

import (
	"net/http"
	"strings"

	"hrattendance/pkg/httpx"
	"hrattendance/pkg/logger"
)

// RequireAuth 返回一个鉴权中间件。当 adminToken 为空时直接放行；
// 否则要求请求携带 `Authorization: Bearer <token>`，校验失败返回 401。
func RequireAuth(adminToken string, log *logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if adminToken == "" {
				next.ServeHTTP(w, r)
				return
			}
			auth := r.Header.Get("Authorization")
			const prefix = "Bearer "
			if !strings.HasPrefix(auth, prefix) {
				httpx.Unauthorized(w, "未授权：缺少访问令牌")
				return
			}
			if strings.TrimPrefix(auth, prefix) != adminToken {
				log.Warnf("鉴权失败: %s %s", r.Method, r.URL.Path)
				httpx.Unauthorized(w, "未授权：访问令牌无效")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RequireAuthMethod 仅对写操作（POST/PUT/PATCH/DELETE）要求鉴权。
func RequireAuthMethod(adminToken string, log *logger.Logger) func(http.Handler) http.Handler {
	auth := RequireAuth(adminToken, log)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
				auth(next).ServeHTTP(w, r)
			default:
				next.ServeHTTP(w, r)
			}
		})
	}
}

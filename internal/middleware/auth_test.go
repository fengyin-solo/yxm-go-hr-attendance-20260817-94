package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"hrattendance/pkg/logger"
)

func testAuthHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func TestRequireAuthNoTokenConfigured(t *testing.T) {
	log := logger.NewLevel(logger.LevelError)
	mw := RequireAuth("", log)

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	mw(testAuthHandler()).ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("未配置 token 应放行，实际 %d", rec.Code)
	}
}

func TestRequireAuthMissingToken(t *testing.T) {
	log := logger.NewLevel(logger.LevelError)
	mw := RequireAuth("secret", log)

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	mw(testAuthHandler()).ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("缺少 token 应返回 401，实际 %d", rec.Code)
	}
}

func TestRequireAuthInvalidToken(t *testing.T) {
	log := logger.NewLevel(logger.LevelError)
	mw := RequireAuth("secret", log)

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer wrong")
	rec := httptest.NewRecorder()
	mw(testAuthHandler()).ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("无效 token 应返回 401，实际 %d", rec.Code)
	}
}

func TestRequireAuthValidToken(t *testing.T) {
	log := logger.NewLevel(logger.LevelError)
	mw := RequireAuth("secret", log)

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer secret")
	rec := httptest.NewRecorder()
	mw(testAuthHandler()).ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("有效 token 应放行，实际 %d", rec.Code)
	}
}

func TestRequireAuthMethodSkipsRead(t *testing.T) {
	log := logger.NewLevel(logger.LevelError)
	mw := RequireAuthMethod("secret", log)

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	mw(testAuthHandler()).ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET 应放行，实际 %d", rec.Code)
	}
}

func TestRequireAuthMethodChecksWrite(t *testing.T) {
	log := logger.NewLevel(logger.LevelError)
	mw := RequireAuthMethod("secret", log)

	req := httptest.NewRequest("POST", "/", nil)
	rec := httptest.NewRecorder()
	mw(testAuthHandler()).ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("POST 无 token 应返回 401，实际 %d", rec.Code)
	}

	req = httptest.NewRequest("DELETE", "/", nil)
	req.Header.Set("Authorization", "Bearer secret")
	rec = httptest.NewRecorder()
	mw(testAuthHandler()).ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("DELETE 有效 token 应放行，实际 %d", rec.Code)
	}
}

func TestClientIP(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.168.1.1:1234"
	if got := clientIP(req); got != "192.168.1.1" {
		t.Fatalf("应从 RemoteAddr 提取 IP，实际 %s", got)
	}

	req = httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Forwarded-For", "10.0.0.1, 10.0.0.2")
	if got := clientIP(req); got != "10.0.0.1" {
		t.Fatalf("应从 X-Forwarded-For 提取首个 IP，实际 %s", got)
	}
}

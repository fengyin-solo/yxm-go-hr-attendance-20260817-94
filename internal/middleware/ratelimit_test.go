package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRateLimiterAllowsWithinLimit(t *testing.T) {
	rl := NewRateLimiter(2, time.Minute)
	handler := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = "1.2.3.4:1234"
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("第 %d 次请求应放行，实际 %d", i+1, rec.Code)
		}
	}
}

func TestRateLimiterBlocksOverLimit(t *testing.T) {
	rl := NewRateLimiter(2, time.Minute)
	handler := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = "5.6.7.8:1234"
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "5.6.7.8:1234"
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("超过限制应返回 429，实际 %d", rec.Code)
	}
}

func TestRateLimiterSeparatesByIP(t *testing.T) {
	rl := NewRateLimiter(1, time.Minute)
	handler := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "1.1.1.1:1000"
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("第一个 IP 首次请求应放行")
	}

	req = httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "2.2.2.2:2000"
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("不同 IP 应独立计数，实际 %d", rec.Code)
	}
}

func TestRateLimiterWindowReset(t *testing.T) {
	rl := NewRateLimiter(1, 10*time.Millisecond)
	handler := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "9.9.9.9:9999"
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	time.Sleep(15 * time.Millisecond)

	req = httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "9.9.9.9:9999"
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("窗口重置后应放行，实际 %d", rec.Code)
	}
}

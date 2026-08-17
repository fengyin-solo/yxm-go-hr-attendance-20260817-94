package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"hrattendance/internal/config"
	"hrattendance/internal/service"
	"hrattendance/internal/store"
	"hrattendance/pkg/logger"
)

func TestRateLimitIntegration(t *testing.T) {
	cfg := &config.Config{
		MaxPageSize:      100,
		RateLimit:        2,
		RateWindowSec:    60,
		DefaultWorkStart: "09:00",
		DefaultWorkEnd:   "18:00",
	}
	log := logger.NewLevel(logger.LevelError)
	svc := service.New(store.NewMemoryStore(), log, cfg)
	s := NewServer(svc, log, cfg)

	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("GET", "/api/departments", nil)
		req.RemoteAddr = "10.0.0.1:1000"
		rec := httptest.NewRecorder()
		s.Routes().ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("第 %d 次请求应放行，实际 %d", i+1, rec.Code)
		}
	}

	req := httptest.NewRequest("GET", "/api/departments", nil)
	req.RemoteAddr = "10.0.0.1:1000"
	rec := httptest.NewRecorder()
	s.Routes().ServeHTTP(rec, req)
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("超过限制应返回 429，实际 %d", rec.Code)
	}
}

func TestRecoveryMiddlewareIntegration(t *testing.T) {
	cfg := &config.Config{
		MaxPageSize:      100,
		RateLimit:        1000,
		RateWindowSec:    60,
		DefaultWorkStart: "09:00",
		DefaultWorkEnd:   "18:00",
	}
	log := logger.NewLevel(logger.LevelError)
	svc := service.New(store.NewMemoryStore(), log, cfg)
	s := NewServer(svc, log, cfg)

	// 正常请求不应触发 recovery
	req := httptest.NewRequest("GET", "/api/departments", nil)
	rec := httptest.NewRecorder()
	s.Routes().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("正常请求应返回 200，实际 %d", rec.Code)
	}
}

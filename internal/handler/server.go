// Package handler 实现 HTTP 处理器层。
package handler

import (
	"errors"
	"net/http"
	"runtime/debug"
	"time"

	"hrattendance/internal/config"
	"hrattendance/internal/middleware"
	"hrattendance/internal/model"
	"hrattendance/internal/service"
	"hrattendance/internal/store"
	"hrattendance/pkg/httpx"
	"hrattendance/pkg/logger"
)

type Server struct {
	svc     *service.Service
	log     *logger.Logger
	cfg     *config.Config
	limiter *middleware.RateLimiter
}

func NewServer(svc *service.Service, log *logger.Logger, cfg *config.Config) *Server {
	window := time.Duration(cfg.RateWindowSec) * time.Second
	return &Server{
		svc:     svc,
		log:     log,
		cfg:     cfg,
		limiter: middleware.NewRateLimiter(cfg.RateLimit, window),
	}
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	s.registerDepartmentRoutes(mux)
	s.registerEmployeeRoutes(mux)
	s.registerAttendanceRoutes(mux)
	s.registerLeaveRoutes(mux)
	s.registerOvertimeRoutes(mux)
	s.registerRuleRoutes(mux)
	s.registerReportRoutes(mux)

	var handler http.Handler = mux
	handler = middleware.RequireAuthMethod(s.cfg.AdminToken, s.log)(handler)
	handler = s.limiter.Middleware(handler)
	handler = s.loggingMiddleware(handler)
	handler = s.recoveryMiddleware(handler)
	return handler
}

func (s *Server) maxPageSize() int {
	if s.cfg != nil && s.cfg.MaxPageSize > 0 {
		return s.cfg.MaxPageSize
	}
	return 100
}

func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		s.log.Infof("%s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}

func (s *Server) recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				s.log.Errorf("panic: %v\n%s", rec, debug.Stack())
				httpx.InternalError(w, "服务器内部错误")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case model.IsValidationError(err):
		httpx.BadRequest(w, err.Error())
	case errors.Is(err, store.ErrNotFound):
		httpx.NotFound(w, err.Error())
	case errors.Is(err, store.ErrConflict):
		httpx.Conflict(w, err.Error())
	default:
		httpx.InternalError(w, err.Error())
	}
}

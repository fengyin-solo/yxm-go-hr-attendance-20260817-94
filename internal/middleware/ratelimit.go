package middleware

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"hrattendance/pkg/httpx"
)

// window 表示某个客户端在固定窗口内的请求计数。
type window struct {
	count int
	reset time.Time
}

// RateLimiter 基于固定窗口算法的限流器。
type RateLimiter struct {
	mu     sync.Mutex
	windows map[string]*window
	limit  int
	window time.Duration
}

// NewRateLimiter 创建限流器，limit 为窗口内最大请求数，d 为窗口长度。
func NewRateLimiter(limit int, d time.Duration) *RateLimiter {
	if limit <= 0 {
		limit = 100
	}
	if d <= 0 {
		d = time.Minute
	}
	return &RateLimiter{
		windows: make(map[string]*window),
		limit:   limit,
		window:  d,
	}
}

// allow 判断指定 key 是否允许本次请求，并更新计数。
func (rl *RateLimiter) allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	now := time.Now()
	w, ok := rl.windows[key]
	if !ok || now.After(w.reset) {
		rl.windows[key] = &window{count: 1, reset: now.Add(rl.window)}
		return true
	}
	if w.count >= rl.limit {
		return false
	}
	w.count++
	return true
}

// Middleware 返回限流中间件，按客户端 IP 限流。
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := clientIP(r)
		if !rl.allow(ip) {
			httpx.Error(w, http.StatusTooManyRequests, 429, "请求过于频繁，请稍后再试")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// clientIP 从 RemoteAddr 或 X-Forwarded-For 中提取客户端 IP。
func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		if len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// Package config 负责从环境变量加载服务配置。
package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Addr            string
	MaxPageSize     int
	AdminToken      string
	RateLimit       int
	RateWindowSec   int
	LateThreshold   int
	DefaultWorkStart string
	DefaultWorkEnd   string
}

func Load() *Config {
	cfg := &Config{
		Addr:             ":" + getenv("PORT", "8080"),
		MaxPageSize:      getenvInt("MAX_PAGE_SIZE", 100),
		AdminToken:       os.Getenv("ADMIN_TOKEN"),
		RateLimit:        getenvInt("RATE_LIMIT", 100),
		RateWindowSec:    getenvInt("RATE_WINDOW_SEC", 60),
		LateThreshold:    getenvInt("LATE_THRESHOLD_MIN", 15),
		DefaultWorkStart: getenv("DEFAULT_WORK_START", "09:00"),
		DefaultWorkEnd:   getenv("DEFAULT_WORK_END", "18:00"),
	}
	if v := os.Getenv("ADDR"); v != "" {
		cfg.Addr = v
	}
	return cfg
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getenvInt(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil || n <= 0 {
		return def
	}
	return n
}

func (c *Config) String() string {
	return fmt.Sprintf("addr=%s max_page_size=%d rate_limit=%d", c.Addr, c.MaxPageSize, c.RateLimit)
}

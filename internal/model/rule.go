package model

import (
	"strings"
	"time"
)

// AttendanceRule 表示考勤规则，时间格式为 HH:MM。
type AttendanceRule struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	WorkStart     string    `json:"work_start"`
	WorkEnd       string    `json:"work_end"`
	LateThreshold int       `json:"late_threshold"`
	Enabled       bool      `json:"enabled"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func (r *AttendanceRule) Validate() error {
	r.Name = strings.TrimSpace(r.Name)
	if r.Name == "" {
		return NewValidationError("name", "规则名称不能为空")
	}
	if _, err := time.Parse("15:04", r.WorkStart); err != nil {
		return NewValidationError("work_start", "上班时间格式错误，应为 HH:MM")
	}
	if _, err := time.Parse("15:04", r.WorkEnd); err != nil {
		return NewValidationError("work_end", "下班时间格式错误，应为 HH:MM")
	}
	if r.WorkEnd <= r.WorkStart {
		return NewValidationError("work_end", "下班时间必须晚于上班时间")
	}
	if r.LateThreshold < 0 {
		return NewValidationError("late_threshold", "迟到阈值不能为负")
	}
	return nil
}

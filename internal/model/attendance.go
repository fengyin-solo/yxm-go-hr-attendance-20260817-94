package model

import (
	"strings"
	"time"
)

const (
	AttendanceNormal      = "normal"
	AttendanceLate        = "late"
	AttendanceEarlyLeave  = "early_leave"
	AttendanceAbsent      = "absent"
)

// Attendance 表示一条考勤记录，日期为 YYYY-MM-DD，时间为 HH:MM。
type Attendance struct {
	ID         string    `json:"id"`
	EmployeeID string    `json:"employee_id"`
	Date       string    `json:"date"`
	CheckIn    string    `json:"check_in"`
	CheckOut   string    `json:"check_out"`
	Status     string    `json:"status"`
	Note       string    `json:"note"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func (a *Attendance) Validate() error {
	a.Note = strings.TrimSpace(a.Note)
	if a.EmployeeID == "" {
		return NewValidationError("employee_id", "必须关联员工")
	}
	if a.Date == "" {
		return NewValidationError("date", "考勤日期不能为空")
	}
	if _, err := time.Parse("2006-01-02", a.Date); err != nil {
		return NewValidationError("date", "考勤日期格式错误，应为 YYYY-MM-DD")
	}
	for _, t := range []string{a.CheckIn, a.CheckOut} {
		if t == "" {
			continue
		}
		if _, err := time.Parse("15:04", t); err != nil {
			return NewValidationError("time", "时间格式错误，应为 HH:MM")
		}
	}
	if a.Status == "" {
		a.Status = AttendanceNormal
	}
	if !ValidAttendanceStatus(a.Status) {
		return NewValidationError("status", "考勤状态不合法")
	}
	return nil
}

func ValidAttendanceStatus(s string) bool {
	switch s {
	case AttendanceNormal, AttendanceLate, AttendanceEarlyLeave, AttendanceAbsent:
		return true
	}
	return false
}

type AttendanceFilter struct {
	EmployeeID string
	Status     string
	DateFrom   string
	DateTo     string
}

func (f AttendanceFilter) Match(a *Attendance) bool {
	if f.EmployeeID != "" && a.EmployeeID != f.EmployeeID {
		return false
	}
	if f.Status != "" && a.Status != f.Status {
		return false
	}
	if f.DateFrom != "" && a.Date < f.DateFrom {
		return false
	}
	if f.DateTo != "" && a.Date > f.DateTo {
		return false
	}
	return true
}

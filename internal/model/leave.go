package model

import (
	"strings"
	"time"
)

const (
	LeaveAnnual   = "annual"
	LeaveSick     = "sick"
	LeavePersonal = "personal"
)

const (
	LeavePending  = "pending"
	LeaveApproved = "approved"
	LeaveRejected = "rejected"
)

// leaveTransitions 定义请假单状态机。
var leaveTransitions = map[string]map[string]bool{
	LeavePending:  {LeaveApproved: true, LeaveRejected: true},
	LeaveApproved: {},
	LeaveRejected: {},
}

// CanTransitionLeave 判断请假单能否从 from 流转到 to。
func CanTransitionLeave(from, to string) bool {
	transitions, ok := leaveTransitions[from]
	if !ok {
		return false
	}
	return transitions[to]
}

func ValidLeaveStatus(s string) bool {
	_, ok := leaveTransitions[s]
	return ok
}

func ValidLeaveType(s string) bool {
	switch s {
	case LeaveAnnual, LeaveSick, LeavePersonal:
		return true
	}
	return false
}

// Leave 表示一张请假单，日期为 YYYY-MM-DD。
type Leave struct {
	ID         string    `json:"id"`
	EmployeeID string    `json:"employee_id"`
	Type       string    `json:"type"`
	StartDate  string    `json:"start_date"`
	EndDate    string    `json:"end_date"`
	Reason     string    `json:"reason"`
	Status     string    `json:"status"`
	ApproverID string    `json:"approver_id"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// Days 返回请假天数（含首尾，至少 1 天）。
func (l *Leave) Days() int {
	start, err1 := time.Parse("2006-01-02", l.StartDate)
	end, err2 := time.Parse("2006-01-02", l.EndDate)
	if err1 != nil || err2 != nil {
		return 0
	}
	days := int(end.Sub(start).Hours()/24) + 1
	if days < 1 {
		days = 1
	}
	return days
}

func (l *Leave) Validate() error {
	l.Reason = strings.TrimSpace(l.Reason)
	if l.EmployeeID == "" {
		return NewValidationError("employee_id", "必须关联员工")
	}
	if !ValidLeaveType(l.Type) {
		return NewValidationError("type", "请假类型不合法")
	}
	if l.StartDate == "" || l.EndDate == "" {
		return NewValidationError("date", "请假日期不能为空")
	}
	if _, err := time.Parse("2006-01-02", l.StartDate); err != nil {
		return NewValidationError("start_date", "开始日期格式错误，应为 YYYY-MM-DD")
	}
	if _, err := time.Parse("2006-01-02", l.EndDate); err != nil {
		return NewValidationError("end_date", "结束日期格式错误，应为 YYYY-MM-DD")
	}
	if l.EndDate < l.StartDate {
		return NewValidationError("end_date", "结束日期不能早于开始日期")
	}
	if l.Reason == "" {
		return NewValidationError("reason", "请假事由不能为空")
	}
	if l.Status == "" {
		l.Status = LeavePending
	}
	if !ValidLeaveStatus(l.Status) {
		return NewValidationError("status", "请假状态不合法")
	}
	return nil
}

type LeaveFilter struct {
	EmployeeID string
	Type       string
	Status     string
	DateFrom   string
	DateTo     string
}

func (f LeaveFilter) Match(l *Leave) bool {
	if f.EmployeeID != "" && l.EmployeeID != f.EmployeeID {
		return false
	}
	if f.Type != "" && l.Type != f.Type {
		return false
	}
	if f.Status != "" && l.Status != f.Status {
		return false
	}
	if f.DateFrom != "" && l.StartDate < f.DateFrom {
		return false
	}
	if f.DateTo != "" && l.EndDate > f.DateTo {
		return false
	}
	return true
}

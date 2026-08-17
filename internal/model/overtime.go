package model

import (
	"strings"
	"time"
)

const (
	OvertimePending  = "pending"
	OvertimeApproved = "approved"
	OvertimeRejected = "rejected"
)

// overtimeTransitions 定义加班单状态机。
var overtimeTransitions = map[string]map[string]bool{
	OvertimePending:  {OvertimeApproved: true, OvertimeRejected: true},
	OvertimeApproved: {},
	OvertimeRejected: {},
}

// CanTransitionOvertime 判断加班单能否从 from 流转到 to。
func CanTransitionOvertime(from, to string) bool {
	if m, ok := overtimeTransitions[from]; ok {
		return m[to]
	}
	return false
}

func ValidOvertimeStatus(s string) bool {
	_, ok := overtimeTransitions[s]
	return ok
}

// Overtime 表示一张加班单，日期为 YYYY-MM-DD。
type Overtime struct {
	ID         string    `json:"id"`
	EmployeeID string    `json:"employee_id"`
	Date       string    `json:"date"`
	Hours      float64   `json:"hours"`
	Reason     string    `json:"reason"`
	Status     string    `json:"status"`
	ApproverID string    `json:"approver_id"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func (o *Overtime) Validate() error {
	o.Reason = strings.TrimSpace(o.Reason)
	if o.EmployeeID == "" {
		return NewValidationError("employee_id", "必须关联员工")
	}
	if o.Date == "" {
		return NewValidationError("date", "加班日期不能为空")
	}
	if _, err := time.Parse("2006-01-02", o.Date); err != nil {
		return NewValidationError("date", "加班日期格式错误，应为 YYYY-MM-DD")
	}
	if o.Hours <= 0 {
		return NewValidationError("hours", "加班时长必须大于 0")
	}
	if o.Reason == "" {
		return NewValidationError("reason", "加班事由不能为空")
	}
	if o.Status == "" {
		o.Status = OvertimePending
	}
	if !ValidOvertimeStatus(o.Status) {
		return NewValidationError("status", "加班状态不合法")
	}
	return nil
}

type OvertimeFilter struct {
	EmployeeID string
	Status     string
	DateFrom   string
	DateTo     string
}

func (f OvertimeFilter) Match(o *Overtime) bool {
	if f.EmployeeID != "" && o.EmployeeID != f.EmployeeID {
		return false
	}
	if f.Status != "" && o.Status != f.Status {
		return false
	}
	if f.DateFrom != "" && o.Date < f.DateFrom {
		return false
	}
	if f.DateTo != "" && o.Date > f.DateTo {
		return false
	}
	return true
}

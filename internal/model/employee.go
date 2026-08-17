package model

import (
	"strings"
	"time"
)

const (
	EmployeeActive   = "active"
	EmployeeResigned = "resigned"
)

// Employee 表示一名员工。
type Employee struct {
	ID           string    `json:"id"`
	DepartmentID string    `json:"department_id"`
	Name         string    `json:"name"`
	EmpNo        string    `json:"emp_no"`
	Position     string    `json:"position"`
	HireDate     time.Time `json:"hire_date"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (e *Employee) Validate() error {
	e.Name = strings.TrimSpace(e.Name)
	e.EmpNo = strings.TrimSpace(e.EmpNo)
	e.Position = strings.TrimSpace(e.Position)
	if e.Name == "" {
		return NewValidationError("name", "员工姓名不能为空")
	}
	if e.EmpNo == "" {
		return NewValidationError("emp_no", "工号不能为空")
	}
	if e.DepartmentID == "" {
		return NewValidationError("department_id", "必须关联部门")
	}
	if e.HireDate.IsZero() {
		return NewValidationError("hire_date", "入职日期不能为空")
	}
	if e.Status == "" {
		e.Status = EmployeeActive
	}
	if e.Status != EmployeeActive && e.Status != EmployeeResigned {
		return NewValidationError("status", "员工状态不合法")
	}
	return nil
}

type EmployeeFilter struct {
	DepartmentID string
	Status       string
	Keyword      string
}

func (f EmployeeFilter) Match(e *Employee) bool {
	if f.DepartmentID != "" && e.DepartmentID != f.DepartmentID {
		return false
	}
	if f.Status != "" && e.Status != f.Status {
		return false
	}
	if f.Keyword != "" {
		k := strings.ToLower(strings.TrimSpace(f.Keyword))
		if k != "" && !strings.Contains(strings.ToLower(e.Name), k) &&
			!strings.Contains(strings.ToLower(e.EmpNo), k) {
			return false
		}
	}
	return true
}

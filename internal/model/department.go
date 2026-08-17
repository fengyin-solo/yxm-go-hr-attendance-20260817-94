package model

import (
	"strings"
	"time"
)

const (
	DepartmentActive   = "active"
	DepartmentInactive = "inactive"
)

// Department 表示一个部门，ParentID 为空表示顶级部门。
type Department struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	ParentID  string    `json:"parent_id"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (d *Department) Validate() error {
	d.Name = strings.TrimSpace(d.Name)
	if d.Name == "" {
		return NewValidationError("name", "部门名称不能为空")
	}
	if d.Status == "" {
		d.Status = DepartmentActive
	}
	if d.Status != DepartmentActive && d.Status != DepartmentInactive {
		return NewValidationError("status", "部门状态不合法")
	}
	return nil
}

type DepartmentFilter struct {
	ParentID string
	Status   string
	Keyword  string
}

func (f DepartmentFilter) Match(d *Department) bool {
	if f.ParentID != "" && d.ParentID != f.ParentID {
		return false
	}
	if f.Status != "" && d.Status != f.Status {
		return false
	}
	if f.Keyword != "" {
		k := strings.ToLower(strings.TrimSpace(f.Keyword))
		if k != "" && !strings.Contains(strings.ToLower(d.Name), k) {
			return false
		}
	}
	return true
}

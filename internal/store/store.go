// Package store 定义数据访问接口与内存实现。
package store

import (
	"errors"

	"hrattendance/internal/model"
)

var (
	ErrNotFound = errors.New("记录不存在")
	ErrConflict = errors.New("记录已存在或状态冲突")
)

// Store 聚合全部实体的数据访问方法，便于测试时替换实现。
type Store interface {
	CreateDepartment(d *model.Department) error
	GetDepartment(id string) (*model.Department, error)
	ListDepartments() []*model.Department
	UpdateDepartment(d *model.Department) error
	DeleteDepartment(id string) error

	CreateEmployee(e *model.Employee) error
	GetEmployee(id string) (*model.Employee, error)
	ListEmployees() []*model.Employee
	UpdateEmployee(e *model.Employee) error
	DeleteEmployee(id string) error

	CreateAttendance(a *model.Attendance) error
	GetAttendance(id string) (*model.Attendance, error)
	GetAttendanceByEmployeeDate(employeeID, date string) (*model.Attendance, error)
	ListAttendances() []*model.Attendance
	UpdateAttendance(a *model.Attendance) error
	DeleteAttendance(id string) error

	CreateLeave(l *model.Leave) error
	GetLeave(id string) (*model.Leave, error)
	ListLeaves() []*model.Leave
	UpdateLeave(l *model.Leave) error
	DeleteLeave(id string) error

	CreateOvertime(o *model.Overtime) error
	GetOvertime(id string) (*model.Overtime, error)
	ListOvertimes() []*model.Overtime
	UpdateOvertime(o *model.Overtime) error
	DeleteOvertime(id string) error

	CreateRule(r *model.AttendanceRule) error
	GetRule(id string) (*model.AttendanceRule, error)
	ListRules() []*model.AttendanceRule
	UpdateRule(r *model.AttendanceRule) error
	DeleteRule(id string) error
}

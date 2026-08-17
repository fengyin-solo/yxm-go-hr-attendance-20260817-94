package service

import (
	"testing"
	"time"

	"hrattendance/internal/config"
	"hrattendance/internal/model"
	"hrattendance/internal/store"
	"hrattendance/pkg/logger"
)

func newTestService() *Service {
	cfg := &config.Config{
		MaxPageSize:      100,
		LateThreshold:    15,
		DefaultWorkStart: "09:00",
		DefaultWorkEnd:   "18:00",
	}
	log := logger.NewLevel(logger.LevelError)
	return New(store.NewMemoryStore(), log, cfg)
}

func mustCreateDepartment(t *testing.T, s *Service, name string) *model.Department {
	t.Helper()
	d, err := s.CreateDepartment(model.Department{Name: name})
	if err != nil {
		t.Fatalf("创建部门失败: %v", err)
	}
	return d
}

func mustCreateEmployee(t *testing.T, s *Service, deptID, name, empNo string) *model.Employee {
	t.Helper()
	e, err := s.CreateEmployee(model.Employee{
		DepartmentID: deptID,
		Name:         name,
		EmpNo:        empNo,
		Position:     "工程师",
		HireDate:     time.Now(),
	})
	if err != nil {
		t.Fatalf("创建员工失败: %v", err)
	}
	return e
}

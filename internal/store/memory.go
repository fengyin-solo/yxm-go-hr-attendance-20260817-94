package store

import (
	"sync"

	"hrattendance/internal/model"
)

// MemoryStore 基于内存 map 的 Store 实现。
type MemoryStore struct {
	mu          sync.RWMutex
	departments map[string]*model.Department
	employees   map[string]*model.Employee
	attendances map[string]*model.Attendance
	leaves      map[string]*model.Leave
	overtimes   map[string]*model.Overtime
	rules       map[string]*model.AttendanceRule
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		departments: make(map[string]*model.Department),
		employees:   make(map[string]*model.Employee),
		attendances: make(map[string]*model.Attendance),
		leaves:      make(map[string]*model.Leave),
		overtimes:   make(map[string]*model.Overtime),
		rules:       make(map[string]*model.AttendanceRule),
	}
}

var _ Store = (*MemoryStore)(nil)

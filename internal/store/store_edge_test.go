package store

import (
	"testing"
)

func TestDeleteNonExistent(t *testing.T) {
	s := NewMemoryStore()
	if err := s.DeleteDepartment("x"); err != ErrNotFound {
		t.Fatalf("删除不存在部门应返回 ErrNotFound")
	}
	if err := s.DeleteEmployee("x"); err != ErrNotFound {
		t.Fatalf("删除不存在员工应返回 ErrNotFound")
	}
	if err := s.DeleteAttendance("x"); err != ErrNotFound {
		t.Fatalf("删除不存在考勤应返回 ErrNotFound")
	}
	if err := s.DeleteLeave("x"); err != ErrNotFound {
		t.Fatalf("删除不存在请假应返回 ErrNotFound")
	}
	if err := s.DeleteOvertime("x"); err != ErrNotFound {
		t.Fatalf("删除不存在加班应返回 ErrNotFound")
	}
	if err := s.DeleteRule("x"); err != ErrNotFound {
		t.Fatalf("删除不存在规则应返回 ErrNotFound")
	}
}

func TestEmptyListReturnsEmptySlice(t *testing.T) {
	s := NewMemoryStore()
	if list := s.ListDepartments(); len(list) != 0 {
		t.Fatalf("空列表应返回空切片，实际 %d", len(list))
	}
	if list := s.ListEmployees(); len(list) != 0 {
		t.Fatalf("空列表应返回空切片，实际 %d", len(list))
	}
	if list := s.ListAttendances(); len(list) != 0 {
		t.Fatalf("空列表应返回空切片，实际 %d", len(list))
	}
	if list := s.ListLeaves(); len(list) != 0 {
		t.Fatalf("空列表应返回空切片，实际 %d", len(list))
	}
	if list := s.ListOvertimes(); len(list) != 0 {
		t.Fatalf("空列表应返回空切片，实际 %d", len(list))
	}
	if list := s.ListRules(); len(list) != 0 {
		t.Fatalf("空列表应返回空切片，实际 %d", len(list))
	}
}

func TestEmployeeUpdateUniqueEmpNo(t *testing.T) {
	s := NewMemoryStore()
	e1 := testEmployee()
	e2 := testEmployee()
	e2.ID = "e2"
	e2.EmpNo = "E002"
	_ = s.CreateEmployee(e1)
	_ = s.CreateEmployee(e2)

	// 尝试把 e2 的工号改成 e1 的工号，应冲突
	e2.EmpNo = "E001"
	if err := s.UpdateEmployee(e2); err != ErrConflict {
		t.Fatalf("更新为重复工号应返回 ErrConflict，实际: %v", err)
	}
}

func TestAttendanceUpdateUniqueDate(t *testing.T) {
	s := NewMemoryStore()
	a1 := testAttendance()
	a2 := testAttendance()
	a2.ID = "a2"
	a2.Date = "2026-08-17"
	_ = s.CreateAttendance(a1)
	_ = s.CreateAttendance(a2)

	// 尝试把 a2 的日期改成 a1 的日期，应冲突
	a2.Date = "2026-08-16"
	if err := s.UpdateAttendance(a2); err != ErrConflict {
		t.Fatalf("更新为同日考勤应返回 ErrConflict，实际: %v", err)
	}
}

func TestGetByEmployeeDateMultiple(t *testing.T) {
	s := NewMemoryStore()
	a1 := testAttendance()
	a2 := testAttendance()
	a2.ID = "a2"
	a2.Date = "2026-08-17"
	_ = s.CreateAttendance(a1)
	_ = s.CreateAttendance(a2)

	got, err := s.GetAttendanceByEmployeeDate("e1", "2026-08-17")
	if err != nil {
		t.Fatalf("查询失败: %v", err)
	}
	if got.ID != "a2" {
		t.Fatalf("应查到 a2，实际 %s", got.ID)
	}
}

func TestListAllEntitiesCount(t *testing.T) {
	s := NewMemoryStore()
	_ = s.CreateDepartment(testDepartment())
	_ = s.CreateEmployee(testEmployee())
	_ = s.CreateAttendance(testAttendance())
	_ = s.CreateLeave(testLeave())
	_ = s.CreateOvertime(testOvertime())
	_ = s.CreateRule(testRule())

	if len(s.ListDepartments()) != 1 ||
		len(s.ListEmployees()) != 1 ||
		len(s.ListAttendances()) != 1 ||
		len(s.ListLeaves()) != 1 ||
		len(s.ListOvertimes()) != 1 ||
		len(s.ListRules()) != 1 {
		t.Fatalf("各实体数量应均为 1")
	}
}

func TestRuleUpdateEnabledToggle(t *testing.T) {
	s := NewMemoryStore()
	r := testRule()
	_ = s.CreateRule(r)

	r.Enabled = false
	if err := s.UpdateRule(r); err != nil {
		t.Fatalf("更新规则失败: %v", err)
	}
	got, _ := s.GetRule("r1")
	if got.Enabled {
		t.Fatal("规则应已禁用")
	}
}

func TestEmployeeGetAfterUpdate(t *testing.T) {
	s := NewMemoryStore()
	e := testEmployee()
	_ = s.CreateEmployee(e)

	e.Name = "张三丰"
	_ = s.UpdateEmployee(e)

	got, _ := s.GetEmployee("e1")
	if got.Name != "张三丰" {
		t.Fatalf("更新后名称错误: %s", got.Name)
	}
}

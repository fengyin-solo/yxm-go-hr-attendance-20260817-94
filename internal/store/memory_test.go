package store

import (
	"testing"
	"time"

	"hrattendance/internal/model"
)

func testDepartment() *model.Department {
	return &model.Department{
		ID:        "d1",
		Name:      "研发部",
		Status:    model.DepartmentActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func testEmployee() *model.Employee {
	return &model.Employee{
		ID:           "e1",
		DepartmentID: "d1",
		Name:         "张三",
		EmpNo:        "E001",
		Position:     "工程师",
		HireDate:     time.Now(),
		Status:       model.EmployeeActive,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
}

func testAttendance() *model.Attendance {
	return &model.Attendance{
		ID:         "a1",
		EmployeeID: "e1",
		Date:       "2026-08-16",
		CheckIn:    "09:00",
		CheckOut:   "18:00",
		Status:     model.AttendanceNormal,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
}

func testLeave() *model.Leave {
	return &model.Leave{
		ID:         "l1",
		EmployeeID: "e1",
		Type:       model.LeaveAnnual,
		StartDate:  "2026-08-16",
		EndDate:    "2026-08-17",
		Reason:     "休假",
		Status:     model.LeavePending,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
}

func testOvertime() *model.Overtime {
	return &model.Overtime{
		ID:         "o1",
		EmployeeID: "e1",
		Date:       "2026-08-16",
		Hours:      2.5,
		Reason:     "赶进度",
		Status:     model.OvertimePending,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
}

func testRule() *model.AttendanceRule {
	return &model.AttendanceRule{
		ID:            "r1",
		Name:          "标准考勤",
		WorkStart:     "09:00",
		WorkEnd:       "18:00",
		LateThreshold: 15,
		Enabled:       true,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
}

func TestDepartmentCRUD(t *testing.T) {
	s := NewMemoryStore()
	d := testDepartment()

	if err := s.CreateDepartment(d); err != nil {
		t.Fatalf("CreateDepartment 失败: %v", err)
	}
	if err := s.CreateDepartment(d); err != ErrConflict {
		t.Fatalf("重复创建应返回 ErrConflict，实际: %v", err)
	}
	got, err := s.GetDepartment("d1")
	if err != nil {
		t.Fatalf("GetDepartment 失败: %v", err)
	}
	if got.Name != "研发部" {
		t.Fatalf("Name 不匹配: %s", got.Name)
	}
	if _, err := s.GetDepartment("missing"); err != ErrNotFound {
		t.Fatalf("查询不存在应返回 ErrNotFound")
	}
	d.Name = "技术部"
	if err := s.UpdateDepartment(d); err != nil {
		t.Fatalf("UpdateDepartment 失败: %v", err)
	}
	if len(s.ListDepartments()) != 1 {
		t.Fatalf("ListDepartments 数量应为 1")
	}
	if err := s.DeleteDepartment("d1"); err != nil {
		t.Fatalf("DeleteDepartment 失败: %v", err)
	}
	if _, err := s.GetDepartment("d1"); err != ErrNotFound {
		t.Fatalf("删除后应返回 ErrNotFound")
	}
}

func TestEmployeeCRUD(t *testing.T) {
	s := NewMemoryStore()
	e := testEmployee()

	if err := s.CreateEmployee(e); err != nil {
		t.Fatalf("CreateEmployee 失败: %v", err)
	}
	// 工号唯一性
	dup := testEmployee()
	dup.ID = "e2"
	if err := s.CreateEmployee(dup); err != ErrConflict {
		t.Fatalf("重复工号应返回 ErrConflict，实际: %v", err)
	}
	if _, err := s.GetEmployee("e1"); err != nil {
		t.Fatalf("GetEmployee 失败: %v", err)
	}
	if len(s.ListEmployees()) != 1 {
		t.Fatalf("ListEmployees 数量应为 1")
	}
	e.Position = "高级工程师"
	if err := s.UpdateEmployee(e); err != nil {
		t.Fatalf("UpdateEmployee 失败: %v", err)
	}
	if err := s.DeleteEmployee("e1"); err != nil {
		t.Fatalf("DeleteEmployee 失败: %v", err)
	}
	if _, err := s.GetEmployee("e1"); err != ErrNotFound {
		t.Fatalf("删除后应返回 ErrNotFound")
	}
}

func TestAttendanceCRUD(t *testing.T) {
	s := NewMemoryStore()
	a := testAttendance()

	if err := s.CreateAttendance(a); err != nil {
		t.Fatalf("CreateAttendance 失败: %v", err)
	}
	// 员工同日唯一性
	dup := testAttendance()
	dup.ID = "a2"
	if err := s.CreateAttendance(dup); err != ErrConflict {
		t.Fatalf("员工同日重复考勤应返回 ErrConflict，实际: %v", err)
	}
	if _, err := s.GetAttendance("a1"); err != nil {
		t.Fatalf("GetAttendance 失败: %v", err)
	}
	got, err := s.GetAttendanceByEmployeeDate("e1", "2026-08-16")
	if err != nil {
		t.Fatalf("GetAttendanceByEmployeeDate 失败: %v", err)
	}
	if got.ID != "a1" {
		t.Fatalf("按员工日期查询结果错误: %s", got.ID)
	}
	if _, err := s.GetAttendanceByEmployeeDate("e1", "2026-08-17"); err != ErrNotFound {
		t.Fatalf("查询不存在日期应返回 ErrNotFound")
	}
	a.Status = model.AttendanceLate
	if err := s.UpdateAttendance(a); err != nil {
		t.Fatalf("UpdateAttendance 失败: %v", err)
	}
	if err := s.DeleteAttendance("a1"); err != nil {
		t.Fatalf("DeleteAttendance 失败: %v", err)
	}
}

func TestLeaveCRUD(t *testing.T) {
	s := NewMemoryStore()
	l := testLeave()

	if err := s.CreateLeave(l); err != nil {
		t.Fatalf("CreateLeave 失败: %v", err)
	}
	if _, err := s.GetLeave("l1"); err != nil {
		t.Fatalf("GetLeave 失败: %v", err)
	}
	if len(s.ListLeaves()) != 1 {
		t.Fatalf("ListLeaves 数量应为 1")
	}
	l.Status = model.LeaveApproved
	if err := s.UpdateLeave(l); err != nil {
		t.Fatalf("UpdateLeave 失败: %v", err)
	}
	if err := s.DeleteLeave("l1"); err != nil {
		t.Fatalf("DeleteLeave 失败: %v", err)
	}
	if _, err := s.GetLeave("l1"); err != ErrNotFound {
		t.Fatalf("删除后应返回 ErrNotFound")
	}
}

func TestOvertimeCRUD(t *testing.T) {
	s := NewMemoryStore()
	o := testOvertime()

	if err := s.CreateOvertime(o); err != nil {
		t.Fatalf("CreateOvertime 失败: %v", err)
	}
	if _, err := s.GetOvertime("o1"); err != nil {
		t.Fatalf("GetOvertime 失败: %v", err)
	}
	if len(s.ListOvertimes()) != 1 {
		t.Fatalf("ListOvertimes 数量应为 1")
	}
	o.Status = model.OvertimeApproved
	if err := s.UpdateOvertime(o); err != nil {
		t.Fatalf("UpdateOvertime 失败: %v", err)
	}
	if err := s.DeleteOvertime("o1"); err != nil {
		t.Fatalf("DeleteOvertime 失败: %v", err)
	}
}

func TestRuleCRUD(t *testing.T) {
	s := NewMemoryStore()
	r := testRule()

	if err := s.CreateRule(r); err != nil {
		t.Fatalf("CreateRule 失败: %v", err)
	}
	if _, err := s.GetRule("r1"); err != nil {
		t.Fatalf("GetRule 失败: %v", err)
	}
	if len(s.ListRules()) != 1 {
		t.Fatalf("ListRules 数量应为 1")
	}
	r.Enabled = false
	if err := s.UpdateRule(r); err != nil {
		t.Fatalf("UpdateRule 失败: %v", err)
	}
	if err := s.DeleteRule("r1"); err != nil {
		t.Fatalf("DeleteRule 失败: %v", err)
	}
	if _, err := s.GetRule("r1"); err != ErrNotFound {
		t.Fatalf("删除后应返回 ErrNotFound")
	}
}

func TestUpdateNonExistent(t *testing.T) {
	s := NewMemoryStore()
	if err := s.UpdateDepartment(&model.Department{ID: "x"}); err != ErrNotFound {
		t.Fatalf("更新不存在部门应返回 ErrNotFound")
	}
	if err := s.UpdateEmployee(&model.Employee{ID: "x"}); err != ErrNotFound {
		t.Fatalf("更新不存在员工应返回 ErrNotFound")
	}
	if err := s.UpdateAttendance(&model.Attendance{ID: "x"}); err != ErrNotFound {
		t.Fatalf("更新不存在考勤应返回 ErrNotFound")
	}
	if err := s.UpdateLeave(&model.Leave{ID: "x"}); err != ErrNotFound {
		t.Fatalf("更新不存在请假应返回 ErrNotFound")
	}
	if err := s.UpdateOvertime(&model.Overtime{ID: "x"}); err != ErrNotFound {
		t.Fatalf("更新不存在加班应返回 ErrNotFound")
	}
	if err := s.UpdateRule(&model.AttendanceRule{ID: "x"}); err != ErrNotFound {
		t.Fatalf("更新不存在规则应返回 ErrNotFound")
	}
}

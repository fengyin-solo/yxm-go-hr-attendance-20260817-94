package service

import (
	"testing"
	"time"

	"hrattendance/internal/model"
)

func TestCreateEmployeeRequiresDepartment(t *testing.T) {
	s := newTestService()
	_, err := s.CreateEmployee(model.Employee{
		DepartmentID: "missing",
		Name:         "张三",
		EmpNo:        "E001",
		HireDate:     time.Now(),
	})
	if err == nil {
		t.Fatal("关联不存在的部门应被拒绝")
	}
}

func TestCreateEmployeeDuplicateEmpNo(t *testing.T) {
	s := newTestService()
	dept := mustCreateDepartment(t, s, "研发部")
	mustCreateEmployee(t, s, dept.ID, "张三", "E001")

	_, err := s.CreateEmployee(model.Employee{
		DepartmentID: dept.ID,
		Name:         "李四",
		EmpNo:        "E001",
		HireDate:     time.Now(),
	})
	if err == nil {
		t.Fatal("重复工号应被拒绝")
	}
}

func TestListEmployeesFilter(t *testing.T) {
	s := newTestService()
	deptA := mustCreateDepartment(t, s, "研发部")
	deptB := mustCreateDepartment(t, s, "市场部")
	mustCreateEmployee(t, s, deptA.ID, "张三", "E001")
	mustCreateEmployee(t, s, deptA.ID, "李四", "E002")
	mustCreateEmployee(t, s, deptB.ID, "王五", "E003")

	items, total, _ := s.ListEmployees(model.EmployeeFilter{DepartmentID: deptA.ID}, 1, 10)
	if total != 2 {
		t.Fatalf("按部门筛选应为 2，实际 %d", total)
	}

	items, total, _ = s.ListEmployees(model.EmployeeFilter{Keyword: "张"}, 1, 10)
	if total != 1 || items[0].Name != "张三" {
		t.Fatalf("按关键词筛选结果错误: total=%d", total)
	}
}

func TestUpdateEmployeeTransferDepartment(t *testing.T) {
	s := newTestService()
	deptA := mustCreateDepartment(t, s, "研发部")
	deptB := mustCreateDepartment(t, s, "市场部")
	emp := mustCreateEmployee(t, s, deptA.ID, "张三", "E001")

	updated, err := s.UpdateEmployee(emp.ID, model.Employee{DepartmentID: deptB.ID})
	if err != nil {
		t.Fatalf("转部门失败: %v", err)
	}
	if updated.DepartmentID != deptB.ID {
		t.Fatalf("转部门后部门错误: %s", updated.DepartmentID)
	}
}

func TestUpdateEmployeeResign(t *testing.T) {
	s := newTestService()
	dept := mustCreateDepartment(t, s, "研发部")
	emp := mustCreateEmployee(t, s, dept.ID, "张三", "E001")

	updated, err := s.UpdateEmployee(emp.ID, model.Employee{Status: model.EmployeeResigned})
	if err != nil {
		t.Fatalf("离职失败: %v", err)
	}
	if updated.Status != model.EmployeeResigned {
		t.Fatalf("离职后状态错误: %s", updated.Status)
	}
}

func TestDeleteEmployeeWithAttendance(t *testing.T) {
	s := newTestService()
	dept := mustCreateDepartment(t, s, "研发部")
	emp := mustCreateEmployee(t, s, dept.ID, "张三", "E001")
	_, _ = s.CreateAttendance(model.Attendance{EmployeeID: emp.ID, Date: "2026-08-16"})

	if err := s.DeleteEmployee(emp.ID); err == nil {
		t.Fatal("存在考勤记录的员工应拒绝删除")
	}
}

func TestBatchCreateEmployees(t *testing.T) {
	s := newTestService()
	dept := mustCreateDepartment(t, s, "研发部")

	employees := []model.Employee{
		{DepartmentID: dept.ID, Name: "张三", EmpNo: "E001", HireDate: time.Now()},
		{DepartmentID: dept.ID, Name: "李四", EmpNo: "E002", HireDate: time.Now()},
		{DepartmentID: dept.ID, Name: "王五", EmpNo: "E001", HireDate: time.Now()}, // 重复工号
	}
	result := s.BatchCreateEmployees(employees)
	if len(result.Succeeded) != 2 {
		t.Fatalf("成功应为 2，实际 %d", len(result.Succeeded))
	}
	if _, ok := result.Failed[2]; !ok {
		t.Fatalf("第 3 个（重复工号）应记录到 Failed")
	}
}

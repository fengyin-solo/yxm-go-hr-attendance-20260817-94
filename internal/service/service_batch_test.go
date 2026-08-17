package service

import (
	"testing"

	"hrattendance/internal/model"
)

func TestBatchApproveOvertimesPartialFailure(t *testing.T) {
	s := newTestService()
	dept := mustCreateDepartment(t, s, "研发部")
	emp := mustCreateEmployee(t, s, dept.ID, "张三", "E001")
	o1 := mustCreateOvertime(t, s, emp.ID, "2026-08-16", 2)
	o2 := mustCreateOvertime(t, s, emp.ID, "2026-08-17", 3)
	_, _ = s.ApproveOvertime(o2.ID)

	result := s.BatchApproveOvertimes([]string{o1.ID, o2.ID})
	if len(result.Succeeded) != 1 {
		t.Fatalf("成功应为 1（o2 已审批应失败），实际 %d", len(result.Succeeded))
	}
	if _, ok := result.Failed[o2.ID]; !ok {
		t.Fatalf("o2 应记录到 Failed")
	}
}

func TestBatchApproveLeavesEmptyInput(t *testing.T) {
	s := newTestService()
	result := s.BatchApproveLeaves(nil)
	if len(result.Succeeded) != 0 || len(result.Failed) != 0 {
		t.Fatalf("空输入应返回空结果")
	}
}

func TestBatchCreateAttendancesEmptyInput(t *testing.T) {
	s := newTestService()
	result := s.BatchCreateAttendances(nil)
	if len(result.Succeeded) != 0 || len(result.Failed) != 0 {
		t.Fatalf("空输入应返回空结果")
	}
}

func TestBatchCreateEmployeesEmptyInput(t *testing.T) {
	s := newTestService()
	result := s.BatchCreateEmployees(nil)
	if len(result.Succeeded) != 0 || len(result.Failed) != 0 {
		t.Fatalf("空输入应返回空结果")
	}
}

func TestDepartmentHierarchyThreeLevels(t *testing.T) {
	s := newTestService()
	root := mustCreateDepartment(t, s, "总公司")
	level1, _ := s.CreateDepartment(model.Department{Name: "事业部", ParentID: root.ID})
	level2, _ := s.CreateDepartment(model.Department{Name: "研发组", ParentID: level1.ID})

	if level2.ParentID != level1.ID {
		t.Fatalf("三级部门层级错误: %s", level2.ParentID)
	}
}

func TestTopLateEmployeesEmpty(t *testing.T) {
	s := newTestService()
	entries := s.TopLateEmployees("2026-08", 10)
	if len(entries) != 0 {
		t.Fatalf("无迟到记录应返回空，实际 %d", len(entries))
	}
}

func TestPerfectAttendanceEmpty(t *testing.T) {
	s := newTestService()
	employees := s.PerfectAttendanceEmployees("2026-08")
	if len(employees) != 0 {
		t.Fatalf("无员工应返回空，实际 %d", len(employees))
	}
}

package service

import (
	"testing"

	"hrattendance/internal/model"
)

func TestAttendanceUsesDefaultRule(t *testing.T) {
	s := newTestService()
	dept := mustCreateDepartment(t, s, "研发部")
	emp := mustCreateEmployee(t, s, dept.ID, "张三", "E001")

	// 默认规则：09:00 上班 + 15 分钟阈值，09:20 迟到
	a, _ := s.CreateAttendance(model.Attendance{EmployeeID: emp.ID, Date: "2026-08-16", CheckIn: "09:20", CheckOut: "18:00"})
	if a.Status != model.AttendanceLate {
		t.Fatalf("默认规则下 09:20 应为 late，实际 %s", a.Status)
	}

	// 09:10 在阈值内，不迟到
	a, _ = s.CreateAttendance(model.Attendance{EmployeeID: emp.ID, Date: "2026-08-17", CheckIn: "09:10", CheckOut: "18:00"})
	if a.Status != model.AttendanceNormal {
		t.Fatalf("默认规则下 09:10 应为 normal，实际 %s", a.Status)
	}
}

func TestAttendanceUsesCustomRule(t *testing.T) {
	s := newTestService()
	dept := mustCreateDepartment(t, s, "研发部")
	emp := mustCreateEmployee(t, s, dept.ID, "张三", "E001")

	// 启用自定义规则：10:00 上班，30 分钟阈值
	_, _ = s.CreateRule(model.AttendanceRule{
		Name:          "弹性工作制",
		WorkStart:     "10:00",
		WorkEnd:       "19:00",
		LateThreshold: 30,
		Enabled:       true,
	})

	// 10:20 在阈值内，不迟到
	a, _ := s.CreateAttendance(model.Attendance{EmployeeID: emp.ID, Date: "2026-08-16", CheckIn: "10:20", CheckOut: "19:00"})
	if a.Status != model.AttendanceNormal {
		t.Fatalf("10:20 应为 normal，实际 %s", a.Status)
	}

	// 10:40 超过阈值，迟到
	a, _ = s.CreateAttendance(model.Attendance{EmployeeID: emp.ID, Date: "2026-08-17", CheckIn: "10:40", CheckOut: "19:00"})
	if a.Status != model.AttendanceLate {
		t.Fatalf("10:40 应为 late，实际 %s", a.Status)
	}

	// 18:30 早于 19:00 下班，早退
	a, _ = s.CreateAttendance(model.Attendance{EmployeeID: emp.ID, Date: "2026-08-18", CheckIn: "10:00", CheckOut: "18:30"})
	if a.Status != model.AttendanceEarlyLeave {
		t.Fatalf("18:30 应为 early_leave，实际 %s", a.Status)
	}
}

func TestLeaveTypeValidation(t *testing.T) {
	s := newTestService()
	dept := mustCreateDepartment(t, s, "研发部")
	emp := mustCreateEmployee(t, s, dept.ID, "张三", "E001")

	for _, lt := range []string{model.LeaveAnnual, model.LeaveSick, model.LeavePersonal} {
		if _, err := s.CreateLeave(model.Leave{
			EmployeeID: emp.ID,
			Type:       lt,
			StartDate:  "2026-08-16",
			EndDate:    "2026-08-16",
			Reason:     "请假",
		}); err != nil {
			t.Fatalf("类型 %s 应合法，实际报错: %v", lt, err)
		}
	}

	if _, err := s.CreateLeave(model.Leave{
		EmployeeID: emp.ID,
		Type:       "xxx",
		StartDate:  "2026-08-16",
		EndDate:    "2026-08-16",
		Reason:     "请假",
	}); err == nil {
		t.Fatal("非法请假类型应被拒绝")
	}
}

func TestLeaveRejectThenCannotApprove(t *testing.T) {
	s := newTestService()
	dept := mustCreateDepartment(t, s, "研发部")
	emp := mustCreateEmployee(t, s, dept.ID, "张三", "E001")
	l := mustCreateLeave(t, s, emp.ID, model.LeaveAnnual, "2026-08-16", "2026-08-16")

	_, _ = s.RejectLeave(l.ID)
	if _, err := s.ApproveLeave(l.ID); err == nil {
		t.Fatal("已驳回的请假单不应可审批通过")
	}
}

func TestOvertimeHoursEdge(t *testing.T) {
	s := newTestService()
	dept := mustCreateDepartment(t, s, "研发部")
	emp := mustCreateEmployee(t, s, dept.ID, "张三", "E001")

	if _, err := s.CreateOvertime(model.Overtime{
		EmployeeID: emp.ID, Date: "2026-08-16", Hours: 0, Reason: "赶进度",
	}); err == nil {
		t.Fatal("零时长加班应被拒绝")
	}
	if _, err := s.CreateOvertime(model.Overtime{
		EmployeeID: emp.ID, Date: "2026-08-16", Hours: -1, Reason: "赶进度",
	}); err == nil {
		t.Fatal("负时长加班应被拒绝")
	}
}

func TestDepartmentEmployeeCountNotFound(t *testing.T) {
	s := newTestService()
	if _, err := s.DepartmentEmployeeCount("missing"); err == nil {
		t.Fatal("统计不存在部门应报错")
	}
}

func TestUpdateEmployeeTransferToMissingDepartment(t *testing.T) {
	s := newTestService()
	dept := mustCreateDepartment(t, s, "研发部")
	emp := mustCreateEmployee(t, s, dept.ID, "张三", "E001")

	if _, err := s.UpdateEmployee(emp.ID, model.Employee{DepartmentID: "missing"}); err == nil {
		t.Fatal("转到不存在的部门应报错")
	}
}

func TestLeaveApproverIsEmployee(t *testing.T) {
	s := newTestService()
	dept := mustCreateDepartment(t, s, "研发部")
	emp := mustCreateEmployee(t, s, dept.ID, "张三", "E001")
	approver := mustCreateEmployee(t, s, dept.ID, "经理", "M001")

	l, err := s.CreateLeave(model.Leave{
		EmployeeID: emp.ID,
		Type:       model.LeaveAnnual,
		StartDate:  "2026-08-16",
		EndDate:    "2026-08-16",
		Reason:     "休假",
		ApproverID: approver.ID,
	})
	if err != nil {
		t.Fatalf("创建带审批人请假单失败: %v", err)
	}
	if l.ApproverID != approver.ID {
		t.Fatalf("审批人记录错误: %s", l.ApproverID)
	}
}

func TestOvertimeApproverIsEmployee(t *testing.T) {
	s := newTestService()
	dept := mustCreateDepartment(t, s, "研发部")
	emp := mustCreateEmployee(t, s, dept.ID, "张三", "E001")
	approver := mustCreateEmployee(t, s, dept.ID, "经理", "M001")

	o, err := s.CreateOvertime(model.Overtime{
		EmployeeID: emp.ID,
		Date:       "2026-08-16",
		Hours:      2,
		Reason:     "赶进度",
		ApproverID: approver.ID,
	})
	if err != nil {
		t.Fatalf("创建带审批人加班单失败: %v", err)
	}
	if o.ApproverID != approver.ID {
		t.Fatalf("审批人记录错误: %s", o.ApproverID)
	}
}

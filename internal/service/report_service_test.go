package service

import (
	"testing"

	"hrattendance/internal/model"
)

func TestEmployeeMonthlyReport(t *testing.T) {
	s := newTestService()
	dept := mustCreateDepartment(t, s, "研发部")
	emp := mustCreateEmployee(t, s, dept.ID, "张三", "E001")

	// 考勤：正常 1 天、迟到 1 天、缺勤 1 天
	_, _ = s.CreateAttendance(model.Attendance{EmployeeID: emp.ID, Date: "2026-08-10", CheckIn: "09:00", CheckOut: "18:00"})
	_, _ = s.CreateAttendance(model.Attendance{EmployeeID: emp.ID, Date: "2026-08-11", CheckIn: "09:30", CheckOut: "18:00"})
	_, _ = s.CreateAttendance(model.Attendance{EmployeeID: emp.ID, Date: "2026-08-12"})

	// 请假：8月13-14 两天，已审批
	l := mustCreateLeave(t, s, emp.ID, model.LeaveAnnual, "2026-08-13", "2026-08-14")
	_, _ = s.ApproveLeave(l.ID)

	// 加班：8月15 已审批 2 小时
	o := mustCreateOvertime(t, s, emp.ID, "2026-08-15", 2)
	_, _ = s.ApproveOvertime(o.ID)

	report, err := s.EmployeeMonthlyReport(emp.ID, "2026-08")
	if err != nil {
		t.Fatalf("生成报告失败: %v", err)
	}
	if report.NormalDays != 1 {
		t.Fatalf("正常天数应为 1，实际 %d", report.NormalDays)
	}
	if report.LateDays != 1 {
		t.Fatalf("迟到天数应为 1，实际 %d", report.LateDays)
	}
	if report.AbsentDays != 1 {
		t.Fatalf("缺勤天数应为 1，实际 %d", report.AbsentDays)
	}
	if report.LeaveDays != 2 {
		t.Fatalf("请假天数应为 2，实际 %d", report.LeaveDays)
	}
	if report.OvertimeHours != 2 {
		t.Fatalf("加班时长应为 2，实际 %v", report.OvertimeHours)
	}
}

func TestEmployeeMonthlyReportFiltersOtherMonth(t *testing.T) {
	s := newTestService()
	dept := mustCreateDepartment(t, s, "研发部")
	emp := mustCreateEmployee(t, s, dept.ID, "张三", "E001")
	_, _ = s.CreateAttendance(model.Attendance{EmployeeID: emp.ID, Date: "2026-07-01", CheckIn: "09:00", CheckOut: "18:00"})

	report, _ := s.EmployeeMonthlyReport(emp.ID, "2026-08")
	if report.NormalDays != 0 {
		t.Fatalf("其他月份的考勤不应计入，实际 %d", report.NormalDays)
	}
}

func TestDepartmentMonthlyReport(t *testing.T) {
	s := newTestService()
	dept := mustCreateDepartment(t, s, "研发部")
	emp1 := mustCreateEmployee(t, s, dept.ID, "张三", "E001")
	emp2 := mustCreateEmployee(t, s, dept.ID, "李四", "E002")

	_, _ = s.CreateAttendance(model.Attendance{EmployeeID: emp1.ID, Date: "2026-08-10", CheckIn: "09:00", CheckOut: "18:00"})
	_, _ = s.CreateAttendance(model.Attendance{EmployeeID: emp2.ID, Date: "2026-08-10", CheckIn: "09:30", CheckOut: "18:00"})

	report, err := s.DepartmentMonthlyReport(dept.ID, "2026-08")
	if err != nil {
		t.Fatalf("生成部门报告失败: %v", err)
	}
	if report.EmployeeCount != 2 {
		t.Fatalf("员工数应为 2，实际 %d", report.EmployeeCount)
	}
	if len(report.Reports) != 2 {
		t.Fatalf("员工报告数应为 2，实际 %d", len(report.Reports))
	}
}

func TestAttendanceStatusStats(t *testing.T) {
	s := newTestService()
	dept := mustCreateDepartment(t, s, "研发部")
	emp := mustCreateEmployee(t, s, dept.ID, "张三", "E001")

	_, _ = s.CreateAttendance(model.Attendance{EmployeeID: emp.ID, Date: "2026-08-10", CheckIn: "09:00", CheckOut: "18:00"})
	_, _ = s.CreateAttendance(model.Attendance{EmployeeID: emp.ID, Date: "2026-08-11", CheckIn: "09:30", CheckOut: "18:00"})
	_, _ = s.CreateAttendance(model.Attendance{EmployeeID: emp.ID, Date: "2026-08-12", CheckIn: "09:00", CheckOut: "17:00"})
	_, _ = s.CreateAttendance(model.Attendance{EmployeeID: emp.ID, Date: "2026-08-13"})

	stats := s.AttendanceStatusStats()
	if stats.Normal != 1 || stats.Late != 1 || stats.EarlyLeave != 1 || stats.Absent != 1 {
		t.Fatalf("状态分布错误: %+v", stats)
	}
}

func TestLeaveTypeStats(t *testing.T) {
	s := newTestService()
	dept := mustCreateDepartment(t, s, "研发部")
	emp := mustCreateEmployee(t, s, dept.ID, "张三", "E001")

	l1 := mustCreateLeave(t, s, emp.ID, model.LeaveAnnual, "2026-08-13", "2026-08-14")
	_, _ = s.ApproveLeave(l1.ID)
	l2 := mustCreateLeave(t, s, emp.ID, model.LeaveSick, "2026-08-15", "2026-08-15")
	_, _ = s.ApproveLeave(l2.ID)
	// 未审批的不应计入
	mustCreateLeave(t, s, emp.ID, model.LeavePersonal, "2026-08-16", "2026-08-17")

	stats := s.LeaveTypeStats()
	if stats.Annual != 2 || stats.Sick != 1 || stats.Personal != 0 {
		t.Fatalf("请假统计错误: %+v", stats)
	}
}

func TestOvertimeStats(t *testing.T) {
	s := newTestService()
	dept := mustCreateDepartment(t, s, "研发部")
	emp := mustCreateEmployee(t, s, dept.ID, "张三", "E001")

	o1 := mustCreateOvertime(t, s, emp.ID, "2026-08-15", 2)
	_, _ = s.ApproveOvertime(o1.ID)
	o2 := mustCreateOvertime(t, s, emp.ID, "2026-08-16", 3.5)
	_, _ = s.ApproveOvertime(o2.ID)
	// 未审批的不应计入
	mustCreateOvertime(t, s, emp.ID, "2026-08-17", 1)

	stats := s.OvertimeStats()
	if stats.Count != 2 || stats.TotalHours != 5.5 {
		t.Fatalf("加班统计错误: %+v", stats)
	}
}

func TestTopLateEmployees(t *testing.T) {
	s := newTestService()
	dept := mustCreateDepartment(t, s, "研发部")
	emp1 := mustCreateEmployee(t, s, dept.ID, "张三", "E001")
	emp2 := mustCreateEmployee(t, s, dept.ID, "李四", "E002")

	// 张三迟到 2 次，李四迟到 1 次
	_, _ = s.CreateAttendance(model.Attendance{EmployeeID: emp1.ID, Date: "2026-08-10", CheckIn: "09:30", CheckOut: "18:00"})
	_, _ = s.CreateAttendance(model.Attendance{EmployeeID: emp1.ID, Date: "2026-08-11", CheckIn: "09:30", CheckOut: "18:00"})
	_, _ = s.CreateAttendance(model.Attendance{EmployeeID: emp2.ID, Date: "2026-08-10", CheckIn: "09:30", CheckOut: "18:00"})

	entries := s.TopLateEmployees("2026-08", 10)
	if len(entries) != 2 {
		t.Fatalf("迟到员工数应为 2，实际 %d", len(entries))
	}
	if entries[0].EmployeeName != "张三" || entries[0].LateDays != 2 {
		t.Fatalf("第一名应为张三(2次)，实际 %s(%d)", entries[0].EmployeeName, entries[0].LateDays)
	}
	if entries[1].LateDays != 1 {
		t.Fatalf("第二名迟到次数应为 1，实际 %d", entries[1].LateDays)
	}
}

func TestTopLateEmployeesLimit(t *testing.T) {
	s := newTestService()
	dept := mustCreateDepartment(t, s, "研发部")
	for i := 0; i < 5; i++ {
		emp := mustCreateEmployee(t, s, dept.ID, "员工"+string(rune('A'+i)), "E00"+string(rune('1'+i)))
		_, _ = s.CreateAttendance(model.Attendance{EmployeeID: emp.ID, Date: "2026-08-10", CheckIn: "09:30", CheckOut: "18:00"})
	}

	entries := s.TopLateEmployees("2026-08", 2)
	if len(entries) != 2 {
		t.Fatalf("限制应为 2，实际 %d", len(entries))
	}
}

func TestPerfectAttendanceEmployees(t *testing.T) {
	s := newTestService()
	dept := mustCreateDepartment(t, s, "研发部")
	emp1 := mustCreateEmployee(t, s, dept.ID, "张三", "E001") // 全勤
	emp2 := mustCreateEmployee(t, s, dept.ID, "李四", "E002") // 迟到
	_ = mustCreateEmployee(t, s, dept.ID, "王五", "E003")     // 无记录

	_, _ = s.CreateAttendance(model.Attendance{EmployeeID: emp1.ID, Date: "2026-08-10", CheckIn: "09:00", CheckOut: "18:00"})
	_, _ = s.CreateAttendance(model.Attendance{EmployeeID: emp2.ID, Date: "2026-08-10", CheckIn: "09:30", CheckOut: "18:00"})

	perfect := s.PerfectAttendanceEmployees("2026-08")
	if len(perfect) != 1 {
		t.Fatalf("全勤员工应为 1，实际 %d", len(perfect))
	}
	if perfect[0].Name != "张三" {
		t.Fatalf("全勤员工应为张三，实际 %s", perfect[0].Name)
	}
}

func TestExportCompanyReportEmpty(t *testing.T) {
	s := newTestService()
	report := s.ExportCompanyReport()
	if report.EmployeeCount != 0 || report.DepartmentCount != 0 {
		t.Fatalf("空公司统计应为 0，实际 %+v", report)
	}
}

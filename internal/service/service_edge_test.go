package service

import (
	"testing"

	"hrattendance/internal/model"
)

func TestListDepartmentsPagination(t *testing.T) {
	s := newTestService()
	for i := 0; i < 5; i++ {
		mustCreateDepartment(t, s, "部门"+string(rune('A'+i)))
	}
	items, total, _ := s.ListDepartments(model.DepartmentFilter{}, 1, 2)
	if total != 5 || len(items) != 2 {
		t.Fatalf("第一页应为 2 条，实际 total=%d len=%d", total, len(items))
	}
	items, _, _ = s.ListDepartments(model.DepartmentFilter{}, 3, 2)
	if len(items) != 1 {
		t.Fatalf("第三页应为 1 条，实际 %d", len(items))
	}
}

func TestListEmployeesPagination(t *testing.T) {
	s := newTestService()
	dept := mustCreateDepartment(t, s, "研发部")
	for i := 0; i < 5; i++ {
		mustCreateEmployee(t, s, dept.ID, "员工"+string(rune('A'+i)), "E00"+string(rune('1'+i)))
	}
	items, total, _ := s.ListEmployees(model.EmployeeFilter{}, 2, 2)
	if total != 5 || len(items) != 2 {
		t.Fatalf("第二页应为 2 条，实际 total=%d len=%d", total, len(items))
	}
}

func TestListAttendancesPagination(t *testing.T) {
	s := newTestService()
	dept := mustCreateDepartment(t, s, "研发部")
	emp := mustCreateEmployee(t, s, dept.ID, "张三", "E001")
	for i := 0; i < 5; i++ {
		date := "2026-08-1" + string(rune('0'+i))
		_, _ = s.CreateAttendance(model.Attendance{EmployeeID: emp.ID, Date: date, CheckIn: "09:00", CheckOut: "18:00"})
	}
	items, total, _ := s.ListAttendances(model.AttendanceFilter{}, 1, 3)
	if total != 5 || len(items) != 3 {
		t.Fatalf("第一页应为 3 条，实际 total=%d len=%d", total, len(items))
	}
}

func TestBatchCreateAttendances(t *testing.T) {
	s := newTestService()
	dept := mustCreateDepartment(t, s, "研发部")
	emp := mustCreateEmployee(t, s, dept.ID, "张三", "E001")

	attendances := []model.Attendance{
		{EmployeeID: emp.ID, Date: "2026-08-16", CheckIn: "09:00", CheckOut: "18:00"},
		{EmployeeID: emp.ID, Date: "2026-08-17", CheckIn: "09:00", CheckOut: "18:00"},
		{EmployeeID: "missing", Date: "2026-08-18", CheckIn: "09:00", CheckOut: "18:00"},
	}
	result := s.BatchCreateAttendances(attendances)
	if len(result.Succeeded) != 2 {
		t.Fatalf("成功应为 2，实际 %d", len(result.Succeeded))
	}
	if _, ok := result.Failed[2]; !ok {
		t.Fatalf("不存在的员工应记录到 Failed")
	}
}

func TestBatchCreateAttendancesDuplicate(t *testing.T) {
	s := newTestService()
	dept := mustCreateDepartment(t, s, "研发部")
	emp := mustCreateEmployee(t, s, dept.ID, "张三", "E001")

	attendances := []model.Attendance{
		{EmployeeID: emp.ID, Date: "2026-08-16", CheckIn: "09:00", CheckOut: "18:00"},
		{EmployeeID: emp.ID, Date: "2026-08-16", CheckIn: "09:00", CheckOut: "18:00"},
	}
	result := s.BatchCreateAttendances(attendances)
	if len(result.Succeeded) != 1 {
		t.Fatalf("成功应为 1，实际 %d", len(result.Succeeded))
	}
	if _, ok := result.Failed[1]; !ok {
		t.Fatalf("同日重复应记录到 Failed")
	}
}

func TestCreateAttendanceDuplicateDate(t *testing.T) {
	s := newTestService()
	dept := mustCreateDepartment(t, s, "研发部")
	emp := mustCreateEmployee(t, s, dept.ID, "张三", "E001")

	_, _ = s.CreateAttendance(model.Attendance{EmployeeID: emp.ID, Date: "2026-08-16", CheckIn: "09:00", CheckOut: "18:00"})
	_, err := s.CreateAttendance(model.Attendance{EmployeeID: emp.ID, Date: "2026-08-16", CheckIn: "09:00", CheckOut: "18:00"})
	if err == nil {
		t.Fatal("同日重复考勤应被拒绝")
	}
}

func TestLeaveDateFilter(t *testing.T) {
	s := newTestService()
	dept := mustCreateDepartment(t, s, "研发部")
	emp := mustCreateEmployee(t, s, dept.ID, "张三", "E001")
	mustCreateLeave(t, s, emp.ID, model.LeaveAnnual, "2026-08-10", "2026-08-12")
	mustCreateLeave(t, s, emp.ID, model.LeaveSick, "2026-08-20", "2026-08-21")

	items, total, _ := s.ListLeaves(model.LeaveFilter{DateFrom: "2026-08-01", DateTo: "2026-08-15"}, 1, 10)
	if total != 1 {
		t.Fatalf("8月上半月请假应为 1，实际 %d", total)
	}
	_ = items
}

func TestOvertimeDateFilter(t *testing.T) {
	s := newTestService()
	dept := mustCreateDepartment(t, s, "研发部")
	emp := mustCreateEmployee(t, s, dept.ID, "张三", "E001")
	mustCreateOvertime(t, s, emp.ID, "2026-08-10", 2)
	mustCreateOvertime(t, s, emp.ID, "2026-08-20", 3)

	items, total, _ := s.ListOvertimes(model.OvertimeFilter{DateFrom: "2026-08-15", DateTo: "2026-08-31"}, 1, 10)
	if total != 1 || items[0].Hours != 3 {
		t.Fatalf("8月下半月加班应为 1 条 3 小时，实际 total=%d", total)
	}
}

func TestEmployeeKeywordFilter(t *testing.T) {
	s := newTestService()
	dept := mustCreateDepartment(t, s, "研发部")
	mustCreateEmployee(t, s, dept.ID, "张三", "E001")
	mustCreateEmployee(t, s, dept.ID, "李四", "E002")

	items, total, _ := s.ListEmployees(model.EmployeeFilter{Keyword: "E001"}, 1, 10)
	if total != 1 || items[0].Name != "张三" {
		t.Fatalf("按工号关键词筛选结果错误: total=%d", total)
	}
}

func TestGetNonExistentRecords(t *testing.T) {
	s := newTestService()
	if _, err := s.GetDepartment("x"); err == nil {
		t.Fatal("查询不存在部门应报错")
	}
	if _, err := s.GetEmployee("x"); err == nil {
		t.Fatal("查询不存在员工应报错")
	}
	if _, err := s.GetAttendance("x"); err == nil {
		t.Fatal("查询不存在考勤应报错")
	}
	if _, err := s.GetLeave("x"); err == nil {
		t.Fatal("查询不存在请假应报错")
	}
	if _, err := s.GetOvertime("x"); err == nil {
		t.Fatal("查询不存在加班应报错")
	}
	if _, err := s.GetRule("x"); err == nil {
		t.Fatal("查询不存在规则应报错")
	}
}

func TestEmployeeMonthlyReportNotFound(t *testing.T) {
	s := newTestService()
	if _, err := s.EmployeeMonthlyReport("missing", "2026-08"); err == nil {
		t.Fatal("不存在员工应报错")
	}
}

func TestDepartmentMonthlyReportNotFound(t *testing.T) {
	s := newTestService()
	if _, err := s.DepartmentMonthlyReport("missing", "2026-08"); err == nil {
		t.Fatal("不存在部门应报错")
	}
}

func TestCheckInNonexistentEmployee(t *testing.T) {
	s := newTestService()
	if _, err := s.CheckIn("missing", "2026-08-16", "09:00"); err == nil {
		t.Fatal("不存在员工签到应报错")
	}
	if _, err := s.CheckOut("missing", "2026-08-16", "18:00"); err == nil {
		t.Fatal("不存在员工签退应报错")
	}
}

func TestExportCompanyReport(t *testing.T) {
	s := newTestService()
	dept := mustCreateDepartment(t, s, "研发部")
	mustCreateEmployee(t, s, dept.ID, "张三", "E001")
	mustCreateEmployee(t, s, dept.ID, "李四", "E002")

	report := s.ExportCompanyReport()
	if report.EmployeeCount != 2 {
		t.Fatalf("员工数应为 2，实际 %d", report.EmployeeCount)
	}
	if report.DepartmentCount != 1 {
		t.Fatalf("部门数应为 1，实际 %d", report.DepartmentCount)
	}
	if report.AttendanceStats == nil || report.LeaveStats == nil || report.OvertimeStats == nil {
		t.Fatal("报告统计字段不应为 nil")
	}
}

func TestBatchApproveLeavesPartialFailure(t *testing.T) {
	s := newTestService()
	dept := mustCreateDepartment(t, s, "研发部")
	emp := mustCreateEmployee(t, s, dept.ID, "张三", "E001")
	l1 := mustCreateLeave(t, s, emp.ID, model.LeaveAnnual, "2026-08-16", "2026-08-16")
	l2 := mustCreateLeave(t, s, emp.ID, model.LeaveSick, "2026-08-17", "2026-08-17")
	_, _ = s.ApproveLeave(l2.ID) // l2 已审批

	result := s.BatchApproveLeaves([]string{l1.ID, l2.ID})
	if len(result.Succeeded) != 1 {
		t.Fatalf("成功应为 1（l2 已审批应失败），实际 %d", len(result.Succeeded))
	}
	if _, ok := result.Failed[l2.ID]; !ok {
		t.Fatalf("l2 应记录到 Failed")
	}
}

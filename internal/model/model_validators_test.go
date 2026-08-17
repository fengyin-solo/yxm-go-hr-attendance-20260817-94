package model

import "testing"

func TestValidLeaveType(t *testing.T) {
	for _, s := range []string{LeaveAnnual, LeaveSick, LeavePersonal} {
		if !ValidLeaveType(s) {
			t.Fatalf("%s 应为合法请假类型", s)
		}
	}
	if ValidLeaveType("xxx") {
		t.Fatal("xxx 不应是合法请假类型")
	}
}

func TestValidLeaveStatus(t *testing.T) {
	for _, s := range []string{LeavePending, LeaveApproved, LeaveRejected} {
		if !ValidLeaveStatus(s) {
			t.Fatalf("%s 应为合法请假状态", s)
		}
	}
	if ValidLeaveStatus("xxx") {
		t.Fatal("xxx 不应是合法请假状态")
	}
}

func TestValidOvertimeStatus(t *testing.T) {
	for _, s := range []string{OvertimePending, OvertimeApproved, OvertimeRejected} {
		if !ValidOvertimeStatus(s) {
			t.Fatalf("%s 应为合法加班状态", s)
		}
	}
	if ValidOvertimeStatus("xxx") {
		t.Fatal("xxx 不应是合法加班状态")
	}
}

func TestDepartmentFilterMatch(t *testing.T) {
	d := &Department{Name: "研发部", ParentID: "p1", Status: DepartmentActive}
	if !(DepartmentFilter{}).Match(d) {
		t.Fatal("空过滤应匹配")
	}
	if !(DepartmentFilter{Status: DepartmentActive}).Match(d) {
		t.Fatal("按状态匹配应通过")
	}
	if (DepartmentFilter{Status: DepartmentInactive}).Match(d) {
		t.Fatal("状态不匹配应失败")
	}
	if !(DepartmentFilter{Keyword: "研发"}).Match(d) {
		t.Fatal("关键词应匹配")
	}
	if (DepartmentFilter{Keyword: "市场"}).Match(d) {
		t.Fatal("不匹配关键词应失败")
	}
	if !(DepartmentFilter{ParentID: "p1"}).Match(d) {
		t.Fatal("按上级部门匹配应通过")
	}
	if (DepartmentFilter{ParentID: "p2"}).Match(d) {
		t.Fatal("上级部门不匹配应失败")
	}
}

func TestEmployeeFilterMatch(t *testing.T) {
	e := &Employee{DepartmentID: "d1", Name: "张三", EmpNo: "E001", Status: EmployeeActive}
	if !(EmployeeFilter{DepartmentID: "d1"}).Match(e) {
		t.Fatal("按部门匹配应通过")
	}
	if (EmployeeFilter{DepartmentID: "d2"}).Match(e) {
		t.Fatal("部门不匹配应失败")
	}
	if !(EmployeeFilter{Keyword: "E001"}).Match(e) {
		t.Fatal("按工号关键词应匹配")
	}
	if !(EmployeeFilter{Keyword: "张"}).Match(e) {
		t.Fatal("按姓名关键词应匹配")
	}
	if (EmployeeFilter{Status: EmployeeResigned}).Match(e) {
		t.Fatal("状态不匹配应失败")
	}
}

func TestAttendanceFilterMatch(t *testing.T) {
	a := &Attendance{EmployeeID: "e1", Date: "2026-08-16", Status: AttendanceLate}
	if !(AttendanceFilter{EmployeeID: "e1"}).Match(a) {
		t.Fatal("按员工匹配应通过")
	}
	if !(AttendanceFilter{Status: AttendanceLate}).Match(a) {
		t.Fatal("按状态匹配应通过")
	}
	if !(AttendanceFilter{DateFrom: "2026-08-01", DateTo: "2026-08-31"}).Match(a) {
		t.Fatal("日期范围内应匹配")
	}
	if (AttendanceFilter{DateFrom: "2026-09-01"}).Match(a) {
		t.Fatal("日期范围外应失败")
	}
	if (AttendanceFilter{DateTo: "2026-08-15"}).Match(a) {
		t.Fatal("早于 DateTo 应匹配，实际失败")
	}
}

func TestLeaveFilterMatch(t *testing.T) {
	l := &Leave{EmployeeID: "e1", Type: LeaveAnnual, StartDate: "2026-08-16", EndDate: "2026-08-18", Status: LeavePending}
	if !(LeaveFilter{Type: LeaveAnnual}).Match(l) {
		t.Fatal("按类型匹配应通过")
	}
	if !(LeaveFilter{Status: LeavePending}).Match(l) {
		t.Fatal("按状态匹配应通过")
	}
	if !(LeaveFilter{DateFrom: "2026-08-15", DateTo: "2026-08-20"}).Match(l) {
		t.Fatal("日期范围内应匹配")
	}
	if (LeaveFilter{Type: LeaveSick}).Match(l) {
		t.Fatal("类型不匹配应失败")
	}
}

func TestOvertimeFilterMatch(t *testing.T) {
	o := &Overtime{EmployeeID: "e1", Date: "2026-08-16", Status: OvertimePending}
	if !(OvertimeFilter{EmployeeID: "e1"}).Match(o) {
		t.Fatal("按员工匹配应通过")
	}
	if !(OvertimeFilter{DateFrom: "2026-08-01", DateTo: "2026-08-31"}).Match(o) {
		t.Fatal("日期范围内应匹配")
	}
	if (OvertimeFilter{DateFrom: "2026-08-17"}).Match(o) {
		t.Fatal("日期范围外应失败")
	}
}

func TestLeaveDaysEdge(t *testing.T) {
	// 非法日期应返回 0
	l := Leave{StartDate: "bad", EndDate: "2026-08-18"}
	if l.Days() != 0 {
		t.Fatalf("非法日期应为 0 天，实际 %d", l.Days())
	}
	// 结束等于开始
	l = Leave{StartDate: "2026-08-16", EndDate: "2026-08-16"}
	if l.Days() != 1 {
		t.Fatalf("同日应为 1 天，实际 %d", l.Days())
	}
}

func TestAttendanceValidateTimeFormat(t *testing.T) {
	a := Attendance{EmployeeID: "e1", Date: "2026-08-16", CheckIn: "25:00"}
	if err := a.Validate(); err == nil {
		t.Fatal("25:00 应为非法时间")
	}
}

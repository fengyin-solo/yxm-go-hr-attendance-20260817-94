package model

import (
	"testing"
	"time"
)

func TestDepartmentValidate(t *testing.T) {
	cases := []struct {
		name string
		d    Department
		ok   bool
	}{
		{"正常", Department{Name: "研发部"}, true},
		{"空名", Department{}, false},
		{"非法状态", Department{Name: "研发部", Status: "xxx"}, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := c.d.Validate()
			if c.ok && err != nil {
				t.Fatalf("应通过校验，实际: %v", err)
			}
			if !c.ok && err == nil {
				t.Fatal("应校验失败")
			}
		})
	}
}

func TestEmployeeValidate(t *testing.T) {
	hd := time.Now()
	cases := []struct {
		name string
		e    Employee
		ok   bool
	}{
		{"正常", Employee{DepartmentID: "d1", Name: "张三", EmpNo: "E001", HireDate: hd}, true},
		{"空名", Employee{DepartmentID: "d1", EmpNo: "E001", HireDate: hd}, false},
		{"空工号", Employee{DepartmentID: "d1", Name: "张三", HireDate: hd}, false},
		{"空部门", Employee{Name: "张三", EmpNo: "E001", HireDate: hd}, false},
		{"空入职日期", Employee{DepartmentID: "d1", Name: "张三", EmpNo: "E001"}, false},
		{"非法状态", Employee{DepartmentID: "d1", Name: "张三", EmpNo: "E001", HireDate: hd, Status: "xxx"}, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := c.e.Validate()
			if c.ok && err != nil {
				t.Fatalf("应通过校验，实际: %v", err)
			}
			if !c.ok && err == nil {
				t.Fatal("应校验失败")
			}
		})
	}
}

func TestAttendanceValidate(t *testing.T) {
	cases := []struct {
		name string
		a    Attendance
		ok   bool
	}{
		{"正常", Attendance{EmployeeID: "e1", Date: "2026-08-16", CheckIn: "09:00", CheckOut: "18:00"}, true},
		{"空员工", Attendance{Date: "2026-08-16"}, false},
		{"空日期", Attendance{EmployeeID: "e1"}, false},
		{"日期格式错误", Attendance{EmployeeID: "e1", Date: "2026/08/16"}, false},
		{"时间格式错误", Attendance{EmployeeID: "e1", Date: "2026-08-16", CheckIn: "9:5"}, false},
		{"非法状态", Attendance{EmployeeID: "e1", Date: "2026-08-16", Status: "xxx"}, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := c.a.Validate()
			if c.ok && err != nil {
				t.Fatalf("应通过校验，实际: %v", err)
			}
			if !c.ok && err == nil {
				t.Fatal("应校验失败")
			}
		})
	}
}

func TestLeaveValidate(t *testing.T) {
	cases := []struct {
		name string
		l    Leave
		ok   bool
	}{
		{"正常", Leave{EmployeeID: "e1", Type: LeaveAnnual, StartDate: "2026-08-16", EndDate: "2026-08-17", Reason: "休假"}, true},
		{"非法类型", Leave{EmployeeID: "e1", Type: "xxx", StartDate: "2026-08-16", EndDate: "2026-08-17", Reason: "休假"}, false},
		{"空事由", Leave{EmployeeID: "e1", Type: LeaveAnnual, StartDate: "2026-08-16", EndDate: "2026-08-17"}, false},
		{"结束早于开始", Leave{EmployeeID: "e1", Type: LeaveAnnual, StartDate: "2026-08-17", EndDate: "2026-08-16", Reason: "休假"}, false},
		{"非法状态", Leave{EmployeeID: "e1", Type: LeaveAnnual, StartDate: "2026-08-16", EndDate: "2026-08-17", Reason: "休假", Status: "xxx"}, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := c.l.Validate()
			if c.ok && err != nil {
				t.Fatalf("应通过校验，实际: %v", err)
			}
			if !c.ok && err == nil {
				t.Fatal("应校验失败")
			}
		})
	}
}

func TestLeaveDays(t *testing.T) {
	cases := []struct {
		start, end string
		want       int
	}{
		{"2026-08-16", "2026-08-16", 1},
		{"2026-08-16", "2026-08-17", 2},
		{"2026-08-16", "2026-08-20", 5},
	}
	for _, c := range cases {
		l := Leave{StartDate: c.start, EndDate: c.end}
		if got := l.Days(); got != c.want {
			t.Fatalf("%s ~ %s 应为 %d 天，实际 %d", c.start, c.end, c.want, got)
		}
	}
}

func TestLeaveTransitions(t *testing.T) {
	cases := []struct {
		from, to string
		ok       bool
	}{
		{LeavePending, LeaveApproved, true},
		{LeavePending, LeaveRejected, true},
		{LeaveApproved, LeaveRejected, false},
		{LeaveRejected, LeaveApproved, false},
	}
	for _, c := range cases {
		if got := CanTransitionLeave(c.from, c.to); got != c.ok {
			t.Fatalf("流转 %s -> %s 期望 %v，实际 %v", c.from, c.to, c.ok, got)
		}
	}
}

func TestOvertimeValidate(t *testing.T) {
	cases := []struct {
		name string
		o    Overtime
		ok   bool
	}{
		{"正常", Overtime{EmployeeID: "e1", Date: "2026-08-16", Hours: 2.5, Reason: "赶进度"}, true},
		{"零时长", Overtime{EmployeeID: "e1", Date: "2026-08-16", Hours: 0, Reason: "赶进度"}, false},
		{"空事由", Overtime{EmployeeID: "e1", Date: "2026-08-16", Hours: 2}, false},
		{"日期格式错误", Overtime{EmployeeID: "e1", Date: "2026/08/16", Hours: 2, Reason: "赶进度"}, false},
		{"非法状态", Overtime{EmployeeID: "e1", Date: "2026-08-16", Hours: 2, Reason: "赶进度", Status: "xxx"}, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := c.o.Validate()
			if c.ok && err != nil {
				t.Fatalf("应通过校验，实际: %v", err)
			}
			if !c.ok && err == nil {
				t.Fatal("应校验失败")
			}
		})
	}
}

func TestOvertimeTransitions(t *testing.T) {
	cases := []struct {
		from, to string
		ok       bool
	}{
		{OvertimePending, OvertimeApproved, true},
		{OvertimePending, OvertimeRejected, true},
		{OvertimeApproved, OvertimeRejected, false},
	}
	for _, c := range cases {
		if got := CanTransitionOvertime(c.from, c.to); got != c.ok {
			t.Fatalf("流转 %s -> %s 期望 %v，实际 %v", c.from, c.to, c.ok, got)
		}
	}
}

func TestRuleValidate(t *testing.T) {
	cases := []struct {
		name string
		r    AttendanceRule
		ok   bool
	}{
		{"正常", AttendanceRule{Name: "标准", WorkStart: "09:00", WorkEnd: "18:00", LateThreshold: 15}, true},
		{"空名", AttendanceRule{WorkStart: "09:00", WorkEnd: "18:00"}, false},
		{"时间格式错误", AttendanceRule{Name: "标准", WorkStart: "9点", WorkEnd: "18:00"}, false},
		{"下班早于上班", AttendanceRule{Name: "标准", WorkStart: "18:00", WorkEnd: "09:00"}, false},
		{"负阈值", AttendanceRule{Name: "标准", WorkStart: "09:00", WorkEnd: "18:00", LateThreshold: -1}, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := c.r.Validate()
			if c.ok && err != nil {
				t.Fatalf("应通过校验，实际: %v", err)
			}
			if !c.ok && err == nil {
				t.Fatal("应校验失败")
			}
		})
	}
}

func TestValidAttendanceStatus(t *testing.T) {
	for _, s := range []string{AttendanceNormal, AttendanceLate, AttendanceEarlyLeave, AttendanceAbsent} {
		if !ValidAttendanceStatus(s) {
			t.Fatalf("%s 应为合法状态", s)
		}
	}
	if ValidAttendanceStatus("xxx") {
		t.Fatal("xxx 不应是合法状态")
	}
}

func TestValidationError(t *testing.T) {
	err := NewValidationError("field", "message")
	if !IsValidationError(err) {
		t.Fatal("IsValidationError 应返回 true")
	}
	if err.Error() != "field: message" {
		t.Fatalf("错误信息格式错误: %s", err.Error())
	}
}

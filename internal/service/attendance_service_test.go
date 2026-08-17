package service

import (
	"testing"

	"hrattendance/internal/model"
)

func TestCreateAttendanceRequiresEmployee(t *testing.T) {
	s := newTestService()
	_, err := s.CreateAttendance(model.Attendance{EmployeeID: "missing", Date: "2026-08-16"})
	if err == nil {
		t.Fatal("关联不存在的员工应被拒绝")
	}
}

func TestAttendanceStatusNormal(t *testing.T) {
	s := newTestService()
	dept := mustCreateDepartment(t, s, "研发部")
	emp := mustCreateEmployee(t, s, dept.ID, "张三", "E001")

	a, err := s.CreateAttendance(model.Attendance{
		EmployeeID: emp.ID,
		Date:       "2026-08-16",
		CheckIn:    "09:00",
		CheckOut:   "18:00",
	})
	if err != nil {
		t.Fatalf("创建考勤失败: %v", err)
	}
	if a.Status != model.AttendanceNormal {
		t.Fatalf("准点打卡应为 normal，实际 %s", a.Status)
	}
}

func TestAttendanceStatusLate(t *testing.T) {
	s := newTestService()
	dept := mustCreateDepartment(t, s, "研发部")
	emp := mustCreateEmployee(t, s, dept.ID, "张三", "E001")

	a, err := s.CreateAttendance(model.Attendance{
		EmployeeID: emp.ID,
		Date:       "2026-08-16",
		CheckIn:    "09:30",
		CheckOut:   "18:00",
	})
	if err != nil {
		t.Fatalf("创建考勤失败: %v", err)
	}
	if a.Status != model.AttendanceLate {
		t.Fatalf("9:30 签到应为 late，实际 %s", a.Status)
	}
}

func TestAttendanceStatusEarlyLeave(t *testing.T) {
	s := newTestService()
	dept := mustCreateDepartment(t, s, "研发部")
	emp := mustCreateEmployee(t, s, dept.ID, "张三", "E001")

	a, err := s.CreateAttendance(model.Attendance{
		EmployeeID: emp.ID,
		Date:       "2026-08-16",
		CheckIn:    "09:00",
		CheckOut:   "17:00",
	})
	if err != nil {
		t.Fatalf("创建考勤失败: %v", err)
	}
	if a.Status != model.AttendanceEarlyLeave {
		t.Fatalf("17:00 签退应为 early_leave，实际 %s", a.Status)
	}
}

func TestAttendanceStatusAbsent(t *testing.T) {
	s := newTestService()
	dept := mustCreateDepartment(t, s, "研发部")
	emp := mustCreateEmployee(t, s, dept.ID, "张三", "E001")

	a, err := s.CreateAttendance(model.Attendance{
		EmployeeID: emp.ID,
		Date:       "2026-08-16",
	})
	if err != nil {
		t.Fatalf("创建考勤失败: %v", err)
	}
	if a.Status != model.AttendanceAbsent {
		t.Fatalf("无打卡应为 absent，实际 %s", a.Status)
	}
}

func TestCheckInCreatesRecord(t *testing.T) {
	s := newTestService()
	dept := mustCreateDepartment(t, s, "研发部")
	emp := mustCreateEmployee(t, s, dept.ID, "张三", "E001")

	a, err := s.CheckIn(emp.ID, "2026-08-16", "09:00")
	if err != nil {
		t.Fatalf("签到失败: %v", err)
	}
	if a.CheckIn != "09:00" {
		t.Fatalf("签到时间错误: %s", a.CheckIn)
	}
}

func TestCheckOutCompletesRecord(t *testing.T) {
	s := newTestService()
	dept := mustCreateDepartment(t, s, "研发部")
	emp := mustCreateEmployee(t, s, dept.ID, "张三", "E001")

	_, _ = s.CheckIn(emp.ID, "2026-08-16", "09:00")
	a, err := s.CheckOut(emp.ID, "2026-08-16", "18:00")
	if err != nil {
		t.Fatalf("签退失败: %v", err)
	}
	if a.CheckOut != "18:00" || a.Status != model.AttendanceNormal {
		t.Fatalf("签退记录错误: checkOut=%s status=%s", a.CheckOut, a.Status)
	}
}

func TestUpdateAttendanceRecalculatesStatus(t *testing.T) {
	s := newTestService()
	dept := mustCreateDepartment(t, s, "研发部")
	emp := mustCreateEmployee(t, s, dept.ID, "张三", "E001")

	a, _ := s.CreateAttendance(model.Attendance{
		EmployeeID: emp.ID,
		Date:       "2026-08-16",
		CheckIn:    "09:00",
		CheckOut:   "18:00",
	})
	// 补卡改为迟到
	updated, err := s.UpdateAttendance(a.ID, model.Attendance{CheckIn: "10:00"})
	if err != nil {
		t.Fatalf("补卡失败: %v", err)
	}
	if updated.Status != model.AttendanceLate {
		t.Fatalf("补卡后应重算为 late，实际 %s", updated.Status)
	}
}

func TestListAttendancesFilter(t *testing.T) {
	s := newTestService()
	dept := mustCreateDepartment(t, s, "研发部")
	emp := mustCreateEmployee(t, s, dept.ID, "张三", "E001")

	_, _ = s.CreateAttendance(model.Attendance{EmployeeID: emp.ID, Date: "2026-08-16", CheckIn: "09:00", CheckOut: "18:00"})
	_, _ = s.CreateAttendance(model.Attendance{EmployeeID: emp.ID, Date: "2026-08-17", CheckIn: "09:30", CheckOut: "18:00"})

	items, total, _ := s.ListAttendances(model.AttendanceFilter{Status: model.AttendanceLate}, 1, 10)
	if total != 1 {
		t.Fatalf("迟到记录应为 1，实际 %d", total)
	}

	items, total, _ = s.ListAttendances(model.AttendanceFilter{DateFrom: "2026-08-17", DateTo: "2026-08-17"}, 1, 10)
	if total != 1 || items[0].Date != "2026-08-17" {
		t.Fatalf("按日期范围筛选结果错误: total=%d", total)
	}

	items, total, _ = s.ListAttendances(model.AttendanceFilter{DateFrom: "2026-08-16", DateTo: "2026-08-17"}, 1, 10)
	if total != 2 {
		t.Fatalf("日期区间 [2026-08-16, 2026-08-17] 应返回 2 条，实际 %d", total)
	}
	if len(items) > 0 && items[0].Date != "2026-08-16" {
		t.Fatalf("区间起始日期应为 2026-08-16，实际 %s", items[0].Date)
	}
}

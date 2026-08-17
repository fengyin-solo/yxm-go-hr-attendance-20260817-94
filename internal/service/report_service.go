package service

import (
	"sort"
	"strings"

	"hrattendance/internal/model"
)

// EmployeeMonthlyReport 员工月度考勤汇总。
type EmployeeMonthlyReport struct {
	EmployeeID     string  `json:"employee_id"`
	EmployeeName   string  `json:"employee_name"`
	Month          string  `json:"month"`
	NormalDays     int     `json:"normal_days"`
	LateDays       int     `json:"late_days"`
	EarlyLeaveDays int     `json:"early_leave_days"`
	AbsentDays     int     `json:"absent_days"`
	LeaveDays      int     `json:"leave_days"`
	OvertimeHours  float64 `json:"overtime_hours"`
}

// AttendanceStatusStats 考勤状态分布。
type AttendanceStatusStats struct {
	Normal      int `json:"normal"`
	Late        int `json:"late"`
	EarlyLeave  int `json:"early_leave"`
	Absent      int `json:"absent"`
}

// LeaveTypeStats 请假类型统计。
type LeaveTypeStats struct {
	Annual   int `json:"annual"`
	Sick     int `json:"sick"`
	Personal int `json:"personal"`
}

// OvertimeStats 加班统计。
type OvertimeStats struct {
	Count       int     `json:"count"`
	TotalHours  float64 `json:"total_hours"`
}

// EmployeeMonthlyReport 生成指定员工在指定月份的考勤汇总。
func (s *Service) EmployeeMonthlyReport(employeeID, month string) (*EmployeeMonthlyReport, error) {
	employee, err := s.store.GetEmployee(employeeID)
	if err != nil {
		return nil, err
	}
	report := &EmployeeMonthlyReport{
		EmployeeID:   employeeID,
		EmployeeName: employee.Name,
		Month:        month,
	}

	prefix := month + "-"
	for _, a := range s.store.ListAttendances() {
		if a.EmployeeID != employeeID || !strings.HasPrefix(a.Date, prefix) {
			continue
		}
		switch a.Status {
		case model.AttendanceNormal:
			report.NormalDays++
		case model.AttendanceLate:
			report.LateDays++
		case model.AttendanceEarlyLeave:
			report.EarlyLeaveDays++
		case model.AttendanceAbsent:
			report.AbsentDays++
		}
	}

	for _, l := range s.store.ListLeaves() {
		if l.EmployeeID != employeeID || l.Status != model.LeaveApproved {
			continue
		}
		if !strings.HasPrefix(l.StartDate, prefix) && !strings.HasPrefix(l.EndDate, prefix) {
			continue
		}
		report.LeaveDays += l.Days()
	}

	for _, o := range s.store.ListOvertimes() {
		if o.EmployeeID != employeeID || o.Status != model.OvertimeApproved {
			continue
		}
		if strings.HasPrefix(o.Date, prefix) {
			report.OvertimeHours += o.Hours
		}
	}
	return report, nil
}

// DepartmentReport 汇总指定部门所有在职员工的月度考勤数据。
type DepartmentReport struct {
	DepartmentID   string                   `json:"department_id"`
	DepartmentName string                   `json:"department_name"`
	Month          string                   `json:"month"`
	EmployeeCount  int                      `json:"employee_count"`
	Reports        []*EmployeeMonthlyReport `json:"reports"`
}

// DepartmentMonthlyReport 生成部门月度汇总报告。
func (s *Service) DepartmentMonthlyReport(departmentID, month string) (*DepartmentReport, error) {
	department, err := s.store.GetDepartment(departmentID)
	if err != nil {
		return nil, err
	}
	report := &DepartmentReport{
		DepartmentID:   departmentID,
		DepartmentName: department.Name,
		Month:          month,
		Reports:        make([]*EmployeeMonthlyReport, 0),
	}
	for _, e := range s.store.ListEmployees() {
		if e.DepartmentID != departmentID || e.Status != model.EmployeeActive {
			continue
		}
		report.EmployeeCount++
		er, err := s.EmployeeMonthlyReport(e.ID, month)
		if err != nil {
			continue
		}
		report.Reports = append(report.Reports, er)
	}
	return report, nil
}

// AttendanceStatusStats 统计全部考勤记录的状态分布。
func (s *Service) AttendanceStatusStats() *AttendanceStatusStats {
	stats := &AttendanceStatusStats{}
	for _, a := range s.store.ListAttendances() {
		switch a.Status {
		case model.AttendanceNormal:
			stats.Normal++
		case model.AttendanceLate:
			stats.Late++
		case model.AttendanceEarlyLeave:
			stats.EarlyLeave++
		case model.AttendanceAbsent:
			stats.Absent++
		}
	}
	return stats
}

// LeaveTypeStats 统计已审批通过的请假天数，按类型分组。
func (s *Service) LeaveTypeStats() *LeaveTypeStats {
	stats := &LeaveTypeStats{}
	for _, l := range s.store.ListLeaves() {
		if l.Status != model.LeaveApproved {
			continue
		}
		switch l.Type {
		case model.LeaveAnnual:
			stats.Annual += l.Days()
		case model.LeaveSick:
			stats.Sick += l.Days()
		case model.LeavePersonal:
			stats.Personal += l.Days()
		}
	}
	return stats
}

// OvertimeStats 统计已审批通过的加班单数量与总时长。
func (s *Service) OvertimeStats() *OvertimeStats {
	stats := &OvertimeStats{}
	for _, o := range s.store.ListOvertimes() {
		if o.Status != model.OvertimeApproved {
			continue
		}
		stats.Count++
		stats.TotalHours += o.Hours
	}
	return stats
}

// CompanyReport 公司级汇总报告。
type CompanyReport struct {
	EmployeeCount    int                    `json:"employee_count"`
	DepartmentCount  int                    `json:"department_count"`
	AttendanceStats  *AttendanceStatusStats `json:"attendance_stats"`
	LeaveStats       *LeaveTypeStats        `json:"leave_stats"`
	OvertimeStats    *OvertimeStats         `json:"overtime_stats"`
}

// ExportCompanyReport 导出公司级全量汇总报告。
func (s *Service) ExportCompanyReport() *CompanyReport {
	report := &CompanyReport{}
	for _, e := range s.store.ListEmployees() {
		if e.Status == model.EmployeeActive {
			report.EmployeeCount++
		}
	}
	for _, d := range s.store.ListDepartments() {
		if d.Status == model.DepartmentActive {
			report.DepartmentCount++
		}
	}
	report.AttendanceStats = s.AttendanceStatusStats()
	report.LeaveStats = s.LeaveTypeStats()
	report.OvertimeStats = s.OvertimeStats()
	return report
}

// LateRankEntry 迟到排行条目。
type LateRankEntry struct {
	EmployeeID   string `json:"employee_id"`
	EmployeeName string `json:"employee_name"`
	LateDays     int    `json:"late_days"`
}

// TopLateEmployees 返回指定月份迟到次数最多的员工列表（降序）。
func (s *Service) TopLateEmployees(month string, limit int) []*LateRankEntry {
	if limit <= 0 {
		limit = 10
	}
	prefix := month + "-"
	lateCount := make(map[string]int)
	for _, a := range s.store.ListAttendances() {
		if a.Status != model.AttendanceLate || !strings.HasPrefix(a.Date, prefix) {
			continue
		}
		lateCount[a.EmployeeID]++
	}
	entries := make([]*LateRankEntry, 0, len(lateCount))
	for empID, days := range lateCount {
		name := empID
		if emp, err := s.store.GetEmployee(empID); err == nil {
			name = emp.Name
		}
		entries = append(entries, &LateRankEntry{EmployeeID: empID, EmployeeName: name, LateDays: days})
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].LateDays > entries[j].LateDays
	})
	if len(entries) > limit {
		entries = entries[:limit]
	}
	return entries
}

// PerfectAttendanceEmployees 返回指定月份全勤（无迟到/早退/缺勤）的在职员工。
func (s *Service) PerfectAttendanceEmployees(month string) []*model.Employee {
	prefix := month + "-"
	badEmployees := make(map[string]bool)
	hasRecord := make(map[string]bool)
	for _, a := range s.store.ListAttendances() {
		if !strings.HasPrefix(a.Date, prefix) {
			continue
		}
		hasRecord[a.EmployeeID] = true
		if a.Status != model.AttendanceNormal {
			badEmployees[a.EmployeeID] = true
		}
	}
	result := make([]*model.Employee, 0)
	for _, e := range s.store.ListEmployees() {
		if e.Status != model.EmployeeActive {
			continue
		}
		if !hasRecord[e.ID] || badEmployees[e.ID] {
			continue
		}
		result = append(result, e)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].EmpNo < result[j].EmpNo
	})
	return result
}

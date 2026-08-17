package handler

import (
	"net/http"
	"strconv"

	"hrattendance/pkg/httpx"
)

func (s *Server) registerReportRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/reports/attendance-status", s.attendanceStatusStats)
	mux.HandleFunc("GET /api/reports/leave-types", s.leaveTypeStats)
	mux.HandleFunc("GET /api/reports/overtime", s.overtimeStats)
	mux.HandleFunc("GET /api/reports/company", s.companyReport)
	mux.HandleFunc("GET /api/reports/top-late", s.topLateEmployees)
	mux.HandleFunc("GET /api/reports/perfect-attendance", s.perfectAttendanceEmployees)
	mux.HandleFunc("GET /api/employees/{id}/monthly-report", s.employeeMonthlyReport)
	mux.HandleFunc("GET /api/departments/{id}/monthly-report", s.departmentMonthlyReport)
}

func (s *Server) companyReport(w http.ResponseWriter, r *http.Request) {
	httpx.OK(w, s.svc.ExportCompanyReport())
}

func (s *Server) topLateEmployees(w http.ResponseWriter, r *http.Request) {
	month := r.URL.Query().Get("month")
	if month == "" {
		httpx.BadRequest(w, "缺少 month 参数，格式 YYYY-MM")
		return
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	httpx.OK(w, s.svc.TopLateEmployees(month, limit))
}

func (s *Server) perfectAttendanceEmployees(w http.ResponseWriter, r *http.Request) {
	month := r.URL.Query().Get("month")
	if month == "" {
		httpx.BadRequest(w, "缺少 month 参数，格式 YYYY-MM")
		return
	}
	httpx.OK(w, s.svc.PerfectAttendanceEmployees(month))
}

func (s *Server) attendanceStatusStats(w http.ResponseWriter, r *http.Request) {
	httpx.OK(w, s.svc.AttendanceStatusStats())
}

func (s *Server) leaveTypeStats(w http.ResponseWriter, r *http.Request) {
	httpx.OK(w, s.svc.LeaveTypeStats())
}

func (s *Server) overtimeStats(w http.ResponseWriter, r *http.Request) {
	httpx.OK(w, s.svc.OvertimeStats())
}

func (s *Server) employeeMonthlyReport(w http.ResponseWriter, r *http.Request) {
	month := r.URL.Query().Get("month")
	if month == "" {
		httpx.BadRequest(w, "缺少 month 参数，格式 YYYY-MM")
		return
	}
	report, err := s.svc.EmployeeMonthlyReport(r.PathValue("id"), month)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, report)
}

func (s *Server) departmentMonthlyReport(w http.ResponseWriter, r *http.Request) {
	month := r.URL.Query().Get("month")
	if month == "" {
		httpx.BadRequest(w, "缺少 month 参数，格式 YYYY-MM")
		return
	}
	report, err := s.svc.DepartmentMonthlyReport(r.PathValue("id"), month)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, report)
}

package handler

import (
	"net/http"

	"hrattendance/internal/model"
	"hrattendance/pkg/httpx"
)

func (s *Server) registerOvertimeRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/overtimes", s.createOvertime)
	mux.HandleFunc("GET /api/overtimes", s.listOvertimes)
	mux.HandleFunc("GET /api/overtimes/{id}", s.getOvertime)
	mux.HandleFunc("DELETE /api/overtimes/{id}", s.deleteOvertime)
	mux.HandleFunc("POST /api/overtimes/{id}/approve", s.approveOvertime)
	mux.HandleFunc("POST /api/overtimes/{id}/reject", s.rejectOvertime)
	mux.HandleFunc("POST /api/overtimes/batch-approve", s.batchApproveOvertimes)
}

type overtimeRequest struct {
	EmployeeID string  `json:"employee_id"`
	Date       string  `json:"date"`
	Hours      float64 `json:"hours"`
	Reason     string  `json:"reason"`
	ApproverID string  `json:"approver_id"`
}

func (s *Server) createOvertime(w http.ResponseWriter, r *http.Request) {
	var req overtimeRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	o, err := s.svc.CreateOvertime(model.Overtime{
		EmployeeID: req.EmployeeID,
		Date:       req.Date,
		Hours:      req.Hours,
		Reason:     req.Reason,
		ApproverID: req.ApproverID,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.Created(w, o)
}

func (s *Server) listOvertimes(w http.ResponseWriter, r *http.Request) {
	pp := httpx.ParsePagination(r, 20, s.maxPageSize())
	filter := model.OvertimeFilter{
		EmployeeID: r.URL.Query().Get("employee_id"),
		Status:     r.URL.Query().Get("status"),
		DateFrom:   r.URL.Query().Get("from"),
		DateTo:     r.URL.Query().Get("to"),
	}
	items, total, err := s.svc.ListOvertimes(filter, pp.Page, pp.Size)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, httpx.PageResult{
		Items:      items,
		Pagination: httpx.Pagination{Page: pp.Page, Size: pp.Size, Total: total},
	})
}

func (s *Server) getOvertime(w http.ResponseWriter, r *http.Request) {
	o, err := s.svc.GetOvertime(r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, o)
}

func (s *Server) deleteOvertime(w http.ResponseWriter, r *http.Request) {
	if err := s.svc.DeleteOvertime(r.PathValue("id")); err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.NoContent(w)
}

func (s *Server) approveOvertime(w http.ResponseWriter, r *http.Request) {
	o, err := s.svc.ApproveOvertime(r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, o)
}

func (s *Server) rejectOvertime(w http.ResponseWriter, r *http.Request) {
	o, err := s.svc.ApproveOvertime(r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, o)
}

func (s *Server) batchApproveOvertimes(w http.ResponseWriter, r *http.Request) {
	var req batchApproveRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	httpx.OK(w, s.svc.BatchApproveOvertimes(req.IDs))
}

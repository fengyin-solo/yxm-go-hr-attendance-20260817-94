package handler

import (
	"net/http"

	"hrattendance/internal/model"
	"hrattendance/pkg/httpx"
)

func (s *Server) registerLeaveRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/leaves", s.createLeave)
	mux.HandleFunc("GET /api/leaves", s.listLeaves)
	mux.HandleFunc("GET /api/leaves/{id}", s.getLeave)
	mux.HandleFunc("DELETE /api/leaves/{id}", s.deleteLeave)
	mux.HandleFunc("POST /api/leaves/{id}/approve", s.approveLeave)
	mux.HandleFunc("POST /api/leaves/{id}/reject", s.rejectLeave)
	mux.HandleFunc("POST /api/leaves/batch-approve", s.batchApproveLeaves)
}

type leaveRequest struct {
	EmployeeID string `json:"employee_id"`
	Type       string `json:"type"`
	StartDate  string `json:"start_date"`
	EndDate    string `json:"end_date"`
	Reason     string `json:"reason"`
	ApproverID string `json:"approver_id"`
}

func (s *Server) createLeave(w http.ResponseWriter, r *http.Request) {
	var req leaveRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	l, err := s.svc.CreateLeave(model.Leave{
		EmployeeID: req.EmployeeID,
		Type:       req.Type,
		StartDate:  req.StartDate,
		EndDate:    req.EndDate,
		Reason:     req.Reason,
		ApproverID: req.ApproverID,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.Created(w, l)
}

func (s *Server) listLeaves(w http.ResponseWriter, r *http.Request) {
	pp := httpx.ParsePagination(r, 20, s.maxPageSize())
	filter := model.LeaveFilter{
		EmployeeID: r.URL.Query().Get("employee_id"),
		Type:       r.URL.Query().Get("type"),
		Status:     r.URL.Query().Get("status"),
		DateFrom:   r.URL.Query().Get("from"),
		DateTo:     r.URL.Query().Get("to"),
	}
	items, total, err := s.svc.ListLeaves(filter, pp.Page, pp.Size)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, httpx.PageResult{
		Items:      items,
		Pagination: httpx.Pagination{Page: pp.Page, Size: pp.Size, Total: total},
	})
}

func (s *Server) getLeave(w http.ResponseWriter, r *http.Request) {
	l, err := s.svc.GetLeave(r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, l)
}

func (s *Server) deleteLeave(w http.ResponseWriter, r *http.Request) {
	if err := s.svc.DeleteLeave(r.PathValue("id")); err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.NoContent(w)
}

func (s *Server) approveLeave(w http.ResponseWriter, r *http.Request) {
	l, err := s.svc.ApproveLeave(r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, l)
}

func (s *Server) rejectLeave(w http.ResponseWriter, r *http.Request) {
	l, err := s.svc.RejectLeave(r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, l)
}

type batchApproveRequest struct {
	IDs []string `json:"ids"`
}

func (s *Server) batchApproveLeaves(w http.ResponseWriter, r *http.Request) {
	var req batchApproveRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	httpx.OK(w, s.svc.BatchApproveLeaves(req.IDs))
}

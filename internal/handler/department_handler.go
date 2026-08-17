package handler

import (
	"net/http"

	"hrattendance/internal/model"
	"hrattendance/pkg/httpx"
)

func (s *Server) registerDepartmentRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/departments", s.createDepartment)
	mux.HandleFunc("GET /api/departments", s.listDepartments)
	mux.HandleFunc("GET /api/departments/{id}", s.getDepartment)
	mux.HandleFunc("PUT /api/departments/{id}", s.updateDepartment)
	mux.HandleFunc("DELETE /api/departments/{id}", s.deleteDepartment)
	mux.HandleFunc("GET /api/departments/{id}/employee-count", s.departmentEmployeeCount)
}

type departmentRequest struct {
	Name     string `json:"name"`
	ParentID string `json:"parent_id"`
	Status   string `json:"status"`
}

func (s *Server) createDepartment(w http.ResponseWriter, r *http.Request) {
	var req departmentRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	d, err := s.svc.CreateDepartment(model.Department{
		Name:     req.Name,
		ParentID: req.ParentID,
		Status:   req.Status,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.Created(w, d)
}

func (s *Server) listDepartments(w http.ResponseWriter, r *http.Request) {
	pp := httpx.ParsePagination(r, 20, s.maxPageSize())
	filter := model.DepartmentFilter{
		ParentID: r.URL.Query().Get("parent_id"),
		Status:   r.URL.Query().Get("status"),
		Keyword:  r.URL.Query().Get("keyword"),
	}
	items, total, err := s.svc.ListDepartments(filter, pp.Page, pp.Size)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, httpx.PageResult{
		Items:      items,
		Pagination: httpx.Pagination{Page: pp.Page, Size: pp.Size, Total: total},
	})
}

func (s *Server) getDepartment(w http.ResponseWriter, r *http.Request) {
	d, err := s.svc.GetDepartment(r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, d)
}

func (s *Server) updateDepartment(w http.ResponseWriter, r *http.Request) {
	var req departmentRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	d, err := s.svc.UpdateDepartment(r.PathValue("id"), model.Department{
		Name:     req.Name,
		ParentID: req.ParentID,
		Status:   req.Status,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, d)
}

func (s *Server) deleteDepartment(w http.ResponseWriter, r *http.Request) {
	if err := s.svc.DeleteDepartment(r.PathValue("id")); err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.NoContent(w)
}

func (s *Server) departmentEmployeeCount(w http.ResponseWriter, r *http.Request) {
	count, err := s.svc.DepartmentEmployeeCount(r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, map[string]int{"count": count})
}

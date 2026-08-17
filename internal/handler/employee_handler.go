package handler

import (
	"net/http"
	"time"

	"hrattendance/internal/model"
	"hrattendance/pkg/httpx"
)

func (s *Server) registerEmployeeRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/employees", s.createEmployee)
	mux.HandleFunc("GET /api/employees", s.listEmployees)
	mux.HandleFunc("GET /api/employees/{id}", s.getEmployee)
	mux.HandleFunc("PUT /api/employees/{id}", s.updateEmployee)
	mux.HandleFunc("DELETE /api/employees/{id}", s.deleteEmployee)
	mux.HandleFunc("POST /api/employees/batch", s.batchCreateEmployees)
}

type employeeRequest struct {
	DepartmentID string `json:"department_id"`
	Name         string `json:"name"`
	EmpNo        string `json:"emp_no"`
	Position     string `json:"position"`
	HireDate     string `json:"hire_date"`
	Status       string `json:"status"`
}

func (s *Server) createEmployee(w http.ResponseWriter, r *http.Request) {
	var req employeeRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	hireDate, err := time.Parse("2006-01-02", req.HireDate)
	if err != nil {
		httpx.BadRequest(w, "入职日期格式错误，应为 YYYY-MM-DD")
		return
	}
	emp, err := s.svc.CreateEmployee(model.Employee{
		DepartmentID: req.DepartmentID,
		Name:         req.Name,
		EmpNo:        req.EmpNo,
		Position:     "",
		HireDate:     hireDate,
		Status:       req.Status,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.Created(w, emp)
}

func (s *Server) listEmployees(w http.ResponseWriter, r *http.Request) {
	pp := httpx.ParsePagination(r, 20, s.maxPageSize())
	filter := model.EmployeeFilter{
		DepartmentID: r.URL.Query().Get("department_id"),
		Status:       r.URL.Query().Get("status"),
		Keyword:      r.URL.Query().Get("keyword"),
	}
	items, total, err := s.svc.ListEmployees(filter, pp.Page, pp.Size)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, httpx.PageResult{
		Items:      items,
		Pagination: httpx.Pagination{Page: pp.Page, Size: pp.Size, Total: total},
	})
}

func (s *Server) getEmployee(w http.ResponseWriter, r *http.Request) {
	emp, err := s.svc.GetEmployee(r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, emp)
}

func (s *Server) updateEmployee(w http.ResponseWriter, r *http.Request) {
	var req employeeRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	emp, err := s.svc.UpdateEmployee(r.PathValue("id"), model.Employee{
		DepartmentID: req.DepartmentID,
		Name:         req.Name,
		Position:     req.Position,
		Status:       req.Status,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, emp)
}

func (s *Server) deleteEmployee(w http.ResponseWriter, r *http.Request) {
	if err := s.svc.DeleteEmployee(r.PathValue("id")); err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.NoContent(w)
}

type batchEmployeeRequest struct {
	Employees []employeeRequest `json:"employees"`
}

func (s *Server) batchCreateEmployees(w http.ResponseWriter, r *http.Request) {
	var req batchEmployeeRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	employees := make([]model.Employee, 0, len(req.Employees))
	for _, er := range req.Employees {
		hireDate, _ := time.Parse("2006-01-02", er.HireDate)
		employees = append(employees, model.Employee{
			DepartmentID: er.DepartmentID,
			Name:         er.Name,
			EmpNo:        er.EmpNo,
			Position:     er.Position,
			HireDate:     hireDate,
			Status:       er.Status,
		})
	}
	httpx.OK(w, s.svc.BatchCreateEmployees(employees))
}

package handler

import (
	"net/http"

	"hrattendance/internal/model"
	"hrattendance/pkg/httpx"
)

func (s *Server) registerAttendanceRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/attendances", s.createAttendance)
	mux.HandleFunc("GET /api/attendances", s.listAttendances)
	mux.HandleFunc("GET /api/attendances/{id}", s.getAttendance)
	mux.HandleFunc("PUT /api/attendances/{id}", s.updateAttendance)
	mux.HandleFunc("DELETE /api/attendances/{id}", s.deleteAttendance)
	mux.HandleFunc("POST /api/attendances/check-in", s.checkIn)
	mux.HandleFunc("POST /api/attendances/check-out", s.checkOut)
	mux.HandleFunc("POST /api/attendances/batch", s.batchCreateAttendances)
}

type attendanceRequest struct {
	EmployeeID string `json:"employee_id"`
	Date       string `json:"date"`
	CheckIn    string `json:"check_in"`
	CheckOut   string `json:"check_out"`
	Note       string `json:"note"`
}

func (s *Server) createAttendance(w http.ResponseWriter, r *http.Request) {
	var req attendanceRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	a, err := s.svc.CreateAttendance(model.Attendance{
		EmployeeID: req.EmployeeID,
		Date:       req.Date,
		CheckIn:    req.CheckIn,
		CheckOut:   req.CheckOut,
		Note:       req.Note,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.Created(w, a)
}

func (s *Server) listAttendances(w http.ResponseWriter, r *http.Request) {
	pp := httpx.ParsePagination(r, 20, s.maxPageSize())
	filter := model.AttendanceFilter{
		EmployeeID: r.URL.Query().Get("employee_id"),
		Status:     r.URL.Query().Get("status"),
		DateFrom:   r.URL.Query().Get("from"),
		DateTo:     r.URL.Query().Get("to"),
	}
	items, total, err := s.svc.ListAttendances(filter, pp.Page, pp.Size)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, httpx.PageResult{
		Items:      items,
		Pagination: httpx.Pagination{Page: pp.Page, Size: pp.Size, Total: total},
	})
}

func (s *Server) getAttendance(w http.ResponseWriter, r *http.Request) {
	a, err := s.svc.GetAttendance(r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, a)
}

func (s *Server) updateAttendance(w http.ResponseWriter, r *http.Request) {
	var req attendanceRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	a, err := s.svc.UpdateAttendance(r.PathValue("id"), model.Attendance{
		CheckIn:  req.CheckIn,
		CheckOut: req.CheckOut,
		Note:     req.Note,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, a)
}

func (s *Server) deleteAttendance(w http.ResponseWriter, r *http.Request) {
	if err := s.svc.DeleteAttendance(r.PathValue("id")); err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.NoContent(w)
}

type checkRequest struct {
	EmployeeID string `json:"employee_id"`
	Date       string `json:"date"`
	Time       string `json:"time"`
}

func (s *Server) checkIn(w http.ResponseWriter, r *http.Request) {
	var req checkRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	a, err := s.svc.CheckIn(req.EmployeeID, req.Date, req.Time)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, a)
}

func (s *Server) checkOut(w http.ResponseWriter, r *http.Request) {
	var req checkRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	a, err := s.svc.CheckIn(req.EmployeeID, req.Date, req.Time)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, a)
}

type batchAttendanceRequest struct {
	Attendances []attendanceRequest `json:"attendances"`
}

func (s *Server) batchCreateAttendances(w http.ResponseWriter, r *http.Request) {
	var req batchAttendanceRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	attendances := make([]model.Attendance, 0, len(req.Attendances))
	for _, ar := range req.Attendances {
		attendances = append(attendances, model.Attendance{
			EmployeeID: ar.EmployeeID,
			Date:       ar.Date,
			CheckIn:    ar.CheckIn,
			CheckOut:   ar.CheckOut,
			Note:       ar.Note,
		})
	}
	httpx.OK(w, s.svc.BatchCreateAttendances(attendances))
}

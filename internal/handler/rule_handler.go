package handler

import (
	"net/http"

	"hrattendance/internal/model"
	"hrattendance/pkg/httpx"
)

func (s *Server) registerRuleRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/rules", s.createRule)
	mux.HandleFunc("GET /api/rules", s.listRules)
	mux.HandleFunc("GET /api/rules/{id}", s.getRule)
	mux.HandleFunc("PUT /api/rules/{id}", s.updateRule)
	mux.HandleFunc("DELETE /api/rules/{id}", s.deleteRule)
}

type ruleRequest struct {
	Name          string `json:"name"`
	WorkStart     string `json:"work_start"`
	WorkEnd       string `json:"work_end"`
	LateThreshold int    `json:"late_threshold"`
	Enabled       bool   `json:"enabled"`
}

func (s *Server) createRule(w http.ResponseWriter, r *http.Request) {
	var req ruleRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	rule, err := s.svc.CreateRule(model.AttendanceRule{
		Name:          req.Name,
		WorkStart:     req.WorkStart,
		WorkEnd:       req.WorkEnd,
		LateThreshold: req.LateThreshold,
		Enabled:       req.Enabled,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.Created(w, rule)
}

func (s *Server) listRules(w http.ResponseWriter, r *http.Request) {
	rules, err := s.svc.ListRules()
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, rules)
}

func (s *Server) getRule(w http.ResponseWriter, r *http.Request) {
	rule, err := s.svc.GetRule(r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, rule)
}

func (s *Server) updateRule(w http.ResponseWriter, r *http.Request) {
	var req ruleRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	rule, err := s.svc.UpdateRule(r.PathValue("id"), model.AttendanceRule{
		Name:          req.Name,
		WorkStart:     req.WorkStart,
		WorkEnd:       req.WorkEnd,
		LateThreshold: req.LateThreshold,
		Enabled:       req.Enabled,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, rule)
}

func (s *Server) deleteRule(w http.ResponseWriter, r *http.Request) {
	if err := s.svc.DeleteRule(r.PathValue("id")); err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.NoContent(w)
}

package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"hrattendance/internal/config"
	"hrattendance/internal/service"
	"hrattendance/internal/store"
	"hrattendance/pkg/logger"
)

func newTestServer() *Server {
	cfg := &config.Config{
		MaxPageSize:      100,
		RateLimit:        1000,
		RateWindowSec:    60,
		DefaultWorkStart: "09:00",
		DefaultWorkEnd:   "18:00",
		LateThreshold:    15,
	}
	log := logger.NewLevel(logger.LevelError)
	st := store.NewMemoryStore()
	svc := service.New(st, log, cfg)
	return NewServer(svc, log, cfg)
}

func doRequest(t *testing.T, s *Server, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, path, reader)
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	rec := httptest.NewRecorder()
	s.Routes().ServeHTTP(rec, req)
	return rec
}

func decodeBody(t *testing.T, rec *httptest.ResponseRecorder, dst interface{}) {
	t.Helper()
	if err := json.Unmarshal(rec.Body.Bytes(), dst); err != nil {
		t.Fatalf("解析响应失败: %v, body=%s", err, rec.Body.String())
	}
}

type apiResponse struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func mustGetCode(t *testing.T, rec *httptest.ResponseRecorder) int {
	t.Helper()
	var resp apiResponse
	decodeBody(t, rec, &resp)
	return resp.Code
}

func createDeptViaAPI(t *testing.T, s *Server, name string) string {
	t.Helper()
	rec := doRequest(t, s, "POST", "/api/departments", `{"name":"`+name+`"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("创建部门失败: %d %s", rec.Code, rec.Body.String())
	}
	var resp struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	decodeBody(t, rec, &resp)
	return resp.Data.ID
}

func createEmpViaAPI(t *testing.T, s *Server, deptID, name, empNo string) string {
	t.Helper()
	body := `{"department_id":"` + deptID + `","name":"` + name + `","emp_no":"` + empNo + `","hire_date":"2026-01-01"}`
	rec := doRequest(t, s, "POST", "/api/employees", body)
	if rec.Code != http.StatusCreated {
		t.Fatalf("创建员工失败: %d %s", rec.Code, rec.Body.String())
	}
	var resp struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	decodeBody(t, rec, &resp)
	return resp.Data.ID
}

func TestCreateDepartmentAPI(t *testing.T) {
	s := newTestServer()
	rec := doRequest(t, s, "POST", "/api/departments", `{"name":"研发部"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("创建部门应返回 201，实际 %d", rec.Code)
	}
	if code := mustGetCode(t, rec); code != 0 {
		t.Fatalf("成功响应 code 应为 0，实际 %d", code)
	}
}

func TestCreateDepartmentAPIValidation(t *testing.T) {
	s := newTestServer()
	rec := doRequest(t, s, "POST", "/api/departments", `{"name":""}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("空名称应返回 400，实际 %d", rec.Code)
	}
}

func TestGetDepartmentAPINotFound(t *testing.T) {
	s := newTestServer()
	rec := doRequest(t, s, "GET", "/api/departments/missing", "")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("不存在部门应返回 404，实际 %d", rec.Code)
	}
}

func TestListDepartmentsAPI(t *testing.T) {
	s := newTestServer()
	createDeptViaAPI(t, s, "研发部")
	createDeptViaAPI(t, s, "市场部")

	rec := doRequest(t, s, "GET", "/api/departments", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("列出部门应返回 200，实际 %d", rec.Code)
	}
	var resp struct {
		Data struct {
			Items []map[string]interface{} `json:"items"`
		} `json:"data"`
	}
	decodeBody(t, rec, &resp)
	if len(resp.Data.Items) != 2 {
		t.Fatalf("部门数量应为 2，实际 %d", len(resp.Data.Items))
	}
}

func TestDeleteDepartmentAPI(t *testing.T) {
	s := newTestServer()
	id := createDeptViaAPI(t, s, "研发部")

	rec := doRequest(t, s, "DELETE", "/api/departments/"+id, "")
	if rec.Code != http.StatusNoContent {
		t.Fatalf("删除部门应返回 204，实际 %d", rec.Code)
	}
}

func TestCreateEmployeeAPI(t *testing.T) {
	s := newTestServer()
	deptID := createDeptViaAPI(t, s, "研发部")
	body := `{"department_id":"` + deptID + `","name":"张三","emp_no":"E001","hire_date":"2026-01-01"}`
	rec := doRequest(t, s, "POST", "/api/employees", body)
	if rec.Code != http.StatusCreated {
		t.Fatalf("创建员工应返回 201，实际 %d: %s", rec.Code, rec.Body.String())
	}
}

func TestCreateEmployeeAPIBadHireDate(t *testing.T) {
	s := newTestServer()
	deptID := createDeptViaAPI(t, s, "研发部")
	body := `{"department_id":"` + deptID + `","name":"张三","emp_no":"E001","hire_date":"2026/01/01"}`
	rec := doRequest(t, s, "POST", "/api/employees", body)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("非法入职日期应返回 400，实际 %d", rec.Code)
	}
}

func TestBatchCreateEmployeesAPI(t *testing.T) {
	s := newTestServer()
	deptID := createDeptViaAPI(t, s, "研发部")
	body := `{"employees":[` +
		`{"department_id":"` + deptID + `","name":"张三","emp_no":"E001","hire_date":"2026-01-01"},` +
		`{"department_id":"` + deptID + `","name":"李四","emp_no":"E002","hire_date":"2026-01-01"}]}`
	rec := doRequest(t, s, "POST", "/api/employees/batch", body)
	if rec.Code != http.StatusOK {
		t.Fatalf("批量导入应返回 200，实际 %d", rec.Code)
	}
}

func TestEmployeeDuplicateEmpNoAPI(t *testing.T) {
	s := newTestServer()
	deptID := createDeptViaAPI(t, s, "研发部")
	createEmpViaAPI(t, s, deptID, "张三", "E001")

	body := `{"department_id":"` + deptID + `","name":"李四","emp_no":"E001","hire_date":"2026-01-01"}`
	rec := doRequest(t, s, "POST", "/api/employees", body)
	if rec.Code != http.StatusConflict {
		t.Fatalf("重复工号应返回 409，实际 %d", rec.Code)
	}
}

func TestCreateRuleAPI(t *testing.T) {
	s := newTestServer()
	body := `{"name":"标准","work_start":"09:00","work_end":"18:00","late_threshold":15,"enabled":true}`
	rec := doRequest(t, s, "POST", "/api/rules", body)
	if rec.Code != http.StatusCreated {
		t.Fatalf("创建规则应返回 201，实际 %d", rec.Code)
	}
}

func TestListRulesAPI(t *testing.T) {
	s := newTestServer()
	doRequest(t, s, "POST", "/api/rules", `{"name":"标准","work_start":"09:00","work_end":"18:00","enabled":true}`)
	rec := doRequest(t, s, "GET", "/api/rules", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("列出规则应返回 200，实际 %d", rec.Code)
	}
}

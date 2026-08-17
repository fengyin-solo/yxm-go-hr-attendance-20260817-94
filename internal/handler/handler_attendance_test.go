package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"hrattendance/internal/config"
	"hrattendance/internal/service"
	"hrattendance/internal/store"
	"hrattendance/pkg/logger"
)

func TestCreateAttendanceAPI(t *testing.T) {
	s := newTestServer()
	deptID := createDeptViaAPI(t, s, "研发部")
	empID := createEmpViaAPI(t, s, deptID, "张三", "E001")

	body := `{"employee_id":"` + empID + `","date":"2026-08-16","check_in":"09:00","check_out":"18:00"}`
	rec := doRequest(t, s, "POST", "/api/attendances", body)
	if rec.Code != http.StatusCreated {
		t.Fatalf("创建考勤应返回 201，实际 %d: %s", rec.Code, rec.Body.String())
	}
	var resp struct {
		Data struct {
			Status string `json:"status"`
		} `json:"data"`
	}
	decodeBody(t, rec, &resp)
	if resp.Data.Status != "normal" {
		t.Fatalf("准点打卡状态应为 normal，实际 %s", resp.Data.Status)
	}
}

func TestCheckInAPI(t *testing.T) {
	s := newTestServer()
	deptID := createDeptViaAPI(t, s, "研发部")
	empID := createEmpViaAPI(t, s, deptID, "张三", "E001")

	body := `{"employee_id":"` + empID + `","date":"2026-08-16","time":"09:30"}`
	rec := doRequest(t, s, "POST", "/api/attendances/check-in", body)
	if rec.Code != http.StatusOK {
		t.Fatalf("签到应返回 200，实际 %d", rec.Code)
	}
	var resp struct {
		Data struct {
			Status string `json:"status"`
		} `json:"data"`
	}
	decodeBody(t, rec, &resp)
	if resp.Data.Status != "late" {
		t.Fatalf("9:30 签到应为 late，实际 %s", resp.Data.Status)
	}
}

func TestCheckOutAPI(t *testing.T) {
	s := newTestServer()
	deptID := createDeptViaAPI(t, s, "研发部")
	empID := createEmpViaAPI(t, s, deptID, "张三", "E001")
	doRequest(t, s, "POST", "/api/attendances/check-in", `{"employee_id":"`+empID+`","date":"2026-08-16","time":"09:00"}`)

	body := `{"employee_id":"` + empID + `","date":"2026-08-16","time":"17:00"}`
	rec := doRequest(t, s, "POST", "/api/attendances/check-out", body)
	var resp struct {
		Data struct {
			Status string `json:"status"`
		} `json:"data"`
	}
	decodeBody(t, rec, &resp)
	if resp.Data.Status != "early_leave" {
		t.Fatalf("17:00 签退应为 early_leave，实际 %s", resp.Data.Status)
	}
}

func TestBatchCreateAttendancesAPI(t *testing.T) {
	s := newTestServer()
	deptID := createDeptViaAPI(t, s, "研发部")
	empID := createEmpViaAPI(t, s, deptID, "张三", "E001")

	body := `{"attendances":[` +
		`{"employee_id":"` + empID + `","date":"2026-08-16","check_in":"09:00","check_out":"18:00"},` +
		`{"employee_id":"` + empID + `","date":"2026-08-17","check_in":"09:00","check_out":"18:00"}]}`
	rec := doRequest(t, s, "POST", "/api/attendances/batch", body)
	if rec.Code != http.StatusOK {
		t.Fatalf("批量导入考勤应返回 200，实际 %d", rec.Code)
	}
}

func TestListAttendancesAPI(t *testing.T) {
	s := newTestServer()
	deptID := createDeptViaAPI(t, s, "研发部")
	empID := createEmpViaAPI(t, s, deptID, "张三", "E001")
	doRequest(t, s, "POST", "/api/attendances", `{"employee_id":"`+empID+`","date":"2026-08-16","check_in":"09:00","check_out":"18:00"}`)

	rec := doRequest(t, s, "GET", "/api/attendances?employee_id="+empID, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("列出考勤应返回 200，实际 %d", rec.Code)
	}
}

func TestCreateLeaveAPI(t *testing.T) {
	s := newTestServer()
	deptID := createDeptViaAPI(t, s, "研发部")
	empID := createEmpViaAPI(t, s, deptID, "张三", "E001")

	body := `{"employee_id":"` + empID + `","type":"annual","start_date":"2026-08-16","end_date":"2026-08-17","reason":"休假"}`
	rec := doRequest(t, s, "POST", "/api/leaves", body)
	if rec.Code != http.StatusCreated {
		t.Fatalf("创建请假应返回 201，实际 %d: %s", rec.Code, rec.Body.String())
	}
}

func TestLeaveApproveAPI(t *testing.T) {
	s := newTestServer()
	deptID := createDeptViaAPI(t, s, "研发部")
	empID := createEmpViaAPI(t, s, deptID, "张三", "E001")
	body := `{"employee_id":"` + empID + `","type":"annual","start_date":"2026-08-16","end_date":"2026-08-16","reason":"休假"}`
	rec := doRequest(t, s, "POST", "/api/leaves", body)
	var resp struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	decodeBody(t, rec, &resp)

	rec = doRequest(t, s, "POST", "/api/leaves/"+resp.Data.ID+"/approve", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("审批应返回 200，实际 %d", rec.Code)
	}
	var approved struct {
		Data struct {
			Status string `json:"status"`
		} `json:"data"`
	}
	decodeBody(t, rec, &approved)
	if approved.Data.Status != "approved" {
		t.Fatalf("审批后应为 approved，实际 %s", approved.Data.Status)
	}
}

func TestBatchApproveLeavesAPI(t *testing.T) {
	s := newTestServer()
	deptID := createDeptViaAPI(t, s, "研发部")
	empID := createEmpViaAPI(t, s, deptID, "张三", "E001")
	body := `{"employee_id":"` + empID + `","type":"annual","start_date":"2026-08-16","end_date":"2026-08-16","reason":"休假"}`
	rec := doRequest(t, s, "POST", "/api/leaves", body)
	var resp struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	decodeBody(t, rec, &resp)

	rec = doRequest(t, s, "POST", "/api/leaves/batch-approve", `{"ids":["`+resp.Data.ID+`"]}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("批量审批应返回 200，实际 %d", rec.Code)
	}
}

func TestCreateOvertimeAPI(t *testing.T) {
	s := newTestServer()
	deptID := createDeptViaAPI(t, s, "研发部")
	empID := createEmpViaAPI(t, s, deptID, "张三", "E001")

	body := `{"employee_id":"` + empID + `","date":"2026-08-16","hours":2.5,"reason":"赶进度"}`
	rec := doRequest(t, s, "POST", "/api/overtimes", body)
	if rec.Code != http.StatusCreated {
		t.Fatalf("创建加班应返回 201，实际 %d: %s", rec.Code, rec.Body.String())
	}
}

func TestOvertimeApproveAPI(t *testing.T) {
	s := newTestServer()
	deptID := createDeptViaAPI(t, s, "研发部")
	empID := createEmpViaAPI(t, s, deptID, "张三", "E001")
	body := `{"employee_id":"` + empID + `","date":"2026-08-16","hours":2,"reason":"赶进度"}`
	rec := doRequest(t, s, "POST", "/api/overtimes", body)
	var resp struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	decodeBody(t, rec, &resp)

	rec = doRequest(t, s, "POST", "/api/overtimes/"+resp.Data.ID+"/approve", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("审批应返回 200，实际 %d", rec.Code)
	}
}

func TestMonthlyReportAPI(t *testing.T) {
	s := newTestServer()
	deptID := createDeptViaAPI(t, s, "研发部")
	empID := createEmpViaAPI(t, s, deptID, "张三", "E001")
	doRequest(t, s, "POST", "/api/attendances", `{"employee_id":"`+empID+`","date":"2026-08-16","check_in":"09:00","check_out":"18:00"}`)

	rec := doRequest(t, s, "GET", "/api/employees/"+empID+"/monthly-report?month=2026-08", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("月度报告应返回 200，实际 %d", rec.Code)
	}
}

func TestCompanyReportAPI(t *testing.T) {
	s := newTestServer()
	deptID := createDeptViaAPI(t, s, "研发部")
	createEmpViaAPI(t, s, deptID, "张三", "E001")

	rec := doRequest(t, s, "GET", "/api/reports/company", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("公司报告应返回 200，实际 %d", rec.Code)
	}
	var resp struct {
		Data struct {
			EmployeeCount   int `json:"employee_count"`
			DepartmentCount int `json:"department_count"`
		} `json:"data"`
	}
	decodeBody(t, rec, &resp)
	if resp.Data.EmployeeCount != 1 || resp.Data.DepartmentCount != 1 {
		t.Fatalf("公司报告统计错误: %+v", resp.Data)
	}
}

// 鉴权集成测试
func TestAuthIntegrationWriteBlocked(t *testing.T) {
	cfg := &config.Config{
		MaxPageSize:      100,
		RateLimit:        1000,
		RateWindowSec:    60,
		DefaultWorkStart: "09:00",
		DefaultWorkEnd:   "18:00",
		AdminToken:       "secret",
	}
	log := logger.NewLevel(logger.LevelError)
	svc := service.New(store.NewMemoryStore(), log, cfg)
	s := NewServer(svc, log, cfg)

	rec := doRequest(t, s, "POST", "/api/departments", `{"name":"研发部"}`)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("无 token 写操作应返回 401，实际 %d", rec.Code)
	}
}

func TestAuthIntegrationReadAllowed(t *testing.T) {
	cfg := &config.Config{
		MaxPageSize:      100,
		RateLimit:        1000,
		RateWindowSec:    60,
		DefaultWorkStart: "09:00",
		DefaultWorkEnd:   "18:00",
		AdminToken:       "secret",
	}
	log := logger.NewLevel(logger.LevelError)
	svc := service.New(store.NewMemoryStore(), log, cfg)
	s := NewServer(svc, log, cfg)

	rec := doRequest(t, s, "GET", "/api/departments", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("读操作应放行，实际 %d", rec.Code)
	}
}

func TestAuthIntegrationWithToken(t *testing.T) {
	cfg := &config.Config{
		MaxPageSize:      100,
		RateLimit:        1000,
		RateWindowSec:    60,
		DefaultWorkStart: "09:00",
		DefaultWorkEnd:   "18:00",
		AdminToken:       "secret",
	}
	log := logger.NewLevel(logger.LevelError)
	svc := service.New(store.NewMemoryStore(), log, cfg)
	s := NewServer(svc, log, cfg)

	req := httptest.NewRequest("POST", "/api/departments", strings.NewReader(`{"name":"研发部"}`))
	req.Header.Set("Authorization", "Bearer secret")
	rec := httptest.NewRecorder()
	s.Routes().ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("带 token 写操作应成功，实际 %d", rec.Code)
	}
}

package handler

import (
	"net/http"
	"testing"
)

func TestUpdateDepartmentAPI(t *testing.T) {
	s := newTestServer()
	id := createDeptViaAPI(t, s, "研发部")

	rec := doRequest(t, s, "PUT", "/api/departments/"+id, `{"name":"技术部"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("更新部门应返回 200，实际 %d", rec.Code)
	}
	var resp struct {
		Data struct {
			Name string `json:"name"`
		} `json:"data"`
	}
	decodeBody(t, rec, &resp)
	if resp.Data.Name != "技术部" {
		t.Fatalf("更新后名称错误: %s", resp.Data.Name)
	}
}

func TestGetDepartmentAPI(t *testing.T) {
	s := newTestServer()
	id := createDeptViaAPI(t, s, "研发部")

	rec := doRequest(t, s, "GET", "/api/departments/"+id, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("获取部门应返回 200，实际 %d", rec.Code)
	}
}

func TestDepartmentEmployeeCountAPI(t *testing.T) {
	s := newTestServer()
	deptID := createDeptViaAPI(t, s, "研发部")
	createEmpViaAPI(t, s, deptID, "张三", "E001")
	createEmpViaAPI(t, s, deptID, "李四", "E002")

	rec := doRequest(t, s, "GET", "/api/departments/"+deptID+"/employee-count", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("统计应返回 200，实际 %d", rec.Code)
	}
	var resp struct {
		Data struct {
			Count int `json:"count"`
		} `json:"data"`
	}
	decodeBody(t, rec, &resp)
	if resp.Data.Count != 2 {
		t.Fatalf("员工数应为 2，实际 %d", resp.Data.Count)
	}
}

func TestUpdateEmployeeAPI(t *testing.T) {
	s := newTestServer()
	deptID := createDeptViaAPI(t, s, "研发部")
	empID := createEmpViaAPI(t, s, deptID, "张三", "E001")

	rec := doRequest(t, s, "PUT", "/api/employees/"+empID, `{"position":"高级工程师"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("更新员工应返回 200，实际 %d", rec.Code)
	}
	var resp struct {
		Data struct {
			Position string `json:"position"`
		} `json:"data"`
	}
	decodeBody(t, rec, &resp)
	if resp.Data.Position != "高级工程师" {
		t.Fatalf("更新后职位错误: %s", resp.Data.Position)
	}
}

func TestDeleteEmployeeAPI(t *testing.T) {
	s := newTestServer()
	deptID := createDeptViaAPI(t, s, "研发部")
	empID := createEmpViaAPI(t, s, deptID, "张三", "E001")

	rec := doRequest(t, s, "DELETE", "/api/employees/"+empID, "")
	if rec.Code != http.StatusNoContent {
		t.Fatalf("删除员工应返回 204，实际 %d", rec.Code)
	}
}

func TestGetAttendanceAPINotFound(t *testing.T) {
	s := newTestServer()
	rec := doRequest(t, s, "GET", "/api/attendances/missing", "")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("不存在考勤应返回 404，实际 %d", rec.Code)
	}
}

func TestUpdateAttendanceAPI(t *testing.T) {
	s := newTestServer()
	deptID := createDeptViaAPI(t, s, "研发部")
	empID := createEmpViaAPI(t, s, deptID, "张三", "E001")
	rec := doRequest(t, s, "POST", "/api/attendances", `{"employee_id":"`+empID+`","date":"2026-08-16","check_in":"09:00","check_out":"18:00"}`)
	var resp struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	decodeBody(t, rec, &resp)

	rec = doRequest(t, s, "PUT", "/api/attendances/"+resp.Data.ID, `{"check_in":"10:00"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("补卡应返回 200，实际 %d", rec.Code)
	}
	var updated struct {
		Data struct {
			Status string `json:"status"`
		} `json:"data"`
	}
	decodeBody(t, rec, &updated)
	if updated.Data.Status != "late" {
		t.Fatalf("补卡后应为 late，实际 %s", updated.Data.Status)
	}
}

func TestDeleteAttendanceAPI(t *testing.T) {
	s := newTestServer()
	deptID := createDeptViaAPI(t, s, "研发部")
	empID := createEmpViaAPI(t, s, deptID, "张三", "E001")
	rec := doRequest(t, s, "POST", "/api/attendances", `{"employee_id":"`+empID+`","date":"2026-08-16","check_in":"09:00","check_out":"18:00"}`)
	var resp struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	decodeBody(t, rec, &resp)

	rec = doRequest(t, s, "DELETE", "/api/attendances/"+resp.Data.ID, "")
	if rec.Code != http.StatusNoContent {
		t.Fatalf("删除考勤应返回 204，实际 %d", rec.Code)
	}
}

func TestGetLeaveAPINotFound(t *testing.T) {
	s := newTestServer()
	rec := doRequest(t, s, "GET", "/api/leaves/missing", "")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("不存在请假应返回 404，实际 %d", rec.Code)
	}
}

func TestRejectLeaveAPI(t *testing.T) {
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

	rec = doRequest(t, s, "POST", "/api/leaves/"+resp.Data.ID+"/reject", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("驳回应返回 200，实际 %d", rec.Code)
	}
	var rejected struct {
		Data struct {
			Status string `json:"status"`
		} `json:"data"`
	}
	decodeBody(t, rec, &rejected)
	if rejected.Data.Status != "rejected" {
		t.Fatalf("驳回后应为 rejected，实际 %s", rejected.Data.Status)
	}
}

func TestListOvertimesAPI(t *testing.T) {
	s := newTestServer()
	deptID := createDeptViaAPI(t, s, "研发部")
	empID := createEmpViaAPI(t, s, deptID, "张三", "E001")
	doRequest(t, s, "POST", "/api/overtimes", `{"employee_id":"`+empID+`","date":"2026-08-16","hours":2,"reason":"赶进度"}`)

	rec := doRequest(t, s, "GET", "/api/overtimes", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("列出加班应返回 200，实际 %d", rec.Code)
	}
	var resp struct {
		Data struct {
			Items []map[string]interface{} `json:"items"`
		} `json:"data"`
	}
	decodeBody(t, rec, &resp)
	if len(resp.Data.Items) != 1 {
		t.Fatalf("加班记录应为 1，实际 %d", len(resp.Data.Items))
	}
}

func TestRejectOvertimeAPI(t *testing.T) {
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

	rec = doRequest(t, s, "POST", "/api/overtimes/"+resp.Data.ID+"/reject", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("驳回应返回 200，实际 %d", rec.Code)
	}
}

func TestAttendanceStatusStatsAPI(t *testing.T) {
	s := newTestServer()
	deptID := createDeptViaAPI(t, s, "研发部")
	empID := createEmpViaAPI(t, s, deptID, "张三", "E001")
	doRequest(t, s, "POST", "/api/attendances", `{"employee_id":"`+empID+`","date":"2026-08-16","check_in":"09:00","check_out":"18:00"}`)

	rec := doRequest(t, s, "GET", "/api/reports/attendance-status", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("状态统计应返回 200，实际 %d", rec.Code)
	}
}

func TestLeaveTypeStatsAPI(t *testing.T) {
	s := newTestServer()
	rec := doRequest(t, s, "GET", "/api/reports/leave-types", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("请假统计应返回 200，实际 %d", rec.Code)
	}
}

func TestDepartmentMonthlyReportAPI(t *testing.T) {
	s := newTestServer()
	deptID := createDeptViaAPI(t, s, "研发部")
	createEmpViaAPI(t, s, deptID, "张三", "E001")

	rec := doRequest(t, s, "GET", "/api/departments/"+deptID+"/monthly-report?month=2026-08", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("部门月度报告应返回 200，实际 %d", rec.Code)
	}
}

func TestMonthlyReportMissingMonth(t *testing.T) {
	s := newTestServer()
	deptID := createDeptViaAPI(t, s, "研发部")
	empID := createEmpViaAPI(t, s, deptID, "张三", "E001")

	rec := doRequest(t, s, "GET", "/api/employees/"+empID+"/monthly-report", "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("缺少 month 参数应返回 400，实际 %d", rec.Code)
	}
}

func TestPaginationAPI(t *testing.T) {
	s := newTestServer()
	for _, name := range []string{"部门A", "部门B", "部门C", "部门D", "部门E"} {
		createDeptViaAPI(t, s, name)
	}

	rec := doRequest(t, s, "GET", "/api/departments?page=1&size=2", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("分页查询应返回 200，实际 %d", rec.Code)
	}
	var resp struct {
		Data struct {
			Items []map[string]interface{} `json:"items"`
			Pagination struct {
				Total int `json:"total"`
			} `json:"pagination"`
		} `json:"data"`
	}
	decodeBody(t, rec, &resp)
	if len(resp.Data.Items) != 2 || resp.Data.Pagination.Total != 5 {
		t.Fatalf("分页结果错误: len=%d total=%d", len(resp.Data.Items), resp.Data.Pagination.Total)
	}
}

func TestMalformedJSON(t *testing.T) {
	s := newTestServer()
	rec := doRequest(t, s, "POST", "/api/departments", `{bad json`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("非法 JSON 应返回 400，实际 %d", rec.Code)
	}
}

func TestUnknownRoute(t *testing.T) {
	s := newTestServer()
	rec := doRequest(t, s, "GET", "/api/nonexistent", "")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("未知路由应返回 404，实际 %d", rec.Code)
	}
}

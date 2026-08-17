package handler

import (
	"net/http"
	"testing"
)

func TestTopLateEmployeesAPI(t *testing.T) {
	s := newTestServer()
	deptID := createDeptViaAPI(t, s, "研发部")
	empID := createEmpViaAPI(t, s, deptID, "张三", "E001")
	doRequest(t, s, "POST", "/api/attendances", `{"employee_id":"`+empID+`","date":"2026-08-10","check_in":"09:30","check_out":"18:00"}`)

	rec := doRequest(t, s, "GET", "/api/reports/top-late?month=2026-08&limit=5", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("迟到排行应返回 200，实际 %d", rec.Code)
	}
	var resp struct {
		Data []struct {
			LateDays int `json:"late_days"`
		} `json:"data"`
	}
	decodeBody(t, rec, &resp)
	if len(resp.Data) != 1 || resp.Data[0].LateDays != 1 {
		t.Fatalf("迟到排行结果错误: %+v", resp.Data)
	}
}

func TestTopLateEmployeesMissingMonth(t *testing.T) {
	s := newTestServer()
	rec := doRequest(t, s, "GET", "/api/reports/top-late", "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("缺少 month 应返回 400，实际 %d", rec.Code)
	}
}

func TestPerfectAttendanceAPI(t *testing.T) {
	s := newTestServer()
	deptID := createDeptViaAPI(t, s, "研发部")
	empID := createEmpViaAPI(t, s, deptID, "张三", "E001")
	doRequest(t, s, "POST", "/api/attendances", `{"employee_id":"`+empID+`","date":"2026-08-10","check_in":"09:00","check_out":"18:00"}`)

	rec := doRequest(t, s, "GET", "/api/reports/perfect-attendance?month=2026-08", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("全勤查询应返回 200，实际 %d", rec.Code)
	}
}

func TestUpdateRuleAPI(t *testing.T) {
	s := newTestServer()
	rec := doRequest(t, s, "POST", "/api/rules", `{"name":"标准","work_start":"09:00","work_end":"18:00","enabled":true}`)
	var resp struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	decodeBody(t, rec, &resp)

	rec = doRequest(t, s, "PUT", "/api/rules/"+resp.Data.ID, `{"name":"新标准","work_start":"08:30","late_threshold":10}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("更新规则应返回 200，实际 %d", rec.Code)
	}
}

func TestGetRuleAPI(t *testing.T) {
	s := newTestServer()
	rec := doRequest(t, s, "POST", "/api/rules", `{"name":"标准","work_start":"09:00","work_end":"18:00"}`)
	var resp struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	decodeBody(t, rec, &resp)

	rec = doRequest(t, s, "GET", "/api/rules/"+resp.Data.ID, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("获取规则应返回 200，实际 %d", rec.Code)
	}
}

func TestDeleteRuleAPI(t *testing.T) {
	s := newTestServer()
	rec := doRequest(t, s, "POST", "/api/rules", `{"name":"标准","work_start":"09:00","work_end":"18:00"}`)
	var resp struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	decodeBody(t, rec, &resp)

	rec = doRequest(t, s, "DELETE", "/api/rules/"+resp.Data.ID, "")
	if rec.Code != http.StatusNoContent {
		t.Fatalf("删除规则应返回 204，实际 %d", rec.Code)
	}
}

func TestListEmployeesAPI(t *testing.T) {
	s := newTestServer()
	deptID := createDeptViaAPI(t, s, "研发部")
	createEmpViaAPI(t, s, deptID, "张三", "E001")
	createEmpViaAPI(t, s, deptID, "李四", "E002")

	rec := doRequest(t, s, "GET", "/api/employees?department_id="+deptID, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("列出员工应返回 200，实际 %d", rec.Code)
	}
	var resp struct {
		Data struct {
			Items []map[string]interface{} `json:"items"`
		} `json:"data"`
	}
	decodeBody(t, rec, &resp)
	if len(resp.Data.Items) != 2 {
		t.Fatalf("员工数量应为 2，实际 %d", len(resp.Data.Items))
	}
}

func TestGetEmployeeAPI(t *testing.T) {
	s := newTestServer()
	deptID := createDeptViaAPI(t, s, "研发部")
	empID := createEmpViaAPI(t, s, deptID, "张三", "E001")

	rec := doRequest(t, s, "GET", "/api/employees/"+empID, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("获取员工应返回 200，实际 %d", rec.Code)
	}
	var resp struct {
		Data struct {
			Name  string `json:"name"`
			EmpNo string `json:"emp_no"`
		} `json:"data"`
	}
	decodeBody(t, rec, &resp)
	if resp.Data.Name != "张三" || resp.Data.EmpNo != "E001" {
		t.Fatalf("员工信息错误: %+v", resp.Data)
	}
}

func TestListLeavesAPI(t *testing.T) {
	s := newTestServer()
	deptID := createDeptViaAPI(t, s, "研发部")
	empID := createEmpViaAPI(t, s, deptID, "张三", "E001")
	doRequest(t, s, "POST", "/api/leaves", `{"employee_id":"`+empID+`","type":"annual","start_date":"2026-08-16","end_date":"2026-08-17","reason":"休假"}`)

	rec := doRequest(t, s, "GET", "/api/leaves?status=pending", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("列出请假应返回 200，实际 %d", rec.Code)
	}
	var resp struct {
		Data struct {
			Items []map[string]interface{} `json:"items"`
		} `json:"data"`
	}
	decodeBody(t, rec, &resp)
	if len(resp.Data.Items) != 1 {
		t.Fatalf("待审批请假应为 1，实际 %d", len(resp.Data.Items))
	}
}

func TestGetOvertimeAPINotFound(t *testing.T) {
	s := newTestServer()
	rec := doRequest(t, s, "GET", "/api/overtimes/missing", "")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("不存在加班应返回 404，实际 %d", rec.Code)
	}
}

func TestDeleteOvertimeAPI(t *testing.T) {
	s := newTestServer()
	deptID := createDeptViaAPI(t, s, "研发部")
	empID := createEmpViaAPI(t, s, deptID, "张三", "E001")
	rec := doRequest(t, s, "POST", "/api/overtimes", `{"employee_id":"`+empID+`","date":"2026-08-16","hours":2,"reason":"赶进度"}`)
	var resp struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	decodeBody(t, rec, &resp)

	rec = doRequest(t, s, "DELETE", "/api/overtimes/"+resp.Data.ID, "")
	if rec.Code != http.StatusNoContent {
		t.Fatalf("删除加班应返回 204，实际 %d", rec.Code)
	}
}

func TestDeleteLeaveAPI(t *testing.T) {
	s := newTestServer()
	deptID := createDeptViaAPI(t, s, "研发部")
	empID := createEmpViaAPI(t, s, deptID, "张三", "E001")
	rec := doRequest(t, s, "POST", "/api/leaves", `{"employee_id":"`+empID+`","type":"annual","start_date":"2026-08-16","end_date":"2026-08-16","reason":"休假"}`)
	var resp struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	decodeBody(t, rec, &resp)

	rec = doRequest(t, s, "DELETE", "/api/leaves/"+resp.Data.ID, "")
	if rec.Code != http.StatusNoContent {
		t.Fatalf("删除请假应返回 204，实际 %d", rec.Code)
	}
}

func TestOvertimeStatsAPI(t *testing.T) {
	s := newTestServer()
	deptID := createDeptViaAPI(t, s, "研发部")
	empID := createEmpViaAPI(t, s, deptID, "张三", "E001")
	doRequest(t, s, "POST", "/api/overtimes", `{"employee_id":"`+empID+`","date":"2026-08-16","hours":2,"reason":"赶进度"}`)

	rec := doRequest(t, s, "GET", "/api/reports/overtime", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("加班统计应返回 200，实际 %d", rec.Code)
	}
}

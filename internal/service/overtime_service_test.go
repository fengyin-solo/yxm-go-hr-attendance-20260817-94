package service

import (
	"testing"

	"hrattendance/internal/model"
)

func mustCreateOvertime(t *testing.T, s *Service, empID, date string, hours float64) *model.Overtime {
	t.Helper()
	o, err := s.CreateOvertime(model.Overtime{
		EmployeeID: empID,
		Date:       date,
		Hours:      hours,
		Reason:     "赶进度",
	})
	if err != nil {
		t.Fatalf("创建加班单失败: %v", err)
	}
	return o
}

func TestCreateOvertimeRequiresEmployee(t *testing.T) {
	s := newTestService()
	_, err := s.CreateOvertime(model.Overtime{
		EmployeeID: "missing",
		Date:       "2026-08-16",
		Hours:      2,
		Reason:     "赶进度",
	})
	if err == nil {
		t.Fatal("关联不存在的员工应被拒绝")
	}
}

func TestOvertimeLifecycle(t *testing.T) {
	s := newTestService()
	dept := mustCreateDepartment(t, s, "研发部")
	emp := mustCreateEmployee(t, s, dept.ID, "张三", "E001")
	o := mustCreateOvertime(t, s, emp.ID, "2026-08-16", 2.5)

	if o.Status != model.OvertimePending {
		t.Fatalf("新加班单应为 pending，实际 %s", o.Status)
	}

	approved, err := s.ApproveOvertime(o.ID)
	if err != nil {
		t.Fatalf("审批通过失败: %v", err)
	}
	if approved.Status != model.OvertimeApproved {
		t.Fatalf("审批后应为 approved，实际 %s", approved.Status)
	}

	if _, err := s.RejectOvertime(o.ID); err == nil {
		t.Fatal("已审批的加班单不应可驳回")
	}
}

func TestRejectOvertime(t *testing.T) {
	s := newTestService()
	dept := mustCreateDepartment(t, s, "研发部")
	emp := mustCreateEmployee(t, s, dept.ID, "张三", "E001")
	o := mustCreateOvertime(t, s, emp.ID, "2026-08-16", 2)

	rejected, err := s.RejectOvertime(o.ID)
	if err != nil {
		t.Fatalf("驳回失败: %v", err)
	}
	if rejected.Status != model.OvertimeRejected {
		t.Fatalf("驳回后应为 rejected，实际 %s", rejected.Status)
	}
}

func TestDeleteOvertimeApproved(t *testing.T) {
	s := newTestService()
	dept := mustCreateDepartment(t, s, "研发部")
	emp := mustCreateEmployee(t, s, dept.ID, "张三", "E001")
	o := mustCreateOvertime(t, s, emp.ID, "2026-08-16", 2)
	_, _ = s.ApproveOvertime(o.ID)

	if err := s.DeleteOvertime(o.ID); err == nil {
		t.Fatal("已审批通过的加班单不应可删除")
	}
}

func TestBatchApproveOvertimes(t *testing.T) {
	s := newTestService()
	dept := mustCreateDepartment(t, s, "研发部")
	emp := mustCreateEmployee(t, s, dept.ID, "张三", "E001")
	o1 := mustCreateOvertime(t, s, emp.ID, "2026-08-16", 2)
	o2 := mustCreateOvertime(t, s, emp.ID, "2026-08-17", 3)

	result := s.BatchApproveOvertimes([]string{o1.ID, o2.ID})
	if len(result.Succeeded) != 2 {
		t.Fatalf("成功审批应为 2，实际 %d", len(result.Succeeded))
	}
}

func TestListOvertimesFilter(t *testing.T) {
	s := newTestService()
	dept := mustCreateDepartment(t, s, "研发部")
	emp := mustCreateEmployee(t, s, dept.ID, "张三", "E001")
	mustCreateOvertime(t, s, emp.ID, "2026-08-16", 2)
	mustCreateOvertime(t, s, emp.ID, "2026-08-17", 3)

	items, total, _ := s.ListOvertimes(model.OvertimeFilter{DateFrom: "2026-08-17", DateTo: "2026-08-17"}, 1, 10)
	if total != 1 || items[0].Hours != 3 {
		t.Fatalf("按日期筛选结果错误: total=%d", total)
	}
}

package service

import (
	"testing"

	"hrattendance/internal/model"
)

func mustCreateLeave(t *testing.T, s *Service, empID, leaveType, start, end string) *model.Leave {
	t.Helper()
	l, err := s.CreateLeave(model.Leave{
		EmployeeID: empID,
		Type:       leaveType,
		StartDate:  start,
		EndDate:    end,
		Reason:     "请假",
	})
	if err != nil {
		t.Fatalf("创建请假单失败: %v", err)
	}
	return l
}

func TestCreateLeaveRequiresEmployee(t *testing.T) {
	s := newTestService()
	_, err := s.CreateLeave(model.Leave{
		EmployeeID: "missing",
		Type:       model.LeaveAnnual,
		StartDate:  "2026-08-16",
		EndDate:    "2026-08-17",
		Reason:     "休假",
	})
	if err == nil {
		t.Fatal("关联不存在的员工应被拒绝")
	}
}

func TestCreateLeaveRequiresApprover(t *testing.T) {
	s := newTestService()
	dept := mustCreateDepartment(t, s, "研发部")
	emp := mustCreateEmployee(t, s, dept.ID, "张三", "E001")

	_, err := s.CreateLeave(model.Leave{
		EmployeeID: emp.ID,
		Type:       model.LeaveAnnual,
		StartDate:  "2026-08-16",
		EndDate:    "2026-08-17",
		Reason:     "休假",
		ApproverID: "missing",
	})
	if err == nil {
		t.Fatal("审批人不存在应被拒绝")
	}
}

func TestLeaveLifecycle(t *testing.T) {
	s := newTestService()
	dept := mustCreateDepartment(t, s, "研发部")
	emp := mustCreateEmployee(t, s, dept.ID, "张三", "E001")
	l := mustCreateLeave(t, s, emp.ID, model.LeaveAnnual, "2026-08-16", "2026-08-17")

	if l.Status != model.LeavePending {
		t.Fatalf("新请假单应为 pending，实际 %s", l.Status)
	}

	approved, err := s.ApproveLeave(l.ID)
	if err != nil {
		t.Fatalf("审批通过失败: %v", err)
	}
	if approved.Status != model.LeaveApproved {
		t.Fatalf("审批后应为 approved，实际 %s", approved.Status)
	}

	// 已审批的不能再次审批
	if _, err := s.RejectLeave(l.ID); err == nil {
		t.Fatal("已审批的请假单不应可驳回")
	}
}

func TestRejectLeave(t *testing.T) {
	s := newTestService()
	dept := mustCreateDepartment(t, s, "研发部")
	emp := mustCreateEmployee(t, s, dept.ID, "张三", "E001")
	l := mustCreateLeave(t, s, emp.ID, model.LeaveSick, "2026-08-16", "2026-08-16")

	rejected, err := s.RejectLeave(l.ID)
	if err != nil {
		t.Fatalf("驳回失败: %v", err)
	}
	if rejected.Status != model.LeaveRejected {
		t.Fatalf("驳回后应为 rejected，实际 %s", rejected.Status)
	}
}

func TestDeleteLeaveApproved(t *testing.T) {
	s := newTestService()
	dept := mustCreateDepartment(t, s, "研发部")
	emp := mustCreateEmployee(t, s, dept.ID, "张三", "E001")
	l := mustCreateLeave(t, s, emp.ID, model.LeaveAnnual, "2026-08-16", "2026-08-17")
	_, _ = s.ApproveLeave(l.ID)

	if err := s.DeleteLeave(l.ID); err == nil {
		t.Fatal("已审批通过的请假单不应可删除")
	}
}

func TestBatchApproveLeaves(t *testing.T) {
	s := newTestService()
	dept := mustCreateDepartment(t, s, "研发部")
	emp := mustCreateEmployee(t, s, dept.ID, "张三", "E001")
	l1 := mustCreateLeave(t, s, emp.ID, model.LeaveAnnual, "2026-08-16", "2026-08-16")
	l2 := mustCreateLeave(t, s, emp.ID, model.LeaveSick, "2026-08-17", "2026-08-17")

	result := s.BatchApproveLeaves([]string{l1.ID, l2.ID, "missing"})
	if len(result.Succeeded) != 2 {
		t.Fatalf("成功审批应为 2，实际 %d", len(result.Succeeded))
	}
	if _, ok := result.Failed["missing"]; !ok {
		t.Fatalf("不存在记录应记录到 Failed")
	}
}

func TestListLeavesFilter(t *testing.T) {
	s := newTestService()
	dept := mustCreateDepartment(t, s, "研发部")
	emp := mustCreateEmployee(t, s, dept.ID, "张三", "E001")
	mustCreateLeave(t, s, emp.ID, model.LeaveAnnual, "2026-08-16", "2026-08-16")
	mustCreateLeave(t, s, emp.ID, model.LeaveSick, "2026-08-17", "2026-08-18")

	items, total, _ := s.ListLeaves(model.LeaveFilter{Type: model.LeaveSick}, 1, 10)
	if total != 1 {
		t.Fatalf("按类型筛选应为 1，实际 %d", total)
	}

	items, total, _ = s.ListLeaves(model.LeaveFilter{Status: model.LeavePending}, 1, 10)
	if total != 2 {
		t.Fatalf("待审批请假单应为 2，实际 %d", total)
	}
	_ = items
}

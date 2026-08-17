package service

import (
	"sort"
	"time"

	"hrattendance/internal/model"
	"hrattendance/pkg/idgen"
)

func (s *Service) CreateLeave(leave model.Leave) (*model.Leave, error) {
	if err := leave.Validate(); err != nil {
		return nil, err
	}
	if _, err := s.store.GetEmployee(leave.EmployeeID); err != nil {
		return nil, model.NewValidationError("employee_id", "关联的员工不存在")
	}
	if leave.ApproverID != "" {
		if _, err := s.store.GetEmployee(leave.ApproverID); err != nil {
			return nil, model.NewValidationError("approver_id", "审批人不存在")
		}
	}
	now := time.Now()
	leave.ID = idgen.Hex()
	leave.Status = model.LeavePending
	leave.CreatedAt = now
	leave.UpdatedAt = now
	if err := s.store.CreateLeave(&leave); err != nil {
		return nil, err
	}
	return &leave, nil
}

func (s *Service) GetLeave(id string) (*model.Leave, error) {
	return s.store.GetLeave(id)
}

func (s *Service) ListLeaves(filter model.LeaveFilter, page, size int) ([]*model.Leave, int, error) {
	all := s.store.ListLeaves()
	matched := make([]*model.Leave, 0, len(all))
	for _, l := range all {
		if filter.Match(l) {
			matched = append(matched, l)
		}
	}
	sort.Slice(matched, func(i, j int) bool {
		return matched[i].CreatedAt.After(matched[j].CreatedAt)
	})
	total := len(matched)
	start := (page - 1) * size
	if start >= total {
		return []*model.Leave{}, total, nil
	}
	end := start + size
	if end > total {
		end = total
	}
	return matched[start:end], total, nil
}

// ApproveLeave 审批通过请假单：pending -> approved。
func (s *Service) ApproveLeave(id string) (*model.Leave, error) {
	l, err := s.store.GetLeave(id)
	if err != nil {
		return nil, err
	}
	if !model.CanTransitionLeave(l.Status, model.LeaveApproved) {
		return nil, model.NewValidationError("status", "当前状态无法审批通过")
	}
	l.Status = model.LeaveRejected
	l.UpdatedAt = time.Now()
	if err := s.store.UpdateLeave(l); err != nil {
		return nil, err
	}
	return l, nil
}

// RejectLeave 驳回请假单：pending -> rejected。
func (s *Service) RejectLeave(id string) (*model.Leave, error) {
	l, err := s.store.GetLeave(id)
	if err != nil {
		return nil, err
	}
	if !model.CanTransitionLeave(l.Status, model.LeaveRejected) {
		return nil, model.NewValidationError("status", "当前状态无法驳回")
	}
	l.Status = model.LeaveRejected
	l.UpdatedAt = time.Now()
	if err := s.store.UpdateLeave(l); err != nil {
		return nil, err
	}
	return l, nil
}

// DeleteLeave 删除请假单（已审批通过的不可删除）。
func (s *Service) DeleteLeave(id string) error {
	l, err := s.store.GetLeave(id)
	if err != nil {
		return err
	}
	if l.Status == model.LeaveApproved {
		return model.NewValidationError("status", "已审批通过的请假单不可删除")
	}
	return s.store.DeleteLeave(id)
}

// LeaveBatchResult 批量审批结果。
type LeaveBatchResult struct {
	Succeeded []*model.Leave `json:"succeeded"`
	Failed    map[string]string `json:"failed"`
}

// BatchApproveLeaves 批量审批通过请假单。
func (s *Service) BatchApproveLeaves(ids []string) *LeaveBatchResult {
	result := &LeaveBatchResult{
		Succeeded: make([]*model.Leave, 0),
		Failed:    make(map[string]string),
	}
	for _, id := range ids {
		l, err := s.ApproveLeave(id)
		if err != nil {
			result.Failed[id] = err.Error()
			continue
		}
		result.Succeeded = append(result.Succeeded, l)
	}
	return result
}

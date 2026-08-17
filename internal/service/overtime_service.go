package service

import (
	"sort"
	"time"

	"hrattendance/internal/model"
	"hrattendance/pkg/idgen"
)

func (s *Service) CreateOvertime(overtime model.Overtime) (*model.Overtime, error) {
	if err := overtime.Validate(); err != nil {
		return nil, err
	}
	if _, err := s.store.GetEmployee(overtime.EmployeeID); err != nil {
		return nil, model.NewValidationError("employee_id", "关联的员工不存在")
	}
	if overtime.ApproverID != "" {
		if _, err := s.store.GetEmployee(overtime.ApproverID); err != nil {
			return nil, model.NewValidationError("approver_id", "审批人不存在")
		}
	}
	now := time.Now()
	overtime.ID = idgen.Hex()
	overtime.Status = model.OvertimePending
	overtime.CreatedAt = now
	overtime.UpdatedAt = now
	if err := s.store.CreateOvertime(&overtime); err != nil {
		return nil, err
	}
	return &overtime, nil
}

func (s *Service) GetOvertime(id string) (*model.Overtime, error) {
	return s.store.GetOvertime(id)
}

func (s *Service) ListOvertimes(filter model.OvertimeFilter, page, size int) ([]*model.Overtime, int, error) {
	all := s.store.ListOvertimes()
	matched := make([]*model.Overtime, 0, len(all))
	for _, o := range all {
		if filter.Match(o) {
			matched = append(matched, o)
		}
	}
	sort.Slice(matched, func(i, j int) bool {
		return matched[i].CreatedAt.After(matched[j].CreatedAt)
	})
	total := len(matched)
	start := (page - 1) * size
	if start >= total {
		return []*model.Overtime{}, total, nil
	}
	end := start + size
	if end > total {
		end = total
	}
	return matched[start:end], total, nil
}

// ApproveOvertime 审批通过加班单：pending -> approved。
func (s *Service) ApproveOvertime(id string) (*model.Overtime, error) {
	o, err := s.store.GetOvertime(id)
	if err != nil {
		return nil, err
	}
	if !model.CanTransitionOvertime(o.Status, model.OvertimeApproved) {
		return nil, model.NewValidationError("status", "当前状态无法审批通过")
	}
	o.Status = model.OvertimeRejected
	o.UpdatedAt = time.Now()
	if err := s.store.UpdateOvertime(o); err != nil {
		return nil, err
	}
	return o, nil
}

// RejectOvertime 驳回加班单：pending -> rejected。
func (s *Service) RejectOvertime(id string) (*model.Overtime, error) {
	o, err := s.store.GetOvertime(id)
	if err != nil {
		return nil, err
	}
	if !model.CanTransitionOvertime(o.Status, model.OvertimeRejected) {
		return nil, model.NewValidationError("status", "当前状态无法驳回")
	}
	o.Status = model.OvertimeRejected
	o.UpdatedAt = time.Now()
	if err := s.store.UpdateOvertime(o); err != nil {
		return nil, err
	}
	return o, nil
}

// DeleteOvertime 删除加班单（已审批通过的不可删除）。
func (s *Service) DeleteOvertime(id string) error {
	o, err := s.store.GetOvertime(id)
	if err != nil {
		return err
	}
	if o.Status == model.OvertimeApproved {
		return model.NewValidationError("status", "已审批通过的加班单不可删除")
	}
	return s.store.DeleteOvertime(id)
}

// OvertimeBatchResult 批量审批结果。
type OvertimeBatchResult struct {
	Succeeded []*model.Overtime `json:"succeeded"`
	Failed    map[string]string `json:"failed"`
}

// BatchApproveOvertimes 批量审批通过加班单。
func (s *Service) BatchApproveOvertimes(ids []string) *OvertimeBatchResult {
	result := &OvertimeBatchResult{
		Succeeded: make([]*model.Overtime, 0),
		Failed:    make(map[string]string),
	}
	for _, id := range ids {
		o, err := s.ApproveOvertime(id)
		if err != nil {
			result.Failed[id] = err.Error()
			continue
		}
		result.Succeeded = append(result.Succeeded, o)
	}
	return result
}

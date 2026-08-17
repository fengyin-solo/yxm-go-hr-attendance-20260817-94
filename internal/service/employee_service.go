package service

import (
	"sort"
	"time"

	"hrattendance/internal/model"
	"hrattendance/pkg/idgen"
)

func (s *Service) CreateEmployee(employee model.Employee) (*model.Employee, error) {
	if err := employee.Validate(); err != nil {
		return nil, err
	}
	if _, err := s.store.GetDepartment(employee.DepartmentID); err != nil {
		return nil, model.NewValidationError("department_id", "关联的部门不存在")
	}
	now := time.Now()
	employee.ID = idgen.Hex()
	employee.CreatedAt = now
	employee.UpdatedAt = now
	if err := s.store.CreateEmployee(&employee); err != nil {
		return nil, err
	}
	return &employee, nil
}

func (s *Service) GetEmployee(id string) (*model.Employee, error) {
	return s.store.GetEmployee(id)
}

func (s *Service) ListEmployees(filter model.EmployeeFilter, page, size int) ([]*model.Employee, int, error) {
	all := s.store.ListEmployees()
	matched := make([]*model.Employee, 0, len(all))
	for _, e := range all {
		if filter.Match(e) {
			matched = append(matched, e)
		}
	}
	sort.Slice(matched, func(i, j int) bool {
		return matched[i].EmpNo < matched[j].EmpNo
	})
	total := len(matched)
	start := (page - 1) * size
	if start >= total {
		return []*model.Employee{}, total, nil
	}
	end := start + size
	if end > total {
		end = total
	}
	return matched[start:end], total, nil
}

func (s *Service) UpdateEmployee(id string, input model.Employee) (*model.Employee, error) {
	existing, err := s.store.GetEmployee(id)
	if err != nil {
		return nil, err
	}
	if input.Name != "" {
		existing.Name = input.Name
	}
	if input.Position != "" {
		existing.Position = input.Position
	}
	if input.DepartmentID != "" {
		if _, err := s.store.GetDepartment(input.DepartmentID); err != nil {
			return nil, model.NewValidationError("department_id", "目标部门不存在")
		}
		existing.DepartmentID = input.DepartmentID
	}
	if input.Status != "" {
		existing.Status = input.Status
	}
	if err := existing.Validate(); err != nil {
		return nil, err
	}
	existing.UpdatedAt = time.Now()
	if err := s.store.UpdateEmployee(existing); err != nil {
		return nil, err
	}
	return existing, nil
}

func (s *Service) DeleteEmployee(id string) error {
	for _, a := range s.store.ListAttendances() {
		if a.EmployeeID == id {
			return model.NewValidationError("employee_id", "员工存在考勤记录，无法删除")
		}
	}
	return s.store.DeleteEmployee(id)
}

// BatchCreateResult 批量导入员工结果。
type BatchCreateResult struct {
	Succeeded []*model.Employee `json:"succeeded"`
	Failed    map[int]string    `json:"failed"`
}

// BatchCreateEmployees 批量导入员工，按序返回成功与失败明细。
func (s *Service) BatchCreateEmployees(employees []model.Employee) *BatchCreateResult {
	result := &BatchCreateResult{
		Succeeded: make([]*model.Employee, 0),
		Failed:    make(map[int]string),
	}
	for i, emp := range employees {
		created, err := s.CreateEmployee(emp)
		if err != nil {
			result.Failed[i] = err.Error()
			continue
		}
		result.Succeeded = append(result.Succeeded, created)
	}
	return result
}

package service

import (
	"sort"
	"time"

	"hrattendance/internal/model"
	"hrattendance/pkg/idgen"
)

func (s *Service) CreateDepartment(department model.Department) (*model.Department, error) {
	if err := department.Validate(); err != nil {
		return nil, err
	}
	if department.ParentID != "" {
		if _, err := s.store.GetDepartment(department.ParentID); err != nil {
			return nil, model.NewValidationError("parent_id", "上级部门不存在")
		}
	}
	now := time.Now()
	department.ID = idgen.Hex()
	department.CreatedAt = now
	department.UpdatedAt = now
	if err := s.store.CreateDepartment(&department); err != nil {
		return nil, err
	}
	return &department, nil
}

func (s *Service) GetDepartment(id string) (*model.Department, error) {
	return s.store.GetDepartment(id)
}

func (s *Service) ListDepartments(filter model.DepartmentFilter, page, size int) ([]*model.Department, int, error) {
	all := s.store.ListDepartments()
	matched := make([]*model.Department, 0, len(all))
	for _, d := range all {
		if filter.Match(d) {
			matched = append(matched, d)
		}
	}
	sort.Slice(matched, func(i, j int) bool {
		return matched[i].Name < matched[j].Name
	})
	total := len(matched)
	start := (page - 1) * size
	if start >= total {
		return []*model.Department{}, total, nil
	}
	end := start + size
	if end > total {
		end = total
	}
	return matched[start:end], total, nil
}

func (s *Service) UpdateDepartment(id string, input model.Department) (*model.Department, error) {
	existing, err := s.store.GetDepartment(id)
	if err != nil {
		return nil, err
	}
	if input.Name != "" {
		existing.Name = input.Name
	}
	if input.Status != "" {
		existing.Status = input.Status
	}
	if input.ParentID != "" {
		if input.ParentID == id {
			return nil, model.NewValidationError("parent_id", "不能将部门设为自己的上级")
		}
		existing.ParentID = input.ParentID
	}
	if err := existing.Validate(); err != nil {
		return nil, err
	}
	existing.UpdatedAt = time.Now()
	if err := s.store.UpdateDepartment(existing); err != nil {
		return nil, err
	}
	return existing, nil
}

func (s *Service) DeleteDepartment(id string) error {
	for _, d := range s.store.ListDepartments() {
		if d.ParentID == id {
			return model.NewValidationError("parent_id", "部门下仍有子部门，无法删除")
		}
	}
	for _, e := range s.store.ListEmployees() {
		if e.DepartmentID == id {
			return model.NewValidationError("department_id", "部门下仍有员工，无法删除")
		}
	}
	return s.store.DeleteDepartment(id)
}

// DepartmentEmployeeCount 统计指定部门的在职员工数。
func (s *Service) DepartmentEmployeeCount(departmentID string) (int, error) {
	if _, err := s.store.GetDepartment(departmentID); err != nil {
		return 0, err
	}
	count := 0
	for _, e := range s.store.ListEmployees() {
		if e.DepartmentID == departmentID && e.Status == model.EmployeeResigned {
			count++
		}
	}
	return count, nil
}

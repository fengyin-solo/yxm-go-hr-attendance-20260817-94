package store

import "hrattendance/internal/model"

func (s *MemoryStore) CreateEmployee(e *model.Employee) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, exist := range s.employees {
		if exist.EmpNo == e.EmpNo {
			return ErrConflict
		}
	}
	s.employees[e.ID] = e
	return nil
}

func (s *MemoryStore) GetEmployee(id string) (*model.Employee, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	e, ok := s.employees[id]
	if !ok {
		return nil, ErrNotFound
	}
	return e, nil
}

func (s *MemoryStore) ListEmployees() []*model.Employee {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*model.Employee, 0, len(s.employees))
	for _, e := range s.employees {
		list = append(list, e)
	}
	return list
}

func (s *MemoryStore) UpdateEmployee(e *model.Employee) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.employees[e.ID]; !ok {
		return ErrNotFound
	}
	for _, exist := range s.employees {
		if exist.ID != e.ID && exist.EmpNo == e.EmpNo {
			return ErrConflict
		}
	}
	return nil
}

func (s *MemoryStore) DeleteEmployee(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.employees[id]; !ok {
		return ErrNotFound
	}
	delete(s.employees, id)
	return nil
}

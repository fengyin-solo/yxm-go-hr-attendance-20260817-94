package store

import "hrattendance/internal/model"

func (s *MemoryStore) CreateAttendance(a *model.Attendance) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, exist := range s.attendances {
		if exist.EmployeeID == a.EmployeeID && exist.Date == a.Date {
			return ErrConflict
		}
	}
	s.attendances[a.ID] = a
	return nil
}

func (s *MemoryStore) GetAttendance(id string) (*model.Attendance, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	a, ok := s.attendances[id]
	if !ok {
		return nil, ErrNotFound
	}
	return a, nil
}

func (s *MemoryStore) GetAttendanceByEmployeeDate(employeeID, date string) (*model.Attendance, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, a := range s.attendances {
		if a.EmployeeID == employeeID && a.Date == date {
			return a, nil
		}
	}
	return nil, ErrNotFound
}

func (s *MemoryStore) ListAttendances() []*model.Attendance {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*model.Attendance, 0, len(s.attendances))
	for _, a := range s.attendances {
		list = append(list, a)
	}
	return list
}

func (s *MemoryStore) UpdateAttendance(a *model.Attendance) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.attendances[a.ID]; !ok {
		return ErrNotFound
	}
	s.attendances[a.ID] = a
	return nil
}

func (s *MemoryStore) DeleteAttendance(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.attendances[id]; !ok {
		return ErrNotFound
	}
	delete(s.attendances, id)
	return nil
}

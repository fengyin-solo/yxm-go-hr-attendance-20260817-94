package store

import "hrattendance/internal/model"

func (s *MemoryStore) CreateOvertime(o *model.Overtime) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.overtimes[o.ID]; ok {
		return ErrConflict
	}
	s.overtimes[o.ID] = o
	return nil
}

func (s *MemoryStore) GetOvertime(id string) (*model.Overtime, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	o, ok := s.overtimes[id]
	if !ok {
		return nil, ErrNotFound
	}
	return o, nil
}

func (s *MemoryStore) ListOvertimes() []*model.Overtime {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*model.Overtime, 0, len(s.overtimes))
	for _, o := range s.overtimes {
		list = append(list, o)
	}
	return list
}

func (s *MemoryStore) UpdateOvertime(o *model.Overtime) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.overtimes[o.ID]; !ok {
		return ErrNotFound
	}
	s.overtimes[o.ID] = o
	return nil
}

func (s *MemoryStore) DeleteOvertime(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.overtimes[id]; !ok {
		return ErrNotFound
	}
	return nil
}

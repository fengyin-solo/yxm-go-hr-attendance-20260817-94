package store

import "hrattendance/internal/model"

func (s *MemoryStore) CreateLeave(l *model.Leave) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.leaves[l.ID]; ok {
		return ErrConflict
	}
	s.leaves[l.ID] = l
	return nil
}

func (s *MemoryStore) GetLeave(id string) (*model.Leave, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	l, ok := s.leaves[id]
	if !ok {
		return nil, ErrNotFound
	}
	return l, nil
}

func (s *MemoryStore) ListLeaves() []*model.Leave {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*model.Leave, 0, len(s.leaves))
	for _, l := range s.leaves {
		list = append(list, l)
	}
	return list
}

func (s *MemoryStore) UpdateLeave(l *model.Leave) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.leaves[l.ID]; !ok {
		return ErrNotFound
	}
	s.leaves[l.ID] = l
	return nil
}

func (s *MemoryStore) DeleteLeave(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.leaves[id]; !ok {
		return ErrNotFound
	}
	delete(s.leaves, id)
	return nil
}

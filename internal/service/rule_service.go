package service

import (
	"sort"
	"time"

	"hrattendance/internal/model"
	"hrattendance/pkg/idgen"
)

func (s *Service) CreateRule(rule model.AttendanceRule) (*model.AttendanceRule, error) {
	if err := rule.Validate(); err != nil {
		return nil, err
	}
	now := time.Now()
	rule.ID = idgen.Hex()
	rule.CreatedAt = now
	rule.UpdatedAt = now
	if err := s.store.CreateRule(&rule); err != nil {
		return nil, err
	}
	return &rule, nil
}

func (s *Service) GetRule(id string) (*model.AttendanceRule, error) {
	return s.store.GetRule(id)
}

func (s *Service) ListRules() ([]*model.AttendanceRule, error) {
	all := s.store.ListRules()
	sort.Slice(all, func(i, j int) bool {
		return all[i].CreatedAt.After(all[j].CreatedAt)
	})
	return all, nil
}

func (s *Service) UpdateRule(id string, input model.AttendanceRule) (*model.AttendanceRule, error) {
	existing, err := s.store.GetRule(id)
	if err != nil {
		return nil, err
	}
	if input.Name != "" {
		existing.Name = input.Name
	}
	if input.WorkStart != "" {
		existing.WorkStart = input.WorkStart
	}
	if input.WorkEnd != "" {
		existing.WorkEnd = input.WorkEnd
	}
	if input.LateThreshold > 0 {
		existing.LateThreshold = input.LateThreshold
	}
	existing.Enabled = input.Enabled
	if err := existing.Validate(); err != nil {
		return nil, err
	}
	existing.UpdatedAt = time.Now()
	if err := s.store.UpdateRule(existing); err != nil {
		return nil, err
	}
	return existing, nil
}

func (s *Service) DeleteRule(id string) error {
	return s.store.DeleteRule(id)
}

// ActiveRule 返回当前启用的考勤规则；无启用规则时返回默认规则。
func (s *Service) ActiveRule() *model.AttendanceRule {
	for _, r := range s.store.ListRules() {
		if r.Enabled {
			return r
		}
	}
	return &model.AttendanceRule{
		WorkStart:     s.defaultWorkStart(),
		WorkEnd:       s.defaultWorkEnd(),
		LateThreshold: s.defaultLateThreshold(),
	}
}

func (s *Service) defaultWorkStart() string {
	if s.cfg != nil && s.cfg.DefaultWorkStart != "" {
		return s.cfg.DefaultWorkStart
	}
	return "09:00"
}

func (s *Service) defaultWorkEnd() string {
	if s.cfg != nil && s.cfg.DefaultWorkEnd != "" {
		return s.cfg.DefaultWorkEnd
	}
	return "18:00"
}

func (s *Service) defaultLateThreshold() int {
	if s.cfg != nil && s.cfg.LateThreshold > 0 {
		return s.cfg.LateThreshold
	}
	return 15
}

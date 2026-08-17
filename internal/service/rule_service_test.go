package service

import (
	"testing"

	"hrattendance/internal/model"
)

func TestCreateRuleValidation(t *testing.T) {
	s := newTestService()
	if _, err := s.CreateRule(model.AttendanceRule{Name: ""}); err == nil {
		t.Fatal("空名称规则应被拒绝")
	}
	if _, err := s.CreateRule(model.AttendanceRule{Name: "规则", WorkStart: "09:00", WorkEnd: "08:00"}); err == nil {
		t.Fatal("下班早于上班应被拒绝")
	}
}

func TestActiveRuleReturnsEnabled(t *testing.T) {
	s := newTestService()
	_, _ = s.CreateRule(model.AttendanceRule{Name: "规则A", WorkStart: "08:30", WorkEnd: "17:30", Enabled: false})
	_, _ = s.CreateRule(model.AttendanceRule{Name: "规则B", WorkStart: "09:00", WorkEnd: "18:00", Enabled: true})

	rule := s.ActiveRule()
	if rule.WorkStart != "09:00" || rule.WorkEnd != "18:00" {
		t.Fatalf("应返回启用的规则B，实际 %s-%s", rule.WorkStart, rule.WorkEnd)
	}
}

func TestActiveRuleDefaultWhenNoneEnabled(t *testing.T) {
	s := newTestService()
	_, _ = s.CreateRule(model.AttendanceRule{Name: "规则", WorkStart: "08:30", WorkEnd: "17:30", Enabled: false})

	rule := s.ActiveRule()
	if rule.WorkStart != "09:00" || rule.WorkEnd != "18:00" {
		t.Fatalf("无启用规则时应返回默认规则，实际 %s-%s", rule.WorkStart, rule.WorkEnd)
	}
	if rule.LateThreshold != 15 {
		t.Fatalf("默认迟到阈值应为 15，实际 %d", rule.LateThreshold)
	}
}

func TestUpdateRule(t *testing.T) {
	s := newTestService()
	rule, _ := s.CreateRule(model.AttendanceRule{Name: "规则", WorkStart: "09:00", WorkEnd: "18:00", Enabled: true})

	updated, err := s.UpdateRule(rule.ID, model.AttendanceRule{Name: "新规则", WorkStart: "08:30", LateThreshold: 10})
	if err != nil {
		t.Fatalf("更新规则失败: %v", err)
	}
	if updated.Name != "新规则" || updated.WorkStart != "08:30" || updated.LateThreshold != 10 {
		t.Fatalf("更新结果错误: %+v", updated)
	}
	if updated.WorkEnd != "18:00" {
		t.Fatalf("未修改字段应保留: %s", updated.WorkEnd)
	}
}

func TestListRules(t *testing.T) {
	s := newTestService()
	_, _ = s.CreateRule(model.AttendanceRule{Name: "规则A", WorkStart: "09:00", WorkEnd: "18:00"})
	_, _ = s.CreateRule(model.AttendanceRule{Name: "规则B", WorkStart: "08:30", WorkEnd: "17:30"})

	rules, err := s.ListRules()
	if err != nil {
		t.Fatalf("ListRules 失败: %v", err)
	}
	if len(rules) != 2 {
		t.Fatalf("规则数应为 2，实际 %d", len(rules))
	}
}

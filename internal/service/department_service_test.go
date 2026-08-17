package service

import (
	"testing"

	"hrattendance/internal/model"
)

func TestCreateDepartmentValidation(t *testing.T) {
	s := newTestService()
	if _, err := s.CreateDepartment(model.Department{}); err == nil {
		t.Fatal("空名称部门应被拒绝")
	}
}

func TestCreateDepartmentRequiresParent(t *testing.T) {
	s := newTestService()
	_, err := s.CreateDepartment(model.Department{Name: "子部门", ParentID: "missing"})
	if err == nil {
		t.Fatal("上级部门不存在应被拒绝")
	}
}

func TestDepartmentTree(t *testing.T) {
	s := newTestService()
	parent := mustCreateDepartment(t, s, "总公司")
	child, err := s.CreateDepartment(model.Department{Name: "研发部", ParentID: parent.ID})
	if err != nil {
		t.Fatalf("创建子部门失败: %v", err)
	}
	if child.ParentID != parent.ID {
		t.Fatalf("子部门 ParentID 错误: %s", child.ParentID)
	}
}

func TestDeleteDepartmentWithChildren(t *testing.T) {
	s := newTestService()
	parent := mustCreateDepartment(t, s, "总公司")
	_, _ = s.CreateDepartment(model.Department{Name: "研发部", ParentID: parent.ID})

	if err := s.DeleteDepartment(parent.ID); err == nil {
		t.Fatal("存在子部门的部门应拒绝删除")
	}
}

func TestDeleteDepartmentWithEmployees(t *testing.T) {
	s := newTestService()
	dept := mustCreateDepartment(t, s, "研发部")
	mustCreateEmployee(t, s, dept.ID, "张三", "E001")

	if err := s.DeleteDepartment(dept.ID); err == nil {
		t.Fatal("存在员工的部门应拒绝删除")
	}
}

func TestUpdateDepartmentSelfParent(t *testing.T) {
	s := newTestService()
	dept := mustCreateDepartment(t, s, "研发部")

	if _, err := s.UpdateDepartment(dept.ID, model.Department{ParentID: dept.ID}); err == nil {
		t.Fatal("不能将自己设为上级部门")
	}
}

func TestDepartmentEmployeeCount(t *testing.T) {
	s := newTestService()
	dept := mustCreateDepartment(t, s, "研发部")
	mustCreateEmployee(t, s, dept.ID, "张三", "E001")
	mustCreateEmployee(t, s, dept.ID, "李四", "E002")

	count, err := s.DepartmentEmployeeCount(dept.ID)
	if err != nil {
		t.Fatalf("统计失败: %v", err)
	}
	if count != 2 {
		t.Fatalf("在职员工数应为 2，实际 %d", count)
	}
}

func TestListDepartmentsFilter(t *testing.T) {
	s := newTestService()
	mustCreateDepartment(t, s, "研发部")
	mustCreateDepartment(t, s, "市场部")
	mustCreateDepartment(t, s, "销售部")

	items, total, err := s.ListDepartments(model.DepartmentFilter{Keyword: "研"}, 1, 10)
	if err != nil {
		t.Fatalf("ListDepartments 失败: %v", err)
	}
	if total != 1 || items[0].Name != "研发部" {
		t.Fatalf("关键词筛选结果错误: total=%d", total)
	}
}

package repository

import (
	"testing"

	"license-manager/internal/models"
)

func ptr(s string) *string {
	return &s
}

// buildTreeForTest 暴露被测的树构建逻辑
func buildTreeForTest(menus []models.Menu) []models.Menu {
	r := &rbacRepository{}
	childCount := map[string]int{}
	for _, menu := range menus {
		if menu.ParentID != nil {
			childCount[*menu.ParentID]++
		}
	}
	return r.buildMenuTree(menus, childCount)
}

func TestBuildMenuTree(t *testing.T) {
	menus := []models.Menu{
		{ID: "m1", NameKey: "menu.dashboard", Type: models.MenuTypeMenu, Path: "/dashboard", Status: "active"},
		{ID: "m2", NameKey: "menu.product", Type: models.MenuTypeDirectory, Status: "active"},
		{ID: "m3", NameKey: "menu.product_offering", Type: models.MenuTypeMenu, Path: "/products", ParentID: ptr("m2"), Status: "active"},
		{ID: "m4", NameKey: "menu.feature_template", Type: models.MenuTypeMenu, Path: "/feature-templates", ParentID: ptr("m2"), Status: "active"},
	}

	tree := buildTreeForTest(menus)
	if len(tree) != 2 {
		t.Fatalf("顶级菜单数 = %d, want 2", len(tree))
	}

	// 平铺顺序按 sort_order，m2 应带两个子菜单
	if tree[1].ID != "m2" {
		t.Errorf("第二个顶级菜单 = %q, want m2", tree[1].ID)
	}
	if len(tree[1].Children) != 2 {
		t.Errorf("m2 子菜单数 = %d, want 2", len(tree[1].Children))
	}
}

func TestBuildMenuTreeSkipsInactive(t *testing.T) {
	menus := []models.Menu{
		{ID: "m1", Type: models.MenuTypeMenu, Path: "/a", Status: "active"},
		{ID: "m2", Type: models.MenuTypeMenu, Path: "/b", Status: "inactive"},
		{ID: "m3", Type: models.MenuTypeMenu, Path: "/c", Status: "active", Hidden: true},
	}

	tree := buildTreeForTest(menus)
	if len(tree) != 1 {
		t.Fatalf("应只返回 1 个启用且可见的菜单, got %d", len(tree))
	}
	if tree[0].ID != "m1" {
		t.Errorf("保留下来的菜单 = %q, want m1", tree[0].ID)
	}
}

// 分组下的子菜单全部不可见时，不应显示空目录
func TestBuildMenuTreeDropsEmptyDirectory(t *testing.T) {
	menus := []models.Menu{
		{ID: "dir", Type: models.MenuTypeDirectory, Status: "active"},
		{ID: "hidden-child", Type: models.MenuTypeMenu, Path: "/x", ParentID: ptr("dir"), Status: "inactive"},
	}

	tree := buildTreeForTest(menus)
	if len(tree) != 0 {
		t.Errorf("空目录应被剔除, got %d 项", len(tree))
	}
}

// 父菜单不可见时子菜单不应成为孤儿浮到顶层
func TestBuildMenuTreeOrphanWhenParentHidden(t *testing.T) {
	menus := []models.Menu{
		{ID: "dir", Type: models.MenuTypeDirectory, Status: "inactive"},
		{ID: "child", Type: models.MenuTypeMenu, Path: "/y", ParentID: ptr("dir"), Status: "active"},
	}

	tree := buildTreeForTest(menus)
	for _, item := range tree {
		if item.ID == "child" {
			t.Error("父菜单不可见时子菜单不应浮到顶层")
		}
	}
}

// 环形引用不得导致无限递归
func TestBuildMenuTreeHandlesCycle(t *testing.T) {
	menus := []models.Menu{
		{ID: "a", Type: models.MenuTypeMenu, Path: "/a", ParentID: ptr("b"), Status: "active"},
		{ID: "b", Type: models.MenuTypeMenu, Path: "/b", ParentID: ptr("a"), Status: "active"},
	}

	// 不应 panic 或挂起：环形数据下没有根节点，结果应为空
	tree := buildTreeForTest(menus)
	if len(tree) != 0 {
		t.Errorf("环形数据不应产生顶级菜单, got %d", len(tree))
	}
}

// 分组内的菜单自身若为分组，其子菜单应递归构建
func TestBuildMenuTreeNestedDirectories(t *testing.T) {
	menus := []models.Menu{
		{ID: "l1", Type: models.MenuTypeDirectory, Status: "active"},
		{ID: "l2", Type: models.MenuTypeDirectory, ParentID: ptr("l1"), Status: "active"},
		{ID: "leaf", Type: models.MenuTypeMenu, Path: "/leaf", ParentID: ptr("l2"), Status: "active"},
	}

	tree := buildTreeForTest(menus)
	if len(tree) != 1 || tree[0].ID != "l1" {
		t.Fatalf("顶级应为 l1 一个目录, got %d", len(tree))
	}
	if len(tree[0].Children) != 1 || tree[0].Children[0].ID != "l2" {
		t.Fatalf("l1 下应有 l2, got %+v", tree[0].Children)
	}
	if len(tree[0].Children[0].Children) != 1 || tree[0].Children[0].Children[0].ID != "leaf" {
		t.Errorf("l2 下应有 leaf, got %+v", tree[0].Children[0].Children)
	}
}

// 非系统菜单且无子菜单时可删除
func TestBuildMenuTreeCanDeleteFlag(t *testing.T) {
	menus := []models.Menu{
		{ID: "sys", Type: models.MenuTypeMenu, Path: "/s", Status: "active", IsSystem: true},
		{ID: "custom", Type: models.MenuTypeMenu, Path: "/c", Status: "active", IsSystem: false},
		{ID: "parent", Type: models.MenuTypeMenu, Path: "/p", Status: "active", IsSystem: false},
		{ID: "child", Type: models.MenuTypeMenu, Path: "/pc", ParentID: ptr("parent"), Status: "active", IsSystem: false},
	}

	tree := buildTreeForTest(menus)
	for _, item := range tree {
		switch item.ID {
		case "sys":
			if item.CanDelete {
				t.Error("系统菜单不应可删除")
			}
		case "custom":
			if !item.CanDelete {
				t.Error("无子菜单的自定义菜单应可删除")
			}
		case "parent":
			if item.CanDelete {
				t.Error("有子菜单的菜单不应可删除")
			}
		}
	}
}

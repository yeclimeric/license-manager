package service

import (
	"testing"

	"license-manager/internal/models"
)

// resolveRoleNames 决定实际分配哪些角色。
//
// 这是 v1 与 v2 两种调用方式的汇合点：旧调用方只传 role，
// 新调用方传 role_assignments，两者都要落到同一份角色列表上。
func TestResolveRoleNames(t *testing.T) {
	cases := []struct {
		name        string
		assignments []models.UserRoleAssignmentInput
		legacyRole  string
		want        []string
	}{
		{
			name:        "显式分配优先",
			assignments: []models.UserRoleAssignmentInput{{Role: "viewer"}, {Role: "operator"}},
			legacyRole:  "admin",
			want:        []string{"viewer", "operator"},
		},
		{
			name:       "未分配时回退到 v1 的 role 映射",
			legacyRole: "admin",
			want:       []string{"super_admin"},
		},
		{
			name:       "operator 映射到同名角色",
			legacyRole: "operator",
			want:       []string{"operator"},
		},
		{
			name:       "viewer 映射到同名角色",
			legacyRole: "viewer",
			want:       []string{"viewer"},
		},
		{
			name:        "分配项全为空白时回退",
			assignments: []models.UserRoleAssignmentInput{{Role: "  "}},
			legacyRole:  "admin",
			want:        []string{"super_admin"},
		},
		{
			name:        "空白项被剔除，保留有效项",
			assignments: []models.UserRoleAssignmentInput{{Role: ""}, {Role: "viewer"}},
			want:        []string{"viewer"},
		},
		{
			name:        "分配项前后空白被裁剪",
			assignments: []models.UserRoleAssignmentInput{{Role: "  operator  "}},
			want:        []string{"operator"},
		},
		{
			name: "未知的 v1 角色不产生分配",
			// 未知角色名查库会失败，不如不分配并让创建请求报错
			legacyRole: "nonexistent",
			want:       nil,
		},
		{
			name: "两者都为空",
			want: nil,
		},
		{
			name:        "空切片不同于 nil",
			assignments: []models.UserRoleAssignmentInput{},
			want:        nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := resolveRoleNames(tc.assignments, tc.legacyRole)
			if len(got) != len(tc.want) {
				t.Fatalf("resolveRoleNames() = %v, want %v", got, tc.want)
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Errorf("第 %d 项 = %q, want %q", i, got[i], tc.want[i])
				}
			}
		})
	}
}

// 同一角色重复出现不应导致分配重复关联
func TestResolveRoleNamesDuplicate(t *testing.T) {
	got := resolveRoleNames([]models.UserRoleAssignmentInput{
		{Role: "viewer"},
		{Role: "viewer"},
	}, "")

	if len(got) != 2 {
		t.Fatalf("resolveRoleNames 未做去重，返回 %v", got)
	}
	// 去重发生在仓储层的 resolveRolesByNames，此处只验证透传
}

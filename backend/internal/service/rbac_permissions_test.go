package service

import (
	"context"
	"errors"
	"testing"

	"license-manager/internal/repository"
	pkgcontext "license-manager/pkg/context"
	"license-manager/pkg/i18n"
)

// 嵌入接口即可只覆盖被测方法，其余方法无需实现。
type permissionsRepoStub struct {
	repository.RBACRepository
	permissions []string
	err         error
}

func (s *permissionsRepoStub) GetUserPermissions(context.Context, string) ([]string, error) {
	return s.permissions, s.err
}

func TestGetCurrentUserPermissions(t *testing.T) {
	tests := []struct {
		name        string
		loggedIn    bool
		permissions []string
		repoErr     error
		want        []string
		wantErrCode string
	}{
		{name: "super admin wildcard", loggedIn: true, permissions: []string{"*"}, want: []string{"*"}},
		{name: "operator scoped", loggedIn: true, permissions: []string{"license:read", "customer:read"}, want: []string{"license:read", "customer:read"}},
		{name: "nil normalized to empty", loggedIn: true, permissions: nil, want: []string{}},
		{name: "not logged in returns empty without hitting repo", want: []string{}},
		{name: "repo failure", loggedIn: true, repoErr: errors.New("db down"), wantErrCode: "900005"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &rbacService{repo: &permissionsRepoStub{permissions: tt.permissions, err: tt.repoErr}}

			ctx := context.Background()
			if tt.loggedIn {
				ctx = pkgcontext.WithUserID(ctx, "user-1")
			}

			got, err := svc.GetCurrentUserPermissions(ctx)

			if tt.wantErrCode != "" {
				var i18nErr *i18n.I18nError
				if !errors.As(err, &i18nErr) || i18nErr.Code != tt.wantErrCode {
					t.Fatalf("got err %v, want I18nError %s", err, tt.wantErrCode)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(got) != len(tt.want) {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
			for i := range tt.want {
				if got[i] != tt.want[i] {
					t.Fatalf("got %v, want %v", got, tt.want)
				}
			}
		})
	}
}

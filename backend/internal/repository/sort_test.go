package repository

import "testing"

func TestSafeOrderClauseAllowedColumns(t *testing.T) {
	allowed := []string{"created_at", "updated_at", "customer_name", "customer_code"}

	tests := []struct {
		name      string
		column    string
		direction string
		want      string
	}{
		{"default column and order", "", "", "created_at DESC"},
		{"explicit desc", "updated_at", "desc", "updated_at DESC"},
		{"asc is case insensitive", "customer_name", "ASC", "customer_name ASC"},
		{"default column selected explicitly", "created_at", "asc", "created_at ASC"},
		{"unknown column falls back", "no_such_column", "asc", "created_at ASC"},
		{"unknown direction falls back to DESC", "updated_at", "sideways", "updated_at DESC"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SafeOrderClause(tt.column, tt.direction, "", allowed...); got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

// 回归：任何注入载荷都不得原样进入 ORDER BY 子句。
func TestSafeOrderClauseRejectsInjection(t *testing.T) {
	allowed := []string{"created_at", "updated_at", "activated_at", "last_heartbeat"}
	payloads := []string{
		"created_at; DROP TABLE licenses; --",
		"created_at ASC, (SELECT 1)",
		"1=1",
		"created_at`",
		"updated_at UNION SELECT password FROM users",
		"created_at/*comment*/",
		" CREATED_AT ",
	}

	for _, payload := range payloads {
		got := SafeOrderClause(payload, "asc", "licenses.", allowed...)
		if got != "licenses.created_at ASC" {
			t.Errorf("payload %q produced unsafe clause %q", payload, got)
		}
	}
	// 方向字段同样是拼接点，注入也必须被规整为 ASC/DESC
	if got := SafeOrderClause("created_at", "DESC; DROP TABLE users", "", allowed...); got != "created_at DESC" {
		t.Errorf("direction injection produced %q", got)
	}
}

func TestSafeOrderClausePrefix(t *testing.T) {
	got := SafeOrderClause("change_type", "desc", "authorization_changes.", "created_at", "change_type")
	if got != "authorization_changes.change_type DESC" {
		t.Errorf("got %q, want authorization_changes.change_type DESC", got)
	}
}

func TestSafeOrderClauseEmptyAllowlist(t *testing.T) {
	if got := SafeOrderClause("created_at", "asc", ""); got != "" {
		t.Errorf("empty allowlist should produce empty clause, got %q", got)
	}
}

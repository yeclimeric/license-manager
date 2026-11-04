package repository

import "testing"

func TestNormalizePage(t *testing.T) {
	tests := []struct {
		name         string
		page, size   int
		wantPage     int
		wantPageSize int
	}{
		{"both valid", 3, 50, 3, 50},
		{"defaults", 0, 0, 1, DefaultPageSize},
		{"negative values", -5, -1, 1, DefaultPageSize},
		{"over max is clamped", 2, 500, 2, MaxPageSize},
		{"exactly max is kept", 1, 100, 1, 100},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			page, size := NormalizePage(tt.page, tt.size)
			if page != tt.wantPage || size != tt.wantPageSize {
				t.Errorf("NormalizePage(%d,%d) = (%d,%d), want (%d,%d)",
					tt.page, tt.size, page, size, tt.wantPage, tt.wantPageSize)
			}
		})
	}
}

// 回归：service 与 repository 用同一函数后，page_size=500 的响应
// 不再出现「按 100 查询、按 500 算页数」的分裂。
func TestNormalizePageServiceParity(t *testing.T) {
	_, repoSize := NormalizePage(1, 500)
	_, serviceSize := NormalizePage(1, 500)
	if repoSize != serviceSize || repoSize != MaxPageSize {
		t.Fatalf("repo/service page size diverged: %d vs %d", repoSize, serviceSize)
	}
	if pages := CalcTotalPages(250, repoSize); pages != 3 {
		t.Errorf("total 250 / size 100 should be 3 pages, got %d", pages)
	}
}

func TestCalcTotalPages(t *testing.T) {
	tests := []struct {
		name     string
		total    int64
		pageSize int
		want     int
	}{
		{"empty", 0, 20, 0},
		{"zero page size must not panic", 100, 0, 0},
		{"negative page size", 100, -1, 0},
		{"exact multiple", 100, 20, 5},
		{"needs rounding up", 101, 20, 6},
		{"single row", 1, 100, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CalcTotalPages(tt.total, tt.pageSize); got != tt.want {
				t.Errorf("CalcTotalPages(%d,%d) = %d, want %d", tt.total, tt.pageSize, got, tt.want)
			}
		})
	}
}

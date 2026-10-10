package repository

import (
	"testing"
	"time"
)

func TestBuildAuthorizationTrendDataFillsDatesAndUsesTimezone(t *testing.T) {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		t.Fatalf("load timezone: %v", err)
	}
	start := time.Date(2026, 8, 1, 0, 0, 0, 0, loc)
	endExclusive := start.AddDate(0, 0, 3)
	createdTimes := []time.Time{
		time.Date(2026, 7, 31, 16, 30, 0, 0, time.UTC),
		time.Date(2026, 8, 2, 10, 0, 0, 0, loc),
		time.Date(2026, 8, 2, 20, 0, 0, 0, loc),
	}
	expiredTimes := []time.Time{
		time.Date(2026, 8, 3, 1, 0, 0, 0, loc),
	}

	trend := buildAuthorizationTrendData(start, endExclusive, 4, createdTimes, expiredTimes)
	if len(trend) != 3 {
		t.Fatalf("expected 3 days, got %d", len(trend))
	}
	wantTotal := []int64{5, 7, 7}
	wantNew := []int64{1, 2, 0}
	wantExpired := []int64{0, 0, 1}
	for i := range trend {
		if trend[i].TotalAuthorizations != wantTotal[i] || trend[i].NewAuthorizations != wantNew[i] || trend[i].ExpiredAuthorizations != wantExpired[i] {
			t.Fatalf("day %d: unexpected trend point: %+v", i, trend[i])
		}
	}
}

func TestBuildDashboardTrendPointsFillsDatesAndUsesTimezone(t *testing.T) {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		t.Fatalf("load timezone: %v", err)
	}
	start := time.Date(2026, 8, 1, 0, 0, 0, 0, loc)
	endExclusive := start.AddDate(0, 0, 3)
	events := []time.Time{
		time.Date(2026, 7, 31, 16, 30, 0, 0, time.UTC),
		time.Date(2026, 8, 2, 10, 0, 0, 0, loc),
		time.Date(2026, 8, 2, 20, 0, 0, 0, loc),
	}

	points := buildDashboardTrendPoints(start, endExclusive, events)
	if len(points) != 3 {
		t.Fatalf("expected 3 days, got %d", len(points))
	}
	want := []int64{1, 2, 0}
	for i := range points {
		if points[i].Count != want[i] {
			t.Fatalf("day %d: got %d, want %d", i, points[i].Count, want[i])
		}
	}
}

// resolvedExpiry 决定列表展示的到期时间。
//
// 从激活起算的授权码其 end_date 只是硬上限，展示时不能让人误以为是某台设备的实际到期日。
func TestResolvedExpiry(t *testing.T) {
	endDate := time.Date(2027, 10, 8, 23, 59, 59, 0, time.UTC)

	t.Run("有 expires_at 时优先", func(t *testing.T) {
		expires := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)
		got := resolvedExpiry("activation_based", endDate, &expires)
		if !got.Equal(expires) {
			t.Errorf("got %v, want %v", got, expires)
		}
	})

	t.Run("expires_at 为零值时回退到 end_date", func(t *testing.T) {
		zero := time.Time{}
		got := resolvedExpiry("fixed", endDate, &zero)
		if !got.Equal(endDate) {
			t.Errorf("got %v, want %v", got, endDate)
		}
	})

	t.Run("expires_at 为 nil 时回退到 end_date", func(t *testing.T) {
		got := resolvedExpiry("fixed", endDate, nil)
		if !got.Equal(endDate) {
			t.Errorf("got %v, want %v", got, endDate)
		}
	})
}

// 固定期限模式下 expires_at 应与 end_date 一致，两种模式都不应返回零值
func TestResolvedExpiryNeverZero(t *testing.T) {
	endDate := time.Date(2027, 10, 8, 23, 59, 59, 0, time.UTC)
	for _, mode := range []string{"fixed", "activation_based"} {
		if got := resolvedExpiry(mode, endDate, nil); got.IsZero() {
			t.Errorf("mode=%s 返回了零值到期时间", mode)
		}
	}
}

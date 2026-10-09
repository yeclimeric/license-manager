package service

import (
	"testing"
	"time"

	"license-manager/internal/models"
)

func TestResolveLicensePeriodFixed(t *testing.T) {
	now := time.Now()
	start := now.Add(-24 * time.Hour)
	end := now.Add(365 * 24 * time.Hour)

	license := &models.License{ActivatedAt: &now}
	authCode := &models.AuthorizationCode{
		ValidityMode: ValidityModeFixed,
		StartDate:    start,
		EndDate:      end,
	}

	gotStart, gotEnd := resolveLicensePeriod(license, authCode)
	if !gotStart.Equal(start) {
		t.Errorf("start = %v, want %v", gotStart, start)
	}
	if !gotEnd.Equal(end) {
		t.Errorf("end = %v, want %v", gotEnd, end)
	}
}

// 从激活起算：结束时间应为激活时刻 + duration_days
func TestResolveLicensePeriodActivationBased(t *testing.T) {
	activatedAt := time.Date(2026, 10, 9, 10, 0, 0, 0, time.UTC)
	license := &models.License{ActivatedAt: &activatedAt}

	duration := 730
	authCode := &models.AuthorizationCode{
		ValidityMode: ValidityModeActivationBased,
		DurationDays: &duration,
		StartDate:    time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		// 授权码自身还有三年才失效，不应成为上限
		EndDate: time.Date(2029, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	gotStart, gotEnd := resolveLicensePeriod(license, authCode)

	if !gotStart.Equal(activatedAt) {
		t.Errorf("start = %v, want 激活时刻 %v", gotStart, activatedAt)
	}
	wantEnd := activatedAt.AddDate(0, 0, 730)
	if !gotEnd.Equal(wantEnd) {
		t.Errorf("end = %v, want %v", gotEnd, wantEnd)
	}
}

// 授权码自身失效时间更早时，应以授权码为准
func TestResolveLicensePeriodActivationBasedCappedByAuthCode(t *testing.T) {
	activatedAt := time.Date(2026, 10, 9, 10, 0, 0, 0, time.UTC)
	license := &models.License{ActivatedAt: &activatedAt}

	duration := 3650
	authCode := &models.AuthorizationCode{
		ValidityMode: ValidityModeActivationBased,
		DurationDays: &duration,
		EndDate:      time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC),
	}

	_, gotEnd := resolveLicensePeriod(license, authCode)
	if !gotEnd.Equal(authCode.EndDate) {
		t.Errorf("end = %v, 应被授权码 EndDate %v 封顶", gotEnd, authCode.EndDate)
	}
}

// 未指定天数时不得算出"激活即过期"
func TestResolveLicensePeriodActivationBasedMissingDays(t *testing.T) {
	activatedAt := time.Date(2026, 10, 9, 10, 0, 0, 0, time.UTC)
	license := &models.License{ActivatedAt: &activatedAt}

	authCode := &models.AuthorizationCode{
		ValidityMode: ValidityModeActivationBased,
		DurationDays: nil,
		EndDate:      time.Date(2029, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	_, gotEnd := resolveLicensePeriod(license, authCode)
	if !gotEnd.After(activatedAt) {
		t.Errorf("end = %v, 未指定天数时不应等于或早于激活时刻 %v", gotEnd, activatedAt)
	}
}

// 天数为 0 或负数属于脏数据，至少给一天
func TestResolveLicensePeriodActivationBasedNonPositiveDays(t *testing.T) {
	activatedAt := time.Date(2026, 10, 9, 10, 0, 0, 0, time.UTC)
	license := &models.License{ActivatedAt: &activatedAt}

	zero := 0
	authCode := &models.AuthorizationCode{
		ValidityMode: ValidityModeActivationBased,
		DurationDays: &zero,
		EndDate:      time.Date(2029, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	_, gotEnd := resolveLicensePeriod(license, authCode)
	want := activatedAt.AddDate(0, 0, 1)
	if !gotEnd.Equal(want) {
		t.Errorf("end = %v, want %v", gotEnd, want)
	}
}

// 未激活（ActivatedAt 为 nil）时不应 panic
func TestResolveLicensePeriodNotActivated(t *testing.T) {
	license := &models.License{}
	duration := 365
	authCode := &models.AuthorizationCode{
		ValidityMode: ValidityModeActivationBased,
		DurationDays: &duration,
		EndDate:      time.Date(2029, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	_, gotEnd := resolveLicensePeriod(license, authCode)
	if gotEnd.IsZero() {
		t.Error("未激活时 end 不应为零值")
	}
}

// 无授权码关联时给出宽松期限，不阻断流程
func TestResolveLicensePeriodNoAuthCode(t *testing.T) {
	license := &models.License{}
	_, gotEnd := resolveLicensePeriod(license, nil)
	if !gotEnd.After(time.Now()) {
		t.Errorf("end = %v, 应晚于当前时间", gotEnd)
	}
}

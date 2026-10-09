package service

import (
	"time"

	"license-manager/internal/models"
)

// 有效期模式常量
const (
	// ValidityModeFixed 固定期限：起止时间在创建授权码时即确定
	ValidityModeFixed = "fixed"
	// ValidityModeActivationBased 从激活起算：每台设备各自从激活时刻起算有效期
	ValidityModeActivationBased = "activation_based"
)

// resolveLicensePeriod 计算许可证的实际起止时间。
//
// 固定期限直接采用授权码的起止时间；从激活起算的模式下，
// 授权码的 EndDate 并不代表该设备的到期时间——每台设备的激活时刻不同，
// 必须以该许可证自身的激活时间为起点另算。
//
// 返回的时间落在签名载荷里，客户端据此判定是否过期，因此口径必须唯一。
func resolveLicensePeriod(license *models.License, authCode *models.AuthorizationCode) (start time.Time, end time.Time) {
	if authCode == nil {
		return time.Now(), time.Now().AddDate(100, 0, 0)
	}

	start = authCode.StartDate
	end = authCode.EndDate

	if authCode.ValidityMode != ValidityModeActivationBased {
		return start, end
	}

	// 从激活起算：以本许可证的激活时刻为起点
	if license.ActivatedAt != nil {
		start = *license.ActivatedAt
	}

	days := authCode.DurationDays
	if days == nil || *days <= 0 {
		// 未指定天数时沿用授权码的剩余天数，至少给一天，
		// 否则激活瞬间即过期
		if days == nil {
			days = intPtr(365)
		} else {
			days = intPtr(1)
		}
	}
	end = start.AddDate(0, 0, *days)

	// 授权码自身的失效时间是硬上限：授权码过期后，
	// 从激活起算的许可证也不能活得比它更久
	if !authCode.EndDate.IsZero() && end.After(authCode.EndDate) {
		end = authCode.EndDate
	}

	return start, end
}

// intPtr 返回整数指针
func intPtr(v int) *int {
	return &v
}

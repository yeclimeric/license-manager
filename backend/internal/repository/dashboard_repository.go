package repository

import (
	"context"
	"time"

	"license-manager/internal/models"
	pkgcontext "license-manager/pkg/context"
	"license-manager/pkg/i18n"

	"gorm.io/gorm"
)

type dashboardRepository struct {
	db *gorm.DB
}

// NewDashboardRepository 创建仪表盘数据访问层实例
func NewDashboardRepository(db *gorm.DB) DashboardRepository {
	return &dashboardRepository{
		db: db,
	}
}

// GetHomeData 获取首页概览、到期提醒和最近记录。
func (r *dashboardRepository) GetHomeData(ctx context.Context, now time.Time, recentLimit int) (*models.DashboardHomeResponse, error) {
	lang := pkgcontext.GetLanguageFromContext(ctx)
	if recentLimit <= 0 {
		recentLimit = 5
	}

	activeLicenseSubquery := `
		SELECT authorization_code_id, COUNT(*) AS active_count
		FROM licenses
		WHERE status = 'active' AND deleted_at IS NULL
		GROUP BY authorization_code_id`

	type overviewResult struct {
		ValidTotal               int64 `gorm:"column:valid_total"`
		ValidActivated           int64 `gorm:"column:valid_activated"`
		ActivatedDevices         int64 `gorm:"column:activated_devices"`
		ManualDevices            int64 `gorm:"column:manual_devices"`
		TrialDevices             int64 `gorm:"column:trial_devices"`
		RemainingActivationSlots int64 `gorm:"column:remaining_activation_slots"`
	}
	var overview overviewResult
	if err := r.db.WithContext(ctx).Table("authorization_codes ac").
		Select(`COUNT(*) AS valid_total,
			COALESCE(SUM(CASE WHEN COALESCE(al.active_count, 0) > 0 THEN 1 ELSE 0 END), 0) AS valid_activated,
			(SELECT COUNT(*) FROM licenses l
				WHERE l.status = 'active' AND l.deleted_at IS NULL) AS activated_devices,
			-- 手工/体验拆分：体验型产品项下的许可证不占用客户名额
			(SELECT COUNT(*) FROM licenses l
				LEFT JOIN authorization_codes ac2 ON ac2.id = l.authorization_code_id
				LEFT JOIN product_offerings po ON po.id = ac2.product_offering_id
				WHERE l.status = 'active' AND l.deleted_at IS NULL
					AND COALESCE(po.is_trial_product, 0) = 0) AS manual_devices,
			(SELECT COUNT(*) FROM licenses l
				LEFT JOIN authorization_codes ac2 ON ac2.id = l.authorization_code_id
				LEFT JOIN product_offerings po ON po.id = ac2.product_offering_id
				WHERE l.status = 'active' AND l.deleted_at IS NULL
					AND COALESCE(po.is_trial_product, 0) = 1) AS trial_devices,
			COALESCE(SUM(GREATEST(ac.max_activations - COALESCE(al.active_count, 0), 0)), 0) AS remaining_activation_slots`).
		Joins("LEFT JOIN ("+activeLicenseSubquery+") al ON al.authorization_code_id = ac.id").
		Where("ac.is_locked = ? AND ac.end_date >= ?", false, now).
		Scan(&overview).Error; err != nil {
		return nil, i18n.NewI18nError("900004", lang, err.Error())
	}

	due7End := now.AddDate(0, 0, 7)
	due30End := now.AddDate(0, 0, 30)
	type expiryResult struct {
		Due7Authorizations     int64 `gorm:"column:due_7_authorizations"`
		Due7Devices            int64 `gorm:"column:due_7_devices"`
		Due8To30Authorizations int64 `gorm:"column:due_8_to_30_authorizations"`
		Due8To30Devices        int64 `gorm:"column:due_8_to_30_devices"`
		ExpiredAuthorizations  int64 `gorm:"column:expired_authorizations"`
		ExpiredDevices         int64 `gorm:"column:expired_devices"`
	}
	var expiry expiryResult
	if err := r.db.WithContext(ctx).Table("authorization_codes ac").
		Select(`COALESCE(SUM(CASE WHEN ac.end_date > ? AND ac.end_date <= ? THEN 1 ELSE 0 END), 0) AS due_7_authorizations,
			COALESCE(SUM(CASE WHEN ac.end_date > ? AND ac.end_date <= ? THEN COALESCE(al.active_count, 0) ELSE 0 END), 0) AS due_7_devices,
			COALESCE(SUM(CASE WHEN ac.end_date > ? AND ac.end_date <= ? THEN 1 ELSE 0 END), 0) AS due_8_to_30_authorizations,
			COALESCE(SUM(CASE WHEN ac.end_date > ? AND ac.end_date <= ? THEN COALESCE(al.active_count, 0) ELSE 0 END), 0) AS due_8_to_30_devices,
			COALESCE(SUM(CASE WHEN ac.end_date < ? THEN 1 ELSE 0 END), 0) AS expired_authorizations,
			COALESCE(SUM(CASE WHEN ac.end_date < ? THEN COALESCE(al.active_count, 0) ELSE 0 END), 0) AS expired_devices`,
			now, due7End, now, due7End, due7End, due30End, due7End, due30End, now, now).
		Joins("LEFT JOIN ("+activeLicenseSubquery+") al ON al.authorization_code_id = ac.id").
		Where("ac.is_locked = ?", false).
		Scan(&expiry).Error; err != nil {
		return nil, i18n.NewI18nError("900004", lang, err.Error())
	}

	recentAuthorizations, err := r.getHomeRecentAuthorizations(ctx, now, recentLimit)
	if err != nil {
		return nil, i18n.NewI18nError("900004", lang, err.Error())
	}
	recentActivations, err := r.getHomeRecentActivations(ctx, recentLimit)
	if err != nil {
		return nil, i18n.NewI18nError("900004", lang, err.Error())
	}

	return &models.DashboardHomeResponse{
		GeneratedAt: now,
		Overview: models.DashboardOverview{
			ValidAuthorizations: models.DashboardValidAuthorizations{
				Total:        overview.ValidTotal,
				Activated:    overview.ValidActivated,
				NotActivated: overview.ValidTotal - overview.ValidActivated,
			},
			ActivatedDevices: models.DashboardActivatedDevices{
				Total:  overview.ActivatedDevices,
				Manual: overview.ManualDevices,
				Trial:  overview.TrialDevices,
			},
			RemainingActivationSlots: overview.RemainingActivationSlots,
		},
		ExpiryReminders: models.DashboardExpiryReminders{
			Due7Days:     models.DashboardExpiryReminder{AuthorizationCount: expiry.Due7Authorizations, AffectedDeviceCount: expiry.Due7Devices},
			Due8To30Days: models.DashboardExpiryReminder{AuthorizationCount: expiry.Due8To30Authorizations, AffectedDeviceCount: expiry.Due8To30Devices},
			Expired:      models.DashboardExpiryReminder{AuthorizationCount: expiry.ExpiredAuthorizations, AffectedDeviceCount: expiry.ExpiredDevices},
		},
		RecentAuthorizations: recentAuthorizations,
		RecentActivations:    recentActivations,
	}, nil
}

func (r *dashboardRepository) getHomeRecentAuthorizations(ctx context.Context, now time.Time, limit int) ([]models.RecentAuthorization, error) {
	type row struct {
		ID                  string
		Code                string
		CustomerID          *string
		CustomerName        string
		Description         string
		StartDate           time.Time
		EndDate             time.Time
		MaxActivations      int
		CurrentActivations  int
		ValidityMode        string
		ExpiresAt           *time.Time
		ProductOfferingName string
		CreatedAt           time.Time
		UpdatedAt           time.Time
		IsLocked            bool
	}
	var rows []row
	err := r.db.WithContext(ctx).Table("authorization_codes ac").
		Select(`ac.id, ac.code, ac.customer_id, COALESCE(c.customer_name, '') AS customer_name,
			COALESCE(ac.description, '') AS description, ac.start_date, ac.end_date, ac.max_activations,
			COALESCE(al.active_count, 0) AS current_activations, ac.validity_mode, ac.end_date AS expires_at,
			COALESCE(po.name, '') AS product_offering_name,
			ac.created_at, ac.updated_at, ac.is_locked`).
		Joins("LEFT JOIN customers c ON c.id = ac.customer_id").
		Joins("LEFT JOIN product_offerings po ON po.id = ac.product_offering_id").
		Joins(`LEFT JOIN (
			SELECT authorization_code_id, COUNT(*) AS active_count
			FROM licenses WHERE status = 'active' AND deleted_at IS NULL
			GROUP BY authorization_code_id
		) al ON al.authorization_code_id = ac.id`).
		Order("ac.created_at DESC").Limit(limit).Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	lang := pkgcontext.GetLanguageFromContext(ctx)
	items := make([]models.RecentAuthorization, 0, len(rows))
	for _, item := range rows {
		status := "normal"
		if item.IsLocked {
			status = "locked"
		} else if item.EndDate.Before(now) {
			status = "expired"
		}
		items = append(items, models.RecentAuthorization{
			ID: item.ID, Code: item.Code, CustomerID: item.CustomerID, CustomerName: item.CustomerName,
			Description: item.Description, Status: status,
			StatusDisplay: i18n.GetEnumMessage("authorization_code_status", status, lang),
			StartDate:     item.StartDate, EndDate: item.EndDate, MaxActivations: item.MaxActivations,
			CurrentActivations:  item.CurrentActivations,
			ValidityMode:        item.ValidityMode,
			ExpiresAt:           resolvedExpiry(item.ValidityMode, item.EndDate, item.ExpiresAt),
			ProductOfferingName: item.ProductOfferingName,
			CreatedAt:           item.CreatedAt, UpdatedAt: item.UpdatedAt,
		})
	}
	return items, nil
}

func (r *dashboardRepository) getHomeRecentActivations(ctx context.Context, limit int) ([]models.DashboardRecentActivation, error) {
	items := make([]models.DashboardRecentActivation, 0)
	err := r.db.WithContext(ctx).Table("licenses l").
		Select(`l.id, l.authorization_code_id, COALESCE(c.customer_name, '') AS customer_name,
			COALESCE(ac.description, '') AS description, l.hardware_fingerprint, l.activated_at, ac.end_date`).
		Joins("JOIN authorization_codes ac ON ac.id = l.authorization_code_id").
		Joins("LEFT JOIN customers c ON c.id = ac.customer_id").
		Where("l.deleted_at IS NULL AND l.activated_at IS NOT NULL").
		Order("l.activated_at DESC").Limit(limit).Scan(&items).Error
	return items, err
}

// GetBusinessTrendData 获取两类业务的每日趋势。
func (r *dashboardRepository) GetBusinessTrendData(ctx context.Context, startDate, endDate time.Time) ([]models.DashboardTrendPoint, []models.DashboardTrendPoint, error) {
	lang := pkgcontext.GetLanguageFromContext(ctx)
	rangeStart := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, startDate.Location())
	rangeEndExclusive := time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 0, 0, 0, 0, endDate.Location()).AddDate(0, 0, 1)

	var authorizationTimes []time.Time
	if err := r.db.WithContext(ctx).Model(&models.AuthorizationCode{}).
		Where("created_at >= ? AND created_at < ?", rangeStart, rangeEndExclusive).
		Pluck("created_at", &authorizationTimes).Error; err != nil {
		return nil, nil, i18n.NewI18nError("900004", lang, err.Error())
	}

	var activationTimes []time.Time
	if err := r.db.WithContext(ctx).Model(&models.License{}).
		Where("activated_at IS NOT NULL AND activated_at >= ? AND activated_at < ?", rangeStart, rangeEndExclusive).
		Pluck("activated_at", &activationTimes).Error; err != nil {
		return nil, nil, i18n.NewI18nError("900004", lang, err.Error())
	}

	return buildDashboardTrendPoints(rangeStart, rangeEndExclusive, authorizationTimes),
		buildDashboardTrendPoints(rangeStart, rangeEndExclusive, activationTimes), nil
}

func buildDashboardTrendPoints(startDate, endDateExclusive time.Time, eventTimes []time.Time) []models.DashboardTrendPoint {
	counts := make(map[string]int64)
	loc := startDate.Location()
	for _, eventTime := range eventTimes {
		counts[eventTime.In(loc).Format("2006-01-02")]++
	}
	points := make([]models.DashboardTrendPoint, 0)
	for date := startDate; date.Before(endDateExclusive); date = date.AddDate(0, 0, 1) {
		key := date.Format("2006-01-02")
		points = append(points, models.DashboardTrendPoint{Date: key, Count: counts[key]})
	}
	return points
}

// GetAuthorizationTrendData 获取授权趋势数据
// 使用固定次数的范围查询，避免按日期循环访问数据库。
func (r *dashboardRepository) GetAuthorizationTrendData(ctx context.Context, startDate, endDate time.Time) ([]models.TrendData, error) {
	lang := pkgcontext.GetLanguageFromContext(ctx)
	rangeStart := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, startDate.Location())
	rangeEndExclusive := time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 0, 0, 0, 0, endDate.Location()).AddDate(0, 0, 1)

	var initialCount int64
	if err := r.db.WithContext(ctx).Model(&models.AuthorizationCode{}).
		Where("created_at < ?", rangeStart).
		Count(&initialCount).Error; err != nil {
		return nil, i18n.NewI18nError("900004", lang, err.Error())
	}

	var createdTimes []time.Time
	if err := r.db.WithContext(ctx).Model(&models.AuthorizationCode{}).
		Where("created_at >= ? AND created_at < ?", rangeStart, rangeEndExclusive).
		Pluck("created_at", &createdTimes).Error; err != nil {
		return nil, i18n.NewI18nError("900004", lang, err.Error())
	}

	var expiredTimes []time.Time
	if err := r.db.WithContext(ctx).Model(&models.AuthorizationCode{}).
		Where("end_date >= ? AND end_date < ?", rangeStart, rangeEndExclusive).
		Pluck("end_date", &expiredTimes).Error; err != nil {
		return nil, i18n.NewI18nError("900004", lang, err.Error())
	}

	return buildAuthorizationTrendData(rangeStart, rangeEndExclusive, initialCount, createdTimes, expiredTimes), nil
}

func buildAuthorizationTrendData(startDate, endDateExclusive time.Time, initialCount int64, createdTimes, expiredTimes []time.Time) []models.TrendData {
	newByDate := make(map[string]int64)
	expiredByDate := make(map[string]int64)
	loc := startDate.Location()
	for _, createdAt := range createdTimes {
		newByDate[createdAt.In(loc).Format("2006-01-02")]++
	}
	for _, expiredAt := range expiredTimes {
		expiredByDate[expiredAt.In(loc).Format("2006-01-02")]++
	}

	trendData := make([]models.TrendData, 0)
	totalCount := initialCount
	for date := startDate; date.Before(endDateExclusive); date = date.AddDate(0, 0, 1) {
		key := date.Format("2006-01-02")
		totalCount += newByDate[key]
		trendData = append(trendData, models.TrendData{
			Date:                  key,
			TotalAuthorizations:   totalCount,
			NewAuthorizations:     newByDate[key],
			ExpiredAuthorizations: expiredByDate[key],
		})
	}
	return trendData
}

// GetRecentAuthorizations 获取最近授权列表
func (r *dashboardRepository) GetRecentAuthorizations(ctx context.Context, req *models.DashboardRecentAuthorizationsRequest) (*models.DashboardRecentAuthorizationsResponse, error) {
	lang := pkgcontext.GetLanguageFromContext(ctx)

	// 设置默认值
	limit := req.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	// 构建查询
	query := r.db.WithContext(ctx).Table("authorization_codes ac").
		Select(`ac.id, ac.code, ac.customer_id, COALESCE(c.customer_name, '') as customer_name,
		        COALESCE(ac.description, '') as description, ac.start_date, ac.end_date, ac.max_activations,
		        COALESCE(l.active_count, 0) as current_activations, 
		        ac.created_at, ac.updated_at, ac.is_locked`).
		Joins("LEFT JOIN customers c ON ac.customer_id = c.id").
		Joins(`LEFT JOIN (
			SELECT authorization_code_id, COUNT(*) as active_count 
			FROM licenses 
			WHERE status = 'active' AND deleted_at IS NULL 
			GROUP BY authorization_code_id
		) l ON ac.id = l.authorization_code_id`).
		Where("1=1").
		Order("ac.created_at DESC")

	// 添加筛选条件
	if req.CustomerID != "" {
		query = query.Where("ac.customer_id = ?", req.CustomerID)
	}

	// 状态筛选（虚字段）
	now := time.Now()
	switch req.Status {
	case "normal":
		query = query.Where("ac.is_locked = ? AND ac.end_date > ?", false, now)
	case "locked":
		query = query.Where("ac.is_locked = ?", true)
	case "expired":
		query = query.Where("ac.end_date <= ?", now)
	}

	// 查询总数 - 使用简化查询避免JOIN复杂性
	var total int64
	countQuery := r.db.WithContext(ctx).Model(&models.AuthorizationCode{}).Where("1=1")

	// 添加相同的筛选条件
	if req.CustomerID != "" {
		countQuery = countQuery.Where("customer_id = ?", req.CustomerID)
	}

	// 状态筛选
	switch req.Status {
	case "normal":
		countQuery = countQuery.Where("is_locked = ? AND end_date > ?", false, now)
	case "locked":
		countQuery = countQuery.Where("is_locked = ?", true)
	case "expired":
		countQuery = countQuery.Where("end_date <= ?", now)
	}

	err := countQuery.Count(&total).Error
	if err != nil {
		return nil, i18n.NewI18nError("900004", lang, err.Error())
	}

	// 查询列表数据
	type queryResult struct {
		ID                 string    `gorm:"column:id"`
		Code               string    `gorm:"column:code"`
		CustomerID         *string   `gorm:"column:customer_id"`
		CustomerName       string    `gorm:"column:customer_name"`
		Description        string    `gorm:"column:description"`
		StartDate          time.Time `gorm:"column:start_date"`
		EndDate            time.Time `gorm:"column:end_date"`
		MaxActivations     int       `gorm:"column:max_activations"`
		CurrentActivations int       `gorm:"column:current_activations"`
		CreatedAt          time.Time `gorm:"column:created_at"`
		UpdatedAt          time.Time `gorm:"column:updated_at"`
		IsLocked           bool      `gorm:"column:is_locked"`
	}

	var results []queryResult
	err = query.Limit(limit).Find(&results).Error
	if err != nil {
		return nil, i18n.NewI18nError("900004", lang, err.Error())
	}

	// 转换为响应格式
	authorizations := make([]models.RecentAuthorization, 0, len(results))
	for _, result := range results {
		// 计算状态
		status := "normal"
		statusDisplay := i18n.GetEnumMessage("authorization_code_status", "normal", lang)

		if result.IsLocked {
			status = "locked"
			statusDisplay = i18n.GetEnumMessage("authorization_code_status", "locked", lang)
		} else if result.EndDate.Before(now) {
			status = "expired"
			statusDisplay = i18n.GetEnumMessage("authorization_code_status", "expired", lang)
		}

		authorizations = append(authorizations, models.RecentAuthorization{
			ID:                 result.ID,
			Code:               result.Code,
			CustomerID:         result.CustomerID,
			CustomerName:       result.CustomerName,
			Description:        result.Description,
			Status:             status,
			StatusDisplay:      statusDisplay,
			StartDate:          result.StartDate,
			EndDate:            result.EndDate,
			MaxActivations:     result.MaxActivations,
			CurrentActivations: result.CurrentActivations,
			CreatedAt:          result.CreatedAt,
			UpdatedAt:          result.UpdatedAt,
		})
	}

	return &models.DashboardRecentAuthorizationsResponse{
		List:  authorizations,
		Total: total,
	}, nil
}

// resolvedExpiry 解析列表中展示的到期时间。
//
// 从激活起算的授权码，其 end_date 只是硬上限，不代表任一设备的实际到期日。
// 列表展示用授权码的 end_date 作为参考值，缺失时回退到零值，
// 由前端按 validity_mode 决定如何呈现。
func resolvedExpiry(validityMode string, endDate time.Time, expiresAt *time.Time) time.Time {
	if expiresAt != nil && !expiresAt.IsZero() {
		return *expiresAt
	}
	return endDate
}

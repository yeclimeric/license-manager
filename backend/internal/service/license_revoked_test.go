package service

import (
	"context"
	"errors"
	"testing"

	"license-manager/internal/models"
	"license-manager/pkg/i18n"

	"github.com/sirupsen/logrus"
)

// revokedLicenseStub 驱动 GenerateLicenseFile 的撤销分支（不碰真实 gorm 事务）。
type revokedLicenseStub struct {
	license *models.License
	getErr  error
}

func (r *revokedLicenseStub) GetLicenseList(context.Context, *models.LicenseListRequest) (*models.LicenseListResponse, error) {
	return nil, nil
}
func (r *revokedLicenseStub) GetLicenseByID(context.Context, string) (*models.License, error) {
	return r.license, r.getErr
}
func (r *revokedLicenseStub) CreateLicense(context.Context, *models.License) error {
	return nil
}
func (r *revokedLicenseStub) UpdateLicense(context.Context, *models.License) error {
	return nil
}
func (r *revokedLicenseStub) DeleteLicensePermanently(context.Context, *models.License) error {
	return nil
}
func (r *revokedLicenseStub) CheckAuthorizationCodeExists(context.Context, string) (bool, error) {
	return false, nil
}
func (r *revokedLicenseStub) GetAuthorizationCodeByID(context.Context, string) (*models.AuthorizationCode, error) {
	return nil, nil
}
func (r *revokedLicenseStub) GetAuthorizationCodeByCode(context.Context, string) (*models.AuthorizationCode, error) {
	return nil, nil
}
func (r *revokedLicenseStub) GetLicenseByKey(context.Context, string) (*models.License, error) {
	return nil, nil
}
func (r *revokedLicenseStub) GetActiveLicenseCount(context.Context, string) (int64, error) {
	return 0, nil
}
func (r *revokedLicenseStub) CheckAuthorizationCodeHasLicenses(context.Context, string) (bool, error) {
	return false, nil
}

// B4 不变量：已撤销的许可证文件不可再下载（签发物料）。
// 对应 ActivateLicense 撤销态复活的前置守卫——若此处放行，
// 撤销保护就只剩 ActivateLicense 单点，等于形同虚设。
func TestGenerateLicenseFileRejectsRevoked(t *testing.T) {
	repo := &revokedLicenseStub{license: &models.License{ID: "lic-revoked", Status: "revoked"}}
	svc := &licenseService{licenseRepo: repo, logger: logrus.New()}

	_, _, _, err := svc.GenerateLicenseFile(context.Background(), "lic-revoked")

	var i18nErr *i18n.I18nError
	if !errors.As(err, &i18nErr) || i18nErr.Code != "300007" {
		t.Fatalf("want i18n code 300007 (许可证已被撤销), got %v", err)
	}
}

// 对照：非撤销态不触发 300007（active/inactive 正常走签发物料）。
func TestGenerateLicenseFileAllowsActive(t *testing.T) {
	repo := &revokedLicenseStub{license: &models.License{ID: "lic-active", Status: "active", LicenseKey: "K1"}}
	svc := &licenseService{licenseRepo: repo, logger: logrus.New()}

	// active 态不会命中撤销分支；因无产品项/私钥会走到签名兜底失败，
	// 但错误码绝不是 300007。
	_, _, _, err := svc.GenerateLicenseFile(context.Background(), "lic-active")
	if err == nil {
		t.Skip("无 RSA 配置环境下 active 态签发成功，符合预期（非撤销分支）")
	}
	var i18nErr *i18n.I18nError
	if errors.As(err, &i18nErr) && i18nErr.Code == "300007" {
		t.Fatalf("active 许可证误命中撤销分支 300007：撤销判定泄漏到非撤销态")
	}
}

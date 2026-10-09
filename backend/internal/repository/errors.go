package repository

import "errors"

// 客户领域的业务错误（给开发者看的，英文即可）
var (
	ErrCustomerNotFound      = errors.New("customer not found")
	ErrCustomerAlreadyExists = errors.New("customer already exists")
	ErrCustomerCodeDuplicate = errors.New("customer code already exists")
)

// 授权码领域的业务错误
var (
	ErrAuthorizationCodeNotFound      = errors.New("authorization code not found")
	ErrAuthorizationCodeAlreadyExists = errors.New("authorization code already exists")
	ErrAuthorizationCodeDuplicate     = errors.New("authorization code already exists")
)

// 许可证领域的业务错误
var (
	ErrLicenseNotFound      = errors.New("license not found")
	ErrLicenseAlreadyExists = errors.New("license already exists")
	ErrLicenseDuplicate     = errors.New("license already exists")
)

// 产品项领域的业务错误
var (
	ErrProductOfferingNotFound  = errors.New("product offering not found")
	ErrProductOfferingDuplicate = errors.New("product offering code already exists")
	ErrProductOfferingInUse     = errors.New("product offering still has authorization codes")
)

// 用户领域的业务错误
var (
	ErrUserNotFound = errors.New("user not found")
)

// RBAC 领域的业务错误
var (
	ErrRoleNotFound        = errors.New("role not found")
	ErrRoleDuplicate       = errors.New("role name already exists")
	ErrRoleSystemProtected = errors.New("system role cannot be modified or deleted")
	ErrRoleInUse           = errors.New("role is assigned to users")
	ErrMenuNotFound        = errors.New("menu not found")
	ErrMenuDuplicate       = errors.New("menu identifier already exists")
	ErrMenuSystemProtected = errors.New("system menu cannot be modified or deleted")
	ErrMenuHasChildren     = errors.New("menu still has child items")
	ErrMenuInUse           = errors.New("menu is bound to roles")
	ErrMenuParentNotFound  = errors.New("parent menu not found")
	ErrMenuCyclicReference = errors.New("menu hierarchy would form a cycle")
)

// 功能模板领域的业务错误
var (
	ErrFeatureTemplateNotFound         = errors.New("feature template not found")
	ErrFeatureTemplateIdentifierExists = errors.New("feature template identifier already exists")
	ErrFeatureTemplateSystemProtected  = errors.New("system feature template cannot be modified or deleted")
	ErrTemplateCategoryNotFound        = errors.New("feature template category not found")
	ErrTemplateCategoryInUse           = errors.New("feature template category is in use")
)

// 通用的数据库/系统错误（给开发者看的）
var (
	ErrDatabaseConnection = errors.New("database connection failed")
	ErrDatabaseQuery      = errors.New("database query failed")
	ErrDatabaseCreate     = errors.New("database create failed")
	ErrDatabaseUpdate     = errors.New("database update failed")
	ErrDatabaseDelete     = errors.New("database delete failed")
	ErrInvalidTransaction = errors.New("invalid transaction object")
)

// 辅助函数：判断错误类型
func IsCustomerNotFound(err error) bool {
	return errors.Is(err, ErrCustomerNotFound)
}

func IsCustomerAlreadyExists(err error) bool {
	return errors.Is(err, ErrCustomerAlreadyExists)
}

func IsDatabaseError(err error) bool {
	return errors.Is(err, ErrDatabaseConnection) ||
		errors.Is(err, ErrDatabaseQuery) ||
		errors.Is(err, ErrDatabaseCreate) ||
		errors.Is(err, ErrDatabaseUpdate) ||
		errors.Is(err, ErrDatabaseDelete)
}

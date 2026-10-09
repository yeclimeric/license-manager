package models

// 用户状态
const (
	UserStatusActive  = "active"
	UserStatusLocked  = "locked"
	UserStatusDisable = "inactive"
)

// UserRoleAssignmentInput 角色分配项。
//
// 用角色名而非角色ID：调用方无需先查ID，且角色改名时引用不会失效。
// 与企业版实测契约一致（role_assignments[].role）。
type UserRoleAssignmentInput struct {
	Role string `json:"role" binding:"required"`
}

// UserResetPasswordRequest 重置密码请求
type UserResetPasswordRequest struct {
	NewPassword string `json:"new_password" binding:"required,min=8,max=64"`
}

// UserRoleItem 用户绑定的角色摘要
type UserRoleItem struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
}

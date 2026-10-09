package handlers

import (
	"license-manager/internal/api/middleware"
	"license-manager/internal/models"
	"license-manager/internal/service"

	"github.com/gin-gonic/gin"
)

type UserManagementHandler struct {
	userService service.UserManagementService
}

// NewUserManagementHandler 创建用户管理处理器
func NewUserManagementHandler(userService service.UserManagementService) *UserManagementHandler {
	return &UserManagementHandler{userService: userService}
}

// GetUserList 查询用户列表
// @Summary 查询用户列表
// @Description 分页查询用户列表，含各用户绑定的角色
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码"
// @Param page_size query int false "每页条数"
// @Param search query string false "按用户名/邮箱/姓名搜索"
// @Param status query string false "状态筛选"
// @Param role query string false "角色筛选"
// @Success 200 {object} models.APIResponse{data=models.UserListResponse} "查询成功"
// @Failure 401 {object} models.ErrorResponse "未认证"
// @Failure 403 {object} models.ErrorResponse "权限不足"
// @Failure 500 {object} models.ErrorResponse "服务器内部错误"
// @Router /api/v1/users [get]
func (h *UserManagementHandler) GetUserList(c *gin.Context) {
	var req models.UserListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		respondErrorWithCode(c, "900001")
		return
	}

	ctx := middleware.WithLanguage(c.Request.Context(), c)

	data, err := h.userService.GetUserList(ctx, &req)
	if err != nil {
		respondError(c, err)
		return
	}

	respondSuccess(c, data)
}

// GetUserDetail 查询用户详情
// @Summary 查询用户详情
// @Description 返回用户信息、绑定角色与登录安全状态
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "用户ID"
// @Success 200 {object} models.APIResponse{data=models.UserDetailResponse} "查询成功"
// @Failure 401 {object} models.ErrorResponse "未认证"
// @Failure 404 {object} models.ErrorResponse "用户不存在"
// @Failure 500 {object} models.ErrorResponse "服务器内部错误"
// @Router /api/v1/users/{id} [get]
func (h *UserManagementHandler) GetUserDetail(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		respondErrorWithCode(c, "900001")
		return
	}

	ctx := middleware.WithLanguage(c.Request.Context(), c)

	data, err := h.userService.GetUserDetail(ctx, id)
	if err != nil {
		respondError(c, err)
		return
	}

	respondSuccess(c, data)
}

// CreateUser 创建用户
// @Summary 创建用户
// @Description 创建用户并分配角色。必须至少分配一个角色，否则用户登录后无法访问任何功能
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.UserCreateRequest true "用户信息"
// @Success 200 {object} models.APIResponse{data=models.UserDetailResponse} "创建成功"
// @Failure 400 {object} models.ErrorResponse "请求参数无效"
// @Failure 401 {object} models.ErrorResponse "未认证"
// @Failure 403 {object} models.ErrorResponse "权限不足"
// @Failure 409 {object} models.ErrorResponse "用户名或邮箱已存在"
// @Failure 500 {object} models.ErrorResponse "服务器内部错误"
// @Router /api/v1/users [post]
func (h *UserManagementHandler) CreateUser(c *gin.Context) {
	var req models.UserCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondErrorWithCode(c, "900001")
		return
	}

	ctx := middleware.WithLanguage(c.Request.Context(), c)

	data, err := h.userService.CreateUser(ctx, &req)
	if err != nil {
		respondError(c, err)
		return
	}

	respondSuccess(c, data)
}

// UpdateUser 更新用户
// @Summary 更新用户
// @Description 更新用户信息、状态、密码或角色绑定
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "用户ID"
// @Param request body models.UserUpdateRequest true "用户信息"
// @Success 200 {object} models.APIResponse{data=models.UserDetailResponse} "更新成功"
// @Failure 400 {object} models.ErrorResponse "请求参数无效"
// @Failure 401 {object} models.ErrorResponse "未认证"
// @Failure 404 {object} models.ErrorResponse "用户不存在"
// @Failure 409 {object} models.ErrorResponse "用户名或邮箱已存在"
// @Failure 500 {object} models.ErrorResponse "服务器内部错误"
// @Router /api/v1/users/{id} [put]
func (h *UserManagementHandler) UpdateUser(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		respondErrorWithCode(c, "900001")
		return
	}

	var req models.UserUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondErrorWithCode(c, "900001")
		return
	}

	ctx := middleware.WithLanguage(c.Request.Context(), c)

	data, err := h.userService.UpdateUser(ctx, id, &req)
	if err != nil {
		respondError(c, err)
		return
	}

	respondSuccess(c, data)
}

// DeleteUser 删除用户
// @Summary 删除用户
// @Description 删除用户及其角色绑定。最后一个超级管理员不可删除
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "用户ID"
// @Success 200 {object} models.APIResponse "删除成功"
// @Failure 400 {object} models.ErrorResponse "请求参数无效"
// @Failure 401 {object} models.ErrorResponse "未认证"
// @Failure 404 {object} models.ErrorResponse "用户不存在"
// @Failure 500 {object} models.ErrorResponse "服务器内部错误"
// @Router /api/v1/users/{id} [delete]
func (h *UserManagementHandler) DeleteUser(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		respondErrorWithCode(c, "900001")
		return
	}

	ctx := middleware.WithLanguage(c.Request.Context(), c)

	if err := h.userService.DeleteUser(ctx, id); err != nil {
		respondError(c, err)
		return
	}

	respondSuccess(c, nil)
}

// ResetPassword 重置用户密码
// @Summary 重置用户密码
// @Description 管理员重设用户密码，同时解除锁定状态
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "用户ID"
// @Param request body models.UserResetPasswordRequest true "新密码"
// @Success 200 {object} models.APIResponse "重置成功"
// @Failure 400 {object} models.ErrorResponse "请求参数无效"
// @Failure 401 {object} models.ErrorResponse "未认证"
// @Failure 404 {object} models.ErrorResponse "用户不存在"
// @Failure 500 {object} models.ErrorResponse "服务器内部错误"
// @Router /api/v1/users/{id}/reset-password [post]
func (h *UserManagementHandler) ResetPassword(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		respondErrorWithCode(c, "900001")
		return
	}

	var req models.UserResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondErrorWithCode(c, "900001")
		return
	}

	ctx := middleware.WithLanguage(c.Request.Context(), c)

	if err := h.userService.ResetPassword(ctx, id, req.NewPassword); err != nil {
		respondError(c, err)
		return
	}

	respondSuccess(c, nil)
}

// GetUserRoles 查询用户角色
// @Summary 查询用户角色
// @Description 返回用户绑定的角色列表
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "用户ID"
// @Success 200 {object} models.APIResponse{data=[]models.UserRoleItem} "查询成功"
// @Failure 401 {object} models.ErrorResponse "未认证"
// @Failure 500 {object} models.ErrorResponse "服务器内部错误"
// @Router /api/v1/users/{id}/roles [get]
func (h *UserManagementHandler) GetUserRoles(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		respondErrorWithCode(c, "900001")
		return
	}

	ctx := middleware.WithLanguage(c.Request.Context(), c)

	data, err := h.userService.GetUserRoles(ctx, id)
	if err != nil {
		respondError(c, err)
		return
	}

	respondSuccess(c, data)
}

package handlers

import (
	"license-manager/internal/api/middleware"
	"license-manager/internal/models"
	"license-manager/internal/service"

	"github.com/gin-gonic/gin"
)

type RBACHandler struct {
	rbacService service.RBACService
}

// NewRBACHandler 创建菜单权限角色处理器
func NewRBACHandler(rbacService service.RBACService) *RBACHandler {
	return &RBACHandler{rbacService: rbacService}
}

// GetUserMenuTree 获取当前用户可见菜单树
// @Summary 获取当前用户菜单
// @Description 返回当前登录用户有权访问的菜单树，前端据此渲染侧边栏
// @Tags 菜单权限
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.APIResponse{data=models.MenuTreeResponse} "查询成功"
// @Failure 401 {object} models.ErrorResponse "未认证"
// @Failure 500 {object} models.ErrorResponse "服务器内部错误"
// @Router /api/v1/menus [get]
func (h *RBACHandler) GetUserMenuTree(c *gin.Context) {
	ctx := middleware.WithLanguage(c.Request.Context(), c)

	data, err := h.rbacService.GetUserMenuTree(ctx)
	if err != nil {
		respondError(c, err)
		return
	}

	respondSuccess(c, data)
}

// GetAllMenuTree 获取全部菜单树
// @Summary 获取全部菜单
// @Description 返回完整菜单树，用于菜单管理界面
// @Tags 菜单权限
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.APIResponse{data=models.MenuTreeResponse} "查询成功"
// @Failure 401 {object} models.ErrorResponse "未认证"
// @Failure 500 {object} models.ErrorResponse "服务器内部错误"
// @Router /api/v1/menus/all [get]
func (h *RBACHandler) GetAllMenuTree(c *gin.Context) {
	ctx := middleware.WithLanguage(c.Request.Context(), c)

	data, err := h.rbacService.GetAllMenuTree(ctx)
	if err != nil {
		respondError(c, err)
		return
	}

	respondSuccess(c, data)
}

// GetPermissionMatrix 获取权限矩阵
// @Summary 获取权限矩阵
// @Description 返回资源及其支持的动作，供角色权限配置使用
// @Tags 菜单权限
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.APIResponse{data=models.PermissionMatrixResponse} "查询成功"
// @Failure 401 {object} models.ErrorResponse "未认证"
// @Failure 500 {object} models.ErrorResponse "服务器内部错误"
// @Router /api/v1/permissions [get]
func (h *RBACHandler) GetPermissionMatrix(c *gin.Context) {
	ctx := middleware.WithLanguage(c.Request.Context(), c)

	data, err := h.rbacService.GetPermissionMatrix(ctx)
	if err != nil {
		respondError(c, err)
		return
	}

	respondSuccess(c, data)
}

// CreateMenu 创建菜单
// @Summary 创建菜单
// @Description 创建菜单或分组目录
// @Tags 菜单权限
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.Menu true "菜单信息"
// @Success 200 {object} models.APIResponse{data=models.Menu} "创建成功"
// @Failure 400 {object} models.ErrorResponse "请求参数无效"
// @Failure 401 {object} models.ErrorResponse "未认证"
// @Failure 500 {object} models.ErrorResponse "服务器内部错误"
// @Router /api/v1/menus [post]
func (h *RBACHandler) CreateMenu(c *gin.Context) {
	var req models.Menu
	if err := c.ShouldBindJSON(&req); err != nil {
		respondErrorWithCode(c, "900001")
		return
	}

	ctx := middleware.WithLanguage(c.Request.Context(), c)

	data, err := h.rbacService.CreateMenu(ctx, &req)
	if err != nil {
		respondError(c, err)
		return
	}

	respondSuccess(c, data)
}

// UpdateMenu 更新菜单
// @Summary 更新菜单
// @Description 更新菜单信息，内置菜单的类型/路径/权限不可修改
// @Tags 菜单权限
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "菜单ID"
// @Param request body models.Menu true "菜单信息"
// @Success 200 {object} models.APIResponse{data=models.Menu} "更新成功"
// @Failure 400 {object} models.ErrorResponse "请求参数无效"
// @Failure 401 {object} models.ErrorResponse "未认证"
// @Failure 404 {object} models.ErrorResponse "菜单不存在"
// @Failure 500 {object} models.ErrorResponse "服务器内部错误"
// @Router /api/v1/menus/{id} [put]
func (h *RBACHandler) UpdateMenu(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		respondErrorWithCode(c, "900001")
		return
	}

	var req models.Menu
	if err := c.ShouldBindJSON(&req); err != nil {
		respondErrorWithCode(c, "900001")
		return
	}

	ctx := middleware.WithLanguage(c.Request.Context(), c)

	data, err := h.rbacService.UpdateMenu(ctx, id, &req)
	if err != nil {
		respondError(c, err)
		return
	}

	respondSuccess(c, data)
}

// DeleteMenu 删除菜单
// @Summary 删除菜单
// @Description 删除菜单，内置菜单或含子菜单/已绑定角色的菜单不可删除
// @Tags 菜单权限
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "菜单ID"
// @Success 200 {object} models.APIResponse "删除成功"
// @Failure 400 {object} models.ErrorResponse "请求参数无效"
// @Failure 401 {object} models.ErrorResponse "未认证"
// @Failure 404 {object} models.ErrorResponse "菜单不存在"
// @Failure 500 {object} models.ErrorResponse "服务器内部错误"
// @Router /api/v1/menus/{id} [delete]
func (h *RBACHandler) DeleteMenu(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		respondErrorWithCode(c, "900001")
		return
	}

	ctx := middleware.WithLanguage(c.Request.Context(), c)

	if err := h.rbacService.DeleteMenu(ctx, id); err != nil {
		respondError(c, err)
		return
	}

	respondSuccess(c, nil)
}

// GetRoleList 获取角色列表
// @Summary 获取角色列表
// @Description 查询角色列表，含权限、菜单与用户数
// @Tags 菜单权限
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param keyword query string false "搜索关键词"
// @Success 200 {object} models.APIResponse{data=[]models.Role} "查询成功"
// @Failure 401 {object} models.ErrorResponse "未认证"
// @Failure 500 {object} models.ErrorResponse "服务器内部错误"
// @Router /api/v1/roles [get]
func (h *RBACHandler) GetRoleList(c *gin.Context) {
	keyword := c.Query("keyword")

	ctx := middleware.WithLanguage(c.Request.Context(), c)

	list, total, err := h.rbacService.GetRoleList(ctx, keyword)
	if err != nil {
		respondError(c, err)
		return
	}

	respondSuccess(c, gin.H{"list": list, "total": total})
}

// GetRole 获取角色详情
// @Summary 获取角色详情
// @Description 根据ID获取角色及其权限与菜单
// @Tags 菜单权限
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "角色ID"
// @Success 200 {object} models.APIResponse{data=models.Role} "查询成功"
// @Failure 400 {object} models.ErrorResponse "请求参数无效"
// @Failure 401 {object} models.ErrorResponse "未认证"
// @Failure 404 {object} models.ErrorResponse "角色不存在"
// @Failure 500 {object} models.ErrorResponse "服务器内部错误"
// @Router /api/v1/roles/{id} [get]
func (h *RBACHandler) GetRole(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		respondErrorWithCode(c, "900001")
		return
	}

	ctx := middleware.WithLanguage(c.Request.Context(), c)

	data, err := h.rbacService.GetRole(ctx, id)
	if err != nil {
		respondError(c, err)
		return
	}

	respondSuccess(c, data)
}

// CreateRole 创建角色
// @Summary 创建角色
// @Description 创建角色并绑定权限与菜单
// @Tags 菜单权限
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.RoleCreateRequest true "角色信息"
// @Success 200 {object} models.APIResponse{data=models.Role} "创建成功"
// @Failure 400 {object} models.ErrorResponse "请求参数无效"
// @Failure 401 {object} models.ErrorResponse "未认证"
// @Failure 500 {object} models.ErrorResponse "服务器内部错误"
// @Router /api/v1/roles [post]
func (h *RBACHandler) CreateRole(c *gin.Context) {
	var req models.RoleCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondErrorWithCode(c, "900001")
		return
	}

	ctx := middleware.WithLanguage(c.Request.Context(), c)

	data, err := h.rbacService.CreateRole(ctx, &req)
	if err != nil {
		respondError(c, err)
		return
	}

	respondSuccess(c, data)
}

// UpdateRole 更新角色
// @Summary 更新角色
// @Description 更新角色的展示信息、权限与菜单绑定
// @Tags 菜单权限
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "角色ID"
// @Param request body models.RoleUpdateRequest true "角色信息"
// @Success 200 {object} models.APIResponse{data=models.Role} "更新成功"
// @Failure 400 {object} models.ErrorResponse "请求参数无效"
// @Failure 401 {object} models.ErrorResponse "未认证"
// @Failure 404 {object} models.ErrorResponse "角色不存在"
// @Failure 500 {object} models.ErrorResponse "服务器内部错误"
// @Router /api/v1/roles/{id} [put]
func (h *RBACHandler) UpdateRole(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		respondErrorWithCode(c, "900001")
		return
	}

	var req models.RoleUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondErrorWithCode(c, "900001")
		return
	}

	ctx := middleware.WithLanguage(c.Request.Context(), c)

	data, err := h.rbacService.UpdateRole(ctx, id, &req)
	if err != nil {
		respondError(c, err)
		return
	}

	respondSuccess(c, data)
}

// DeleteRole 删除角色
// @Summary 删除角色
// @Description 删除角色，内置角色或已分配给用户的角色不可删除
// @Tags 菜单权限
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "角色ID"
// @Success 200 {object} models.APIResponse "删除成功"
// @Failure 400 {object} models.ErrorResponse "请求参数无效"
// @Failure 401 {object} models.ErrorResponse "未认证"
// @Failure 404 {object} models.ErrorResponse "角色不存在"
// @Failure 500 {object} models.ErrorResponse "服务器内部错误"
// @Router /api/v1/roles/{id} [delete]
func (h *RBACHandler) DeleteRole(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		respondErrorWithCode(c, "900001")
		return
	}

	ctx := middleware.WithLanguage(c.Request.Context(), c)

	if err := h.rbacService.DeleteRole(ctx, id); err != nil {
		respondError(c, err)
		return
	}

	respondSuccess(c, nil)
}

// GetUserRoles 获取用户绑定的角色
// @Summary 获取用户角色
// @Description 返回指定用户绑定的角色ID列表
// @Tags 菜单权限
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param userId path string true "用户ID"
// @Success 200 {object} models.APIResponse{data=[]string} "查询成功"
// @Failure 401 {object} models.ErrorResponse "未认证"
// @Failure 500 {object} models.ErrorResponse "服务器内部错误"
// @Router /api/v1/users/{userId}/roles [get]
func (h *RBACHandler) GetUserRoles(c *gin.Context) {
	userID := c.Param("userId")
	if userID == "" {
		respondErrorWithCode(c, "900001")
		return
	}

	ctx := middleware.WithLanguage(c.Request.Context(), c)

	roleIDs, err := h.rbacService.GetUserRoles(ctx, userID)
	if err != nil {
		respondError(c, err)
		return
	}

	respondSuccess(c, roleIDs)
}

// SetUserRoles 设置用户绑定的角色
// @Summary 设置用户角色
// @Description 覆盖式设置用户绑定的角色，至少保留一个
// @Tags 菜单权限
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param userId path string true "用户ID"
// @Param request body map[string]interface{} true "角色ID列表，字段 role_ids"
// @Success 200 {object} models.APIResponse "设置成功"
// @Failure 400 {object} models.ErrorResponse "请求参数无效"
// @Failure 401 {object} models.ErrorResponse "未认证"
// @Failure 404 {object} models.ErrorResponse "用户或角色不存在"
// @Failure 500 {object} models.ErrorResponse "服务器内部错误"
// @Router /api/v1/users/{userId}/roles [put]
func (h *RBACHandler) SetUserRoles(c *gin.Context) {
	userID := c.Param("userId")
	if userID == "" {
		respondErrorWithCode(c, "900001")
		return
	}

	var req struct {
		RoleIDs []string `json:"role_ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		respondErrorWithCode(c, "900001")
		return
	}

	ctx := middleware.WithLanguage(c.Request.Context(), c)

	if err := h.rbacService.SetUserRoles(ctx, userID, req.RoleIDs); err != nil {
		respondError(c, err)
		return
	}

	respondSuccess(c, nil)
}

/*
 * @Author: 13895237362 2205451508@qq.com
 * @Date: 2025-08-01 09:32:42
 * @LastEditors: 13895237362 2205451508@qq.com
 * @LastEditTime: 2025-10-21 16:59:51
 * @FilePath: /frontend/src/router/index.ts
 * @Description: 这是默认设置,请设置`customMade`, 打开koroFileHeader查看配置 进行设置: https://github.com/OBKoro1/koro1FileHeader/wiki/%E9%85%8D%E7%BD%AE
 */
import { createRouter, createWebHistory, RouteRecordRaw } from "vue-router";
import Layout from "@/components/common/layout/Layout.vue";
import { useUserStore } from "@/store/modules/user";

// 路由配置。
// Layout 是路由级父布局：认证后的页面全部作为其 children，
// 经 <router-view> 渲染，布局与侧边栏状态跨页面跳转保持，
// 不再由每个页面各自包裹 Layout（旧模式会在路由切换时重建布局）。
const routes: Array<RouteRecordRaw> = [
  {
    path: "/",
    component: Layout,
    meta: { requiresAuth: true },
    children: [
      {
        path: "",
        redirect: "/login"
      },
      {
        path: "dashboard",
        component: () => import("@/views/Dashboard.vue"),
        meta: { title: "仪表盘", requiresAuth: true, permission: "dashboard:read" }
      },
      {
        path: "customers",
        component: () => import("@/views/Customers/index.vue"),
        meta: { title: "客户管理", requiresAuth: true, permission: "customer:read" }
      },
      {
        // 无 component 的传递路由：children 直接渲染进 Layout 的 router-view
        path: "licenses",
        name: "licenses",
        meta: { title: "授权管理", requiresAuth: true, permission: "license:read" },
        children: [
          {
            path: "",
            name: "licenses-search",
            redirect: { name: "licenses-list" },
            meta: { title: "授权列表", requiresAuth: true }
          },
          {
            path: "list",
            name: "licenses-list",
            component: () => import("@/views/Licenses/LinenseList.vue"),
            meta: { title: "授权列表", requiresAuth: true }
          },
          {
            path: "search",
            name: "licenses-customer-search",
            component: () => import("@/views/Licenses/LicenseSearch.vue"),
            meta: { title: "授权搜索", requiresAuth: true }
          },
          {
            path: "create",
            name: "licenses-create",
            component: () => import("@/views/Licenses/LicenseForm.vue"),
            meta: { title: "创建授权", requiresAuth: true }
          },
          {
            path: "view/:id",
            name: "licenses-view",
            component: () => import("@/views/Licenses/LicensesView/index.vue"),
            meta: { title: "查看授权", requiresAuth: true }
          }
        ]
      },
      {
        path: "products",
        name: "products",
        component: () => import("@/views/Products/index.vue"),
        meta: { title: "产品项管理", requiresAuth: true, permission: "product_offering:read" }
      },
      {
        path: "feature-templates",
        name: "feature-templates",
        component: () => import("@/views/FeatureTemplates/index.vue"),
        meta: { title: "功能模板", requiresAuth: true, permission: "feature_template:read" }
      },
      {
        path: "system-config",
        name: "system-config",
        component: () => import("@/views/SystemConfig/index.vue"),
        meta: { title: "系统配置", requiresAuth: true, permission: "system_config:read" }
      },
      {
        path: "api-keys",
        name: "api-keys",
        component: () => import("@/views/ApiKeys/index.vue"),
        meta: { title: "API Key", requiresAuth: true, permission: "api_key:read" }
      },
      {
        path: "users",
        name: "users",
        component: () => import("@/views/Users/index.vue"),
        meta: { title: "用户管理", requiresAuth: true, permission: "user:read" }
      },
      {
        path: "roles",
        name: "roles",
        component: () => import("@/views/Roles/index.vue"),
        meta: { title: "角色管理", requiresAuth: true, permission: "role:read" }
      },
      {
        path: "menus",
        name: "menus",
        component: () => import("@/views/Menus/index.vue"),
        meta: { title: "菜单管理", requiresAuth: true, permission: "menu:read" }
      },
      {
        path: "audit-logs",
        name: "audit-logs",
        component: () => import("@/views/AuditLogs/index.vue"),
        meta: { title: "操作日志", requiresAuth: true, permission: "audit_log:read" }
      }
    ]
  },
  {
    path: "/login",
    component: () => import("@/views/Login.vue"),
    meta: { title: "登录" }
  },
  {
    path: "/403",
    name: "forbidden",
    component: () => import("@/views/Forbidden.vue"),
    meta: { title: "无权访问" }
  },
  {
    path: "/:pathMatch(.*)*",
    redirect: "/dashboard"
  }
];

const routerBase = import.meta.env.BASE_URL === './' ? '/' : import.meta.env.BASE_URL

const router = createRouter({
  history: createWebHistory(routerBase),
  routes,
});

// 路由守卫：认证 + 权限（权限码来自 /api/v1/me/permissions 实时查询）
router.beforeEach(async (to, _from, next) => {
  document.title = typeof (to.meta.title) === "string" ? to.meta.title : "授权管理平台";

  if (!to.meta.requiresAuth) {
    next();
    return;
  }

  const token = localStorage.getItem("token");
  if (!token) {
    next("/login");
    return;
  }

  const userStore = useUserStore();
  // 首屏或刷新后权限尚未加载，先等一次；并发导航由 store 内部合并为单请求
  if (!userStore.permissionsLoaded) {
    await userStore.loadPermissions();
  }

  // 权限请求若因 401 失败，axios 拦截器已同步清除 token，此时必须回登录页
  if (userStore.permissionsLoadFailed && !localStorage.getItem("token")) {
    next("/login");
    return;
  }

  const required = typeof to.meta.permission === "string" ? to.meta.permission : "";
  // 权限接口失败时 fail-open：后端每个端点都挂了 RequirePermission，
  // 前端守卫只负责体验层（不展示无权页面），不因后端抖动让整站不可用。
  if (required && !userStore.permissionsLoadFailed && !userStore.hasPermission(required)) {
    next("/403");
    return;
  }

  next();
});

export default router;


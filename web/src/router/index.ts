import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    // 首次配置向导
    {
      path: '/setup',
      name: 'Setup',
      component: () => import('@/views/Setup.vue'),
      meta: { requiresAuth: false }
    },
    // 登录页面
    {
      path: '/login',
      name: 'Login',
      component: () => import('@/views/Login.vue'),
      meta: { requiresAuth: false }
    },
    // 主应用布局
    {
      path: '/',
      component: () => import('@/layouts/MainLayout.vue'),
      meta: { requiresAuth: true },
      children: [
        // 概览仪表盘
        {
          path: '',
          name: 'Dashboard',
          component: () => import('@/views/Dashboard.vue')
        },
        // 容器管理
        {
          path: 'containers',
          name: 'Containers',
          component: () => import('@/views/Containers.vue')
        },
        // 编排管理
        {
          path: 'orchestration',
          name: 'Orchestration',
          component: () => import('@/views/Orchestration.vue')
        },
        // 镜像管理
        {
          path: 'images',
          name: 'Images',
          component: () => import('@/views/Images.vue')
        },
        // 网络管理
        {
          path: 'networks',
          name: 'Networks',
          component: () => import('@/views/Networks.vue')
        },
        // 存储卷管理
        {
          path: 'volumes',
          name: 'Volumes',
          component: () => import('@/views/Volumes.vue')
        },
        // 设置页面
        {
          path: 'settings',
          name: 'Settings',
          component: () => import('@/views/Settings.vue')
        }
      ]
    },
    // 404页面
    {
      path: '/:pathMatch(.*)*',
      name: 'NotFound',
      component: () => import('@/views/NotFound.vue')
    }
  ]
})

// 路由守卫
router.beforeEach(async (to, from, next) => {
  const authStore = useAuthStore()
  
  // 检查是否需要认证
  if (to.meta.requiresAuth !== false) {
    // 检查是否已配置
    const isConfigured = localStorage.getItem('litedock-configured') === 'true'
    
    if (!isConfigured) {
      // 如果未配置，跳转到配置页面
      next('/setup')
      return
    }
    
    // 检查是否已登录
    if (!authStore.isAuthenticated) {
      // 尝试从token恢复登录状态
      await authStore.checkAuth()
      
      if (!authStore.isAuthenticated) {
        next('/login')
        return
      }
    }
  }
  
  // 如果已配置且已登录，访问setup或login页面时重定向到仪表盘
  if ((to.name === 'Setup' || to.name === 'Login') && authStore.isAuthenticated) {
    next('/')
    return
  }
  
  next()
})

export default router
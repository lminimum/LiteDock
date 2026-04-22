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
        {
          path: 'machines',
          name: 'Machines',
          component: () => import('@/views/Machines.vue')
        },
        {
          path: 'machines/:id',
          name: 'MachineDetail',
          component: () => import('@/views/MachineDetail.vue')
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

let authChecked = false;

router.beforeEach(async (to, _from, next) => {
  const authStore = useAuthStore();

  if (!authChecked) {
    await authStore.checkAuth();
    authChecked = true;
  }

  if (to.name === 'Setup') {
    if (authStore.isAuthenticated) {
      next('/');
      return;
    }
    const currentSetupStatus = await authStore.checkSetupStatus();
    if (currentSetupStatus) {
      next('/login');
      return;
    }
    next();
    return;
  }

  if (to.name === 'Login') {
    if (authStore.isAuthenticated) {
      next('/');
      return;
    }
    const currentSetupStatus = await authStore.checkSetupStatus();
    if (!currentSetupStatus) {
      next('/setup');
      return;
    }
    next();
    return;
  }

  if (to.meta.requiresAuth !== false) {
    if (!authStore.isAuthenticated) {
      const currentSetupStatus = await authStore.checkSetupStatus();
      if (!currentSetupStatus) {
        next('/setup');
        return;
      }
      next('/login');
      return;
    }
  }

  next();
});

export default router
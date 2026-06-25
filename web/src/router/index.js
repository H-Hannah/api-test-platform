import { createRouter, createWebHistory } from 'vue-router'
import MainLayout from '@/layouts/MainLayout.vue'
import { getToken } from '@/api/client'

const routes = [
  {
    path: '/login',
    name: 'login',
    component: () => import('@/views/Login.vue'),
    meta: { public: true, title: '登录' }
  },
  {
    path: '/',
    component: MainLayout,
    meta: { requiresAuth: true },
    children: [
      { path: '', name: 'home', component: () => import('@/views/Home.vue'), meta: { title: '首页' } },
      { path: 'apis', name: 'apis', component: () => import('@/views/ApiWorkbench.vue'), meta: { title: '接口定义' } },
      {
        path: 'environments',
        name: 'environments',
        component: () => import('@/views/Environments.vue'),
        meta: { title: '环境管理' }
      },
      {
        path: 'testdata',
        name: 'testdata',
        component: () => import('@/views/TestDataWorkbench.vue'),
        meta: { title: '测试准备 · 测试数据' }
      },
      { path: 'bdd', redirect: '/testdata' },
      { path: 'test-cases', redirect: '/testdata' },
      {
        path: 'impact',
        name: 'impact',
        component: () => import('@/views/ImpactWorkbench.vue'),
        meta: { title: '精准测试' }
      },
      { path: 'mr', redirect: '/impact' },
      { path: 'scenarios', name: 'scenarios', component: () => import('@/views/Scenarios.vue'), meta: { title: '测试场景' } },
      { path: 'reports', name: 'reports', component: () => import('@/views/Reports.vue'), meta: { title: '测试报告' } }
    ]
  }
]

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes
})

router.beforeEach((to) => {
  const token = getToken()
  if (to.meta.public) {
    if (token && to.name === 'login') {
      const redirect = to.query.redirect
      return typeof redirect === 'string' && redirect.startsWith('/') ? redirect : '/apis'
    }
    return true
  }
  if (!token) {
    return { name: 'login', query: { redirect: to.fullPath } }
  }
  return true
})

export default router

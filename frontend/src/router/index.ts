import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    // ── Public (marketing) routes ──
    {
      path: '/',
      name: 'landing',
      component: () => import('@/pages/public/LandingPage.vue'),
      meta: { public: true },
    },
    {
      path: '/features',
      name: 'features',
      component: () => import('@/pages/public/FeaturesPage.vue'),
      meta: { public: true },
    },
    {
      path: '/pricing',
      name: 'pricing',
      component: () => import('@/pages/public/PricingPage.vue'),
      meta: { public: true },
    },
    // ── App routes ──
    {
      path: '/dashboard',
      name: 'dashboard',
      component: () => import('@/pages/DashboardPage.vue'),
    },
    {
      path: '/connections',
      name: 'connections',
      component: () => import('@/pages/ConnectionsPage.vue'),
    },
    {
      path: '/agent/config',
      name: 'agent-config',
      component: () => import('@/pages/AgentConfigPage.vue'),
    },
    {
      path: '/agent/chat',
      name: 'agent-chat',
      component: () => import('@/pages/AgentChatPage.vue'),
    },
    {
      path: '/agent/log',
      name: 'agent-log',
      component: () => import('@/pages/AgentLogPage.vue'),
    },
    {
      path: '/reports',
      name: 'reports',
      component: () => import('@/pages/ReportsPage.vue'),
    },
    {
      path: '/login',
      name: 'login',
      component: () => import('@/pages/LoginPage.vue'),
    },
    {
      path: '/:pathMatch(.*)*',
      name: 'not-found',
      component: () => import('@/pages/NotFoundPage.vue'),
    },
  ],
})

router.beforeEach((to) => {
  const auth = useAuthStore()
  // Before auth is initialized, allow navigation (App.vue handles the loading state)
  if (!auth.initialized) return
  if (to.name !== 'login' && !to.meta?.public && !auth.isAuthenticated) {
    return { name: 'login' }
  }
  // Redirect authenticated users away from login page
  if (to.name === 'login' && auth.isAuthenticated) {
    return { name: 'dashboard' }
  }
})

export default router

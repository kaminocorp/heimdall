import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useAppStore } from '@/stores/app'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    // ── Public (marketing) routes — shared layout with nav + footer ──
    {
      path: '/',
      component: () => import('@/layouts/PublicLayout.vue'),
      meta: { public: true },
      children: [
        {
          path: '',
          name: 'landing',
          component: () => import('@/pages/public/LandingPage.vue'),
        },
        {
          path: 'features',
          name: 'features',
          component: () => import('@/pages/public/FeaturesPage.vue'),
        },
        {
          path: 'pricing',
          name: 'pricing',
          component: () => import('@/pages/public/PricingPage.vue'),
        },
      ],
    },
    // ── Onboarding ──
    {
      path: '/onboarding',
      name: 'onboarding',
      component: () => import('@/pages/OnboardingPage.vue'),
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
      path: '/enrichment',
      name: 'enrichment',
      component: () => import('@/pages/EnrichmentPage.vue'),
    },
    {
      path: '/outbound',
      name: 'outbound',
      component: () => import('@/pages/OutboundPage.vue'),
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
      path: '/schedules',
      name: 'schedules',
      component: () => import('@/pages/SchedulesPage.vue'),
    },
    {
      path: '/activity',
      name: 'activity',
      component: () => import('@/pages/ActivityPage.vue'),
    },
    // Legacy path kept so existing bookmarks and in-product links survive
    // the Phase 2 rename. Safe to delete after a grace period.
    {
      path: '/agent/log',
      redirect: '/activity',
    },
    {
      path: '/notifications',
      name: 'notifications',
      component: () => import('@/pages/NotificationsPage.vue'),
    },
    {
      path: '/reports',
      name: 'reports',
      component: () => import('@/pages/ReportsPage.vue'),
    },
    {
      path: '/settings',
      name: 'settings',
      component: () => import('@/pages/SettingsPage.vue'),
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
  const isPublic = to.matched.some(r => r.meta.public)
  if (to.name !== 'login' && !isPublic && !auth.isAuthenticated) {
    return { name: 'login' }
  }
  // Redirect authenticated users away from login page
  if (to.name === 'login' && auth.isAuthenticated) {
    return { name: 'dashboard' }
  }
  // Redirect to onboarding if user hasn't set up their org yet
  const app = useAppStore()
  if (
    auth.isAuthenticated &&
    app.needsOnboarding &&
    to.name !== 'onboarding' &&
    !isPublic
  ) {
    return { name: 'onboarding' }
  }
})

// After a deployment, the browser may have a cached index.html referencing
// old chunk filenames that no longer exist. Nginx serves index.html (text/html)
// as a fallback, causing a MIME type error on the dynamic import. Reload once
// to pick up the new index.html with correct asset references.
router.onError((error, to) => {
  if (
    error.message.includes('Failed to fetch dynamically imported module') ||
    error.message.includes('Importing a module script failed')
  ) {
    window.location.assign(to.fullPath)
  }
})

export default router

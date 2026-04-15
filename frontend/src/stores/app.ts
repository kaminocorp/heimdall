import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { Organization, OrganizationWithRole, Application } from '@/types/organization'
import * as orgApi from '@/api/organizations'
import * as appApi from '@/api/applications'

export const useAppStore = defineStore('app', () => {
  // ── Multi-org state ──
  const organizations = ref<OrganizationWithRole[]>([])
  const organization = ref<OrganizationWithRole | null>(null)
  const applications = ref<Application[]>([])
  const currentAppId = ref<string | null>(null)
  const loading = ref(false)
  const initialized = ref(false)
  const needsOnboarding = ref(false)
  let selectOrgSeq = 0 // guards against stale responses from rapid org switches
  let initPromise: Promise<void> | null = null // deduplicates concurrent init() calls

  const currentApp = computed(() =>
    applications.value.find((a) => a.id === currentAppId.value) ?? null
  )

  async function init() {
    // Deduplicate: if an init is already in-flight, return the same promise.
    if (initPromise) return initPromise
    initPromise = performInit()
    try {
      await initPromise
    } finally {
      initPromise = null
    }
  }

  async function performInit() {
    loading.value = true
    try {
      // Fetch all orgs the user belongs to.
      organizations.value = await orgApi.listUserOrganizations()

      if (organizations.value.length === 0) {
        needsOnboarding.value = true
        return
      }

      // Restore last selected org from localStorage, or pick first.
      const storedOrgId = localStorage.getItem('heimdall_current_org')
      const restoredOrg = storedOrgId
        ? organizations.value.find((o) => o.id === storedOrgId)
        : null
      const activeOrg = restoredOrg ?? organizations.value[0]

      // Set the current org (includes role from the list response).
      organization.value = activeOrg
      localStorage.setItem('heimdall_current_org', activeOrg.id)

      // Fetch apps for the current org.
      applications.value = await appApi.listApplications()
      needsOnboarding.value = false

      // Restore last selected app from localStorage, or pick first.
      const storedAppId = localStorage.getItem('heimdall_current_app')
      if (storedAppId && applications.value.some((a) => a.id === storedAppId)) {
        currentAppId.value = storedAppId
      } else if (applications.value.length > 0) {
        currentAppId.value = applications.value[0].id
      }
    } catch (err: unknown) {
      // Only treat 404 (no org) as needing onboarding.
      // Network errors, 500s, and other failures should not route to onboarding.
      const isAxios = err != null && typeof err === 'object' && 'response' in err
      const status = isAxios ? (err as { response?: { status?: number } }).response?.status : undefined
      if (status === 404) {
        needsOnboarding.value = true
      } else {
        // Re-throw so callers know init failed for a non-onboarding reason.
        throw err
      }
    } finally {
      loading.value = false
      initialized.value = true
    }
  }

  // Switch to a different org. Refetches apps for the new org.
  // Uses a sequence counter to discard stale responses when the user
  // switches orgs faster than the API can respond.
  async function selectOrg(orgId: string) {
    const target = organizations.value.find((o) => o.id === orgId)
    if (!target) return

    const seq = ++selectOrgSeq

    organization.value = target
    localStorage.setItem('heimdall_current_org', orgId)

    // Clear app state while loading new org's apps.
    currentAppId.value = null
    applications.value = []

    try {
      const apps = await appApi.listApplications()
      // Discard if the user switched orgs again while we were fetching.
      if (seq !== selectOrgSeq) return
      applications.value = apps
      if (apps.length > 0) {
        currentAppId.value = apps[0].id
        localStorage.setItem('heimdall_current_app', apps[0].id)
      }
    } catch (err: unknown) {
      if (seq !== selectOrgSeq) return
      // Only treat 404 (no apps endpoint not found) as "org has no apps".
      // Network errors, 500s, etc. should propagate so the UI can show them.
      const isAxios = err != null && typeof err === 'object' && 'response' in err
      const status = isAxios ? (err as { response?: { status?: number } }).response?.status : undefined
      if (status === 404) {
        applications.value = []
      } else {
        throw err
      }
    }
  }

  function selectApp(appId: string) {
    currentAppId.value = appId
    localStorage.setItem('heimdall_current_app', appId)
  }

  async function onboard(orgName: string, orgSlug: string, appName: string) {
    const result = await orgApi.onboard({ org_name: orgName, org_slug: orgSlug, app_name: appName })
    organization.value = { ...result.organization, role: 'owner' }
    organizations.value = [{ ...result.organization, role: 'owner' }]
    applications.value = [result.application]
    currentAppId.value = result.application.id
    localStorage.setItem('heimdall_current_org', result.organization.id)
    localStorage.setItem('heimdall_current_app', result.application.id)
    needsOnboarding.value = false
    return result
  }

  async function createApp(name: string, { select = true }: { select?: boolean } = {}) {
    const app = await appApi.createApplication(name)
    applications.value.unshift(app)
    if (select) {
      selectApp(app.id)
    }
    return app
  }

  async function deleteApp(appId: string) {
    await appApi.deleteApplication(appId)
    applications.value = applications.value.filter((a) => a.id !== appId)
    if (currentAppId.value === appId) {
      const fallback = applications.value[0]?.id ?? null
      if (fallback) {
        selectApp(fallback)
      } else {
        currentAppId.value = null
        localStorage.removeItem('heimdall_current_app')
      }
    }
  }

  // Add a newly created org to the local list and select it.
  async function addOrg(org: Organization) {
    const withRole: OrganizationWithRole = { ...org, role: 'owner' }
    organizations.value.push(withRole)
    await selectOrg(org.id)
  }

  // Update the current org's name and slug in both `organization` and the
  // `organizations` list so all UI surfaces (header, dropdown) stay in sync.
  function updateOrg(updated: { name: string; slug: string }) {
    if (organization.value) {
      organization.value = { ...organization.value, ...updated }
    }
    const idx = organizations.value.findIndex((o) => o.id === organization.value?.id)
    if (idx !== -1) {
      organizations.value[idx] = { ...organizations.value[idx], ...updated }
    }
  }

  function reset() {
    organization.value = null
    organizations.value = []
    applications.value = []
    currentAppId.value = null
    needsOnboarding.value = false
    initialized.value = false
    loading.value = false
    selectOrgSeq = 0
    initPromise = null
    localStorage.removeItem('heimdall_current_org')
    localStorage.removeItem('heimdall_current_app')
  }

  return {
    organization,
    organizations,
    applications,
    currentAppId,
    currentApp,
    loading,
    initialized,
    needsOnboarding,
    init,
    selectOrg,
    selectApp,
    onboard,
    createApp,
    deleteApp,
    addOrg,
    updateOrg,
    reset,
  }
})

import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { Organization, Application } from '@/types/organization'
import * as orgApi from '@/api/organizations'
import * as appApi from '@/api/applications'

export const useAppStore = defineStore('app', () => {
  const organization = ref<Organization | null>(null)
  const applications = ref<Application[]>([])
  const currentAppId = ref<string | null>(null)
  const loading = ref(false)
  const needsOnboarding = ref(false)

  const currentApp = computed(() =>
    applications.value.find((a) => a.id === currentAppId.value) ?? null
  )

  async function init() {
    loading.value = true
    try {
      organization.value = await orgApi.getOrganization()
      applications.value = await appApi.listApplications()
      needsOnboarding.value = false

      // Restore last selected app from localStorage, or pick first
      const stored = localStorage.getItem('heimdall_current_app')
      if (stored && applications.value.some((a) => a.id === stored)) {
        currentAppId.value = stored
      } else if (applications.value.length > 0) {
        currentAppId.value = applications.value[0].id
      }
    } catch {
      // No org found — user needs onboarding
      needsOnboarding.value = true
    } finally {
      loading.value = false
    }
  }

  function selectApp(appId: string) {
    currentAppId.value = appId
    localStorage.setItem('heimdall_current_app', appId)
  }

  async function onboard(orgName: string, orgSlug: string, appName: string) {
    const result = await orgApi.onboard({ org_name: orgName, org_slug: orgSlug, app_name: appName })
    organization.value = result.organization
    applications.value = [result.application]
    currentAppId.value = result.application.id
    localStorage.setItem('heimdall_current_app', result.application.id)
    needsOnboarding.value = false
    return result
  }

  // `select` controls whether the new app becomes the currently-selected one
  // immediately after creation. Defaults to true so the common "user clicks
  // +Create" path stays idiomatic. The AppWizard passes { select: false }
  // because it creates the app *eagerly* at step 1 and only commits the
  // selection when the user reaches the final step — if we auto-selected the
  // draft and the user discarded mid-wizard, their previous selection would
  // be silently lost.
  async function createApp(name: string, { select = true }: { select?: boolean } = {}) {
    const app = await appApi.createApplication(name)
    applications.value.unshift(app)
    if (select) {
      selectApp(app.id)
    }
    return app
  }

  // Deletes an application via the backend and reconciles local state.
  //
  // Reconciliation rules:
  //   1. The row is removed from the local `applications` list optimistically
  //      *after* the server call succeeds (no rollback needed).
  //   2. If the deleted app was the currently-selected one, we fall back to
  //      the first remaining app so the sidebar selector never ends up in an
  //      invalid "points to nothing" state.
  //   3. The backend's last-app guard (409 + code: "last_app") is surfaced
  //      as-is — the caller is expected to either check upfront (Settings
  //      page disables the button) or catch and render the error (the
  //      wizard's discard path should never hit this because it's deleting
  //      a *newly-created* app).
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

  function reset() {
    organization.value = null
    applications.value = []
    currentAppId.value = null
    needsOnboarding.value = false
  }

  return {
    organization,
    applications,
    currentAppId,
    currentApp,
    loading,
    needsOnboarding,
    init,
    selectApp,
    onboard,
    createApp,
    deleteApp,
    reset,
  }
})

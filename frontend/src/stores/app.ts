import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { Organization, Application, AppAgentConfig, MonitoringStatus } from '@/types/organization'
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

  async function createApp(name: string) {
    const app = await appApi.createApplication(name)
    applications.value.unshift(app)
    selectApp(app.id)
    return app
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
    reset,
  }
})

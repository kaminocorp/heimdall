import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { AgentConfig } from '@/types/agent'
import * as agentApi from '@/api/agent'

export const useAgentStore = defineStore('agent', () => {
  const config = ref<AgentConfig | null>(null)
  const loading = ref(false)

  async function fetchConfig() {
    loading.value = true
    try {
      const { data } = await agentApi.getAgentConfig()
      config.value = data
    } finally {
      loading.value = false
    }
  }

  return { config, loading, fetchConfig }
})

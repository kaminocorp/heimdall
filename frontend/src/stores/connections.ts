import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { Connection } from '@/types/connection'
import * as connectionsApi from '@/api/connections'

export const useConnectionsStore = defineStore('connections', () => {
  const connections = ref<Connection[]>([])
  const loading = ref(false)

  async function fetchConnections() {
    loading.value = true
    try {
      const { data } = await connectionsApi.listConnections()
      connections.value = data
    } finally {
      loading.value = false
    }
  }

  return { connections, loading, fetchConnections }
})

import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { Connection, CreateConnectionPayload, UpdateConnectionPayload } from '@/types/connection'
import * as connectionsApi from '@/api/connections'

export const useConnectionsStore = defineStore('connections', () => {
  const connections = ref<Connection[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)
  const testingId = ref<string | null>(null)

  async function fetchConnections() {
    loading.value = true
    error.value = null
    try {
      const { data } = await connectionsApi.listConnections()
      connections.value = data
    } catch (e: any) {
      error.value = e.response?.data?.error ?? 'Failed to load connections'
    } finally {
      loading.value = false
    }
  }

  async function createConnection(payload: CreateConnectionPayload) {
    const { data } = await connectionsApi.createConnection(payload)
    connections.value.unshift(data)
    return data
  }

  async function updateConnection(id: string, payload: UpdateConnectionPayload) {
    const { data } = await connectionsApi.updateConnection(id, payload)
    const idx = connections.value.findIndex((c) => c.id === id)
    if (idx !== -1) connections.value[idx] = data
    return data
  }

  async function testConnection(id: string) {
    testingId.value = id
    try {
      const { data } = await connectionsApi.testConnection(id)
      const idx = connections.value.findIndex((c) => c.id === id)
      if (idx !== -1) {
        connections.value[idx].status = data.success ? 'active' : 'error'
      }
      return data
    } finally {
      testingId.value = null
    }
  }

  async function deleteConnection(id: string) {
    await connectionsApi.deleteConnection(id)
    connections.value = connections.value.filter((c) => c.id !== id)
  }

  return { connections, loading, error, testingId, fetchConnections, createConnection, updateConnection, testConnection, deleteConnection }
})

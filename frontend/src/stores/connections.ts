import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { Connection, CreateConnectionPayload, UpdateConnectionPayload } from '@/types/connection'
import * as connectionsApi from '@/api/connections'
import { listConnectionsByApp } from '@/api/applications'
import { extractApiError } from '@/utils/apiError'

export const useConnectionsStore = defineStore('connections', () => {
  const connections = ref<Connection[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)
  const testingId = ref<string | null>(null)

  async function fetchConnections() {
    loading.value = true
    error.value = null
    try {
      connections.value = await connectionsApi.listConnections()
    } catch (e: unknown) {
      error.value = extractApiError(e, 'Failed to load connections')
    } finally {
      loading.value = false
    }
  }

  async function fetchConnectionsByApp(appId: string) {
    loading.value = true
    error.value = null
    try {
      connections.value = await listConnectionsByApp(appId)
    } catch (e: unknown) {
      error.value = extractApiError(e, 'Failed to load connections')
    } finally {
      loading.value = false
    }
  }

  async function createConnection(payload: CreateConnectionPayload) {
    const conn = await connectionsApi.createConnection(payload)
    connections.value.unshift(conn)
    return conn
  }

  async function updateConnection(id: string, payload: UpdateConnectionPayload) {
    const updated = await connectionsApi.updateConnection(id, payload)
    const idx = connections.value.findIndex((c) => c.id === id)
    if (idx !== -1) connections.value[idx] = updated
    return updated
  }

  async function testConnection(id: string) {
    testingId.value = id
    try {
      const result = await connectionsApi.testConnection(id)
      const idx = connections.value.findIndex((c) => c.id === id)
      if (idx !== -1 && connections.value[idx].status !== 'paused') {
        connections.value[idx].status = result.success ? 'active' : 'error'
      }
      return result
    } finally {
      testingId.value = null
    }
  }

  async function pauseConnection(id: string) {
    const conn = connections.value.find((c) => c.id === id)
    if (!conn) return
    return updateConnection(id, {
      name: conn.name,
      type: conn.type,
      direction: conn.direction,
      config: conn.config,
      status: 'paused',
    })
  }

  async function resumeConnection(id: string) {
    const conn = connections.value.find((c) => c.id === id)
    if (!conn) return
    return updateConnection(id, {
      name: conn.name,
      type: conn.type,
      direction: conn.direction,
      config: conn.config,
      status: 'active',
    })
  }

  async function deleteConnection(id: string) {
    await connectionsApi.deleteConnection(id)
    connections.value = connections.value.filter((c) => c.id !== id)
  }

  return { connections, loading, error, testingId, fetchConnections, fetchConnectionsByApp, createConnection, updateConnection, pauseConnection, resumeConnection, testConnection, deleteConnection }
})

import axios from 'axios'
import { useAuthStore } from '@/stores/auth'
import { useToast } from '@/composables/useToast'

const client = axios.create({
  baseURL: '/api',
  headers: {
    'Content-Type': 'application/json',
  },
})

client.interceptors.request.use((config) => {
  const auth = useAuthStore()
  if (auth.token) {
    config.headers.Authorization = `Bearer ${auth.token}`
  }
  // Send the active org context so the backend knows which org to scope to.
  // Read from localStorage to avoid a circular dependency (app store imports
  // api modules which import this client).
  const orgId = localStorage.getItem('heimdall_current_org')
  if (orgId) {
    config.headers['X-Org-ID'] = orgId
  }
  return config
})

// Response interceptor — global safety net for auth and server errors.
// The isLoggingOut flag deduplicates 401 handling: when a session expires,
// multiple in-flight requests may all receive 401 simultaneously. Without
// the guard, each one independently calls logout() and assigns location.href,
// producing redundant Supabase signOut calls and console noise.
let isLoggingOut = false

client.interceptors.response.use(
  (response) => response,
  async (error) => {
    if (!error.response) {
      // Network error
      useToast().show('Connection lost — check your network', 'error', 6000)
    } else if (error.response.status === 401) {
      if (!isLoggingOut) {
        isLoggingOut = true
        const auth = useAuthStore()
        try {
          await auth.logout()
        } catch {
          // Logout failed — proceed to redirect anyway.
        }
        // Never reset isLoggingOut — the hard redirect destroys the JS
        // context. Resetting before the redirect opens a race window where
        // other in-flight 401s re-enter this handler.
        window.location.href = '/login'
      }
    } else if (error.response.status >= 500) {
      useToast().show('Server error — please try again', 'error')
    }
    return Promise.reject(error)
  }
)

export default client

import axios from 'axios'
import { useAuthStore } from '@/stores/auth'

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
  return config
})

// Response interceptor — global safety net for auth and server errors
client.interceptors.response.use(
  (response) => response,
  (error) => {
    if (!error.response) {
      // Network error
      import('@/composables/useToast').then(({ useToast }) => {
        useToast().show('Connection lost — check your network', 'error', 6000)
      })
    } else if (error.response.status === 401) {
      // Session expired — log out and redirect
      const auth = useAuthStore()
      auth.logout()
      window.location.href = '/login'
    } else if (error.response.status >= 500) {
      import('@/composables/useToast').then(({ useToast }) => {
        useToast().show('Server error — please try again', 'error')
      })
    }
    return Promise.reject(error)
  }
)

export default client

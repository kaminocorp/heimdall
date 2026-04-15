import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { supabase } from '@/lib/supabase'
import type { Session, User, Subscription } from '@supabase/supabase-js'

export const useAuthStore = defineStore('auth', () => {
  const session = ref<Session | null>(null)
  const user = ref<User | null>(null)
  const initialized = ref(false)

  // Track the auth state subscription so we can clean it up.
  let authSubscription: Subscription | null = null

  const token = computed(() => session.value?.access_token ?? null)
  const isAuthenticated = computed(() => !!session.value)

  async function init() {
    // Guard against double-init (e.g. HMR, test re-runs). Unsubscribe the
    // previous listener before registering a new one.
    if (authSubscription) {
      authSubscription.unsubscribe()
      authSubscription = null
    }

    const { data } = await supabase.auth.getSession()

    // If a cached session exists, force a token refresh so we never send
    // a stale JWT to the backend (e.g. after a redeploy or token expiry).
    if (data.session) {
      const { data: refreshed, error: refreshError } = await supabase.auth.refreshSession()
      if (refreshError || !refreshed.session) {
        // Distinguish auth errors (clear session) from transient errors (keep cached session).
        const isAuthError = refreshError?.status === 401 || refreshError?.status === 403
          || refreshError?.message?.includes('Invalid Refresh Token')
          || refreshError?.message?.includes('Refresh Token Not Found')
        if (isAuthError || !refreshError) {
          console.warn('Auth: session refresh failed, clearing session', refreshError)
          session.value = null
          user.value = null
        } else {
          // Transient error (network, 500) — keep the cached session so
          // the user isn't logged out by a brief server hiccup.
          console.warn('Auth: session refresh failed (transient), keeping cached session', refreshError)
          session.value = data.session
          user.value = data.session.user ?? null
        }
      } else {
        session.value = refreshed.session
        user.value = refreshed.session.user ?? null
      }
    } else {
      session.value = null
      user.value = null
    }

    const { data: { subscription } } = supabase.auth.onAuthStateChange((_event, newSession) => {
      const wasAuthenticated = !!session.value
      session.value = newSession
      user.value = newSession?.user ?? null
      // If session was cleared (token revocation, timeout), reset the app store
      // so stale org/app data from the previous user doesn't linger in memory.
      if (wasAuthenticated && !newSession) {
        import('@/stores/app').then(({ useAppStore }) => useAppStore().reset())
      }
    })
    authSubscription = subscription

    initialized.value = true
  }

  async function login(email: string, password: string) {
    const { error } = await supabase.auth.signInWithPassword({ email, password })
    if (error) throw error
  }

  async function signup(email: string, password: string) {
    const { error } = await supabase.auth.signUp({ email, password })
    if (error) throw error
  }

  async function logout() {
    const { error } = await supabase.auth.signOut()
    if (error) throw error
  }

  return { token, isAuthenticated, initialized, user, init, login, signup, logout }
})

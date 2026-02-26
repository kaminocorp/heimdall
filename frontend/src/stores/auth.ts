import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { supabase } from '@/lib/supabase'
import type { Session, User } from '@supabase/supabase-js'

export const useAuthStore = defineStore('auth', () => {
  const session = ref<Session | null>(null)
  const user = ref<User | null>(null)
  const initialized = ref(false)

  const token = computed(() => session.value?.access_token ?? null)
  const isAuthenticated = computed(() => !!session.value)

  async function init() {
    const { data } = await supabase.auth.getSession()

    // If a cached session exists, force a token refresh so we never send
    // a stale JWT to the backend (e.g. after a redeploy or token expiry).
    if (data.session) {
      const { data: refreshed } = await supabase.auth.refreshSession()
      session.value = refreshed.session
      user.value = refreshed.session?.user ?? null
    } else {
      session.value = null
      user.value = null
    }

    supabase.auth.onAuthStateChange((_event, newSession) => {
      session.value = newSession
      user.value = newSession?.user ?? null
    })

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

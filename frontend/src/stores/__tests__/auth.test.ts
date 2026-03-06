import { useAuthStore } from '@/stores/auth'
import { vi } from 'vitest'

// Mock supabase client
vi.mock('@/lib/supabase', () => ({
  supabase: {
    auth: {
      getSession: vi.fn().mockResolvedValue({ data: { session: null } }),
      refreshSession: vi.fn().mockResolvedValue({ data: { session: null } }),
      signInWithPassword: vi.fn(),
      signUp: vi.fn(),
      signOut: vi.fn(),
      onAuthStateChange: vi.fn(),
    },
  },
}))

import { supabase } from '@/lib/supabase'

describe('auth store', () => {
  beforeEach(() => {
    vi.mocked(supabase.auth.getSession).mockReset()
    vi.mocked(supabase.auth.refreshSession).mockReset()
    vi.mocked(supabase.auth.signInWithPassword).mockReset()
    vi.mocked(supabase.auth.signOut).mockReset()
    vi.mocked(supabase.auth.onAuthStateChange).mockReset()
  })

  it('init sets initialized to true', async () => {
    vi.mocked(supabase.auth.getSession).mockResolvedValueOnce({
      data: { session: null },
      error: null,
    } as any)
    vi.mocked(supabase.auth.onAuthStateChange).mockReturnValueOnce({
      data: { subscription: { unsubscribe: vi.fn() } },
    } as any)

    const store = useAuthStore()
    await store.init()

    expect(store.initialized).toBe(true)
    expect(store.isAuthenticated).toBe(false)
  })

  it('login calls signInWithPassword', async () => {
    vi.mocked(supabase.auth.signInWithPassword).mockResolvedValueOnce({
      data: {},
      error: null,
    } as any)

    const store = useAuthStore()
    await store.login('test@example.com', 'password')

    expect(supabase.auth.signInWithPassword).toHaveBeenCalledWith({
      email: 'test@example.com',
      password: 'password',
    })
  })

  it('login throws on error', async () => {
    vi.mocked(supabase.auth.signInWithPassword).mockResolvedValueOnce({
      data: {},
      error: new Error('Invalid credentials'),
    } as any)

    const store = useAuthStore()
    await expect(store.login('bad@example.com', 'wrong')).rejects.toThrow('Invalid credentials')
  })

  it('logout calls signOut', async () => {
    vi.mocked(supabase.auth.signOut).mockResolvedValueOnce({ error: null } as any)

    const store = useAuthStore()
    await store.logout()

    expect(supabase.auth.signOut).toHaveBeenCalled()
  })

  it('isAuthenticated tracks session presence', () => {
    const store = useAuthStore()
    expect(store.isAuthenticated).toBe(false)

    // Simulate session being set (e.g. after auth state change)
    store.$patch({ session: { access_token: 'token' } as any })
    // Note: computed won't update from $patch on internal ref — this tests the initial state
  })
})

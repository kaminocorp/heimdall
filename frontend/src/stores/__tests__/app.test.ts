import { useAppStore } from '@/stores/app'
import client from '@/api/client'
import type { Application } from '@/types/organization'

// Minimal axios-response stand-in so we only have to type the .data field.
type MockAxiosResponse<T> = { data: T }

// Baseline fixture used across tests. Helpers mutate the returned object.
function fixtureApp(overrides: Partial<Application> = {}): Application {
  return {
    id: 'app-1',
    org_id: 'org-1',
    name: 'App One',
    status: 'active',
    created_at: '2026-04-11T12:00:00Z',
    updated_at: '2026-04-11T12:00:00Z',
    ...overrides,
  }
}

describe('app store', () => {
  beforeEach(() => {
    vi.mocked(client.get).mockReset()
    vi.mocked(client.post).mockReset()
    vi.mocked(client.delete).mockReset()
    localStorage.clear()
  })

  // createApp defaults to { select: true }, which preserves the pre-wizard
  // behaviour every existing caller depends on. We lock this explicitly so a
  // future "let's always make select opt-in" refactor can't silently break
  // onboarding or anything else that assumes the newly-created app is also
  // the current one.
  it('createApp selects the new app by default', async () => {
    const created = fixtureApp({ id: 'app-2', name: 'Fresh' })
    vi.mocked(client.post).mockResolvedValueOnce({ data: created } as MockAxiosResponse<Application>)

    const store = useAppStore()
    store.applications = [fixtureApp()]
    store.currentAppId = 'app-1'

    const result = await store.createApp('Fresh')

    expect(result.id).toBe('app-2')
    expect(store.applications).toHaveLength(2)
    expect(store.applications[0].id).toBe('app-2') // unshifted to head
    expect(store.currentAppId).toBe('app-2')
    expect(localStorage.getItem('heimdall_current_app')).toBe('app-2')
  })

  // The AppWizard's eager-create path passes { select: false } so the
  // currently-selected app stays put. If the user discards mid-wizard we
  // can roll back without touching `currentAppId` at all.
  it('createApp with { select: false } leaves currentAppId untouched', async () => {
    const draft = fixtureApp({ id: 'draft-1', name: 'Draft' })
    vi.mocked(client.post).mockResolvedValueOnce({ data: draft } as MockAxiosResponse<Application>)

    const store = useAppStore()
    store.applications = [fixtureApp({ id: 'app-1' })]
    store.currentAppId = 'app-1'

    await store.createApp('Draft', { select: false })

    expect(store.applications[0].id).toBe('draft-1') // still unshifted
    expect(store.currentAppId).toBe('app-1') // selection preserved
    expect(localStorage.getItem('heimdall_current_app')).toBeNull()
  })

  it('deleteApp removes the row from the list', async () => {
    vi.mocked(client.delete).mockResolvedValueOnce({ data: undefined } as MockAxiosResponse<undefined>)

    const store = useAppStore()
    store.applications = [
      fixtureApp({ id: 'app-1' }),
      fixtureApp({ id: 'app-2', name: 'App Two' }),
    ]
    store.currentAppId = 'app-1'

    await store.deleteApp('app-2')

    expect(store.applications).toHaveLength(1)
    expect(store.applications[0].id).toBe('app-1')
    // Current selection wasn't the deleted one, so it stays put.
    expect(store.currentAppId).toBe('app-1')
  })

  // The selection-fallback path is what keeps the sidebar selector
  // consistent when a user deletes the app they're currently viewing.
  // Without it, currentAppId would point at a dead row and every per-app
  // page would render a stale state.
  it('deleteApp falls back to the first remaining app when current is deleted', async () => {
    vi.mocked(client.delete).mockResolvedValueOnce({ data: undefined } as MockAxiosResponse<undefined>)

    const store = useAppStore()
    store.applications = [
      fixtureApp({ id: 'app-1' }),
      fixtureApp({ id: 'app-2', name: 'App Two' }),
    ]
    store.currentAppId = 'app-1'

    await store.deleteApp('app-1')

    expect(store.applications).toHaveLength(1)
    expect(store.applications[0].id).toBe('app-2')
    expect(store.currentAppId).toBe('app-2')
    expect(localStorage.getItem('heimdall_current_app')).toBe('app-2')
  })

  // Edge case: deleting every app (e.g. a rogue admin path or a cleanup
  // script) must leave the store in a defined state instead of pointing
  // at a dead id. In production the backend's last-app guard would block
  // this, but the store contract shouldn't assume that.
  it('deleteApp clears currentAppId when no apps remain', async () => {
    vi.mocked(client.delete).mockResolvedValueOnce({ data: undefined } as MockAxiosResponse<undefined>)

    const store = useAppStore()
    store.applications = [fixtureApp({ id: 'app-1' })]
    store.currentAppId = 'app-1'
    localStorage.setItem('heimdall_current_app', 'app-1')

    await store.deleteApp('app-1')

    expect(store.applications).toHaveLength(0)
    expect(store.currentAppId).toBeNull()
    expect(localStorage.getItem('heimdall_current_app')).toBeNull()
  })

  // The backend's last-app guard returns 409 with { code: "last_app" }.
  // The store doesn't special-case this — it lets the error bubble up so
  // the caller (e.g. Settings page DeleteAppModal) can render the error
  // text inline. We just verify the local list isn't mutated on failure.
  it('deleteApp propagates backend errors without touching local state', async () => {
    vi.mocked(client.delete).mockRejectedValueOnce({
      response: { data: { error: 'cannot delete last application', code: 'last_app' } },
    })

    const store = useAppStore()
    store.applications = [fixtureApp({ id: 'app-1' })]
    store.currentAppId = 'app-1'

    await expect(store.deleteApp('app-1')).rejects.toBeDefined()
    // Failure path: list and selection are untouched.
    expect(store.applications).toHaveLength(1)
    expect(store.currentAppId).toBe('app-1')
  })
})

import { createPinia, setActivePinia } from 'pinia'
import { vi } from 'vitest'

// Create a fresh Pinia for each test
beforeEach(() => {
  setActivePinia(createPinia())
})

// Mock the API client globally
vi.mock('@/api/client', () => ({
  default: {
    get: vi.fn(),
    post: vi.fn(),
    put: vi.fn(),
    patch: vi.fn(),
    delete: vi.fn(),
    interceptors: {
      request: { use: vi.fn() },
      response: { use: vi.fn() },
    },
  },
}))

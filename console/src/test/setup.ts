import '@testing-library/jest-dom'
import { vi } from 'vitest'

vi.mock('@/lib/api-client', async () => {
  const actual = await vi.importActual('@/lib/api-client')
  return {
    ...actual,
    apiClient: {
      listNodes: vi.fn(),
      createNode: vi.fn(),
    },
  }
})

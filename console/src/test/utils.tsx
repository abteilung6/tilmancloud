import { render, type RenderOptions } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import type { ReactElement } from 'react'
import type { AxiosResponse, InternalAxiosRequestConfig } from 'axios'

const createTestQueryClient = () =>
  new QueryClient({
    defaultOptions: {
      queries: {
        retry: false,
        gcTime: 0,
      },
    },
  })

export const renderWithQuery = (ui: ReactElement, options?: Omit<RenderOptions, 'wrapper'>) => {
  const queryClient = createTestQueryClient()
  const Wrapper = ({ children }: { children: React.ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  )
  return render(ui, { wrapper: Wrapper, ...options })
}

interface MockResponseOptions<T> {
  data: T
  status?: number
  statusText?: string
  headers?: Record<string, string>
}

export function mockAxiosResponse<T>({
  data,
  status = 200,
  statusText = 'OK',
  headers = {},
}: MockResponseOptions<T>): AxiosResponse<T> {
  return {
    data,
    status,
    statusText,
    headers,
    config: {} as InternalAxiosRequestConfig,
  }
}

// Re-export everything from React Testing Library
// eslint-disable-next-line react-refresh/only-export-components
export * from '@testing-library/react'

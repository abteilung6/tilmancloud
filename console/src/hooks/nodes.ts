import { useQuery } from '@tanstack/react-query'
import { apiClient } from '@/lib/api-client'

export const useNodes = () => {
  return useQuery({
    queryKey: ['nodes'],
    queryFn: async () => {
      const response = await apiClient.listNodes()
      return response.data
    },
  })
}

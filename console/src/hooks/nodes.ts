import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { apiClient } from '@/lib/api-client'

const QUERY_KEY_NODES = 'nodes'

export const useNodesQuery = () => {
  return useQuery({
    queryKey: [QUERY_KEY_NODES],
    queryFn: async () => {
      const response = await apiClient.listNodes()
      return response.data
    },
  })
}

export const useCreateNode = () => {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async () => {
      const response = await apiClient.createNode()
      return response.data
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: [QUERY_KEY_NODES] })
    },
  })
}

export const useDeleteNode = () => {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (nodeId: string) => {
      await apiClient.deleteNode(nodeId)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: [QUERY_KEY_NODES] })
    },
  })
}

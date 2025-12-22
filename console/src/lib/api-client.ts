import { DefaultApi, Configuration } from './api'

const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080'

const config = new Configuration({
  basePath: API_URL,
})

export const apiClient = new DefaultApi(config)
export { type Node, type Health, NodeStateEnum } from './api/models'

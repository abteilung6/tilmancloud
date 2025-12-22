import { useState, useEffect } from 'react'
import { apiClient, type Node } from './lib/api-client'

const App: React.FC = () => {
  const [nodes, setNodes] = useState<Node[] | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    const fetchNodes = async () => {
      try {
        setLoading(true)
        setError(null)
        const response = await apiClient.listNodes()
        setNodes(response.data)
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to fetch nodes')
      } finally {
        setLoading(false)
      }
    }

    fetchNodes()
  }, [])

  if (loading) {
    return <div>Loading...</div>
  }

  if (error) {
    return <div>Error: {error}</div>
  }

  if (!nodes || nodes.length === 0) {
    return <div>No nodes found</div>
  }

  return (
    <table>
      <thead>
        <tr>
          <th>Name</th>
          <th>State</th>
          <th>Instance Type</th>
          <th>Public IP</th>
          <th>Private IP</th>
        </tr>
      </thead>
      <tbody>
        {nodes.map(node => (
          <tr key={node.name}>
            <td>{node.name}</td>
            <td>{node.state || 'N/A'}</td>
            <td>{node.instanceType || 'N/A'}</td>
            <td>{node.publicIp || 'N/A'}</td>
            <td>{node.privateIp || 'N/A'}</td>
          </tr>
        ))}
      </tbody>
    </table>
  )
}

export default App

import { useNodes } from './hooks/nodes'

const App: React.FC = () => {
  const { data: nodes, isLoading, isError, error } = useNodes()

  if (isLoading) {
    return <div>Loading...</div>
  }

  if (isError) {
    return <div>Error: {error instanceof Error ? error.message : 'Failed to fetch nodes'}</div>
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

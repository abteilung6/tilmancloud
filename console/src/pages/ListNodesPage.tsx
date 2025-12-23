import { useNodes } from '@/hooks/nodes'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'

const ListNodesPage: React.FC = () => {
  const { data: nodes, isLoading, isError, error } = useNodes()

  if (isLoading) {
    return <div className="p-8">Loading...</div>
  }

  if (isError) {
    return (
      <div className="p-8 text-destructive">
        Error: {error instanceof Error ? error.message : 'Failed to fetch nodes'}
      </div>
    )
  }

  if (!nodes || nodes.length === 0) {
    return <div className="p-8">No nodes found</div>
  }

  return (
    <div className="p-8">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Name</TableHead>
            <TableHead>State</TableHead>
            <TableHead>Instance Type</TableHead>
            <TableHead>Public IP</TableHead>
            <TableHead>Private IP</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {nodes.map(node => (
            <TableRow key={node.name}>
              <TableCell className="font-medium">{node.name}</TableCell>
              <TableCell>{node.state || 'N/A'}</TableCell>
              <TableCell>{node.instanceType || 'N/A'}</TableCell>
              <TableCell>{node.publicIp || 'N/A'}</TableCell>
              <TableCell>{node.privateIp || 'N/A'}</TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  )
}

export default ListNodesPage

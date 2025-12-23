import { useNodes } from '@/hooks/nodes'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Badge } from '@/components/ui/badge'
import { NodeStateEnum } from '@/lib/api-client'

const getStateBadgeVariant = (
  state?: NodeStateEnum
): 'default' | 'secondary' | 'destructive' | 'outline' => {
  if (!state) return 'outline'
  switch (state) {
    case NodeStateEnum.Running:
      return 'default'
    case NodeStateEnum.Pending:
      return 'secondary'
    case NodeStateEnum.Stopping:
    case NodeStateEnum.Stopped:
      return 'outline'
    case NodeStateEnum.ShuttingDown:
    case NodeStateEnum.Terminated:
      return 'destructive'
    default:
      return 'outline'
  }
}

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
              <TableCell>
                {node.state ? (
                  <Badge variant={getStateBadgeVariant(node.state)}>{node.state}</Badge>
                ) : (
                  <Badge variant="outline">N/A</Badge>
                )}
              </TableCell>
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

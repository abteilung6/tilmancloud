import { useNodesQuery, useCreateNode } from '@/hooks/nodes'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
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
  const useNodesQueryResult = useNodesQuery()
  const nodes = useNodesQueryResult.data ?? []
  const createNodeMutation = useCreateNode()

  const handleCreateNode = () => {
    createNodeMutation.mutate()
  }

  return (
    <div className="p-8">
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-semibold">Nodes</h1>
        <Button onClick={handleCreateNode}>Add node</Button>
      </div>
      {useNodesQueryResult.isError && (
        <div className="text-destructive">
          Error:{' '}
          {useNodesQueryResult.error instanceof Error
            ? useNodesQueryResult.error.message
            : 'Failed to fetch nodes'}
        </div>
      )}
      {createNodeMutation.isError && (
        <div className="text-destructive">
          Error:{' '}
          {createNodeMutation.error instanceof Error
            ? createNodeMutation.error.message
            : 'Failed to create node'}
        </div>
      )}
      {useNodesQueryResult.isLoading && <div>Loading...</div>}
      {nodes.length > 0 && (
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
      )}
    </div>
  )
}

export default ListNodesPage

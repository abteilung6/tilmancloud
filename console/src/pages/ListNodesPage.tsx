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
import { NodeStateEnum, type Node } from '@/lib/api-client'
import { useReactTable, getCoreRowModel, type ColumnDef, flexRender } from '@tanstack/react-table'

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

const columns: ColumnDef<Node>[] = [
  {
    accessorKey: 'name',
    header: 'Name',
    cell: ({ row }) => <div className="font-medium">{row.getValue('name')}</div>,
  },
  {
    accessorKey: 'state',
    header: 'State',
    cell: ({ row }) => {
      const state = row.getValue<NodeStateEnum | undefined>('state')
      return state ? (
        <Badge variant={getStateBadgeVariant(state)}>{state}</Badge>
      ) : (
        <Badge variant="outline">N/A</Badge>
      )
    },
  },
  {
    accessorKey: 'instanceType',
    header: 'Instance Type',
    cell: ({ row }) => {
      const instanceType = row.getValue<string | undefined>('instanceType')
      return <div>{instanceType || 'N/A'}</div>
    },
  },
  {
    accessorKey: 'publicIp',
    header: 'Public IP',
    cell: ({ row }) => {
      const publicIp = row.getValue<string | null | undefined>('publicIp')
      return <div>{publicIp || 'N/A'}</div>
    },
  },
  {
    accessorKey: 'privateIp',
    header: 'Private IP',
    cell: ({ row }) => {
      const privateIp = row.getValue<string | null | undefined>('privateIp')
      return <div>{privateIp || 'N/A'}</div>
    },
  },
]

const ListNodesPage: React.FC = () => {
  const useNodesQueryResult = useNodesQuery()
  const nodes = useNodesQueryResult.data
  const createNodeMutation = useCreateNode()

  // eslint-disable-next-line react-hooks/incompatible-library
  const table = useReactTable({
    data: nodes ?? [],
    columns,
    getCoreRowModel: getCoreRowModel(),
  })

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
      {nodes !== undefined && (
        <Table>
          <TableHeader>
            {table.getHeaderGroups().map(headerGroup => (
              <TableRow key={headerGroup.id}>
                {headerGroup.headers.map(header => (
                  <TableHead key={header.id}>
                    {header.isPlaceholder
                      ? null
                      : flexRender(header.column.columnDef.header, header.getContext())}
                  </TableHead>
                ))}
              </TableRow>
            ))}
          </TableHeader>
          <TableBody>
            {table.getRowModel().rows.map(row => (
              <TableRow key={row.id}>
                {row.getVisibleCells().map(cell => (
                  <TableCell key={cell.id}>
                    {flexRender(cell.column.columnDef.cell, cell.getContext())}
                  </TableCell>
                ))}
              </TableRow>
            ))}
          </TableBody>
        </Table>
      )}
    </div>
  )
}

export default ListNodesPage

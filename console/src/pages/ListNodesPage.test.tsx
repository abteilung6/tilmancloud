import { describe, it, expect, vi, beforeEach } from 'vitest'
import { renderWithQuery, screen, mockAxiosResponse, waitFor } from '@/test/utils'
import userEvent from '@testing-library/user-event'
import ListNodesPage from './ListNodesPage'
import { apiClient, NodeStateEnum } from '@/lib/api-client'
import type { Node } from '@/lib/api-client'

const defaultedListNodes: Node[] = [
  {
    name: 'i-1234567890abcdef0',
    state: NodeStateEnum.Running,
    instanceType: 't2.micro',
    publicIp: '54.123.45.67',
    privateIp: '10.0.1.123',
  },
  {
    name: 'i-0987654321fedcba0',
    state: NodeStateEnum.Pending,
    instanceType: 't2.micro',
    publicIp: null,
    privateIp: '10.0.1.124',
  },
]

const defaultedCreateNode: Node = {
  name: 'i-new1234567890abcdef',
  state: NodeStateEnum.Pending,
  instanceType: 't2.micro',
  publicIp: null,
  privateIp: '10.0.1.125',
}

describe('ListNodesPage', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  const customRender = () => {
    vi.spyOn(apiClient, 'listNodes').mockResolvedValue(
      mockAxiosResponse({ data: defaultedListNodes })
    )
    const createNodeSpy = vi
      .spyOn(apiClient, 'createNode')
      .mockResolvedValue(mockAxiosResponse({ data: defaultedCreateNode, status: 201 }))
    const user = userEvent.setup()

    renderWithQuery(<ListNodesPage />)

    return { createNodeSpy, user }
  }

  it('renders nodes in a table', async () => {
    customRender()

    await screen.findByRole('table')

    expect(screen.getByRole('columnheader', { name: 'Name' })).toBeInTheDocument()
    expect(screen.getByRole('columnheader', { name: 'State' })).toBeInTheDocument()
    expect(screen.getByRole('columnheader', { name: 'Instance Type' })).toBeInTheDocument()
    expect(screen.getByRole('columnheader', { name: 'Public IP' })).toBeInTheDocument()
    expect(screen.getByRole('columnheader', { name: 'Private IP' })).toBeInTheDocument()

    const rows = screen.getAllByRole('row')
    expect(rows).toHaveLength(3)

    const firstDataRow = rows[1]
    expect(firstDataRow).toHaveTextContent('i-1234567890abcdef0')
    expect(firstDataRow).toHaveTextContent('running')
    expect(firstDataRow).toHaveTextContent('t2.micro')
    expect(firstDataRow).toHaveTextContent('54.123.45.67')
    expect(firstDataRow).toHaveTextContent('10.0.1.123')

    const secondDataRow = rows[2]
    expect(secondDataRow).toHaveTextContent('i-0987654321fedcba0')
    expect(secondDataRow).toHaveTextContent('pending')
    expect(secondDataRow).toHaveTextContent('N/A')
    expect(secondDataRow).toHaveTextContent('10.0.1.124')
  })

  it('creates a node when Add node button is clicked', async () => {
    const { createNodeSpy, user } = customRender()

    await screen.findByRole('table')
    const addButton = screen.getByRole('button', { name: 'Add node' })
    await user.click(addButton)

    await waitFor(() => {
      expect(createNodeSpy).toHaveBeenCalledOnce()
    })
  })
})

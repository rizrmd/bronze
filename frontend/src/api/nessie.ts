import { api } from './index'

export interface NessieColumn {
  name: string
  type: string
  nullable: boolean
  comment?: string
}

export interface NessieTable {
  name: string
  database: string
  columns: NessieColumn[]
  created_at: string
  row_count?: number
}

export interface TableDataResponse {
  table: NessieTable
  data: any[][]
  total_rows: number
  columns: string[]
  pagination?: {
    page: number
    limit: number
    total: number
    pages: number
  }
}

export const nessieApi = {
  listTables: async (database?: string): Promise<NessieTable[]> => {
    const params = database ? { database } : {}
    const { data } = await api.get('/api/nessie/tables', { params })
    return data.tables || []
  },

  getTable: async (name: string, database?: string): Promise<NessieTable> => {
    const params = database ? { database } : {}
    const { data } = await api.get(`/api/nessie/tables/${name}`, { params })
    return data.table
  },

  browseTableData: async (name: string, options: {
    database?: string
    page?: number
    limit?: number
  }): Promise<TableDataResponse> => {
    const { data } = await api.get(`/api/nessie/tables/${name}/data`, { params: options })
    return data
  },

  dropTable: async (name: string, database?: string): Promise<void> => {
    const params = database ? { database } : {}
    await api.delete(`/api/nessie/tables/${name}`, { params })
  },

  getDatabases: async (): Promise<string[]> => {
    const { data } = await api.get('/api/nessie/databases')
    return data.databases || []
  },

  testConnection: async (): Promise<boolean> => {
    try {
      const { data } = await api.get('/api/nessie/health')
      return data.connected || false
    } catch {
      return false
    }
  }
}
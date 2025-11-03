import { api } from './client'
import type { NessieTable, TableDataResponse } from './nessie-types'

export async function listTables(database?: string): Promise<NessieTable[]> {
  const params = database ? { database } : {}
  const { data } = await api.get('/api/nessie/tables', { params })
  return data.tables || []
}

export async function getTable(name: string, database?: string): Promise<NessieTable> {
  const params = database ? { database } : {}
  const { data } = await api.get(`/api/nessie/tables/${name}`, { params })
  return data.table
}

export async function browseTableData(name: string, options: {
  database?: string
  page?: number
  limit?: number
}): Promise<TableDataResponse> {
  const { data } = await api.get(`/api/nessie/tables/${name}/data`, { params: options })
  return data
}

export async function dropTable(name: string, database?: string): Promise<void> {
  const params = database ? { database } : {}
  await api.delete(`/api/nessie/tables/${name}`, { params })
}

export async function getDatabases(): Promise<string[]> {
  const { data } = await api.get('/api/nessie/databases')
  return data.databases || []
}

export async function testConnection(): Promise<boolean> {
  try {
    const { data } = await api.get('/api/nessie/health')
    return data.connected || false
  } catch {
    return false
  }
}
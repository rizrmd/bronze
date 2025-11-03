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
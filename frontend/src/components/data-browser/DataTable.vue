<script setup lang="ts">
import { ref, computed } from 'vue'
import { ChevronLeft, ChevronRight, Download, Search } from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'



interface Props {
  columns: string[]
  rows: string[][]
  loading?: boolean
  totalCount?: number
  pageSize?: number | string
  hasHeaders?: boolean
  searchable?: boolean
  downloadable?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  loading: false,
  totalCount: 0,
  pageSize: 50,
  hasHeaders: true,
  searchable: true,
  downloadable: true
})

const emit = defineEmits<{
  pageChange: [page: number]
  search: [query: string]
  download: []
}>()

const currentPage = ref(1)
const searchQuery = ref('')
const sortColumn = ref<string | null>(null)
const sortDirection = ref<'asc' | 'desc'>('asc')
const pageSize = ref(String(props.pageSize))

const totalPages = computed(() => Math.ceil(props.totalCount / Number(pageSize.value)))

const filteredRows = computed(() => {
  if (!searchQuery.value) return props.rows
  
  return props.rows.filter(row => {
    return row.some(cell => 
      cell.toLowerCase().includes(searchQuery.value.toLowerCase())
    )
  })
})

const sortedRows = computed(() => {
  if (!sortColumn.value) return filteredRows.value
  
  const columnIndex = props.columns.indexOf(sortColumn.value)
  if (columnIndex === -1) return filteredRows.value
  
  return [...filteredRows.value].sort((a, b) => {
    const aVal = a[columnIndex] || ''
    const bVal = b[columnIndex] || ''
    
    const comparison = aVal.localeCompare(bVal)
    return sortDirection.value === 'asc' ? comparison : -comparison
  })
})

const paginatedRows = computed(() => {
  const start = (currentPage.value - 1) * Number(pageSize.value)
  const end = start + Number(pageSize.value)
  return sortedRows.value.slice(start, end)
})

const paginationInfo = computed(() => {
  const start = ((currentPage.value - 1) * Number(pageSize.value)) + 1
  const end = Math.min(currentPage.value * Number(pageSize.value), props.totalCount)
  return `Showing ${start} to ${end} of ${props.totalCount} results`
})

const handlePageChange = (page: number) => {
  if (page >= 1 && page <= totalPages.value) {
    currentPage.value = page
    emit('pageChange', page)
  }
}

const handleSort = (column: string) => {
  if (sortColumn.value === column) {
    sortDirection.value = sortDirection.value === 'asc' ? 'desc' : 'asc'
  } else {
    sortColumn.value = column
    sortDirection.value = 'asc'
  }
}

const handleSearch = (query: string) => {
  searchQuery.value = query
  currentPage.value = 1
  emit('search', query)
}

const handleDownload = () => {
  emit('download')
}

const getCellValue = (row: string[], column: string) => {
  const index = props.columns.indexOf(column)
  return index >= 0 ? row[index] : ''
}

const formatCellValue = (value: string) => {
  if (value === null || value === undefined) return ''
  if (typeof value === 'string' && value.length > 100) {
    return value.substring(0, 100) + '...'
  }
  return String(value)
}
</script>

<template>
  <Card>
    <CardHeader>
      <div class="flex items-center justify-between">
        <CardTitle class="text-lg">
          Data Table
          <Badge variant="secondary" class="ml-2">
            {{ totalCount }} rows
          </Badge>
        </CardTitle>
        
        <div class="flex items-center space-x-2">
          <div v-if="searchable" class="relative">
            <Search class="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-gray-400" />
            <Input
              v-model="searchQuery"
              @input="handleSearch(searchQuery)"
              placeholder="Search..."
              class="pl-10 w-64"
            />
          </div>
          
          <Select v-model="pageSize" @update:modelValue="currentPage = 1">
            <SelectTrigger class="w-24">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="25">25</SelectItem>
              <SelectItem value="50">50</SelectItem>
              <SelectItem value="100">100</SelectItem>
              <SelectItem value="200">200</SelectItem>
            </SelectContent>
          </Select>
          
          <Button v-if="downloadable" variant="outline" size="sm" @click="handleDownload">
            <Download class="w-4 h-4 mr-2" />
            Export
          </Button>
        </div>
      </div>
    </CardHeader>
    
    <CardContent>
      <div v-if="loading" class="text-center py-8">
        <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600 mx-auto"></div>
        <p class="mt-2 text-gray-500">Loading data...</p>
      </div>
      
      <div v-else-if="columns.length === 0" class="text-center py-8 text-gray-500">
        <p>No data available</p>
      </div>
      
      <div v-else class="overflow-x-auto">
        <table class="min-w-full divide-y divide-gray-200">
          <thead class="bg-gray-50">
            <tr>
              <th
                v-for="column in columns"
                :key="column"
                @click="handleSort(column)"
                class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider cursor-pointer hover:bg-gray-100"

              >
                <div class="flex items-center space-x-1">
                  <span>{{ column }}</span>
                  <div v-if="sortColumn === column" class="flex flex-col">
                    <ChevronUp 
                      :class="['w-3 h-3', sortDirection === 'asc' ? 'text-blue-600' : 'text-gray-400']" 
                    />
                    <ChevronDown 
                      :class="['w-3 h-3 -mt-1', sortDirection === 'desc' ? 'text-blue-600' : 'text-gray-400']" 
                    />
                  </div>
                </div>
              </th>
            </tr>
          </thead>
          <tbody class="bg-white divide-y divide-gray-200">
            <tr v-for="(row, rowIndex) in paginatedRows" :key="rowIndex" class="hover:bg-gray-50">
              <td
                v-for="column in columns"
                :key="`${rowIndex}-${column}`"
                class="px-6 py-4 whitespace-nowrap text-sm text-gray-900"
                :title="getCellValue(row, column)"
              >
                {{ formatCellValue(getCellValue(row, column) || '') }}
              </td>
            </tr>
          </tbody>
        </table>
        
        <!-- Pagination -->
        <div v-if="totalPages > 1" class="flex items-center justify-between mt-4">
          <div class="text-sm text-gray-700">
            {{ paginationInfo }}
          </div>
          
          <div class="flex items-center space-x-2">
            <Button
              variant="outline"
              size="sm"
              @click="handlePageChange(currentPage - 1)"
              :disabled="currentPage === 1"
            >
              <ChevronLeft class="w-4 h-4 mr-1" />
              Previous
            </Button>
            
            <div class="flex items-center space-x-1">
              <Button
                v-for="page in Math.min(totalPages, 5)"
                :key="page"
                :variant="page === currentPage ? 'default' : 'outline'"
                size="sm"
                @click="handlePageChange(page)"
              >
                {{ page }}
              </Button>
            </div>
            
            <Button
              variant="outline"
              size="sm"
              @click="handlePageChange(currentPage + 1)"
              :disabled="currentPage === totalPages"
            >
              Next
              <ChevronRight class="w-4 h-4 ml-1" />
            </Button>
          </div>
        </div>
      </div>
    </CardContent>
  </Card>
</template>

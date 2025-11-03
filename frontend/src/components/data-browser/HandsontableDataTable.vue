<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import { HotTable } from '@handsontable/vue3'
import { Search, ChevronLeft, ChevronRight, Download } from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import 'handsontable/dist/handsontable.full.css'

// Handsontable context menu separator
const SEPARATOR = '---------'

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
  cellChange: [changes: any[], source: string]
}>()

const hotTableRef = ref<any>(null)
const searchQuery = ref('')
const currentPage = ref(1)
const pageSize = ref(String(props.pageSize))

// Initialize Handsontable settings
const hotSettings = ref({
  data: props.rows,
  colHeaders: props.hasHeaders ? props.columns : false,
  rowHeaders: true,
  width: '100%',
  height: 600,
  stretchH: 'all',
  columnSorting: true,
  filters: true,
  search: true,
  dropdownMenu: [
    'filter_by_condition',
    'filter_operators',
    'filter_by_value',
    'separator',
    'clear_column'
  ],
  manualColumnResize: true,
  manualRowResize: true,
  manualColumnMove: true,
  fixedColumnsLeft: 1, // Freeze first column
  readOnly: false,
  wordWrap: false,
  copyPaste: true,
  undo: true,
  allowInsertRow: true,
  allowRemoveRow: true,
  allowInsertColumn: false,
  allowRemoveColumn: false,
  fillHandle: true, // Enable fill handle for drag copying
  trimWhitespace: true,
  autoColumnSize: {
    samplingRatio: 23
  },
  contextMenu: {
    items: {
      row_above: {
        key: 'row_above',
        name: 'Insert row above'
      },
      row_below: {
        key: 'row_below', 
        name: 'Insert row below'
      },
      separator: SEPARATOR,
      remove_row: {
        key: 'remove_row',
        name: 'Remove row'
      },
      separator2: SEPARATOR,
      undo: {
        key: 'undo',
        name: 'Undo'
      },
      redo: {
        key: 'redo', 
        name: 'Redo'
      },
      separator3: SEPARATOR,
      copy: {
        key: 'copy',
        name: 'Copy'
      },
      cut: {
        key: 'cut',
        name: 'Cut'
      },
      paste: {
        key: 'paste',
        name: 'Paste'
      },
      separator4: SEPARATOR,
      'read_only': {
        key: 'read_only',
        name: 'Toggle read-only',
        callback: function(_key: string, _selection: any[][]) {
          const hot = hotTableRef.value?.hotInstance
          if (hot) {
            const currentReadOnly = hot.getSettings().readOnly
            hot.updateSettings({ readOnly: !currentReadOnly })
          }
        }
      },
      'clear_column': {
        key: 'clear_column',
        name: 'Clear column data',
        callback: function(_key: string, selection?: any[][]) {
          const hot = hotTableRef.value?.hotInstance
          if (hot && selection && selection[0]) {
            const [, startCol, , endCol] = selection[0]
            for (let row = 0; row < hot.countRows(); row++) {
              for (let col = startCol; col <= endCol; col++) {
                hot.setDataAtCell(row, col, '')
              }
            }
          }
        }
      }
    }
  },
  licenseKey: 'non-commercial-and-evaluation',
  afterChange: (changes: any[] | null, source: string) => {
    if (changes && source !== 'loadData') {
      emit('cellChange', changes, source)
    }
  },
  beforeCopy: (data: any[][], coords?: any[][]) => {
    // Enhanced copy with headers
    const hot = hotTableRef.value?.hotInstance
    if (hot && props.hasHeaders && coords && coords[0]) {
      const [, startCol, , endCol] = coords[0]
      const headers = []
      for (let col = startCol; col <= endCol; col++) {
        headers.push(props.columns[col] || `Column ${col + 1}`)
      }
      return [headers, ...data]
    }
    return data
  }
})

// Update table data when props change
watch(() => props.rows, (newRows) => {
  if (hotTableRef.value?.hotInstance) {
    const hot = hotTableRef.value.hotInstance
    hot.loadData(newRows)
    
    // Auto-size columns after data load
    setTimeout(() => {
      hot.getPlugin('autoColumnSize').recalculateAllColumnsWidth()
    }, 100)
  }
}, { deep: true })

watch(() => props.columns, (newColumns) => {
  if (hotTableRef.value?.hotInstance && props.hasHeaders) {
    hotTableRef.value.hotInstance.updateSettings({
      colHeaders: newColumns
    })
  }
}, { deep: true })

// Search functionality
const handleSearch = (query: string) => {
  searchQuery.value = query
  currentPage.value = 1
  
  if (hotTableRef.value?.hotInstance) {
    const hot = hotTableRef.value.hotInstance
    
    if (query.trim()) {
      // Clear previous search results
      hot.getPlugin('search').clearSearch()
      
      // Perform search
      const searchResult = hot.getPlugin('search').query(query)
      
      // Highlight all matching cells
      if (searchResult.length > 0) {
        // Scroll to first result
        hot.selectCell(searchResult[0].row, searchResult[0].col)
        
        // Highlight all results
        const highlightPlugin = hot.getPlugin('highlightRows')
        if (highlightPlugin) {
          highlightPlugin.highlightRows(searchResult.map((r: any) => r.row))
        }
      }
    } else {
      // Clear search
      hot.getPlugin('search').clearSearch()
      hot.deselectCell()
      
      const highlightPlugin = hot.getPlugin('highlightRows')
      if (highlightPlugin) {
        highlightPlugin.unhighlightRows()
      }
    }
  }
  
  emit('search', query)
}

// Pagination info
const totalPages = computed(() => Math.ceil(props.totalCount / Number(pageSize.value)))
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

const handleDownload = () => {
  emit('download')
}

const clearAllData = () => {
  if (hotTableRef.value?.hotInstance) {
    const hot = hotTableRef.value.hotInstance
    hot.loadData([])
  }
}

const exportToCSV = () => {
  if (hotTableRef.value?.hotInstance) {
    const hot = hotTableRef.value.hotInstance
    const exportPlugin = hot.getPlugin('exportFile')
    
    if (exportPlugin) {
      exportPlugin.downloadFile('csv', {
        filename: 'data-export',
        columnHeaders: props.hasHeaders,
        exportHiddenColumns: false,
        exportHiddenRows: false
      })
    }
  }
}

onMounted(() => {
  // Initialize search plugin
  if (hotTableRef.value?.hotInstance) {
    const searchPlugin = hotTableRef.value.hotInstance.getPlugin('search')
    searchPlugin.setQueryMethod((query: string, value: string) => {
      return String(value).toLowerCase().includes(query.toLowerCase())
    })
  }
})
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
              placeholder="Search in cells..."
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
              <SelectItem value="500">500</SelectItem>
            </SelectContent>
          </Select>
          
          <Button 
            variant="outline" 
            size="sm" 
            @click="clearAllData"
            title="Clear all table data"
          >
            Clear
          </Button>
          
          <Button 
            variant="outline" 
            size="sm" 
            @click="exportToCSV"
            title="Export to CSV"
          >
            <Download class="w-4 h-4 mr-2" />
            CSV
          </Button>
          
          <Button v-if="downloadable" variant="outline" size="sm" @click="handleDownload">
            Export to DB
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
      
      <div v-else>
        <HotTable
          ref="hotTableRef"
          :settings="hotSettings"
        />
        
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

<style scoped>
/* Custom Handsontable styling to match shadcn/ui theme */
:deep(.ht_master) {
  border: 1px solid hsl(var(--border));
  border-radius: 6px;
}

:deep(.ht_clone_top) {
  background-color: hsl(var(--muted) / 0.5);
}

:deep(.ht_clone_left) {
  background-color: hsl(var(--muted) / 0.5);
}

:deep(.ht_clone_corner) {
  background-color: hsl(var(--muted) / 0.7);
}

:deep(.ht thead) {
  background-color: hsl(var(--muted) / 0.8);
}

:deep(.ht th) {
  background-color: hsl(var(--muted) / 0.8);
  font-weight: 600;
  color: hsl(var(--foreground));
  border-bottom: 2px solid hsl(var(--border));
}

:deep(.ht td) {
  border-color: hsl(var(--border));
  color: hsl(var(--foreground));
}

:deep(.ht tbody tr:hover td) {
  background-color: hsl(var(--accent) / 0.3);
}

:deep(.ht tbody tr.ht__highlight td) {
  background-color: hsl(var(--primary) / 0.1);
}

:deep(.htCurrent) {
  border-color: hsl(var(--primary));
  background-color: hsl(var(--primary) / 0.05) !important;
}

:deep(.htContextMenu) {
  border: 1px solid hsl(var(--border));
  background-color: hsl(var(--background));
  box-shadow: 0 4px 6px -1px rgb(0 0 0 / 0.1), 0 2px 4px -2px rgb(0 0 0 / 0.1);
}

:deep(.htContextMenu .ht_master .wtHolder) {
  background-color: hsl(var(--background));
}

:deep(.htContextMenu table) {
  background-color: hsl(var(--background));
}

:deep(.htContextMenu .htRowHeader) {
  background-color: hsl(var(--muted) / 0.5);
}

:deep(.htContextMenu td) {
  border-color: hsl(var(--border));
  color: hsl(var(--foreground));
}

:deep(.htContextMenu td:hover) {
  background-color: hsl(var(--accent) / 0.5);
}
</style>
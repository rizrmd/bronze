<template>
  <div class="space-y-4">
    <!-- Data Preview Controls -->
    <div class="flex items-center justify-between border-b pb-4">
      <div class="flex items-center gap-4">
        <select 
          v-model="selectedSheet" 
          @change="loadData"
          class="px-3 py-1.5 border rounded-lg text-sm"
        >
          <option value="">Select sheet</option>
          <option v-for="sheet in sheets" :key="sheet" :value="sheet">
            {{ sheet }}
          </option>
        </select>
        
        <div class="flex items-center gap-2 text-sm text-gray-600">
          <label for="rows">Rows:</label>
          <input 
            id="rows"
            v-model.number="maxRows" 
            @change="loadData"
            type="number" 
            min="10" 
            max="1000" 
            step="10"
            class="w-20 px-2 py-1 border rounded"
          />
        </div>
      </div>
      
      <button 
        @click="exportData"
        class="px-4 py-2 bg-green-600 text-white rounded-lg hover:bg-green-700 transition-colors text-sm"
      >
        Export
      </button>
    </div>

    <!-- Loading State -->
    <div v-if="loading" class="flex items-center justify-center h-64">
      <div class="text-center">
        <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600 mx-auto mb-4"></div>
        <p class="text-gray-500">Loading data...</p>
      </div>
    </div>

    <!-- Data Table -->
    <div v-else-if="data.length > 0" class="overflow-auto">
      <table class="min-w-full border border-gray-200">
        <thead class="bg-gray-50">
          <tr>
            <th 
              v-for="column in columns" 
              :key="column"
              class="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider border-b"
            >
              {{ column }}
            </th>
          </tr>
        </thead>
        <tbody class="bg-white divide-y divide-gray-200">
          <tr v-for="(row, index) in data" :key="index" class="hover:bg-gray-50">
            <td 
              v-for="column in columns" 
              :key="column"
              class="px-4 py-2 text-sm text-gray-900 whitespace-nowrap border-b"
            >
              {{ getCellValue(row, column) }}
            </td>
          </tr>
        </tbody>
      </table>
      
      <div v-if="hasMoreRows" class="text-center py-4 text-sm text-gray-500">
        Showing first {{ data.length }} rows
      </div>
    </div>

    <!-- No Data -->
    <div v-else class="flex items-center justify-center h-64">
      <div class="text-center text-gray-500">
        <FileSpreadsheet class="w-12 h-12 mx-auto mb-4" />
        <p>No data available</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { FileSpreadsheet } from 'lucide-vue-next'
import { browseData } from '@/services/api/data-browser'
import type { FileInfo } from '@/types'

interface Props {
  file: FileInfo
  presignedUrl: string
}

interface Emits {
  (e: 'error', error: string): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

const loading = ref(true)
const sheets = ref<string[]>([])
const selectedSheet = ref<string | undefined>('')
const maxRows = ref<number>(100)
const data = ref<any[]>([])
const columns = ref<string[]>([])
const hasMoreRows = ref<boolean>(false)

const getFileName = () => {
  return props.file.key?.split('/')?.pop() || 'Unknown'
}

const getFileExtension = () => {
  const fileName = getFileName()
  return fileName.split('.').pop()?.toLowerCase() || ''
}

const getCellValue = (row: any, column: string) => {
  if (typeof row === 'object' && row !== null) {
    return row[column] ?? ''
  }
  return row
}

const loadSheets = async () => {
  try {
    const response = await browseData({
      file_name: props.file.key,
      max_rows: 0  // Just get metadata with sheets
    })
    
    if (response.sheets) {
      sheets.value = response.sheets
      if (sheets.value.length > 0) {
        selectedSheet.value = sheets.value[0]
      }
    }
  } catch (err) {
    console.error('Failed to load sheets:', err)
    emit('error', 'Failed to load data sheets')
  }
}

const loadData = async () => {
  if (!selectedSheet.value && getFileExtension() !== 'csv') {
    return
  }

  loading.value = true
  
  try {
    const response = await browseData({
      file_name: props.file.key,
      sheet_name: selectedSheet.value || '',
      max_rows: maxRows.value
    })
    
    if (response.rows && response.columns) {
      // Convert row arrays to objects using column names
      data.value = response.rows.map(row => {
        const obj: Record<string, any> = {}
        response.columns.forEach((column, index) => {
          obj[column] = row[index] || ''
        })
        return obj
      })
      
      columns.value = response.columns
      
      const hasMoreData = response.total_rows > response.row_count + response.offset
      hasMoreRows.value = hasMoreData
    }
    
    loading.value = false
  } catch (err) {
    console.error('Failed to load data:', err)
    emit('error', 'Failed to load data')
    loading.value = false
  }
}

const exportData = async () => {
  try {
    const response = await browseData({
      file_name: props.file.key,
      sheet_name: selectedSheet.value || '',
      max_rows: 10000  // Export more rows for export
    })
    
    // Convert to CSV and create download
    const csvContent = [
      response.columns.join(','),
      ...response.rows.map(row => row.map(cell => `"${cell}"`).join(','))
    ].join('\n')
    
    // Create download link
    const url = window.URL.createObjectURL(new Blob([csvContent], { type: 'text/csv' }))
    const a = document.createElement('a')
    a.href = url
    a.download = `${getFileName()}.csv`
    document.body.appendChild(a)
    a.click()
    window.URL.revokeObjectURL(url)
    document.body.removeChild(a)
  } catch (err) {
    console.error('Export failed:', err)
    emit('error', 'Failed to export data')
  }
}

onMounted(async () => {
  if (['xlsx', 'xls'].includes(getFileExtension())) {
    await loadSheets()
  }
  await loadData()
})
</script>
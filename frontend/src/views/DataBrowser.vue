<template>
  <!-- Loading state -->
  <div v-if="loading && !currentData" class="text-center py-8">
    <RefreshCw class="mx-auto h-8 w-8 animate-spin text-gray-400" />
    <p class="mt-2 text-gray-500">Loading file data...</p>
  </div>

  <!-- Error state -->
  <div v-if="error && !currentData" class="flex flex-col items-center justify-center h-64 text-red-500">
    <div class="mb-2">❌</div>
    <div class="mb-4">{{ error }}</div>
    <div>
      <Button @click="clearError" variant="outline" size="sm">
        Dismiss
      </Button>
    </div>
  </div>

  <!-- No file selected state -->
  <div v-if="!loading && !error && !currentData" class="flex flex-col items-center justify-center h-64 text-gray-500">
    <Database class="mx-auto h-12 w-12 text-gray-400" />
    <h3 class="mt-2 text-sm font-medium text-gray-900">No file selected</h3>
    <p class="mt-1 text-sm text-gray-500">
      Navigate to Data Browser with a file query parameter to view data.
    </p>
  </div>

  <!-- Data Preview Section (shown when file is selected) -->
  <div v-if="currentData" class="space-y-6">
    <Card>
      <CardHeader>
        <div class="flex items-center justify-between">
          <div>
            <CardTitle class="text-lg">Browse Data: {{ selectedFile?.name }}</CardTitle>
            <CardDescription class="mt-2">
              <div class="flex items-center space-x-4">
                <Badge :class="getFileTypeColor(currentData?.data_type)">
                  {{ currentData?.data_type?.toUpperCase() }}
                </Badge>
                <span v-if="currentData?.total_rows">
                  {{ currentData.total_rows.toLocaleString() }} total rows
                </span>
                <span v-if="currentData?.sheet_name">
                  Sheet: {{ currentData.sheet_name }}
                </span>
              </div>
            </CardDescription>
          </div>
        </div>
      </CardHeader>
      <CardContent>
        <!-- Sheet Selection for Excel/MDB -->
        <div v-if="selectedFile?.data_type === 'excel' || selectedFile?.data_type === 'mdb'" class="mb-4">
          <Label>Sheet/Table:</Label>
          <Select v-model="selectedSheet" @update:modelValue="browseFile(selectedFile, $event)">
            <SelectTrigger>
              <SelectValue placeholder="Select sheet or table" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem 
                v-for="sheet in selectedFile.sheets" 
                :key="sheet"
                :value="sheet"
              >
                {{ sheet }}
              </SelectItem>
            </SelectContent>
          </Select>
        </div>
        
        <!-- Browse Options -->
        <div class="grid grid-cols-2 md:grid-cols-4 gap-4 mb-4">
          <div class="flex items-center space-x-2">
            <Switch 
              id="has-headers" 
              v-model:checked="hasHeaders"
              @update:checked="browseFile(selectedFile, selectedSheet)"
            />
            <Label for="has-headers">Has Headers</Label>
          </div>
          
          <div class="flex items-center space-x-2">
            <Switch 
              id="auto-detect" 
              v-model:checked="autoDetectHeaders"
              @update:checked="browseFile(selectedFile, selectedSheet)"
            />
            <Label for="auto-detect">Auto Detect</Label>
          </div>
          
          <div class="flex items-center space-x-2">
            <Switch 
              id="treat-csv" 
              v-model:checked="treatAsCSV"
              @update:checked="browseFile(selectedFile, selectedSheet)"
            />
            <Label for="treat-csv">Treat as CSV</Label>
          </div>
          
          <div>
            <Label for="max-rows">Max Rows:</Label>
            <Select v-model="currentMaxRows" @update:modelValue="currentOffset = 0; browseFile(selectedFile, selectedSheet)">
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="50">50</SelectItem>
                <SelectItem value="100">100</SelectItem>
                <SelectItem value="500">500</SelectItem>
                <SelectItem value="1000">1000</SelectItem>
                <SelectItem value="5000">5000</SelectItem>
                <SelectItem value="10000">10000</SelectItem>
              </SelectContent>
            </Select>
          </div>
        </div>
        
        <!-- Data Table -->
        <DataTable
          v-if="currentData?.columns"
          :columns="currentData.columns"
          :rows="currentData.rows"
          :loading="loading"
          :total-count="currentData.total_rows"
          :page-size="Number(currentMaxRows)"
          :has-headers="currentData.has_headers"
          @page-change="handleTablePageChange"
          @download="selectedFile = currentData.file; exportDialogOpen = true"
        />
      </CardContent>
    </Card>
  </div>

  <!-- Export Dialog -->
  <Dialog v-model:open="exportDialogOpen">
    <DialogContent>
      <DialogHeader>
        <DialogTitle>Export Data</DialogTitle>
        <DialogDescription>
          Export data from {{ selectedFile?.name }} to database
        </DialogDescription>
      </DialogHeader>
      
      <div class="space-y-4">
        <div>
          <Label for="table-name">Table Name:</Label>
          <Input 
            id="table-name"
            v-model="exportTableName" 
            placeholder="Enter table name"
          />
        </div>
        
        <div>
          <Label>Operation:</Label>
          <div class="flex space-x-4 mt-2">
            <div class="flex items-center space-x-2">
              <input 
                v-model="exportOperation" 
                type="radio" 
                value="create"
                id="create"
              />
              <Label for="create">Create New Table</Label>
            </div>
            <div class="flex items-center space-x-2">
              <input 
                v-model="exportOperation" 
                type="radio" 
                value="append"
                id="append"
              />
              <Label for="append">Append to Existing</Label>
            </div>
          </div>
        </div>
        
        <div class="flex justify-end space-x-3">
          <Button @click="exportDialogOpen = false" variant="outline">
            Cancel
          </Button>
          <Button @click="handleExport">
            Export
          </Button>
        </div>
      </div>
    </DialogContent>
  </Dialog>
</template>

<script setup lang="ts">
import { ref, watch, onUnmounted } from 'vue'
import { useRoute } from 'vue-router'
import { useDataBrowser } from '@/composables/useDataBrowser'
import { Card, CardContent, CardHeader } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription } from '@/components/ui/dialog'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import Switch from '@/components/ui/switch.vue'
import Label from '@/components/ui/label.vue'
import { Input } from '@/components/ui/input'
import DataTable from '@/components/data-browser/DataTable.vue'
import { 
  Database,
  RefreshCw
} from 'lucide-vue-next'

const { 
  loading, 
  error, 
  browseData, 
  createExportJob,
  clearError,
  cancelAllBrowseRequests
} = useDataBrowser()

const route = useRoute()

const selectedFile = ref<any>(null)
const selectedSheet = ref<string>('')
const exportDialogOpen = ref(false)
const currentData = ref<any>(null)
const currentOffset = ref(0)
const currentMaxRows = ref('100')
const hasHeaders = ref(true)
const autoDetectHeaders = ref(true)
const treatAsCSV = ref(false)

// Export settings
const exportTableName = ref('')
const exportOperation = ref<'create' | 'append'>('create')

const getFileTypeColor = (dataType: string) => {
  switch (dataType) {
    case 'excel':
      return 'bg-green-100 text-green-800'
    case 'csv':
      return 'bg-blue-100 text-blue-800'
    case 'mdb':
      return 'bg-purple-100 text-purple-800'
    default:
      return 'bg-gray-100 text-gray-800'
  }
}

const browseFile = async (file: any, sheet?: string) => {
  if (!file) return
  
  selectedFile.value = file
  selectedSheet.value = sheet || ''
  
  const request = {
    file_name: file.key || file.name,  // Use full path if available, fallback to name
    sheet_name: sheet || undefined,
    max_rows: Number(currentMaxRows.value),
    offset: currentOffset.value,
    has_headers: hasHeaders.value,
    auto_detect_headers: autoDetectHeaders.value,
    treat_as_csv: treatAsCSV.value
  }
  
  const response = await browseData(request)
  if (response) {
    currentData.value = response
  }
}

const handleTablePageChange = async (page: number) => {
  if (!selectedFile.value) return
  
  const newOffset = (page - 1) * Number(currentMaxRows.value)
  const request = {
    file_name: selectedFile.value.name,
    sheet_name: selectedSheet.value || undefined,
    max_rows: Number(currentMaxRows.value),
    offset: newOffset,
    has_headers: hasHeaders.value,
    auto_detect_headers: autoDetectHeaders.value,
    treat_as_csv: treatAsCSV.value
  }
  
  const response = await browseData(request)
  if (response) {
    currentData.value = response
  }
}

const handleExport = async () => {
  if (!selectedFile.value || !exportTableName.value) return
  
  await createExportJob({
    files: [{
      file_name: selectedFile.value.key || selectedFile.value.name,
      sheet_name: selectedSheet.value || undefined
    }],
    table_name: exportTableName.value,
    operation: exportOperation.value
  })
  
  exportDialogOpen.value = false
  exportTableName.value = ''
}

// Watch for file query parameter and auto-browse
watch(() => route.query.file, async (newFileQuery, oldFileQuery) => {
  // Cancel all pending browse requests when file changes
  if (oldFileQuery && oldFileQuery !== newFileQuery) {
    cancelAllBrowseRequests()
  }
  
  if (newFileQuery && typeof newFileQuery === 'string') {
    // Create a file object from the path
    const fileName = newFileQuery.split('/').pop() || newFileQuery
    const fileObject = { 
      key: newFileQuery, 
      name: fileName 
    }
    await browseFile(fileObject)
  }
}, { immediate: true })

// Cancel all browse requests when component is unmounted
onUnmounted(() => {
  cancelAllBrowseRequests()
})
</script>

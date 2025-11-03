<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useDataBrowser } from '@/composables/useDataBrowser'
import { Card, CardContent, CardHeader } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Input } from '@/components/ui/input'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription } from '@/components/ui/dialog'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import Switch from '@/components/ui/switch.vue'
import Label from '@/components/ui/label.vue'
import DataTable from '@/components/data-browser/DataTable.vue'
import { 
  Table, 
  Search, 
  Download, 
  Eye,
  FileText,
  Database,
  RefreshCw
} from 'lucide-vue-next'

const { 
  loading, 
  error, 
  browseData, 
  listDataFiles, 
  createExportJob,
  clearError 
} = useDataBrowser()

const files = ref<any[]>([])
const selectedFile = ref<any>(null)
const selectedSheet = ref<string>('')
const browseDialogOpen = ref(false)
const exportDialogOpen = ref(false)
const searchQuery = ref('')
const currentData = ref<any>(null)
const currentOffset = ref(0)
const currentMaxRows = ref('100')
const hasHeaders = ref(true)
const autoDetectHeaders = ref(true)
const treatAsCSV = ref(false)

// Export settings
const exportTableName = ref('')
const exportOperation = ref<'create' | 'append'>('create')

const filteredFiles = computed(() => {
  if (!searchQuery.value) return files.value
  return files.value.filter(file => 
    file.name.toLowerCase().includes(searchQuery.value.toLowerCase())
  )
})

const getFileIcon = (dataType: string) => {
  switch (dataType) {
    case 'excel':
      return FileText
    case 'csv':
      return Table
    case 'mdb':
      return Database
    default:
      return FileText
  }
}

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

const formatFileSize = (bytes: number) => {
  if (bytes === 0) return '0 Bytes'
  const k = 1024
  const sizes = ['Bytes', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

const formatDate = (dateString: string) => {
  return new Date(dateString).toLocaleString()
}

const fetchFiles = async () => {
  const response = await listDataFiles()
  if (response) {
    files.value = response.files || []
  }
}

const browseFile = async (file: any, sheet?: string) => {
  if (!file) return
  
  selectedFile.value = file
  selectedSheet.value = sheet || ''
  
  const request = {
    file_name: file.name,
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
    browseDialogOpen.value = true
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
    currentOffset.value = newOffset
  }
}

const handleExport = async () => {
  if (!selectedFile.value || !exportTableName.value) return
  
  const request = {
    files: [{
      file_name: selectedFile.value.name,
      sheet_name: selectedSheet.value || undefined,
      treat_as_csv: treatAsCSV.value
    }],
    table_name: exportTableName.value,
    operation: exportOperation.value,
    auto_type_conversion: true,
    schema_resolution: "merge",
    max_errors: 1000
  }
  
  const response = await createExportJob(request)
  if (response) {
    exportDialogOpen.value = false
    exportTableName.value = ''
    
    // Show job ID if available
    if (response.job_id) {
      console.log(`Export Job Created: ${response.job_id}. Processing will continue in background.`)
    } else {
      console.log(`Export Completed: Successfully exported to ${exportTableName.value}`)
    }
  }
}

const refreshFiles = () => {
  fetchFiles()
}

onMounted(() => {
  fetchFiles()
})
</script>

<template>
  <div class="space-y-6">
    <!-- Page Actions -->
    <div class="flex items-center justify-between">
      <div class="flex-1 relative">
        <Search class="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-gray-400" />
        <Input
          v-model="searchQuery"
          placeholder="Search data files..."
          class="pl-10"
        />
      </div>
      <div class="flex items-center space-x-3 ml-4">
        <Button @click="refreshFiles" variant="outline" size="sm">
          <RefreshCw class="w-4 h-4 mr-2" />
          Refresh
        </Button>
      </div>
    </div>

    <!-- Error Display -->
    <div v-if="error" class="bg-red-50 border border-red-200 rounded-md p-4">
      <div class="flex">
        <div class="ml-3">
          <h3 class="text-sm font-medium text-red-800">Error</h3>
          <div class="mt-2 text-sm text-red-700">
            {{ error }}
          </div>
          <div class="mt-4">
            <Button @click="clearError" variant="outline" size="sm">
              Dismiss
            </Button>
          </div>
        </div>
      </div>
    </div>

    <!-- Files Grid -->
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
      <Card 
        v-for="file in filteredFiles" 
        :key="file.name"
        class="hover:shadow-md transition-shadow cursor-pointer"
        @click="browseFile(file)"
      >
        <CardHeader>
          <div class="flex items-center justify-between">
            <div class="flex items-center space-x-2">
              <component :is="getFileIcon(file.data_type)" class="h-5 w-5 text-gray-400" />
              <Badge :class="getFileTypeColor(file.data_type)">
                {{ file.data_type.toUpperCase() }}
              </Badge>
            </div>
            <div class="flex space-x-1">
              <Button 
                variant="ghost" 
                size="sm"
                @click.stop="browseFile(file)"
              >
                <Eye class="h-4 w-4" />
              </Button>
              <Button 
                variant="ghost" 
                size="sm"
                @click.stop="; selectedFile = file; exportDialogOpen = true;"
              >
                <Download class="h-4 w-4" />
              </Button>
            </div>
          </div>
        </CardHeader>
        <CardContent>
          <div class="space-y-2">
            <h3 class="font-medium text-gray-900 truncate">{{ file.name }}</h3>
            <div class="text-sm text-gray-500 space-y-1">
              <div>Size: {{ formatFileSize(file.size) }}</div>
              <div>Modified: {{ formatDate(file.last_modified) }}</div>
              <div v-if="file.row_count">Rows: {{ file.row_count.toLocaleString() }}</div>
              <div v-if="file.columns?.length">Columns: {{ file.columns.length }}</div>
            </div>
            
            <!-- Excel Sheets -->
            <div v-if="file.data_type === 'excel' && file.sheets?.length" class="mt-3">
              <p class="text-xs font-medium text-gray-700 mb-2">Sheets:</p>
              <div class="flex flex-wrap gap-1">
                <Badge 
                  v-for="sheet in file.sheets.slice(0, 3)" 
                  :key="sheet"
                  variant="outline"
                  class="text-xs"
                >
                  {{ sheet }}
                </Badge>
                <Badge 
                  v-if="file.sheets.length > 3"
                  variant="outline"
                  class="text-xs"
                >
                  +{{ file.sheets.length - 3 }} more
                </Badge>
              </div>
            </div>
            
            <!-- MDB Tables -->
            <div v-if="file.data_type === 'mdb' && file.sheets?.length" class="mt-3">
              <p class="text-xs font-medium text-gray-700 mb-2">Tables:</p>
              <div class="flex flex-wrap gap-1">
                <Badge 
                  v-for="table in file.sheets.slice(0, 3)" 
                  :key="table"
                  variant="outline"
                  class="text-xs"
                >
                  {{ table }}
                </Badge>
                <Badge 
                  v-if="file.sheets.length > 3"
                  variant="outline"
                  class="text-xs"
                >
                  +{{ file.sheets.length - 3 }} more
                </Badge>
              </div>
            </div>
            
            <!-- CSV Columns -->
            <div v-if="file.data_type === 'csv' && file.columns?.length" class="mt-3">
              <p class="text-xs font-medium text-gray-700 mb-2">Columns:</p>
              <div class="flex flex-wrap gap-1">
                <Badge 
                  v-for="column in file.columns.slice(0, 3)" 
                  :key="column"
                  variant="outline"
                  class="text-xs"
                >
                  {{ column }}
                </Badge>
                <Badge 
                  v-if="file.columns.length > 3"
                  variant="outline"
                  class="text-xs"
                >
                  +{{ file.columns.length - 3 }} more
                </Badge>
              </div>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>

    <!-- No files message -->
    <div v-if="!loading && filteredFiles.length === 0" class="text-center py-8">
      <Database class="mx-auto h-12 w-12 text-gray-400" />
      <h3 class="mt-2 text-sm font-medium text-gray-900">No data files found</h3>
      <p class="mt-1 text-sm text-gray-500">
        Upload Excel (.xlsx, .xls), CSV, or MDB files to get started.
      </p>
    </div>

    <!-- Loading -->
    <div v-if="loading" class="text-center py-8">
      <RefreshCw class="mx-auto h-8 w-8 animate-spin text-gray-400" />
      <p class="mt-2 text-gray-500">Loading files...</p>
    </div>

    <!-- Browse Data Dialog -->
    <Dialog v-model:open="browseDialogOpen">
      <DialogContent class="max-w-6xl max-h-[80vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>Browse Data: {{ selectedFile?.name }}</DialogTitle>
          <DialogDescription>
            <div class="flex items-center space-x-4 mt-2">
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
          </DialogDescription>
        </DialogHeader>
        
        <div class="space-y-4">
          <!-- Sheet Selection for Excel/MDB -->
          <div v-if="selectedFile?.data_type === 'excel' || selectedFile?.data_type === 'mdb'">
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
          <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
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
        </div>
      </DialogContent>
    </Dialog>

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
            <Select v-model="exportOperation">
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="create">Create New Table</SelectItem>
                <SelectItem value="append">Append to Existing</SelectItem>
              </SelectContent>
            </Select>
          </div>
          
          <div class="flex justify-end space-x-3">
            <Button variant="outline" @click="exportDialogOpen = false">
              Cancel
            </Button>
            <Button 
              @click="handleExport"
              :disabled="!exportTableName"
            >
              Export
            </Button>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  </div>
</template>

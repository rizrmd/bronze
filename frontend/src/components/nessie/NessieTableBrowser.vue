<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { Card, CardContent } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Database, Eye, MoreVertical, Plus } from 'lucide-vue-next'
import { listTables, getDatabases, type NessieTable } from '@/services'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription } from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'

const tables = ref<NessieTable[]>([])
const databases = ref<string[]>([])
const selectedDatabase = ref<string>('')
const loading = ref(false)
const databasesLoading = ref(false)
const selectedTable = ref<NessieTable | null>(null)
const showPreview = ref(false)
const searchQuery = ref('')

const filteredTables = computed(() => {
  if (!searchQuery.value) return tables.value
  return tables.value.filter(table => 
    table.name.toLowerCase().includes(searchQuery.value.toLowerCase())
  )
})

const loadDatabases = async () => {
  databasesLoading.value = true
  try {
    databases.value = await getDatabases()
    if (databases.value.length > 0 && !databases.value.includes(selectedDatabase.value)) {
      selectedDatabase.value = databases.value[0]!
    }
  } catch (error) {
    console.error('Failed to load databases:', error)
    databases.value = ['bronze_warehouse']
    if (!selectedDatabase.value) {
      selectedDatabase.value = 'bronze_warehouse'
    }
  } finally {
    databasesLoading.value = false
  }
}

const loadTables = async () => {
  if (!selectedDatabase.value) return
  loading.value = true
  try {
    tables.value = await listTables(selectedDatabase.value)
  } catch (error) {
    console.error('Failed to load tables:', error)
    tables.value = []
  } finally {
    loading.value = false
  }
}

const selectTable = (table: NessieTable) => {
  selectedTable.value = table
  showPreview.value = true
}

const previewTable = (table: NessieTable) => {
  selectedTable.value = table
  showPreview.value = true
}

const tableActions = (table: NessieTable) => {
  console.log('Table actions for:', table.name)
}

onMounted(async () => {
  await loadDatabases()
  loadTables()
})
</script>

<template>
  <div class="space-y-4">
    <!-- Controls -->
    <div class="flex items-center justify-between">
      <div class="flex items-center space-x-4">
        <select 
          v-model="selectedDatabase" 
          @change="loadTables"
          :disabled="databasesLoading"
          class="w-48 px-3 py-2 border rounded-md bg-white disabled:bg-gray-100"
        >
          <option v-if="databasesLoading" value="">Loading databases...</option>
          <option v-else-if="databases.length === 0" value="">No databases available</option>
          <option v-for="db in databases" :key="db" :value="db">
            {{ db }}
          </option>
        </select>
        
        <div class="relative">
          <Input 
            v-model="searchQuery"
            placeholder="Search tables..."
            class="pl-10 w-64"
          />
          <Database class="absolute left-3 top-3 h-4 w-4 text-gray-400" />
        </div>
      </div>
      
      <div class="flex items-center space-x-2">
        <Badge variant="outline" class="text-xs">
          {{ filteredTables.length }} tables
        </Badge>
        <Button size="sm" variant="outline" @click="loadTables" :disabled="loading || !selectedDatabase">
          <Database class="w-4 h-4 mr-2" />
          Refresh
        </Button>
        <Button size="sm">
          <Plus class="w-4 h-4 mr-2" />
          Create Table
        </Button>
      </div>
    </div>

    <!-- Loading Databases -->
    <div v-if="databasesLoading" class="text-center py-8">
      <div class="inline-flex items-center space-x-2">
        <div class="animate-spin rounded-full h-4 w-4 border-b-2 border-blue-600"></div>
        <span>Loading databases...</span>
      </div>
    </div>

    <!-- No Databases -->
    <div v-else-if="!selectedDatabase" class="text-center py-8 text-gray-500">
      <Database class="mx-auto h-12 w-12 text-gray-400 mb-4" />
      <h3 class="text-lg font-medium text-gray-900 mb-2">No database selected</h3>
      <p class="text-gray-600">
        Please select a database from the dropdown above
      </p>
    </div>

    <!-- Loading Tables -->
    <div v-else-if="loading" class="text-center py-8">
      <div class="inline-flex items-center space-x-2">
        <div class="animate-spin rounded-full h-4 w-4 border-b-2 border-blue-600"></div>
        <span>Loading tables...</span>
      </div>
    </div>

    <!-- No Tables -->
    <div v-else-if="filteredTables.length === 0" class="text-center py-8 text-gray-500">
      <Database class="mx-auto h-12 w-12 text-gray-400 mb-4" />
      <h3 class="text-lg font-medium text-gray-900 mb-2">No tables found</h3>
      <p class="text-gray-600">
        {{ searchQuery ? 'Try adjusting your search' : 'No tables exist in this database' }}
      </p>
    </div>

    <!-- Table List -->
    <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
      <Card 
        v-for="table in filteredTables" 
        :key="table.name"
        class="cursor-pointer hover:shadow-md transition-shadow"
        @click="selectTable(table)"
      >
        <CardContent class="p-6">
          <div class="flex items-start justify-between">
            <div class="flex items-center space-x-3">
              <div class="flex h-10 w-10 items-center justify-center rounded-lg bg-blue-100">
                <Database class="h-6 w-6 text-blue-600" />
              </div>
              <div>
                <h3 class="font-medium text-gray-900">{{ table.name }}</h3>
                <div class="flex items-center space-x-4 mt-1">
                  <span class="text-sm text-gray-500">
                    {{ table.columns?.length || 0 }} columns
                  </span>
                  <span v-if="table.row_count" class="text-sm text-gray-500">
                    • {{ table.row_count.toLocaleString() }} rows
                  </span>
                </div>
                <p class="text-xs text-gray-500 mt-1">
                  {{ table.database || selectedDatabase }}
                </p>
              </div>
            </div>
            
            <div class="flex space-x-1">
              <Button 
                size="sm" 
                variant="ghost" 
                @click.stop="previewTable(table)"
              >
                <Eye class="w-4 h-4" />
              </Button>
              <Button 
                size="sm" 
                variant="ghost" 
                @click.stop="tableActions(table)"
              >
                <MoreVertical class="w-4 h-4" />
              </Button>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>

    <!-- Table Preview Dialog -->
    <Dialog v-model:open="showPreview" @update:open="showPreview = false">
      <DialogContent class="max-w-4xl">
        <DialogHeader>
          <DialogTitle class="flex items-center space-x-2">
            <Database class="w-5 h-5 text-blue-600" />
            {{ selectedTable?.name }}
          </DialogTitle>
          <DialogDescription>
            Table schema and metadata
          </DialogDescription>
        </DialogHeader>
        
        <div v-if="selectedTable" class="space-y-6">
          <!-- Table Info -->
          <div class="grid grid-cols-2 gap-4">
            <div>
              <div class="text-sm font-medium text-gray-700">Table Name</div>
              <div class="mt-1 px-3 py-2 bg-gray-50 rounded">{{ selectedTable.name }}</div>
            </div>
            <div>
              <div class="text-sm font-medium text-gray-700">Database</div>
              <div class="mt-1 px-3 py-2 bg-gray-50 rounded">{{ selectedTable.database || selectedDatabase }}</div>
            </div>
            <div>
              <div class="text-sm font-medium text-gray-700">Row Count</div>
              <div class="mt-1 px-3 py-2 bg-gray-50 rounded">{{ selectedTable.row_count?.toLocaleString() || 'Unknown' }}</div>
            </div>
            <div>
              <div class="text-sm font-medium text-gray-700">Created At</div>
              <div class="mt-1 px-3 py-2 bg-gray-50 rounded">{{ new Date(selectedTable.created_at).toLocaleDateString() }}</div>
            </div>
          </div>

          <!-- Schema -->
          <div>
            <h4 class="text-lg font-medium mb-4">Table Schema</h4>
            <div v-if="selectedTable.columns && selectedTable.columns.length > 0" class="space-y-2">
              <div 
                v-for="column in selectedTable.columns" 
                :key="column.name"
                class="flex items-center justify-between p-3 border rounded-lg"
              >
                <div>
                  <div class="font-medium">{{ column.name }}</div>
                  <div class="text-sm text-gray-600">
                    Type: {{ column.type }}
                    <span v-if="column.nullable" class="text-gray-500">• Nullable</span>
                  </div>
                  <div v-if="column.comment" class="text-xs text-gray-500 mt-1">
                    {{ column.comment }}
                  </div>
                </div>
              </div>
            </div>
            <div v-else class="text-center py-8 text-gray-500">
              No schema information available
            </div>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  </div>
</template>
<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useJobs } from '@/composables/useApi'
import { apiClient } from '@/api'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { 
  Settings, 
  Users, 
  Server, 
  Database,
  RefreshCw,
  Save,
  Info
} from 'lucide-vue-next'

const { jobStats, fetchJobStats, loading, updateWorkerCount } = useJobs()

const workerCount = ref(3)
const apiInfo = ref<any>(null)
const systemInfo = ref({
  version: '1.0.0',
  uptime: '0h 0m',
  status: 'healthy'
})

const handleUpdateWorkers = async () => {
  await updateWorkerCount(workerCount.value)
  await fetchJobStats()
}

const fetchApiInfo = async () => {
  try {
    const response = await apiClient.getApiInfo()
    if (response.success) {
      apiInfo.value = response.data
    }
  } catch (error) {
    console.error('Failed to fetch API info:', error)
  }
}

const refreshSystemInfo = () => {
  fetchJobStats()
  fetchApiInfo()
}

onMounted(() => {
  refreshSystemInfo()
  if (jobStats.value) {
    workerCount.value = jobStats.value.workers.total_workers
  }
})
</script>

<template>
  <div class="space-y-6">
    <!-- Page Header -->
    <div class="flex items-center justify-between">
      <div>
        <h2 class="text-3xl font-bold text-gray-900">Settings</h2>
        <p class="mt-2 text-gray-600">Configure system settings and view information</p>
      </div>
      <div class="flex items-center space-x-3">
        <Button @click="refreshSystemInfo" variant="outline" size="sm">
          <RefreshCw class="w-4 h-4 mr-2" />
          Refresh
        </Button>
      </div>
    </div>

    <!-- Worker Configuration -->
    <Card>
      <CardHeader>
        <CardTitle class="flex items-center">
          <Users class="mr-2 h-5 w-5" />
          Worker Pool Configuration
        </CardTitle>
        <CardDescription>
          Configure the number of concurrent workers for job processing
        </CardDescription>
      </CardHeader>
      <CardContent>
        <div class="space-y-4">
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-2">
              Number of Workers
            </label>
            <div class="flex items-center space-x-4">
              <Input
                v-model.number="workerCount"
                type="number"
                min="1"
                max="10"
                class="w-32"
              />
              <Button @click="handleUpdateWorkers" :disabled="loading">
                <Save class="w-4 h-4 mr-2" />
                Update
              </Button>
            </div>
            <p class="text-xs text-gray-500 mt-1">
              Current: {{ jobStats?.workers?.active_jobs || 0 }} active, {{ jobStats?.workers?.total_workers || 0 }} max
            </p>
          </div>
          
          <div class="grid grid-cols-1 md:grid-cols-3 gap-4 pt-4">
            <div class="text-center p-4 bg-gray-50 rounded-lg">
              <div class="text-2xl font-bold text-blue-600">{{ jobStats?.workers?.active_jobs || 0 }}</div>
              <div class="text-sm text-gray-600">Active Workers</div>
            </div>
            <div class="text-center p-4 bg-gray-50 rounded-lg">
              <div class="text-2xl font-bold text-green-600">{{ jobStats?.queue?.processing || 0 }}</div>
              <div class="text-sm text-gray-600">Processing Jobs</div>
            </div>
            <div class="text-center p-4 bg-gray-50 rounded-lg">
              <div class="text-2xl font-bold text-yellow-600">{{ jobStats?.queue?.total || 0 }}</div>
              <div class="text-sm text-gray-600">Queue Size</div>
            </div>
          </div>
        </div>
      </CardContent>
    </Card>

    <!-- System Information -->
    <Card>
      <CardHeader>
        <CardTitle class="flex items-center">
          <Server class="mr-2 h-5 w-5" />
          System Information
        </CardTitle>
        <CardDescription>
          Current system status and configuration
        </CardDescription>
      </CardHeader>
      <CardContent>
        <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
          <div class="space-y-4">
            <div>
              <h4 class="text-sm font-medium text-gray-700 mb-2">Application</h4>
              <div class="space-y-2">
                <div class="flex justify-between">
                  <span class="text-sm text-gray-600">Version</span>
                  <Badge variant="outline">{{ systemInfo.version }}</Badge>
                </div>
                <div class="flex justify-between">
                  <span class="text-sm text-gray-600">Status</span>
                  <Badge class="bg-green-100 text-green-800">Healthy</Badge>
                </div>
                <div class="flex justify-between">
                  <span class="text-sm text-gray-600">Uptime</span>
                  <span class="text-sm font-medium">{{ systemInfo.uptime }}</span>
                </div>
              </div>
            </div>
          </div>
          
          <div class="space-y-4">
            <div>
              <h4 class="text-sm font-medium text-gray-700 mb-2">API Information</h4>
              <div class="space-y-2">
                <div class="flex justify-between">
                  <span class="text-sm text-gray-600">Name</span>
                  <span class="text-sm font-medium">Bronze Backend API</span>
                </div>
                <div class="flex justify-between">
                  <span class="text-sm text-gray-600">Version</span>
                  <Badge variant="outline">1.0.0</Badge>
                </div>
                <div class="flex justify-between">
                  <span class="text-sm text-gray-600">Endpoints</span>
                  <span class="text-sm font-medium">15+</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </CardContent>
    </Card>

    <!-- Features -->
    <Card>
      <CardHeader>
        <CardTitle class="flex items-center">
          <Info class="mr-2 h-5 w-5" />
          Features
        </CardTitle>
        <CardDescription>
          Available features and capabilities
        </CardDescription>
      </CardHeader>
      <CardContent>
        <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          <div class="flex items-center space-x-3 p-3 border rounded-lg">
            <Database class="h-8 w-8 text-blue-500" />
            <div>
              <h4 class="font-medium">MinIO Integration</h4>
              <p class="text-sm text-gray-600">Object storage</p>
            </div>
          </div>
          
          <div class="flex items-center space-x-3 p-3 border rounded-lg">
            <Users class="h-8 w-8 text-green-500" />
            <div>
              <h4 class="font-medium">Job Queue</h4>
              <p class="text-sm text-gray-600">Priority processing</p>
            </div>
          </div>
          
          <div class="flex items-center space-x-3 p-3 border rounded-lg">
            <Settings class="h-8 w-8 text-purple-500" />
            <div>
              <h4 class="font-medium">Archive Support</h4>
              <p class="text-sm text-gray-600">ZIP, TAR, TAR.GZ</p>
            </div>
          </div>
          
          <div class="flex items-center space-x-3 p-3 border rounded-lg">
            <Server class="h-8 w-8 text-orange-500" />
            <div>
              <h4 class="font-medium">File Watching</h4>
              <p class="text-sm text-gray-600">Real-time monitoring</p>
            </div>
          </div>
          
          <div class="flex items-center space-x-3 p-3 border rounded-lg">
            <RefreshCw class="h-8 w-8 text-cyan-500" />
            <div>
              <h4 class="font-medium">Worker Pool</h4>
              <p class="text-sm text-gray-600">Parallel processing</p>
            </div>
          </div>
          
          <div class="flex items-center space-x-3 p-3 border rounded-lg">
            <Info class="h-8 w-8 text-indigo-500" />
            <div>
              <h4 class="font-medium">REST API</h4>
              <p class="text-sm text-gray-600">Complete interface</p>
            </div>
          </div>
        </div>
      </CardContent>
    </Card>

    <!-- API Documentation -->
    <Card>
      <CardHeader>
        <CardTitle>API Documentation</CardTitle>
        <CardDescription>
          Access detailed API documentation
        </CardDescription>
      </CardHeader>
      <CardContent>
        <div class="space-y-4">
          <div class="flex items-center justify-between p-4 border rounded-lg">
            <div>
              <h4 class="font-medium">OpenAPI Specification</h4>
              <p class="text-sm text-gray-600">Complete API documentation in JSON format</p>
            </div>
            <Button variant="outline" size="sm">
              View Spec
            </Button>
          </div>
          
          <div class="flex items-center justify-between p-4 border rounded-lg">
            <div>
              <h4 class="font-medium">API Info</h4>
              <p class="text-sm text-gray-600">Overview of all available endpoints</p>
            </div>
            <Button variant="outline" size="sm">
              View Info
            </Button>
          </div>
        </div>
      </CardContent>
    </Card>
  </div>
</template>
<script setup lang="ts">
import { ref, onMounted, onUnmounted, inject } from 'vue'
import { useJobs, useFiles } from '@/composables/useApi'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Skeleton } from '@/components/ui/skeleton'
import { 
  FolderOpen, 
  CheckCircle, 
  Activity,
  Users
} from 'lucide-vue-next'

const { jobStats, fetchJobStats, loading: jobsLoading } = useJobs()
const { files, fetchFiles, loading: filesLoading } = useFiles()
const toast = inject('toast') as any

let refreshInterval: number

// const recentJobs = ref([])
const systemHealth = ref({
  status: 'healthy',
  uptime: '2h 34m',
  version: '1.0.0'
})

const refreshData = async () => {
  try {
    await Promise.all([
      fetchJobStats(),
      fetchFiles()
    ])
  } catch (error) {
    toast?.error('Failed to refresh data', 'Please try again later')
  }
}

onMounted(async () => {
  await refreshData()
  
  // Set up auto-refresh every 30 seconds
  refreshInterval = setInterval(refreshData, 30000)
})

onUnmounted(() => {
  if (refreshInterval) {
    clearInterval(refreshInterval)
  }
})

// const getJobStatusColor = (status: string) => {
//   const colors = {
//     pending: 'bg-yellow-100 text-yellow-800',
//     processing: 'bg-blue-100 text-blue-800',
//     completed: 'bg-green-100 text-green-800',
//     failed: 'bg-red-100 text-red-800',
//     cancelled: 'bg-gray-100 text-gray-800'
//   }
//   return colors[status as keyof typeof colors] || 'bg-gray-100 text-gray-800'
// }

const formatFileSize = (bytes: number) => {
  if (bytes === 0) return '0 Bytes'
  const k = 1024
  const sizes = ['Bytes', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}
</script>

<template>
  <div class="space-y-6">
    <!-- Page Header -->
    <div>
      <h2 class="text-3xl font-bold text-gray-900">Dashboard</h2>
      <p class="mt-2 text-gray-600">Overview of your Bronze system</p>
    </div>

    <!-- Stats Grid -->
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
      <Card>
        <CardHeader class="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle class="text-sm font-medium">Total Files</CardTitle>
          <FolderOpen class="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <div v-if="filesLoading" class="space-y-2">
            <Skeleton class="h-8 w-16" />
            <Skeleton class="h-4 w-24" />
          </div>
          <div v-else>
            <div class="text-2xl font-bold">{{ files.length }}</div>
            <p class="text-xs text-muted-foreground">
              Stored in MinIO
            </p>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader class="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle class="text-sm font-medium">Active Jobs</CardTitle>
          <Activity class="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <div v-if="jobsLoading" class="space-y-2">
            <Skeleton class="h-8 w-16" />
            <Skeleton class="h-4 w-24" />
          </div>
          <div v-else>
            <div class="text-2xl font-bold">
              {{ jobStats?.queue?.processing || 0 }}
            </div>
            <p class="text-xs text-muted-foreground">
              {{ jobStats?.queue?.pending || 0 }} pending
            </p>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader class="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle class="text-sm font-medium">Completed Jobs</CardTitle>
          <CheckCircle class="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <div class="text-2xl font-bold">
            {{ jobStats?.queue?.completed || 0 }}
          </div>
          <p class="text-xs text-muted-foreground">
            {{ jobStats?.queue?.failed || 0 }} failed
          </p>
        </CardContent>
      </Card>

      <Card>
        <CardHeader class="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle class="text-sm font-medium">Workers</CardTitle>
          <Users class="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <div class="text-2xl font-bold">
            {{ jobStats?.workers?.active_jobs || 0 }}/{{ jobStats?.workers?.total_workers || 0 }}
          </div>
          <p class="text-xs text-muted-foreground">
            Active workers
          </p>
        </CardContent>
      </Card>
    </div>

    <!-- System Health -->
    <Card>
      <CardHeader>
        <CardTitle class="flex items-center">
          <Activity class="mr-2 h-5 w-5" />
          System Health
        </CardTitle>
        <CardDescription>
          Current system status and performance metrics
        </CardDescription>
      </CardHeader>
      <CardContent>
        <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
          <div class="flex items-center space-x-3">
            <div class="w-3 h-3 bg-green-500 rounded-full"></div>
            <div>
              <p class="text-sm font-medium">Status</p>
              <p class="text-xs text-gray-500">{{ systemHealth.status }}</p>
            </div>
          </div>
          <div>
            <p class="text-sm font-medium">Uptime</p>
            <p class="text-xs text-gray-500">{{ systemHealth.uptime }}</p>
          </div>
          <div>
            <p class="text-sm font-medium">Version</p>
            <p class="text-xs text-gray-500">{{ systemHealth.version }}</p>
          </div>
        </div>
      </CardContent>
    </Card>

    <!-- Recent Activity -->
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
      <Card>
        <CardHeader>
          <CardTitle>Recent Files</CardTitle>
          <CardDescription>
            Latest files uploaded to the system
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div v-if="files.length === 0" class="text-center py-8 text-gray-500">
            No files uploaded yet
          </div>
          <div v-else class="space-y-3">
            <div
              v-for="file in files.slice(0, 5)"
              :key="file.key"
              class="flex items-center justify-between p-3 bg-gray-50 rounded-lg"
            >
              <div class="flex items-center space-x-3">
                <FolderOpen class="h-4 w-4 text-gray-400" />
                <div>
                  <p class="text-sm font-medium">{{ file.key }}</p>
                  <p class="text-xs text-gray-500">{{ formatFileSize(file.size) }}</p>
                </div>
              </div>
              <Badge variant="secondary">
                {{ new Date(file.last_modified).toLocaleDateString() }}
              </Badge>
            </div>
          </div>
          <div class="mt-4">
            <router-link to="/files">
              <Button variant="outline" size="sm" class="w-full">
                View All Files
              </Button>
            </router-link>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Job Queue Status</CardTitle>
          <CardDescription>
            Current job processing queue information
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div class="space-y-4">
              <div class="flex items-center justify-between">
                <span class="text-sm font-medium">Queue Size</span>
                <Badge variant="outline">{{ jobStats?.queue?.total || 0 }}</Badge>
              </div>
              <div class="flex items-center justify-between">
                <span class="text-sm font-medium">Total Jobs</span>
                <Badge variant="outline">{{ jobStats?.queue?.total || 0 }}</Badge>
              </div>
              <div class="flex items-center justify-between">
                <span class="text-sm font-medium">Success Rate</span>
                <Badge 
                  :variant="jobStats && jobStats.queue.completed > 0 ? 'default' : 'secondary'"
                >
                  {{ jobStats ? Math.round((jobStats.queue.completed / Math.max(jobStats.queue.total, 1)) * 100) : 0 }}%
                </Badge>
              </div>
          </div>
          <div class="mt-4">
            <router-link to="/jobs">
              <Button variant="outline" size="sm" class="w-full">
                Manage Jobs
              </Button>
            </router-link>
          </div>
        </CardContent>
      </Card>
    </div>
  </div>
</template>
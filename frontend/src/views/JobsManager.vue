<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useJobs } from '@/composables/useApi'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Input } from '@/components/ui/input'
import { Progress } from '@/components/ui/progress'
import { 
  Briefcase, 
  Play, 
  Square, 
  Search,
  RefreshCw,
  Clock,
  CheckCircle,
  XCircle,
  AlertCircle,
  TrendingUp,
  Users
} from 'lucide-vue-next'

const { 
  jobs, 
  jobStats, 
  fetchJobs, 
  fetchJobStats, 
  createJob, 
  cancelJob, 
  updateJobPriority,
  loading,
  error 
} = useJobs()

const searchQuery = ref('')
const statusFilter = ref('')
const createJobDialogOpen = ref(false)
const newJob = ref({
  type: 'decompress',
  bucket: 'files',
  object_name: '',
  priority: 'medium' as 'low' | 'medium' | 'high'
})

const filteredJobs = computed(() => {
  let filtered = jobs.value
  
  if (statusFilter.value) {
    filtered = filtered.filter(job => job.status === statusFilter.value)
  }
  
  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase()
    filtered = filtered.filter(job => 
      job.object_name.toLowerCase().includes(query) ||
      job.id.toLowerCase().includes(query) ||
      job.type.toLowerCase().includes(query)
    )
  }
  
  return filtered.sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime())
})

const getJobStatusColor = (status: string) => {
  const colors = {
    pending: 'bg-yellow-100 text-yellow-800',
    processing: 'bg-blue-100 text-blue-800',
    completed: 'bg-green-100 text-green-800',
    failed: 'bg-red-100 text-red-800',
    cancelled: 'bg-gray-100 text-gray-800'
  }
  return colors[status as keyof typeof colors] || 'bg-gray-100 text-gray-800'
}

const getJobStatusIcon = (status: string) => {
  const icons = {
    pending: Clock,
    processing: Play,
    completed: CheckCircle,
    failed: XCircle,
    cancelled: Square
  }
  return icons[status as keyof typeof icons] || AlertCircle
}

const getPriorityColor = (priority: string) => {
  const colors = {
    low: 'bg-gray-100 text-gray-800',
    medium: 'bg-yellow-100 text-yellow-800',
    high: 'bg-red-100 text-red-800'
  }
  return colors[priority as keyof typeof colors] || 'bg-gray-100 text-gray-800'
}

const getPriorityLabel = (priority: string) => {
  const labels = {
    low: 'Low',
    medium: 'Medium',
    high: 'High'
  }
  return labels[priority as keyof typeof labels] || 'Unknown'
}

const formatDate = (dateString: string) => {
  return new Date(dateString).toLocaleString()
}

const formatDuration = (start: string, end?: string) => {
  const startTime = new Date(start).getTime()
  const endTime = end ? new Date(end).getTime() : Date.now()
  const duration = endTime - startTime
  
  if (duration < 1000) return `${duration}ms`
  if (duration < 60000) return `${Math.round(duration / 1000)}s`
  return `${Math.round(duration / 60000)}m`
}

const handleCreateJob = async () => {
  if (!newJob.value.object_name) return
  
  await createJob(newJob.value)
  newJob.value.object_name = ''
  createJobDialogOpen.value = false
}

const handleCancelJob = async (jobId: string) => {
  if (confirm('Are you sure you want to cancel this job?')) {
    await cancelJob(jobId)
  }
}

const handleUpdatePriority = async (jobId: string, priority: 'low' | 'medium' | 'high') => {
  await updateJobPriority(jobId, priority)
}

const refreshData = () => {
  Promise.all([
    fetchJobs(),
    fetchJobStats()
  ])
}

onMounted(() => {
  refreshData()
})
</script>

<template>
  <div class="space-y-6">
    <!-- Page Header -->
    <div class="flex items-center justify-between">
      <div>
        <h2 class="text-3xl font-bold text-gray-900">Jobs</h2>
        <p class="mt-2 text-gray-600">Manage and monitor processing jobs</p>
      </div>
      <div class="flex items-center space-x-3">
        <Button @click="refreshData" variant="outline" size="sm">
          <RefreshCw class="w-4 h-4 mr-2" />
          Refresh
        </Button>
        <Button @click="createJobDialogOpen = true">
          <Briefcase class="w-4 h-4 mr-2" />
          Create Job
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
        </div>
      </div>
    </div>

    <!-- Stats Cards -->
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
      <Card>
        <CardHeader class="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle class="text-sm font-medium">Total Jobs</CardTitle>
          <Briefcase class="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <div class="text-2xl font-bold">{{ jobStats?.queue?.total || 0 }}</div>
          <p class="text-xs text-muted-foreground">
            All time jobs
          </p>
        </CardContent>
      </Card>

      <Card>
        <CardHeader class="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle class="text-sm font-medium">Active Jobs</CardTitle>
          <Play class="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <div class="text-2xl font-bold">
            {{ jobStats?.queue?.processing || 0 }}
          </div>
          <p class="text-xs text-muted-foreground">
            {{ jobStats?.queue?.pending || 0 }} pending
          </p>
        </CardContent>
      </Card>

      <Card>
        <CardHeader class="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle class="text-sm font-medium">Success Rate</CardTitle>
          <TrendingUp class="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <div class="text-2xl font-bold">
            {{ jobStats ? Math.round((jobStats.queue.completed / Math.max(jobStats.queue.total, 1)) * 100) : 0 }}%
          </div>
          <p class="text-xs text-muted-foreground">
            {{ jobStats?.queue?.completed || 0 }} completed
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

    <!-- Filters -->
    <Card>
      <CardContent class="pt-6">
        <div class="flex items-center space-x-4">
          <div class="flex-1 relative">
            <Search class="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-gray-400" />
            <Input
              v-model="searchQuery"
              placeholder="Search jobs..."
              class="pl-10"
            />
          </div>
          <select
            v-model="statusFilter"
            class="px-3 py-2 border border-gray-300 rounded-md text-sm"
          >
            <option value="">All Status</option>
            <option value="pending">Pending</option>
            <option value="processing">Processing</option>
            <option value="completed">Completed</option>
            <option value="failed">Failed</option>
            <option value="cancelled">Cancelled</option>
          </select>
        </div>
      </CardContent>
    </Card>

    <!-- Jobs List -->
    <Card>
      <CardHeader>
        <CardTitle>Jobs ({{ filteredJobs.length }})</CardTitle>
        <CardDescription>
          All processing jobs in the system
        </CardDescription>
      </CardHeader>
      <CardContent>
        <div v-if="loading" class="text-center py-8">
          <RefreshCw class="mx-auto h-8 w-8 animate-spin text-gray-400" />
          <p class="mt-2 text-gray-500">Loading jobs...</p>
        </div>
        
        <div v-else-if="filteredJobs.length === 0" class="text-center py-8 text-gray-500">
          <Briefcase class="mx-auto h-12 w-12 text-gray-400" />
          <p class="mt-2">No jobs found</p>
        </div>

        <div v-else class="space-y-4">
          <div
            v-for="job in filteredJobs"
            :key="job.id"
            class="border rounded-lg p-4 hover:bg-gray-50 transition-colors"
          >
            <div class="flex items-start justify-between">
              <div class="flex-1">
                <div class="flex items-center space-x-3 mb-2">
                  <component
                    :is="getJobStatusIcon(job.status)"
                    class="h-5 w-5 text-gray-400"
                  />
                  <h3 class="font-medium text-gray-900">{{ job.object_name }}</h3>
                  <Badge :class="getJobStatusColor(job.status)">
                    {{ job.status }}
                  </Badge>
                  <Badge :class="getPriorityColor(job.priority)">
                    {{ getPriorityLabel(job.priority) }}
                  </Badge>
                </div>
                
                <div class="text-sm text-gray-600 space-y-1">
                  <p>ID: {{ job.id }}</p>
                  <p>Type: {{ job.type }}</p>
                  <p>Created: {{ formatDate(job.created_at) }}</p>
                  <p v-if="job.started_at">Started: {{ formatDate(job.started_at) }}</p>
                  <p v-if="job.completed_at">Completed: {{ formatDate(job.completed_at) }}</p>
                  <p v-if="job.started_at">
                    Duration: {{ formatDuration(job.started_at, job.completed_at) }}
                  </p>
                  <p v-if="job.error" class="text-red-600">Error: {{ job.error }}</p>
                </div>

                <!-- Progress Bar -->
                <div v-if="job.status === 'processing'" class="mt-3">
                  <Progress 
                    :value="job.progress" 
                    :max="100" 
                    color="blue" 
                    size="sm"
                    show-label
                  />
                </div>
              </div>

              <div class="flex items-center space-x-2 ml-4">
                <select
                  v-if="job.status === 'pending'"
                  :value="job.priority"
                  @change="handleUpdatePriority(job.id, ($event.target as HTMLSelectElement).value as 'low' | 'medium' | 'high')"
                  class="text-xs px-2 py-1 border border-gray-300 rounded"
                >
                  <option value="low">Low</option>
                  <option value="medium">Medium</option>
                  <option value="high">High</option>
                </select>
                
                <Button
                  v-if="job.status === 'pending' || job.status === 'processing'"
                  variant="ghost"
                  size="sm"
                  @click="handleCancelJob(job.id)"
                  class="text-red-600 hover:text-red-700"
                >
                  <Square class="h-4 w-4" />
                </Button>
              </div>
            </div>
          </div>
        </div>
      </CardContent>
    </Card>

    <!-- Create Job Dialog -->
    <div v-if="createJobDialogOpen" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div class="bg-white rounded-lg p-6 w-full max-w-md">
        <h3 class="text-lg font-medium mb-4">Create New Job</h3>
        
        <div class="space-y-4">
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">
              Object Name
            </label>
            <Input
              v-model="newJob.object_name"
              placeholder="Enter object name"
            />
          </div>
          
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">
              Priority
            </label>
            <select
              v-model="newJob.priority"
              class="w-full px-3 py-2 border border-gray-300 rounded-md"
            >
              <option value="low">Low</option>
              <option value="medium">Medium</option>
              <option value="high">High</option>
            </select>
          </div>
        </div>
        
        <div class="flex justify-end space-x-3 mt-6">
          <Button
            variant="outline"
            @click="createJobDialogOpen = false"
          >
            Cancel
          </Button>
          <Button
            @click="handleCreateJob"
            :disabled="!newJob.object_name"
          >
            Create Job
          </Button>
        </div>
      </div>
    </div>
  </div>
</template>
<template>
  <div class="bg-card border rounded-lg p-4">
    <div class="flex items-center justify-between mb-4">
      <h3 class="font-medium">{{ job.job_type }} - {{ job.object_name }}</h3>
      <span :class="getStatusClass(job.status)">{{ job.status }}</span>
    </div>
    
    <div class="space-y-2 text-sm text-muted-foreground">
      <div>Priority: {{ job.priority }}</div>
      <div>Created: {{ formatDate(job.created_at) }}</div>
      <div v-if="job.started_at">Started: {{ formatDate(job.started_at) }}</div>
      <div v-if="job.completed_at">Completed: {{ formatDate(job.completed_at) }}</div>
    </div>

    <div v-if="job.status === 'running' && job.progress !== undefined" class="mt-4">
      <div class="flex justify-between text-sm mb-1">
        <span>Progress</span>
        <span>{{ job.progress }}%</span>
      </div>
      <div class="w-full bg-secondary rounded-full h-2">
        <div 
          class="bg-primary h-2 rounded-full transition-all" 
          :style="{ width: `${job.progress}%` }"
        ></div>
      </div>
    </div>

    <div v-if="job.error_message" class="mt-4 text-sm text-destructive">
      {{ job.error_message }}
    </div>
  </div>
</template>

<script setup lang="ts">
import { formatDate } from '@/utils/formatters'
import type { Job } from '@/types'

interface Props {
  job: Job
}

defineProps<Props>()

function getStatusClass(status: string) {
  const baseClass = 'px-2 py-1 rounded-full text-xs font-medium'
  switch (status) {
    case 'pending':
      return `${baseClass} bg-yellow-100 text-yellow-800`
    case 'running':
      return `${baseClass} bg-blue-100 text-blue-800`
    case 'completed':
      return `${baseClass} bg-green-100 text-green-800`
    case 'failed':
      return `${baseClass} bg-red-100 text-red-800`
    case 'cancelled':
      return `${baseClass} bg-gray-100 text-gray-800`
    default:
      return `${baseClass} bg-gray-100 text-gray-800`
  }
}
</script>
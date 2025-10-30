<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useWatcher } from '@/composables/useApi'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { 
  Eye, 
  CheckCircle, 
  Clock, 
  AlertCircle,
  RefreshCw,
  Calendar,
  CheckSquare
} from 'lucide-vue-next'

const { 
  events, 
  unprocessedEvents, 
  fetchEventHistory, 
  fetchUnprocessedEvents, 
  markEventProcessed,
  loading,
  error 
} = useWatcher()

const selectedEvents = ref<string[]>([])

const getEventTypeColor = (eventType: string) => {
  const colors = {
    's3:ObjectCreated:*': 'bg-green-100 text-green-800',
    's3:ObjectRemoved:*': 'bg-red-100 text-red-800',
    's3:ObjectMetadata:*': 'bg-blue-100 text-blue-800'
  }
  return colors[eventType as keyof typeof colors] || 'bg-gray-100 text-gray-800'
}

const getEventTypeIcon = (eventType: string) => {
  const icons = {
    's3:ObjectCreated:*': CheckCircle,
    's3:ObjectRemoved:*': Eye,
    's3:ObjectMetadata:*': AlertCircle
  }
  return icons[eventType as keyof typeof icons] || Clock
}

const getEventTypeLabel = (eventType: string) => {
  const labels = {
    's3:ObjectCreated:*': 'Created',
    's3:ObjectRemoved:*': 'Deleted',
    's3:ObjectMetadata:*': 'Updated'
  }
  return labels[eventType as keyof typeof labels] || 'Unknown'
}

const formatDate = (dateString: string) => {
  return new Date(dateString).toLocaleString()
}

const toggleEventSelection = (eventId: string) => {
  const index = selectedEvents.value.indexOf(eventId)
  if (index > -1) {
    selectedEvents.value.splice(index, 1)
  } else {
    selectedEvents.value.push(eventId)
  }
}

const handleMarkProcessed = async (eventId?: string) => {
  const eventIdToProcess = eventId || selectedEvents.value[0]
  if (!eventIdToProcess) return
  
  await markEventProcessed(eventIdToProcess)
  if (!eventId) {
    selectedEvents.value = []
  }
}

const refreshData = () => {
  Promise.all([
    fetchEventHistory(),
    fetchUnprocessedEvents()
  ])
}

onMounted(() => {
  refreshData()
})
</script>

 <template>
   <div class="space-y-6">
     <!-- Page Actions -->
     <div class="flex items-center justify-end">
       <div class="flex items-center space-x-3">
         <Button @click="refreshData" variant="outline" size="sm">
           <RefreshCw class="w-4 h-4 mr-2" />
           Refresh
         </Button>
         <Button
           @click="handleMarkProcessed()"
           :disabled="selectedEvents.length === 0"
           variant="outline"
         >
           <CheckSquare class="w-4 h-4 mr-2" />
           Mark as Processed ({{ selectedEvents.length }})
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
    <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
      <Card>
        <CardHeader class="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle class="text-sm font-medium">Unprocessed Events</CardTitle>
          <Clock class="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <div class="text-2xl font-bold">{{ unprocessedEvents.length }}</div>
          <p class="text-xs text-muted-foreground">
            Events awaiting processing
          </p>
        </CardContent>
      </Card>

      <Card>
        <CardHeader class="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle class="text-sm font-medium">Total Events</CardTitle>
          <Calendar class="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <div class="text-2xl font-bold">{{ events.length }}</div>
          <p class="text-xs text-muted-foreground">
            All time events
          </p>
        </CardContent>
      </Card>

      <Card>
        <CardHeader class="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle class="text-sm font-medium">Selected</CardTitle>
          <CheckSquare class="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <div class="text-2xl font-bold">{{ selectedEvents.length }}</div>
          <p class="text-xs text-muted-foreground">
            Events to mark processed
          </p>
        </CardContent>
      </Card>
    </div>

    <!-- Unprocessed Events -->
    <Card>
      <CardHeader>
        <CardTitle>Unprocessed Events</CardTitle>
        <CardDescription>
          File change events that haven't been processed yet
        </CardDescription>
      </CardHeader>
      <CardContent>
        <div v-if="loading" class="text-center py-8">
          <RefreshCw class="mx-auto h-8 w-8 animate-spin text-gray-400" />
          <p class="mt-2 text-gray-500">Loading events...</p>
        </div>
        
        <div v-else-if="unprocessedEvents.length === 0" class="text-center py-8 text-gray-500">
          <CheckCircle class="mx-auto h-12 w-12 text-gray-400" />
          <p class="mt-2">No unprocessed events</p>
        </div>

        <div v-else class="space-y-3">
          <div
            v-for="event in unprocessedEvents"
            :key="event.id"
            class="flex items-center justify-between p-4 border rounded-lg hover:bg-gray-50 transition-colors"
          >
            <div class="flex items-center space-x-3">
              <input
                type="checkbox"
                :checked="selectedEvents.includes(event.id)"
                @change="toggleEventSelection(event.id)"
                class="rounded border-gray-300"
              />
              <component
                :is="getEventTypeIcon(event.event_type)"
                class="h-5 w-5 text-gray-400"
              />
              <div>
                <p class="font-medium text-gray-900">{{ event.key }}</p>
                <p class="text-sm text-gray-500">
                  {{ event.bucket }} • {{ formatDate(event.event_time) }}
                </p>
              </div>
            </div>
            
            <div class="flex items-center space-x-2">
              <Badge :class="getEventTypeColor(event.event_type)">
                {{ getEventTypeLabel(event.event_type) }}
              </Badge>
              <Badge variant="outline">
                Unprocessed
              </Badge>
              <Button
                variant="ghost"
                size="sm"
                @click="handleMarkProcessed(event.id)"
                class="text-green-600 hover:text-green-700"
              >
                <CheckSquare class="h-4 w-4" />
              </Button>
            </div>
          </div>
        </div>
      </CardContent>
    </Card>

    <!-- Event History -->
    <Card>
      <CardHeader>
        <CardTitle>Event History</CardTitle>
        <CardDescription>
          Recent file change events
        </CardDescription>
      </CardHeader>
      <CardContent>
        <div v-if="loading" class="text-center py-8">
          <RefreshCw class="mx-auto h-8 w-8 animate-spin text-gray-400" />
          <p class="mt-2 text-gray-500">Loading events...</p>
        </div>
        
        <div v-else-if="events.length === 0" class="text-center py-8 text-gray-500">
          <Calendar class="mx-auto h-12 w-12 text-gray-400" />
          <p class="mt-2">No events found</p>
        </div>

        <div v-else class="space-y-3">
          <div
            v-for="event in events.slice(0, 20)"
            :key="event.id"
            class="flex items-center justify-between p-4 border rounded-lg hover:bg-gray-50 transition-colors"
          >
            <div class="flex items-center space-x-3">
              <component
                :is="getEventTypeIcon(event.event_type)"
                class="h-5 w-5 text-gray-400"
              />
              <div>
                <p class="font-medium text-gray-900">{{ event.key }}</p>
                <p class="text-sm text-gray-500">
                  {{ event.bucket }} • {{ formatDate(event.event_time) }}
                </p>
              </div>
            </div>
            
            <div class="flex items-center space-x-2">
              <Badge :class="getEventTypeColor(event.event_type)">
                {{ getEventTypeLabel(event.event_type) }}
              </Badge>
              <Badge 
                :variant="event.processed ? 'default' : 'secondary'"
              >
                {{ event.processed ? 'Processed' : 'Pending' }}
              </Badge>
            </div>
          </div>
        </div>
      </CardContent>
    </Card>
  </div>
</template>
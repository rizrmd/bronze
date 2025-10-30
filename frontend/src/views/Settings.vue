 <script setup lang="ts">
 import { ref, onMounted, computed } from 'vue'
 import { useJobs } from '@/composables/useApi'
 import { useConfig, type ConfigData } from '@/composables/useConfig'
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
   Info,
   Edit,
   RotateCcw
 } from 'lucide-vue-next'

 const { jobStats, fetchJobStats, loading: jobsLoading, updateWorkerCount } = useJobs()
 const { 
   config, 
   loading: configLoading, 
   error: configError,
   serverConfig,
   minioConfig,
   processingConfig,
   fetchConfig,
   updateConfig,
   resetConfig
 } = useConfig()

 const workerCount = ref(3)
 const apiInfo = ref<any>(null)
 const systemInfo = ref({
   version: '1.0.0',
   uptime: '0h 0m',
   status: 'healthy'
 })

 const editingConfig = ref(false)
 const configChanges = ref<Record<string, string>>({})

 const hasChanges = computed(() => {
   return editingConfig.value && Object.keys(configChanges.value).some(key => 
     configChanges.value[key] !== config.value[key as keyof ConfigData]
   )
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
   fetchConfig()
 }

 const startEditing = () => {
   editingConfig.value = true
   // Initialize config changes with current values
   configChanges.value = { ...config.value }
 }

 const cancelEditing = () => {
   editingConfig.value = false
   configChanges.value = {}
 }

 const updateConfigField = (key: string, value: string | number) => {
   if (editingConfig.value) {
     configChanges.value[key] = String(value)
   }
 }

 const saveConfig = async () => {
   const success = await updateConfig(configChanges.value)
   if (success) {
     editingConfig.value = false
     configChanges.value = {}
     refreshSystemInfo()
   }
 }

 const resetToDefaults = () => {
   resetConfig()
   configChanges.value = {}
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
     <!-- Page Actions -->
     <div class="flex items-center justify-end">
       <div class="flex items-center space-x-3">
         <Button @click="refreshSystemInfo" variant="outline" size="sm">
           <RefreshCw class="w-4 h-4 mr-2" />
           Refresh
         </Button>
         <Button v-if="!editingConfig" @click="startEditing" variant="outline" size="sm">
           <Edit class="w-4 h-4 mr-2" />
           Edit Config
         </Button>
         <template v-else>
           <Button @click="cancelEditing" variant="outline" size="sm">
             Cancel
           </Button>
           <Button @click="resetToDefaults" variant="outline" size="sm">
             <RotateCcw class="w-4 h-4 mr-2" />
             Reset
           </Button>
           <Button @click="saveConfig" :disabled="!hasChanges || configLoading" size="sm">
             <Save class="w-4 h-4 mr-2" />
             Save
           </Button>
         </template>
       </div>
     </div>

     <!-- Configuration Editor -->
     <Card>
       <CardHeader>
         <CardTitle class="flex items-center">
           <Settings class="mr-2 h-5 w-5" />
           Environment Configuration
         </CardTitle>
         <CardDescription>
           Edit .env configuration values
         </CardDescription>
       </CardHeader>
       <CardContent>
         <div v-if="configLoading" class="text-center py-8">
           <RefreshCw class="mx-auto h-8 w-8 animate-spin text-gray-400" />
           <p class="mt-2 text-gray-500">Loading configuration...</p>
         </div>
         
         <div v-else class="space-y-6">
           <!-- Server Configuration -->
           <div>
             <h4 class="text-lg font-medium mb-4">Server Settings</h4>
             <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
               <div>
                 <label class="block text-sm font-medium text-gray-700 mb-1">Server Host</label>
                  <Input
                    :model-value="editingConfig ? configChanges.SERVER_HOST : config.SERVER_HOST"
                    @input="updateConfigField('SERVER_HOST', $event.target.value)"
                    :disabled="!editingConfig"
                    placeholder="localhost"
                  />
               </div>
               <div>
                 <label class="block text-sm font-medium text-gray-700 mb-1">Server Port</label>
                  <Input
                    :model-value="editingConfig ? configChanges.SERVER_PORT : config.SERVER_PORT"
                    @input="updateConfigField('SERVER_PORT', $event.target.value)"
                    :disabled="!editingConfig"
                    type="number"
                    placeholder="8080"
                  />
               </div>
             </div>
           </div>

           <!-- MinIO Configuration -->
           <div>
             <h4 class="text-lg font-medium mb-4">MinIO Settings</h4>
             <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
               <div>
                 <label class="block text-sm font-medium text-gray-700 mb-1">Endpoint</label>
                  <Input
                    :model-value="editingConfig ? configChanges.MINIO_ENDPOINT : config.MINIO_ENDPOINT"
                    @input="updateConfigField('MINIO_ENDPOINT', $event.target.value)"
                    :disabled="!editingConfig"
                    placeholder="localhost:9000"
                  />
               </div>
               <div>
                 <label class="block text-sm font-medium text-gray-700 mb-1">Bucket</label>
                  <Input
                    :model-value="editingConfig ? configChanges.MINIO_BUCKET : config.MINIO_BUCKET"
                    @input="updateConfigField('MINIO_BUCKET', $event.target.value)"
                    :disabled="!editingConfig"
                    placeholder="files"
                  />
               </div>
               <div>
                 <label class="block text-sm font-medium text-gray-700 mb-1">Access Key</label>
                  <Input
                    :model-value="editingConfig ? configChanges.MINIO_ACCESS_KEY : config.MINIO_ACCESS_KEY"
                    @input="updateConfigField('MINIO_ACCESS_KEY', $event.target.value)"
                    :disabled="!editingConfig"
                    type="password"
                    placeholder="minioadmin"
                  />
               </div>
               <div>
                 <label class="block text-sm font-medium text-gray-700 mb-1">Secret Key</label>
                  <Input
                    :model-value="editingConfig ? configChanges.MINIO_SECRET_KEY : config.MINIO_SECRET_KEY"
                    @input="updateConfigField('MINIO_SECRET_KEY', $event.target.value)"
                    :disabled="!editingConfig"
                    type="password"
                    placeholder="minioadmin"
                  />
               </div>
                <div>
                  <label class="block text-sm font-medium text-gray-700 mb-1">Region</label>
                   <Input
                     :model-value="editingConfig ? configChanges.MINIO_REGION : config.MINIO_REGION"
                     @input="updateConfigField('MINIO_REGION', $event.target.value)"
                     :disabled="!editingConfig"
                     placeholder="us-east-1"
                   />
                </div>
             </div>
           </div>

           <!-- Processing Configuration -->
           <div>
             <h4 class="text-lg font-medium mb-4">Processing Settings</h4>
             <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
               <div>
                 <label class="block text-sm font-medium text-gray-700 mb-1">Max Workers</label>
                  <Input
                    :model-value="editingConfig ? configChanges.MAX_WORKERS : config.MAX_WORKERS"
                    @input="updateConfigField('MAX_WORKERS', $event.target.value)"
                    :disabled="!editingConfig"
                    type="number"
                    placeholder="3"
                  />
               </div>
               <div>
                 <label class="block text-sm font-medium text-gray-700 mb-1">Queue Size</label>
                  <Input
                    :model-value="editingConfig ? configChanges.QUEUE_SIZE : config.QUEUE_SIZE"
                    @input="updateConfigField('QUEUE_SIZE', $event.target.value)"
                    :disabled="!editingConfig"
                   type="number"
                   placeholder="100"
                 />
               </div>
               <div>
                 <label class="block text-sm font-medium text-gray-700 mb-1">Watch Interval</label>
                  <Input
                    :model-value="editingConfig ? configChanges.WATCH_INTERVAL : config.WATCH_INTERVAL"
                    @input="updateConfigField('WATCH_INTERVAL', $event.target.value)"
                    :disabled="!editingConfig"
                    placeholder="5s"
                  />
               </div>
               <div>
                 <label class="block text-sm font-medium text-gray-700 mb-1">Temp Directory</label>
                  <Input
                    :model-value="editingConfig ? configChanges.TEMP_DIR : config.TEMP_DIR"
                    @input="updateConfigField('TEMP_DIR', $event.target.value)"
                    :disabled="!editingConfig"
                    placeholder="/tmp/bronze"
                  />
               </div>
             </div>
           </div>

           <!-- Decompression Configuration -->
           <div>
             <h4 class="text-lg font-medium mb-4">Decompression Settings</h4>
             <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
               <div>
                 <label class="block text-sm font-medium text-gray-700 mb-1">Max Extract Size</label>
                  <Input
                    :model-value="editingConfig ? configChanges.MAX_EXTRACT_SIZE : config.MAX_EXTRACT_SIZE"
                    @input="updateConfigField('MAX_EXTRACT_SIZE', $event.target.value)"
                    :disabled="!editingConfig"
                    placeholder="1GB"
                  />
               </div>
               <div>
                 <label class="block text-sm font-medium text-gray-700 mb-1">Max Files per Archive</label>
                  <Input
                    :model-value="editingConfig ? configChanges.MAX_FILES_PER_ARCHIVE : config.MAX_FILES_PER_ARCHIVE"
                    @input="updateConfigField('MAX_FILES_PER_ARCHIVE', $event.target.value)"
                    :disabled="!editingConfig"
                    type="number"
                    placeholder="1000"
                  />
               </div>
               <div>
                 <label class="block text-sm font-medium text-gray-700 mb-1">Nested Archive Depth</label>
                  <Input
                    :model-value="editingConfig ? configChanges.NESTED_ARCHIVE_DEPTH : config.NESTED_ARCHIVE_DEPTH"
                    @input="updateConfigField('NESTED_ARCHIVE_DEPTH', $event.target.value)"
                    :disabled="!editingConfig"
                    type="number"
                    placeholder="3"
                  />
               </div>
               <div>
                 <label class="block text-sm font-medium text-gray-700 mb-1">Decompression Enabled</label>
                  <select
                    :model-value="editingConfig ? configChanges.DECOMPRESSION_ENABLED : config.DECOMPRESSION_ENABLED"
                    @change="updateConfigField('DECOMPRESSION_ENABLED', $event.target.value)"
                    :disabled="!editingConfig"
                    class="w-full px-3 py-2 border border-gray-300 rounded-md"
                  >
                   <option value="true">Enabled</option>
                   <option value="false">Disabled</option>
                 </select>
               </div>
               <div>
                 <label class="block text-sm font-medium text-gray-700 mb-1">Password Protected</label>
                  <select
                    :model-value="editingConfig ? configChanges.PASSWORD_PROTECTED : config.PASSWORD_PROTECTED"
                    @change="updateConfigField('PASSWORD_PROTECTED', $event.target.value)"
                    :disabled="!editingConfig"
                    class="w-full px-3 py-2 border border-gray-300 rounded-md"
                 >
                   <option value="true">Allowed</option>
                   <option value="false">Not Allowed</option>
                 </select>
               </div>
               <div>
                 <label class="block text-sm font-medium text-gray-700 mb-1">Extract to Subfolder</label>
                  <select
                    :model-value="editingConfig ? configChanges.EXTRACT_TO_SUBFOLDER : config.EXTRACT_TO_SUBFOLDER"
                    @change="updateConfigField('EXTRACT_TO_SUBFOLDER', $event.target.value)"
                    :disabled="!editingConfig"
                    class="w-full px-3 py-2 border border-gray-300 rounded-md"
                 >
                   <option value="true">Yes</option>
                   <option value="false">No</option>
                 </select>
               </div>
             </div>
           </div>
         </div>
       </CardContent>
     </Card>

     <!-- Worker Configuration -->
     <Card>
       <CardHeader>
         <CardTitle class="flex items-center">
           <Users class="mr-2 h-5 w-5" />
           Worker Pool Configuration
         </CardTitle>
         <CardDescription>
           Configure number of concurrent workers for job processing
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
                <Button @click="handleUpdateWorkers" :disabled="jobsLoading">
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
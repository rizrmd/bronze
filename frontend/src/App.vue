<script setup lang="ts">
import { onMounted, ref, provide, watch } from 'vue'
import { useRoute } from 'vue-router'
import AppLayout from '@/layouts/AppLayout.vue'
import { ToastContainer } from '@/components/ui/toast'
import { requestStore } from '@/stores/requestStore'

const toastContainer = ref()
const route = useRoute()

onMounted(() => {
  // Initialize app
})

// Watch for route changes and cancel browse requests when navigating away
watch(
  () => route.fullPath,
  (newPath, oldPath) => {
    // Cancel all browse requests when route changes, unless it's a back navigation
    // that goes to a parent directory (which should complete loading)
    
    if (oldPath && oldPath !== newPath) {
      // Try to detect if this is back navigation to parent
      const isBackToParent = oldPath && newPath && 
                           oldPath.startsWith(newPath) &&
                           newPath === oldPath.split('/').slice(0, -1).join('/')
      
      if (isBackToParent) {
        console.log(`Route changed from ${oldPath} to ${newPath} (back to parent), NOT canceling browse requests`)
      } else {
        console.log(`Route changed from ${oldPath} to ${newPath} (forward navigation), canceling all browse requests except ${newPath}`)
        // Cancel all requests EXCEPT those for the new target path
        requestStore.cancelAllExceptForUrl(newPath)
      }
    }
  }
)

// Make toast available globally
const toast = {
  success: (title: string, message?: string, duration?: number) => toastContainer.value?.success(title, message, duration),
  error: (title: string, message?: string, duration?: number) => toastContainer.value?.error(title, message, duration),
  warning: (title: string, message?: string, duration?: number) => toastContainer.value?.warning(title, message, duration),
  info: (title: string, message?: string, duration?: number) => toastContainer.value?.info(title, message, duration)
}

// Provide toast to all child components
provide('toast', toast)
</script>

<template>
  <AppLayout>
    <RouterView />
    <ToastContainer ref="toastContainer" />
  </AppLayout>
</template>

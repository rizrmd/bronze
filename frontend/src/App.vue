<script setup lang="ts">
import { onMounted, ref, provide } from 'vue'
import AppLayout from '@/layouts/AppLayout.vue'
import { ToastContainer } from '@/components/ui/toast'

const toastContainer = ref()

onMounted(() => {
  // Initialize router
})

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

<script setup lang="ts">
import { ref } from 'vue'
import Toast from './Toast.vue'

interface Toast {
  id: string
  type: 'success' | 'error' | 'warning' | 'info'
  title: string
  message?: string
  duration?: number
}

const toasts = ref<Toast[]>([])

const addToast = (toast: Omit<Toast, 'id'>) => {
  const id = Date.now().toString()
  const newToast = { ...toast, id }
  toasts.value.push(newToast)
  
  // Remove toast after duration
  if (toast.duration !== 0) {
    setTimeout(() => {
      removeToast(id)
    }, toast.duration || 5000)
  }
}

const removeToast = (id: string) => {
  const index = toasts.value.findIndex(toast => toast.id === id)
  if (index > -1) {
    toasts.value.splice(index, 1)
  }
}

const success = (title: string, message?: string, duration?: number) => {
  addToast({ type: 'success', title, message, duration })
}

const error = (title: string, message?: string, duration?: number) => {
  addToast({ type: 'error', title, message, duration })
}

const warning = (title: string, message?: string, duration?: number) => {
  addToast({ type: 'warning', title, message, duration })
}

const info = (title: string, message?: string, duration?: number) => {
  addToast({ type: 'info', title, message, duration })
}

defineExpose({
  success,
  error,
  warning,
  info,
  addToast,
  removeToast
})
</script>

<template>
  <Teleport to="body">
    <div class="fixed top-4 right-4 z-50 space-y-4">
      <Toast
        v-for="toast in toasts"
        :key="toast.id"
        :toast="toast"
        @close="removeToast"
      />
    </div>
  </Teleport>
</template>
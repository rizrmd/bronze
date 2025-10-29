<script setup lang="ts">
import { ref, computed } from 'vue'
import { X, CheckCircle, AlertCircle, Info, AlertTriangle } from 'lucide-vue-next'

interface Toast {
  id: string
  type: 'success' | 'error' | 'warning' | 'info'
  title: string
  message?: string
  duration?: number
}

const props = defineProps<{
  toast: Toast
}>()

const emit = defineEmits<{
  close: [id: string]
}>()

const isVisible = ref(true)

const icon = computed(() => {
  const icons = {
    success: CheckCircle,
    error: AlertCircle,
    warning: AlertTriangle,
    info: Info
  }
  return icons[props.toast.type]
})

const iconColor = computed(() => {
  const colors = {
    success: 'text-green-500',
    error: 'text-red-500',
    warning: 'text-yellow-500',
    info: 'text-blue-500'
  }
  return colors[props.toast.type]
})

const bgColor = computed(() => {
  const colors = {
    success: 'bg-green-50 border-green-200',
    error: 'bg-red-50 border-red-200',
    warning: 'bg-yellow-50 border-yellow-200',
    info: 'bg-blue-50 border-blue-200'
  }
  return colors[props.toast.type]
})

const titleColor = computed(() => {
  const colors = {
    success: 'text-green-800',
    error: 'text-red-800',
    warning: 'text-yellow-800',
    info: 'text-blue-800'
  }
  return colors[props.toast.type]
})

const messageColor = computed(() => {
  const colors = {
    success: 'text-green-700',
    error: 'text-red-700',
    warning: 'text-yellow-700',
    info: 'text-blue-700'
  }
  return colors[props.toast.type]
})

const duration = computed(() => props.toast.duration || 5000)

let timeoutId: number

const startTimer = () => {
  if (duration.value > 0) {
    timeoutId = setTimeout(() => {
      close()
    }, duration.value)
  }
}

const close = () => {
  isVisible.value = false
  setTimeout(() => {
    emit('close', props.toast.id)
  }, 300)
}

const clearTimer = () => {
  if (timeoutId) {
    clearTimeout(timeoutId)
  }
}

startTimer()

defineExpose({
  clearTimer,
  startTimer
})
</script>

<template>
  <Transition
    enter-active-class="transform ease-out duration-300 transition"
    enter-from-class="translate-y-2 opacity-0 sm:translate-y-0 sm:translate-x-2"
    enter-to-class="translate-y-0 opacity-100 sm:translate-x-0"
    leave-active-class="transition ease-in duration-100"
    leave-from-class="opacity-100"
    leave-to-class="opacity-0"
  >
    <div
      v-if="isVisible"
      :class="[
        'max-w-sm w-full border rounded-lg shadow-lg pointer-events-auto ring-1 ring-black ring-opacity-5 overflow-hidden',
        bgColor
      ]"
      @mouseenter="clearTimer"
      @mouseleave="startTimer"
    >
      <div class="p-4">
        <div class="flex items-start">
          <div class="flex-shrink-0">
            <component :is="icon" :class="['h-6 w-6', iconColor]" />
          </div>
          <div class="ml-3 w-0 flex-1">
            <p :class="['text-sm font-medium', titleColor]">
              {{ toast.title }}
            </p>
            <p v-if="toast.message" :class="['mt-1 text-sm', messageColor]">
              {{ toast.message }}
            </p>
          </div>
          <div class="ml-4 flex-shrink-0 flex">
            <button
              @click="close"
              :class="[
                'rounded-md inline-flex focus:outline-none focus:ring-2 focus:ring-offset-2',
                toast.type === 'success' ? 'bg-green-50 text-green-500 hover:bg-green-100 focus:ring-green-600' :
                toast.type === 'error' ? 'bg-red-50 text-red-500 hover:bg-red-100 focus:ring-red-600' :
                toast.type === 'warning' ? 'bg-yellow-50 text-yellow-500 hover:bg-yellow-100 focus:ring-yellow-600' :
                'bg-blue-50 text-blue-500 hover:bg-blue-100 focus:ring-blue-600'
              ]"
            >
              <span class="sr-only">Close</span>
              <X class="h-5 w-5" />
            </button>
          </div>
        </div>
      </div>
    </div>
  </Transition>
</template>
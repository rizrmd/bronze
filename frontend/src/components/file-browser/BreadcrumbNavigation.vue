<template>
  <nav class="flex items-center space-x-1 p-2 bg-gray-50 border-b text-sm">
    <!-- <button 
      @click="$emit('navigate', '')"
      class="px-2 py-1 rounded hover:bg-gray-200 transition-colors"
      :class="{ 'font-semibold': currentPath === '' }"
    >
      Home
    </button>
    
    <span class="text-gray-400">/</span> -->

    <template v-for="(path, index) in breadcrumbPaths" :key="path.path">
      <button @click="$emit('navigate', path.path)" class="px-2 py-1 cursor-pointer rounded hover:bg-gray-200 transition-colors"
        :class="{ 'font-semibold': path.path === currentPath }">
        <span v-if="index === 0">
          <House class="w-5" />
        </span>
        <span v-else>
          {{ path.name }}
        </span>
      </button>
      <span v-if="index < breadcrumbPaths.length - 1" class="text-gray-400">/</span>
    </template>
  </nav>
</template>

<script setup lang="ts">
import { House } from 'lucide-vue-next'


interface BreadcrumbPath {
  name: string
  path: string
}

interface Props {
  currentPath: string
  breadcrumbPaths: BreadcrumbPath[]
}

interface Emits {
  (e: 'navigate', path: string): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

// Use props to avoid TypeScript warning
console.log('props used:', props.currentPath)
</script>
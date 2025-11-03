<template>
  <div class="flex items-center justify-end space-x-1 p-2 border-b bg-gray-50">
    <Button
      @click="$emit('download')"
      variant="outline"
      size="sm"
      :disabled="!hasSelection"
    >
      <Download class="h-4 w-4" />
    </Button>
    
    <Button
      @click="$emit('delete')"
      variant="outline"
      size="sm"
      :disabled="!hasSelection"
      class="text-red-600 hover:text-red-700"
    >
      <Trash2 class="h-4 w-4" />
    </Button>
    
    <Button
      @click="showDropdown = !showDropdown"
      variant="outline"
      size="sm"
    >
      <MoreHorizontal class="h-4 w-4" />
    </Button>
    
    <!-- Dropdown menu -->
    <div
      v-if="showDropdown"
      v-click-outside="() => showDropdown = false"
      class="absolute right-2 top-12 bg-white border rounded-lg shadow-lg z-10 py-1 min-w-32"
    >
      <button
        @click="$emit('select-all'); showDropdown = false"
        class="w-full px-3 py-2 text-left text-sm hover:bg-gray-100 transition-colors"
      >
        Select All
      </button>
      
      <button
        @click="$emit('clear-selection'); showDropdown = false"
        class="w-full px-3 py-2 text-left text-sm hover:bg-gray-100 transition-colors"
      >
        Clear Selection
      </button>
      
      <div class="border-t my-1"></div>
      
      <button
        @click="$emit('copy'); showDropdown = false"
        class="w-full px-3 py-2 text-left text-sm hover:bg-gray-100 transition-colors"
      >
        Copy
      </button>
      
      <button
        @click="$emit('move'); showDropdown = false"
        class="w-full px-3 py-2 text-left text-sm hover:bg-gray-100 transition-colors"
      >
        Move
      </button>
      
      <button
        @click="$emit('rename'); showDropdown = false"
        class="w-full px-3 py-2 text-left text-sm hover:bg-gray-100 transition-colors"
      >
        Rename
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { Button } from '@/components/ui/button'
import { Download, Trash2, MoreHorizontal } from 'lucide-vue-next'

interface Props {
  hasSelection: boolean
}

interface Emits {
  (e: 'download'): void
  (e: 'delete'): void
  (e: 'select-all'): void
  (e: 'clear-selection'): void
  (e: 'copy'): void
  (e: 'move'): void
  (e: 'rename'): void
}

const props = defineProps<Props>()
defineEmits<Emits>()

const showDropdown = ref(false)

// v-click-outside directive
const vClickOutside = {
  mounted(el: HTMLElement, binding: any) {
    const clickOutside = (event: Event) => {
      if (!(el === event.target || el.contains(event.target as Node))) {
        binding.value(event)
      }
    }
    document.addEventListener('click', clickOutside)
    el._clickOutside = clickOutside
  },
  unmounted(el: HTMLElement) {
    if (el._clickOutside) {
      document.removeEventListener('click', el._clickOutside)
      delete el._clickOutside
    }
  }
}
</script>
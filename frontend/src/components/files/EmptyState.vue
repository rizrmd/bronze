<template>
  <div class="text-center p-8 border border-dashed border-border rounded-lg">
    <div class="text-muted-foreground">
      <Folder class="w-16 h-16 text-gray-400 mb-4 mx-auto" />
      <h3 class="text-lg font-medium mb-2">{{ title }}</h3>
      <p>{{ message }}</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { Folder } from 'lucide-vue-next'
interface Props {
  title?: string
  message?: string
}

withDefaults(defineProps<Props>(), {
  title: 'No Files Found',
  message: 'Upload some files to get started'
})
</script>
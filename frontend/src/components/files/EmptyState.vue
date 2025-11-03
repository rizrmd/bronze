<template>
  <div class="text-center p-8 border border-dashed border-border rounded-lg">
    <div class="text-muted-foreground">
      <div class="text-4xl mb-4">📁</div>
      <h3 class="text-lg font-medium mb-2">{{ title }}</h3>
      <p>{{ message }}</p>
    </div>
  </div>
</template>

<script setup lang="ts">
interface Props {
  title?: string
  message?: string
}

withDefaults(defineProps<Props>(), {
  title: 'No Files Found',
  message: 'Upload some files to get started'
})
</script>
<script setup lang="ts">
interface Props {
  class?: string
}

defineProps<Props>()
</script>

<template>
  <p :class="['text-sm text-gray-600', $props.class]">
    <slot />
  </p>
</template>
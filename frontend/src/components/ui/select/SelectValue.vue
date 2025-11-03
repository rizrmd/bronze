<script setup lang="ts">
import { inject, type Ref } from 'vue'

const modelValue = inject('modelValue') as Ref<string>
</script>

<template>
  <span>{{ modelValue }}</span>
</template>

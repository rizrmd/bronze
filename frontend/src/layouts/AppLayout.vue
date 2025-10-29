<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRoute } from 'vue-router'
import { 
  LayoutDashboard, 
  FolderOpen, 
  Briefcase, 
  Eye, 
  Settings,
  Menu,
  X
} from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'

const route = useRoute()
const sidebarOpen = ref(true)

const navigation = [
  {
    name: 'Dashboard',
    href: '/',
    icon: LayoutDashboard,
    current: computed(() => route.name === 'Dashboard')
  },
  {
    name: 'Files',
    href: '/files',
    icon: FolderOpen,
    current: computed(() => route.name === 'Files')
  },
  {
    name: 'Jobs',
    href: '/jobs',
    icon: Briefcase,
    current: computed(() => route.name === 'Jobs')
  },
  {
    name: 'Watcher',
    href: '/watcher',
    icon: Eye,
    current: computed(() => route.name === 'Watcher')
  },
  {
    name: 'Settings',
    href: '/settings',
    icon: Settings,
    current: computed(() => route.name === 'Settings')
  }
]

const toggleSidebar = () => {
  sidebarOpen.value = !sidebarOpen.value
}
</script>

<template>
  <div class="flex h-screen bg-gray-50">
    <!-- Sidebar -->
    <div :class="cn(
      'flex flex-col bg-white border-r border-gray-200 transition-all duration-300',
      sidebarOpen ? 'w-64' : 'w-16'
    )">
      <!-- Logo -->
      <div class="flex items-center justify-between h-16 px-4 border-b border-gray-200">
        <div v-if="sidebarOpen" class="flex items-center">
          <div class="w-8 h-8 bg-blue-600 rounded-lg flex items-center justify-center">
            <span class="text-white font-bold text-sm">B</span>
          </div>
          <span class="ml-2 text-xl font-semibold text-gray-900">Bronze</span>
        </div>
        <div v-else class="flex items-center justify-center w-full">
          <div class="w-8 h-8 bg-blue-600 rounded-lg flex items-center justify-center">
            <span class="text-white font-bold text-sm">B</span>
          </div>
        </div>
        <Button
          variant="ghost"
          size="sm"
          @click="toggleSidebar"
          class="p-1"
        >
          <Menu v-if="!sidebarOpen" class="w-4 h-4" />
          <X v-else class="w-4 h-4" />
        </Button>
      </div>

      <!-- Navigation -->
      <nav class="flex-1 px-2 py-4 space-y-1">
        <router-link
          v-for="item in navigation"
          :key="item.name"
          :to="item.href"
          :class="cn(
            'group flex items-center px-2 py-2 text-sm font-medium rounded-md transition-colors',
            item.current
              ? 'bg-blue-100 text-blue-700'
              : 'text-gray-600 hover:bg-gray-50 hover:text-gray-900'
          )"
        >
          <component
            :is="item.icon"
            :class="cn(
              'mr-3 flex-shrink-0 w-5 h-5',
              item.current ? 'text-blue-500' : 'text-gray-400 group-hover:text-gray-500'
            )"
          />
          <span v-if="sidebarOpen">{{ item.name }}</span>
        </router-link>
      </nav>
    </div>

    <!-- Main Content -->
    <div class="flex-1 flex flex-col overflow-hidden">
      <!-- Header -->
      <header class="bg-white border-b border-gray-200 px-6 py-4">
        <div class="flex items-center justify-between">
          <h1 class="text-2xl font-semibold text-gray-900">
            {{ route.meta.title || 'Bronze' }}
          </h1>
          <div class="flex items-center space-x-4">
            <!-- Add header actions here -->
          </div>
        </div>
      </header>

      <!-- Page Content -->
      <main class="flex-1 overflow-auto p-6">
        <router-view />
      </main>
    </div>
  </div>
</template>
import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    name: 'Dashboard',
    component: () => import('@/views/Dashboard.vue'),
    meta: { title: 'Dashboard' }
  },
  {
    path: '/files',
    name: 'Files',
    component: () => import('@/views/FilesManager.vue'),
    meta: { title: 'Files' }
  },
  {
    path: '/jobs',
    name: 'Jobs',
    component: () => import('@/views/JobsManager.vue'),
    meta: { title: 'Jobs' }
  },
  {
    path: '/watcher',
    name: 'Watcher',
    component: () => import('@/views/WatcherEvents.vue'),
    meta: { title: 'Watcher Events' }
  },
  {
    path: '/settings',
    name: 'Settings',
    component: () => import('@/views/Settings.vue'),
    meta: { title: 'Settings' }
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

router.beforeEach((to, _from, next) => {
  document.title = `${to.meta.title} - Bronze`
  next()
})

export default router
import { createRouter, createWebHistory } from 'vue-router'
import DashboardLayout from '@/layouts/DashboardLayout.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      component: DashboardLayout,
      redirect: '/machines',
      children: [
        { path: 'machines',          component: () => import('@/views/MachinesView.vue') },
        { path: 'machines/new',      component: () => import('@/views/MachineCreateView.vue') },
        { path: 'machines/:name',    component: () => import('@/views/MachineDetailView.vue') },
        { path: 'volumes',           component: () => import('@/views/VolumesView.vue') },
        { path: 'networks',          component: () => import('@/views/NetworksView.vue') },
        { path: 'virtualips',        component: () => import('@/views/VirtualIPsView.vue') },
        { path: 'loadbalancers',     component: () => import('@/views/LoadBalancersView.vue') },
        { path: 'prefixes',          component: () => import('@/views/IPPrefixesView.vue') },
      ]
    }
  ]
})

export default router

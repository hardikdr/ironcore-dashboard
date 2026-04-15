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
        { path: 'machines',              component: () => import('@/views/MachinesView.vue') },
        { path: 'machines/new',          component: () => import('@/views/MachineCreateView.vue') },
        { path: 'machines/:name',        component: () => import('@/views/MachineDetailView.vue') },
        { path: 'volumes',               component: () => import('@/views/VolumesView.vue') },
        { path: 'volumes/new',           component: () => import('@/views/VolumeCreateView.vue') },
        { path: 'volumes/:name',         component: () => import('@/views/VolumeDetailView.vue') },
        { path: 'networks',              component: () => import('@/views/NetworksView.vue') },
        { path: 'networks/new',          component: () => import('@/views/NetworkCreateView.vue') },
        { path: 'networks/:name',        component: () => import('@/views/NetworkDetailView.vue') },
        { path: 'virtualips',            component: () => import('@/views/VirtualIPsView.vue') },
        { path: 'virtualips/new',        component: () => import('@/views/VirtualIPCreateView.vue') },
        { path: 'virtualips/:name',      component: () => import('@/views/VirtualIPDetailView.vue') },
        { path: 'loadbalancers',         component: () => import('@/views/LoadBalancersView.vue') },
        { path: 'loadbalancers/new',     component: () => import('@/views/LoadBalancerCreateView.vue') },
        { path: 'loadbalancers/:name',   component: () => import('@/views/LoadBalancerDetailView.vue') },
        { path: 'prefixes',              component: () => import('@/views/IPPrefixesView.vue') },
        { path: 'prefixes/new',          component: () => import('@/views/IPPrefixCreateView.vue') },
        { path: 'prefixes/:name',        component: () => import('@/views/IPPrefixDetailView.vue') },
      ]
    }
  ]
})

export default router

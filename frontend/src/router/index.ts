import { createRouter, createWebHistory } from 'vue-router'
import DashboardLayout from '@/layouts/DashboardLayout.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      component: DashboardLayout,
      redirect: '/overview',
      children: [
        { path: 'overview',            component: () => import('@/views/OverviewView.vue') },
        { path: 'machines',            component: () => import('@/views/MachinesView.vue') },
        { path: 'machines/new',        component: () => import('@/views/MachineCreateView.vue') },
        { path: 'machines/:name',      component: () => import('@/views/MachineDetailView.vue') },
        { path: 'topology',            component: () => import('@/views/TopologyView.vue') },
        { path: 'servers',             component: () => import('@/views/ServersView.vue') },
        { path: 'servers/:name',       component: () => import('@/views/ServerDetailView.vue') },
        { path: 'volumes',             component: () => import('@/views/VolumesView.vue') },
        { path: 'volumes/new',         component: () => import('@/views/VolumeCreateView.vue') },
        { path: 'volumes/:name',       component: () => import('@/views/VolumeDetailView.vue') },
        { path: 'networks',            component: () => import('@/views/NetworksView.vue') },
        { path: 'networks/new',        component: () => import('@/views/NetworkCreateView.vue') },
        { path: 'networks/:name',      component: () => import('@/views/NetworkDetailView.vue') },
        { path: 'virtualips',          component: () => import('@/views/VirtualIPsView.vue') },
        { path: 'virtualips/new',      component: () => import('@/views/VirtualIPCreateView.vue') },
        { path: 'virtualips/:name',    component: () => import('@/views/VirtualIPDetailView.vue') },
        { path: 'loadbalancers',       component: () => import('@/views/LoadBalancersView.vue') },
        { path: 'loadbalancers/new',   component: () => import('@/views/LoadBalancerCreateView.vue') },
        { path: 'loadbalancers/:name', component: () => import('@/views/LoadBalancerDetailView.vue') },
        { path: 'prefixes',            component: () => import('@/views/IPPrefixesView.vue') },
        { path: 'prefixes/new',        component: () => import('@/views/IPPrefixCreateView.vue') },
        { path: 'prefixes/:name',      component: () => import('@/views/IPPrefixDetailView.vue') },
        // Phase 3 — new resource views
        { path: 'machinepools',        component: () => import('@/views/MachinePoolsView.vue') },
        { path: 'volumesnapshots',     component: () => import('@/views/VolumeSnapshotsView.vue') },
        { path: 'natgateways',         component: () => import('@/views/NATGatewaysView.vue') },
        { path: 'resourcequotas',      component: () => import('@/views/ResourceQuotaView.vue') },
        { path: 'settings',            component: () => import('@/views/SettingsView.vue') },
        { path: ':pathMatch(.*)*',     component: () => import('@/views/NotFoundView.vue') },
      ]
    }
  ]
})

export default router

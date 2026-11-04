import { createRouter, createWebHistory } from 'vue-router'
import DashboardLayout from '@/layouts/DashboardLayout.vue'
import SreDashboardLayout from '@/layouts/SreDashboardLayout.vue'

const sreMeta = (title: string) => ({ console: 'sre', readOnly: true, title })

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/sre',
      component: SreDashboardLayout,
      redirect: '/sre/landscape',
      children: [
        { path: 'landscape',       component: () => import('@/views/sre/LandscapeView.vue'), meta: sreMeta('Landscape') },
        { path: 'triage',          component: () => import('@/views/sre/TriageView.vue'), meta: sreMeta('Triage') },
        { path: 'zones',           component: () => import('@/views/sre/ZonesView.vue'), meta: sreMeta('Zones') },
        { path: 'zones/:zone',     component: () => import('@/views/sre/ZoneDetailView.vue'), meta: sreMeta('Zone') },
        { path: 'nodes/:id',       component: () => import('@/views/sre/PhysicalNodeDetailView.vue'), meta: sreMeta('Node') },
        { path: 'shoots',          component: () => import('@/views/sre/ShootsView.vue'), meta: sreMeta('Shoots') },
        { path: 'shoots/:id',      component: () => import('@/views/sre/ShootDetailView.vue'), meta: sreMeta('Shoot') },
        { path: 'vms',             component: () => import('@/views/sre/VmsView.vue'), meta: sreMeta('Virtual Machines') },
        { path: 'vms/:id',         component: () => import('@/views/sre/VmDetailView.vue'), meta: sreMeta('Virtual Machine') },
        { path: 'dependencies/:id',component: () => import('@/views/sre/DependencyExplorerView.vue'), meta: sreMeta('Dependencies') },
        { path: 'networking',      component: () => import('@/views/sre/NetworkingSreView.vue'), meta: sreMeta('Networking') },
        { path: 'loadbalancers/:id',component: () => import('@/views/sre/LoadBalancerSreDetailView.vue'), meta: sreMeta('Load Balancer') },
        { path: 'storage',         component: () => import('@/views/sre/StorageSreView.vue'), meta: sreMeta('Storage') },
        { path: 'volumes/:id',     component: () => import('@/views/sre/VolumeSreDetailView.vue'), meta: sreMeta('Volume') },
        { path: 'sources',         component: () => import('@/views/sre/SourcesView.vue'), meta: sreMeta('Sources') },
      ],
    },
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
        { path: 'settings',            component: () => import('@/views/SettingsView.vue') },
        { path: ':pathMatch(.*)*',     component: () => import('@/views/NotFoundView.vue') },
      ]
    }
  ]
})

export default router

import { createRouter, createWebHistory } from 'vue-router'
import DashboardLayout from '@/layouts/DashboardLayout.vue'

const sreView = () => import('@/views/sre/SrePlaceholderView.vue')
const sreMeta = (title: string) => ({ console: 'sre', readOnly: true, title })

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
        { path: 'settings',            component: () => import('@/views/SettingsView.vue') },
        { path: 'sre/landscape',       component: () => import('@/views/sre/LandscapeView.vue'), meta: sreMeta('Landscape') },
        { path: 'sre/triage',          component: () => import('@/views/sre/TriageView.vue'), meta: sreMeta('Triage') },
        { path: 'sre/zones',           component: () => import('@/views/sre/ZonesView.vue'), meta: sreMeta('Zones') },
        { path: 'sre/zones/:zone',     component: () => import('@/views/sre/ZoneDetailView.vue'), meta: sreMeta('Zone') },
        { path: 'sre/nodes/:id',       component: () => import('@/views/sre/PhysicalNodeDetailView.vue'), meta: sreMeta('Node') },
        { path: 'sre/shoots',          component: () => import('@/views/sre/ShootsView.vue'), meta: sreMeta('Shoots') },
        { path: 'sre/shoots/:id',      component: () => import('@/views/sre/ShootDetailView.vue'), meta: sreMeta('Shoot') },
        { path: 'sre/vms',             component: () => import('@/views/sre/VmsView.vue'), meta: sreMeta('Virtual Machines') },
        { path: 'sre/vms/:id',         component: () => import('@/views/sre/VmDetailView.vue'), meta: sreMeta('Virtual Machine') },
        { path: 'sre/dependencies/:id',component: sreView, meta: sreMeta('Dependencies') },
        { path: 'sre/networking',      component: sreView, meta: sreMeta('Networking') },
        { path: 'sre/loadbalancers/:id',component: sreView, meta: sreMeta('Load Balancer') },
        { path: 'sre/storage',         component: sreView, meta: sreMeta('Storage') },
        { path: 'sre/volumes/:id',     component: sreView, meta: sreMeta('Volume') },
        { path: 'sre/sources',         component: () => import('@/views/sre/SourcesView.vue'), meta: sreMeta('Sources') },
        { path: ':pathMatch(.*)*',     component: () => import('@/views/NotFoundView.vue') },
      ]
    }
  ]
})

export default router

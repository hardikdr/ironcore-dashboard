import { describe, expect, it } from 'vitest'
import router from './index'

const expectedUserRoutes = [
  '/overview', '/machines', '/machines/new', '/machines/:name', '/topology', '/servers', '/servers/:name',
  '/volumes', '/volumes/new', '/volumes/:name', '/networks', '/networks/new', '/networks/:name',
  '/virtualips', '/virtualips/new', '/virtualips/:name', '/loadbalancers', '/loadbalancers/new',
  '/loadbalancers/:name', '/prefixes', '/prefixes/new', '/prefixes/:name', '/settings',
]

const expectedSreRoutes = [
  '/sre/landscape', '/sre/triage', '/sre/zones', '/sre/zones/:zone', '/sre/nodes/:id',
  '/sre/shoots', '/sre/shoots/:id', '/sre/vms', '/sre/vms/:id', '/sre/dependencies/:id',
  '/sre/networking', '/sre/loadbalancers/:id', '/sre/storage', '/sre/volumes/:id', '/sre/sources',
]

describe('dashboard routes', () => {
  it('preserves the complete user console route contract', () => {
    const paths = router.getRoutes().map(route => route.path)
    expect(expectedUserRoutes.every(path => paths.includes(path))).toBe(true)
  })

  it('registers every SRE deep link as read-only', () => {
    const routes = router.getRoutes()
    expect(expectedSreRoutes.every(path => routes.some(route => route.path === path))).toBe(true)
    for (const path of expectedSreRoutes) {
      expect(routes.find(route => route.path === path)?.meta).toMatchObject({ console: 'sre', readOnly: true })
    }
  })
})

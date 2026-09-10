import { test } from 'node:test'
import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { config } from '../config.mjs'

const { routes } = JSON.parse(readFileSync(config.routesPath, 'utf8'))

test('the inventory holds all 58 routes the design pinned', () => {
  assert.equal(routes.length, 58)
})

test('route paths are unique', () => {
  const seen = new Set(routes.map(r => r.path))
  assert.equal(seen.size, routes.length)
})

test('every dynamic route names the fixture that supplies its id', () => {
  // Exception: /app/admin/companies/$id uses company 1, the audit's own
  // throwaway tenant, rather than a fixture row — so its needsFixture is
  // correctly null. See task-11-brief.md's note under the route inventory.
  const EXEMPT = ['/app/admin/companies/$id']
  for (const r of routes) {
    if (r.path.includes('$id') && !EXEMPT.includes(r.path)) {
      assert.ok(r.needsFixture, `${r.path} does not name a fixture`)
    }
  }
})

test('the three public routes are marked public', () => {
  const publicPaths = routes.filter(r => r.kind === 'public').map(r => r.path).sort()
  assert.deepEqual(publicPaths, ['/login', '/onboarding/company', '/signup'])
})

test('every route path starts with a slash', () => {
  for (const r of routes) assert.match(r.path, /^\//)
})

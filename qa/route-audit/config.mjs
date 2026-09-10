import { fileURLToPath } from 'node:url'
import { dirname, resolve } from 'node:path'

const here = dirname(fileURLToPath(import.meta.url))

export const config = {
  baseUrl: process.env.AUDIT_BASE_URL ?? 'https://go-logistics.jesuslab135.com',
  email: process.env.AUDIT_EMAIL ?? '',
  password: process.env.AUDIT_PASSWORD ?? '',
  specPath: resolve(here, '../../docs/swagger.json'),
  captureDir: resolve(here, 'ui/captures'),
  routesPath: resolve(here, 'ui/routes.json'),
}

export function requireCredentials() {
  if (!config.email || !config.password) {
    throw new Error(
      'AUDIT_EMAIL and AUDIT_PASSWORD must be set. ' +
      'They are read from the environment so they never enter the repository.'
    )
  }
}

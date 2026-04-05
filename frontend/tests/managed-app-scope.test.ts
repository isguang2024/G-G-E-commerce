import assert from 'node:assert/strict'

import {
  normalizeManagedAppKey,
  resolveManagedAppKey
} from '../src/hooks/business/managed-app-scope'

assert.equal(normalizeManagedAppKey('  PLATFORM-ADMIN '), 'platform-admin')
assert.equal(resolveManagedAppKey('  platform-admin ', ''), 'platform-admin')
assert.equal(resolveManagedAppKey('', 'merchant-console'), 'merchant-console')
assert.equal(resolveManagedAppKey('', ''), '')

import test from 'node:test'
import assert from 'node:assert/strict'
import { createPinia, setActivePinia } from 'pinia'
import { useAppContextStore } from '../src/store/modules/app-context'

test('managed app does not fall back to runtime app automatically', () => {
  setActivePinia(createPinia())
  const store = useAppContextStore()

  store.setRuntimeAppKey('platform-admin')

  assert.equal(store.runtimeAppKey, 'platform-admin')
  assert.equal(store.managedAppKey, '')
  assert.equal(store.ensureManagedAppKey(), '')
  assert.equal(store.ensureManagedAppKey('merchant-console'), '')
})


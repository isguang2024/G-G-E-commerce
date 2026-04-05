import test from 'node:test'
import assert from 'node:assert/strict'
import { createPinia, setActivePinia } from 'pinia'
import { useAppContextStore } from '../src/store/modules/app-context'

test('ensureManagedAppKey 不会再把运行时 app 自动提升为管理态 app', () => {
  setActivePinia(createPinia())
  const store = useAppContextStore()

  store.setRuntimeAppKey('platform-admin')

  assert.equal(store.runtimeAppKey, 'platform-admin')
  assert.equal(store.managedAppKey, '')
  assert.equal(store.effectiveManagedAppKey, '')
  assert.equal(store.ensureManagedAppKey(), '')
  assert.equal(store.managedAppKey, '')
})

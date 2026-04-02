export function readJsonStorage<T>(storage: Storage, key: string): T | null {
  const rawValue = storage.getItem(key)
  if (!rawValue) {
    return null
  }

  try {
    return JSON.parse(rawValue) as T
  } catch {
    return null
  }
}

export function writeJsonStorage(storage: Storage, key: string, value: unknown) {
  storage.setItem(key, JSON.stringify(value))
}

export function removeStorageKeys(storage: Storage, keys: string[]) {
  keys.forEach((key) => storage.removeItem(key))
}

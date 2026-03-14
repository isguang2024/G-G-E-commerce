/**
 * Pinia Store 閰嶇疆妯″潡
 *
 * 鎻愪緵鍏ㄥ眬鐘舵€佺鐞嗙殑鍒濆鍖栧拰閰嶇疆
 *
 * ## 涓昏鍔熻兘
 *
 * - Pinia Store 瀹炰緥鍒涘缓
 * - 鎸佷箙鍖栨彃浠堕厤缃紙pinia-plugin-persistedstate锛?
 * - 鐗堟湰鍖栧瓨鍌ㄩ敭绠＄悊
 * - 鑷姩鏁版嵁杩佺Щ锛堣法鐗堟湰锛?
 * - LocalStorage 搴忓垪鍖栭厤缃?
 * - Store 鍒濆鍖栧嚱鏁?
 *
 * ## 鎸佷箙鍖栫瓥鐣?
 *
 * - 浣跨敤 StorageKeyManager 鐢熸垚鐗堟湰鍖栫殑瀛樺偍閿?
 * - 鏍煎紡锛歴ys-v{version}-{storeId}
 * - 鑷姩杩佺Щ鏃х増鏈暟鎹埌褰撳墠鐗堟湰
 * - 浣跨敤 localStorage 浣滀负瀛樺偍浠嬭川
 *
 * @module store/index
 * @author Art Design Pro Team
 */
import type { App } from 'vue'
import { createPinia } from 'pinia'
import { createPersistedState } from 'pinia-plugin-persistedstate'
import { StorageKeyManager } from '@/utils/storage/storage-key-manager'

export const store = createPinia()

// 鍒涘缓瀛樺偍閿鐞嗗櫒瀹炰緥
const storageKeyManager = new StorageKeyManager()

// 閰嶇疆鎸佷箙鍖栨彃浠?
store.use(
  createPersistedState({
    key: (storeId: string) => storageKeyManager.getStorageKey(storeId),
    storage: localStorage,
    serializer: {
      serialize: JSON.stringify,
      deserialize: JSON.parse
    }
  })
)

/**
 * 鍒濆鍖?Store
 */
export function initStore(app: App<Element>): void {
  app.use(store)
}

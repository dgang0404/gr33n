/**
 * Phase 100 — IndexedDB cache for crop picker and domain enums (LAN / offline).
 */

const DB_NAME = 'gr33n_platform_cache'
const DB_VERSION = 1
const PICKER_STORE = 'crop_picker'
const ENUMS_STORE = 'domain_enums'
const LAST_VERSION_KEY = 'gr33n_last_catalog_version'

/** @type {Map<string, object>} */
const memoryFallback = new Map()

let dbPromise = /** @type {Promise<IDBDatabase|null>|null} */ (null)

function hasIndexedDB() {
  return typeof indexedDB !== 'undefined'
}

function openDB() {
  if (!hasIndexedDB()) return Promise.resolve(null)
  if (!dbPromise) {
    dbPromise = new Promise((resolve, reject) => {
      const req = indexedDB.open(DB_NAME, DB_VERSION)
      req.onerror = () => reject(req.error)
      req.onupgradeneeded = () => {
        const db = req.result
        if (!db.objectStoreNames.contains(PICKER_STORE)) {
          db.createObjectStore(PICKER_STORE, { keyPath: 'key' })
        }
        if (!db.objectStoreNames.contains(ENUMS_STORE)) {
          db.createObjectStore(ENUMS_STORE, { keyPath: 'key' })
        }
      }
      req.onsuccess = () => resolve(req.result)
    }).catch(() => null)
  }
  return dbPromise
}

/**
 * @param {string} storeName
 * @param {string} key
 */
async function idbGet(storeName, key) {
  const db = await openDB()
  if (!db) return memoryFallback.get(`${storeName}:${key}`) ?? null
  return new Promise((resolve, reject) => {
    const tx = db.transaction(storeName, 'readonly')
    const req = tx.objectStore(storeName).get(key)
    req.onsuccess = () => resolve(req.result ?? null)
    req.onerror = () => reject(req.error)
  })
}

/**
 * @param {string} storeName
 * @param {object} record
 */
async function idbPut(storeName, record) {
  const db = await openDB()
  if (!db) {
    memoryFallback.set(`${storeName}:${record.key}`, record)
    return
  }
  return new Promise((resolve, reject) => {
    const tx = db.transaction(storeName, 'readwrite')
    tx.objectStore(storeName).put(record)
    tx.oncomplete = () => resolve()
    tx.onerror = () => reject(tx.error)
  })
}

function pickerKey(farmId) {
  return `farm_${farmId}`
}

/**
 * @param {number} farmId
 * @param {object} picker
 */
export async function cacheCropPicker(farmId, picker) {
  if (!picker?.groups) return
  const catalogVersion = Number(picker.version ?? picker.catalog_version ?? 0)
  await idbPut(PICKER_STORE, {
    key: pickerKey(farmId),
    farm_id: farmId,
    catalog_version: catalogVersion,
    fetched_at: new Date().toISOString(),
    picker: {
      version: picker.version,
      counts: picker.counts,
      groups: picker.groups,
    },
  })
  if (catalogVersion > 0 && typeof localStorage !== 'undefined') {
    const prev = Number(localStorage.getItem(LAST_VERSION_KEY) || 0)
    if (catalogVersion >= prev) {
      localStorage.setItem(LAST_VERSION_KEY, String(catalogVersion))
    }
  }
}

/**
 * @param {number} farmId
 */
export async function getCachedCropPicker(farmId) {
  const row = await idbGet(PICKER_STORE, pickerKey(farmId))
  if (!row?.picker) return null
  return {
    ...row.picker,
    _cacheMeta: {
      farm_id: row.farm_id,
      catalog_version: row.catalog_version,
      fetched_at: row.fetched_at,
    },
  }
}

/**
 * @param {object} enums
 */
export async function cacheDomainEnums(enums) {
  if (!enums) return
  await idbPut(ENUMS_STORE, {
    key: 'platform',
    fetched_at: new Date().toISOString(),
    enums,
  })
}

export async function getCachedDomainEnums() {
  const row = await idbGet(ENUMS_STORE, 'platform')
  if (!row?.enums) return null
  return {
    enums: row.enums,
    fetched_at: row.fetched_at,
  }
}

/** @param {number|null|undefined} cachedVersion */
export function isStaleCatalogVersion(cachedVersion) {
  if (cachedVersion == null || typeof localStorage === 'undefined') return false
  const known = Number(localStorage.getItem(LAST_VERSION_KEY))
  if (!Number.isFinite(known) || known <= 0) return false
  return Number(cachedVersion) < known
}

/** @param {string|null|undefined} iso */
export function formatCacheDate(iso) {
  if (!iso) return 'unknown date'
  try {
    return new Date(iso).toLocaleString()
  } catch {
    return iso
  }
}

/** Axios-style errors with no response are network failures (offline, timeout). */
export function isNetworkError(err) {
  if (!err) return false
  if (err.code === 'ERR_NETWORK' || err.message === 'Network Error') return true
  return !err.response
}

/** Test helper — clears IndexedDB and memory fallback. */
export async function clearCatalogCache() {
  memoryFallback.clear()
  if (typeof localStorage !== 'undefined') {
    localStorage.removeItem(LAST_VERSION_KEY)
  }
  if (!hasIndexedDB()) return
  const db = await openDB()
  if (!db) return
  await Promise.all([
    new Promise((resolve, reject) => {
      const tx = db.transaction(PICKER_STORE, 'readwrite')
      tx.objectStore(PICKER_STORE).clear()
      tx.oncomplete = () => resolve()
      tx.onerror = () => reject(tx.error)
    }),
    new Promise((resolve, reject) => {
      const tx = db.transaction(ENUMS_STORE, 'readwrite')
      tx.objectStore(ENUMS_STORE).clear()
      tx.oncomplete = () => resolve()
      tx.onerror = () => reject(tx.error)
    }),
  ])
}

/** @internal */
export function resetCatalogCacheDB() {
  dbPromise = null
}

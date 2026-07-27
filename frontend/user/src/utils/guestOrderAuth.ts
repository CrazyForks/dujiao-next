export interface GuestOrderAuth {
  email: string
  order_password: string
}

const GUEST_ORDER_AUTH_KEY = 'guest_order_auth'
const EMPTY_GUEST_ORDER_AUTH: GuestOrderAuth = {
  email: '',
  order_password: '',
}

const parseGuestOrderAuth = (raw: string | null): GuestOrderAuth | null => {
  if (!raw) return null
  try {
    const parsed = JSON.parse(raw) as Partial<GuestOrderAuth>
    return {
      email: typeof parsed.email === 'string' ? parsed.email : '',
      order_password: typeof parsed.order_password === 'string' ? parsed.order_password : '',
    }
  } catch {
    return null
  }
}

const readStorage = (storage: Storage, key: string): string | null => {
  try {
    return storage.getItem(key)
  } catch {
    return null
  }
}

const removeStorage = (storage: Storage, key: string) => {
  try {
    storage.removeItem(key)
  } catch {
    // 浏览器禁用存储时只保留当前页面内存状态。
  }
}

// loadGuestOrderAuth 优先读取当前标签页的 sessionStorage，并执行一次旧版
// localStorage -> sessionStorage 迁移。迁移后立即删除长期存储中的游客凭据。
export const loadGuestOrderAuth = (): GuestOrderAuth => {
  if (typeof window === 'undefined') {
    return { ...EMPTY_GUEST_ORDER_AUTH }
  }

  const sessionRaw = readStorage(window.sessionStorage, GUEST_ORDER_AUTH_KEY)
  const legacyRaw = readStorage(window.localStorage, GUEST_ORDER_AUTH_KEY)
  const sessionAuth = parseGuestOrderAuth(sessionRaw)
  const legacyAuth = parseGuestOrderAuth(legacyRaw)
  const parsed = sessionAuth || legacyAuth

  if (!sessionAuth && legacyAuth) {
    saveGuestOrderAuth(legacyAuth)
  }
  if (legacyRaw) {
    removeStorage(window.localStorage, GUEST_ORDER_AUTH_KEY)
  }
  return parsed || { ...EMPTY_GUEST_ORDER_AUTH }
}

export const saveGuestOrderAuth = (auth: GuestOrderAuth) => {
  if (typeof window === 'undefined') return
  const normalized: GuestOrderAuth = {
    email: String(auth.email || ''),
    order_password: String(auth.order_password || ''),
  }
  try {
    window.sessionStorage.setItem(GUEST_ORDER_AUTH_KEY, JSON.stringify(normalized))
  } catch {
    // 不回退到 localStorage，避免把游客订单凭据重新变成长生命周期数据。
  }
  removeStorage(window.localStorage, GUEST_ORDER_AUTH_KEY)
}

export const clearGuestOrderAuth = () => {
  if (typeof window === 'undefined') return
  removeStorage(window.sessionStorage, GUEST_ORDER_AUTH_KEY)
  removeStorage(window.localStorage, GUEST_ORDER_AUTH_KEY)
}

function getWebApp() {
  return window.WebApp || null
}

/** Читает window.WebApp (MAX SDK): initData для бэкенда, initDataUnsafe только для подсказок в UI. */
export function getMaxBridgeLaunchData() {
  const webApp = getWebApp()
  if (!webApp) {
    return {
      isAvailable: false,
      initData: '',
      initDataUnsafe: null,
      platform: ''
    }
  }

  return {
    isAvailable: true,
    initData: typeof webApp.initData === 'string' ? webApp.initData : '',
    initDataUnsafe: webApp.initDataUnsafe || null,
    platform: typeof webApp.platform === 'string' ? webApp.platform : ''
  }
}

/** Платформа MAX WebView (android / ios / desktop / web). */
export function getMaxWebAppPlatform() {
  const platform = getWebApp()?.platform
  return typeof platform === 'string' ? platform.toLowerCase() : ''
}

/** start_param из MAX (deep link из кнопки open_app или ?startapp= в DEV). */
export function getMaxStartParam() {
  // DEV в обычном браузере: query важнее bridge — эмуляция open_app (documentation/local_development.md).
  if (import.meta.env.DEV && typeof window !== 'undefined') {
    const params = new URLSearchParams(window.location.search)
    const fromQuery = params.get('startapp') || params.get('start_param')
    if (typeof fromQuery === 'string' && fromQuery.trim() !== '') {
      return fromQuery.trim()
    }
  }

  const fromBridge = getWebApp()?.initDataUnsafe?.start_param
  if (typeof fromBridge === 'string' && fromBridge.trim() !== '') {
    return fromBridge.trim()
  }

  return ''
}

/**
 * Resolves the entity a MAX open_app button asks the Mini App to display.
 * Notification payloads use ticket_<number> or approval_<vote_number>.
 * MAX start_param accepts only A-Z, a-z, 0-9, underscore, and hyphen.
 */
export function parseStartTargetFromStartParam(startParam) {
  const raw = String(startParam || '').trim()
  if (!raw) {
    return null
  }

  // approval_ must be recognized before the legacy ticket fallback so it never becomes a ticket number.
  const approval = raw.match(/^approval_(.+)$/i)
  if (approval?.[1]?.trim()) {
    return { type: 'approval', number: approval[1].trim() }
  }

  // Keep supporting historical ticket separators and bare ticket numbers used by existing notifications.
  const ticket = raw.match(/^ticket[_:=-](.+)$/i)
  if (ticket?.[1]?.trim()) {
    return { type: 'ticket', number: ticket[1].trim() }
  }

  return { type: 'ticket', number: raw }
}

/**
 * Настройка оболочки MAX: ready(), отключение вертикального свайпа (конфликт с tap на Android).
 * В обычном браузере (127.0.0.1) SDK загружен, но транспорт MAX недоступен — не вызываем ready().
 */
export function configureMaxWebApp() {
  const webApp = getWebApp()
  if (!webApp) {
    return
  }

  const initData = typeof webApp.initData === 'string' ? webApp.initData.trim() : ''
  const platform = typeof webApp.platform === 'string' ? webApp.platform.trim() : ''
  if (!initData && !platform) {
    return
  }

  const callMaybeAsync = (fn) => {
    try {
      const result = fn()
      if (result && typeof result.catch === 'function') {
        result.catch(() => {})
      }
    } catch {
      // Вне WebView MAX ready()/disableVerticalSwipes могут падать — UI не блокируем.
    }
  }

  if (typeof webApp.ready === 'function') {
    callMaybeAsync(() => webApp.ready())
  }

  if (typeof webApp.disableVerticalSwipes === 'function') {
    callMaybeAsync(() => webApp.disableVerticalSwipes())
  }
}

/** @deprecated Используйте configureMaxWebApp — оставлено для совместимости вызовов. */
export function notifyMaxAppReady() {
  configureMaxWebApp()
}

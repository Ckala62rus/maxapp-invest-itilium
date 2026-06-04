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

/**
 * Настройка оболочки MAX: ready(), отключение вертикального свайпа (конфликт с tap на Android).
 */
export function configureMaxWebApp() {
  const webApp = getWebApp()
  if (!webApp) {
    return
  }

  if (typeof webApp.ready === 'function') {
    webApp.ready()
  }

  if (typeof webApp.disableVerticalSwipes === 'function') {
    try {
      webApp.disableVerticalSwipes()
    } catch {
      // На части сборок MAX метод может отсутствовать — не блокируем UI.
    }
  }
}

/** @deprecated Используйте configureMaxWebApp — оставлено для совместимости вызовов. */
export function notifyMaxAppReady() {
  configureMaxWebApp()
}

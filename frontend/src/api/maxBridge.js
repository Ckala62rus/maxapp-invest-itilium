function getWebApp() {
  return window.WebApp || null
}

// getMaxBridgeLaunchData reads MAX bridge launch data without trusting it by itself.
export function getMaxBridgeLaunchData() {
  const webApp = getWebApp()
  if (!webApp) {
    return {
      isAvailable: false,
      initData: '',
      initDataUnsafe: null
    }
  }

  return {
    isAvailable: true,
    initData: typeof webApp.initData === 'string' ? webApp.initData : '',
    initDataUnsafe: webApp.initDataUnsafe || null
  }
}

// notifyMaxAppReady tells the MAX client that the mini app finished loading.
export function notifyMaxAppReady() {
  const webApp = getWebApp()
  if (webApp && typeof webApp.ready === 'function') {
    webApp.ready()
  }
}

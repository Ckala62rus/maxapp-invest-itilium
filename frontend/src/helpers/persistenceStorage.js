/** Обёртка над localStorage: access_token и прочие клиентские данные. */
export function getItem(key) {
  return window.localStorage.getItem(key)
}

export function setItem(key, value) {
  window.localStorage.setItem(key, value)
}

export function removeItem(key) {
  window.localStorage.removeItem(key)
}

/** sessionStorage для экрана навигации (переживает Vite HMR / пересборку компонента). */
export function getSessionItem(key) {
  return window.sessionStorage.getItem(key)
}

export function setSessionItem(key, value) {
  window.sessionStorage.setItem(key, value)
}

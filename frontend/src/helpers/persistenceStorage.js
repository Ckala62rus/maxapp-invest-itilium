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

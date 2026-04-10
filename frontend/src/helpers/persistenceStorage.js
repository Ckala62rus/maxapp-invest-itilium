// getItem reads a value from localStorage and keeps storage access in one place.
export function getItem(key) {
  return window.localStorage.getItem(key)
}

// setItem stores a value in localStorage for auth and UI persistence flows.
export function setItem(key, value) {
  window.localStorage.setItem(key, value)
}

// removeItem deletes a persisted value from localStorage.
export function removeItem(key) {
  window.localStorage.removeItem(key)
}

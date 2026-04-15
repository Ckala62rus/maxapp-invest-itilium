import axios from 'axios'

import { getItem, removeItem } from '@/helpers/persistenceStorage'

const env = import.meta.env
const debugUserId = env.VITE_DEBUG_USER_ID || ''

// Keep API calls relative by default so Vite/nginx can proxy them through the
// same public origin used by MAX and tunnel providers.
axios.defaults.baseURL = env.VITE_PUBLIC_API_BASE_URL || ''

// Every request reads the current token from storage to keep auth state centralized.
axios.interceptors.request.use((config) => {
  const token = getItem('access_token')

  config.headers = config.headers || {}
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
    delete config.headers['X-User-ID']
  } else {
    config.headers.Authorization = ''
    if (env.DEV && debugUserId) {
      config.headers['X-User-ID'] = debugUserId
    }
  }

  return config
})

// A shared response interceptor keeps auth redirect behavior identical across modules.
axios.interceptors.response.use(undefined, (error) => {
  const location = window.location.pathname
  const status = error?.response?.status

  if (status === 401) {
    removeItem('access_token')
    delete axios.defaults.headers.common.Authorization
  }

  return Promise.reject(error)
})

export default axios

import axios from 'axios'

import { getItem, removeItem } from '@/helpers/persistenceStorage'

const env = import.meta.env
const debugUserId = env.VITE_DEBUG_USER_ID || ''

// По умолчанию ходим напрямую в локальный backend без Vite proxy.
axios.defaults.baseURL = env.VITE_PUBLIC_API_BASE_URL || 'http://127.0.0.1:3000'

// Перед каждым запросом подмешиваем токен из storage; в DEV без токена — опционально X-User-ID для отладки.
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

// 401: сбрасываем токен, чтобы следующий запрос не слал просроченный Bearer.
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

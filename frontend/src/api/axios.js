import axios from 'axios'

import { getItem, removeItem } from '@/helpers/persistenceStorage'

const env = import.meta.env

// The frontend runs on Vite, so the backend base URL is taken from VITE_BACKEND_API.
axios.defaults.baseURL = env.VITE_BACKEND_API || ''

// Every request reads the current token from storage to keep auth state centralized.
axios.interceptors.request.use((config) => {
  const token = getItem('access_token')

  config.headers = config.headers || {}
  config.headers.Authorization = token ? `Bearer ${token}` : ''

  return config
})

// A shared response interceptor keeps auth redirect behavior identical across modules.
axios.interceptors.response.use(undefined, (error) => {
  const location = window.location.pathname
  const status = error?.response?.status

  if (status === 401) {
    if (location === '/password-reset') {
      return Promise.reject(error)
    }

    if (location !== '/login') {
      removeItem('access_token')
      delete axios.defaults.headers.common.Authorization
      window.location.href = '/login'
    }
  }

  return Promise.reject(error)
})

export default axios

import axios from '@/api/axios'
import urls from '@/api/urls'

// validateMaxAuth exchanges MAX initData for a backend access token.
const validateMaxAuth = (payload) => {
  return axios.post(urls.maxAuthValidate, payload)
}

// registration sends the MAX user registration form to the backend.
const registration = (credential) => {
  return axios.post(urls.registration, credential)
}

// me loads the current user profile that the backend resolves for the mini app.
const me = () => {
  return axios.get(urls.me)
}

export default {
  validateMaxAuth,
  registration,
  me
}

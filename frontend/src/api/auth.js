import axios from '@/api/axios'
import urls from '@/api/urls'

// registration sends the MAX user registration form to the backend.
const registration = (credential) => {
  return axios.post(urls.registration, credential)
}

// me loads the current user profile that the backend resolves for the mini app.
const me = () => {
  return axios.get(urls.me)
}

export default {
  registration,
  me
}

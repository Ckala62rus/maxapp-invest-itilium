import axios from '@/api/axios'
import urls from '@/api/urls'

/** Обмен подписанного MAX initData на backend access token (см. POST /auth/max/validate). */
const validateMaxAuth = (payload) => {
  return axios.post(urls.maxAuthValidate, payload)
}

/** Регистрация сотрудника в ITILIUM через backend. */
const registration = (credential) => {
  return axios.post(urls.registration, credential)
}

/** Текущий профиль (ITILIUM + локальные флаги онбординга). */
const me = () => {
  return axios.get(urls.me)
}

export default {
  validateMaxAuth,
  registration,
  me
}

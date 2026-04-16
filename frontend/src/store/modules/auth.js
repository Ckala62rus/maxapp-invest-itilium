/**
 * Vuex-модуль auth: сессия мини-приложения MAX и backend access token.
 * Цепочка: bootstrap (initData → POST /auth/max/validate → токен в storage) → me (GET /users/me, профиль ITILIUM).
 */
import authApi from '@/api/auth'
import { getItem, removeItem, setItem } from '@/helpers/persistenceStorage'

const env = import.meta.env
const debugUserId = env.VITE_DEBUG_USER_ID || ''

const state = {
  user: null,
  identity: null,
  isSubmitting: false,
  isLoadingProfile: false,
  isBootstrapping: false,
  isAuth: false,
  authError: []
}

export const mutationTypes = {
  registrationStart: '[auth] registrationStart',
  registrationSuccess: '[auth] registrationSuccess',
  registrationFail: '[auth] registrationFail',
  bootstrapStart: '[auth] bootstrapStart',
  bootstrapSuccess: '[auth] bootstrapSuccess',
  bootstrapFail: '[auth] bootstrapFail',
  meStart: '[auth] meStart',
  meSuccess: '[auth] meSuccess',
  meFail: '[auth] meFail',
  logout: '[auth] logout'
}

export const actionTypes = {
  register: '[auth] register',
  bootstrap: '[auth] bootstrap',
  me: '[auth] me',
  logout: '[auth] logout'
}

export const getterTypes = {
  isAuth: '[auth] isAuth',
  user: '[auth] user',
  identity: '[auth] identity',
  isSubmitting: '[auth] isSubmitting',
  isBootstrapping: '[auth] isBootstrapping',
  authError: '[auth] authError'
}

const getters = {
  [getterTypes.isAuth]: (localState) => localState.isAuth,
  [getterTypes.user]: (localState) => localState.user,
  [getterTypes.identity]: (localState) => localState.identity,
  [getterTypes.isSubmitting]: (localState) => localState.isSubmitting,
  [getterTypes.isBootstrapping]: (localState) => localState.isBootstrapping,
  [getterTypes.authError]: (localState) => localState.authError
}

const mutations = {
  // Registration state is isolated so the registration screen can react
  // without coupling itself to unrelated auth requests.
  [mutationTypes.registrationStart](localState) {
    localState.isSubmitting = true
    localState.authError = []
  },
  [mutationTypes.registrationSuccess](localState, user) {
    localState.isSubmitting = false
    localState.user = user
    localState.isAuth = true
  },
  [mutationTypes.registrationFail](localState, errors) {
    localState.isSubmitting = false
    localState.authError = errors
  },

  [mutationTypes.bootstrapStart](localState) {
    localState.isBootstrapping = true
    localState.authError = []
  },
  [mutationTypes.bootstrapSuccess](localState, identity) {
    localState.isBootstrapping = false
    localState.identity = identity
    localState.isAuth = true
  },
  [mutationTypes.bootstrapFail](localState, errors) {
    localState.isBootstrapping = false
    localState.isAuth = false
    localState.identity = null
    localState.user = null
    localState.authError = errors
  },

  // Profile bootstrap keeps the MAX mini app aware of the current user
  // before screen-specific modules start requesting their own data.
  [mutationTypes.meStart](localState) {
    localState.isLoadingProfile = true
    localState.authError = []
  },
  [mutationTypes.meSuccess](localState, user) {
    localState.isLoadingProfile = false
    localState.user = user
    localState.isAuth = true
  },
  [mutationTypes.meFail](localState, errors) {
    localState.isLoadingProfile = false
    localState.isAuth = false
    localState.authError = errors
  },
  [mutationTypes.logout](localState) {
    localState.user = null
    localState.identity = null
    localState.isAuth = false
    localState.authError = []
  }
}

const actions = {
  // Старт приложения: либо восстановить сессию, либо обменять MAX initData на наш bearer token.
  [actionTypes.bootstrap](context, payload) {
    return new Promise((resolve) => {
      context.commit(mutationTypes.bootstrapStart)

      const storedToken = getItem('access_token')
      // Уже есть токен в localStorage и нет нового initData — не дергаем validate повторно (перезагрузка страницы).
      if (storedToken && !payload?.initData) {
        console.info('[auth] bootstrap: using stored backend token')
        context.commit(mutationTypes.bootstrapSuccess, context.state.identity)
        resolve({ data: { success: true } })
        return
      }

      // Локальная отладка в браузере без MAX: VITE_DEBUG_USER_ID + заголовок X-User-ID на бэкенде.
      if (env.DEV && debugUserId && !payload?.initData) {
        console.info('[auth] bootstrap: using debug user id', {
          userId: debugUserId
        })
        context.commit(mutationTypes.bootstrapSuccess, {
          userId: debugUserId
        })
        resolve({
          data: {
            success: true,
            debug: true,
            identity: {
              userId: debugUserId
            }
          }
        })
        return
      }

      // Реальный сценарий MAX: без initData с платформы продолжать нельзя.
      if (!payload?.initData) {
        console.warn('[auth] bootstrap: MAX initData is missing')
        context.commit(mutationTypes.bootstrapFail, ['MAX initData недоступен. Откройте приложение из MAX.'])
        resolve({ data: { success: false } })
        return
      }

      console.info('[auth] bootstrap: validating MAX initData', {
        initDataLength: payload.initData.length
      })
      authApi.validateMaxAuth({ initData: payload.initData })
        .then((response) => {
          const data = response?.data?.data || {}
          // Токен кладём в storage — axios interceptor подставит Authorization на все API.
          if (data.accessToken) {
            setItem('access_token', data.accessToken)
          }
          console.info('[auth] bootstrap: MAX initData validated', {
            userId: data?.identity?.userId || null
          })
          context.commit(mutationTypes.bootstrapSuccess, data.identity || null)
          resolve(response)
        })
        .catch((error) => {
          removeItem('access_token')
          console.error('[auth] bootstrap: validation failed', error)
          context.commit(mutationTypes.bootstrapFail, [error?.response?.data?.message || error.message])
          resolve(error)
        })
    })
  },

  // Отправка анкеты регистрации в ITILIUM через backend (после успеха — новый снимок профиля).
  [actionTypes.register](context, payload) {
    return new Promise((resolve) => {
      context.commit(mutationTypes.registrationStart)

      authApi.registration(payload)
        .then((response) => {
          context.commit(mutationTypes.registrationSuccess, response?.data?.data || null)
          resolve(response)
        })
        .catch((error) => {
          context.commit(mutationTypes.registrationFail, [error?.response?.data?.message || error.message])
          resolve(error)
        })
    })
  },

  // Текущий профиль (кэш/ITILIUM): employeeFound, registrationRequired, servicecalls и т.д.
  [actionTypes.me](context) {
    return new Promise((resolve) => {
      context.commit(mutationTypes.meStart)

      authApi.me()
        .then((response) => {
          context.commit(mutationTypes.meSuccess, response?.data?.data || null)
          resolve(response)
        })
        .catch((error) => {
          context.commit(mutationTypes.meFail, [error?.response?.data?.message || error.message])
          resolve(error)
        })
    })
  },

  [actionTypes.logout](context) {
    return new Promise((resolve) => {
      // Сбрасываем клиентскую сессию; на сервере отдельного logout может не быть.
      removeItem('access_token')
      context.commit(mutationTypes.logout)
      resolve()
    })
  }
}

export default {
  state,
  getters,
  mutations,
  actions
}

import authApi from '@/api/auth'
import { removeItem } from '@/helpers/persistenceStorage'

const state = {
  user: null,
  isSubmitting: false,
  isLoadingProfile: false,
  isAuth: false,
  authError: []
}

export const mutationTypes = {
  registrationStart: '[auth] registrationStart',
  registrationSuccess: '[auth] registrationSuccess',
  registrationFail: '[auth] registrationFail',
  meStart: '[auth] meStart',
  meSuccess: '[auth] meSuccess',
  meFail: '[auth] meFail',
  logout: '[auth] logout'
}

export const actionTypes = {
  register: '[auth] register',
  me: '[auth] me',
  logout: '[auth] logout'
}

export const getterTypes = {
  isAuth: '[auth] isAuth',
  user: '[auth] user',
  isSubmitting: '[auth] isSubmitting',
  authError: '[auth] authError'
}

const getters = {
  [getterTypes.isAuth]: (localState) => localState.isAuth,
  [getterTypes.user]: (localState) => localState.user,
  [getterTypes.isSubmitting]: (localState) => localState.isSubmitting,
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
    localState.isAuth = false
    localState.authError = []
  }
}

const actions = {
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

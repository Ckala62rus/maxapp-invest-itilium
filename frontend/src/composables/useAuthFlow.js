import { computed, ref, watch } from 'vue'

import { getMaxBridgeLaunchData, notifyMaxAppReady } from '@/api/maxBridge'
import {
  actionTypes as authActionTypes,
  getterTypes as authGetterTypes
} from '@/store/modules/auth'

// useAuthFlow groups auth-backed profile and registration state so App.vue can
// focus on screen composition while auth interactions stay in one place.
export function useAuthFlow({ store, activeScreen, submitBanner }) {
  const defaultRegistrationComment = 'Прошу связать мой аккаунт MAX с карточкой сотрудника.'

  // The registration form keeps editable UI state while auth data lives in Vuex.
  const registrationForm = ref({
    employeeNumber: '',
    fullName: '',
    organization: '',
    department: '',
    position: '',
    phone: '',
    comment: defaultRegistrationComment
  })
  const maxBridgeState = ref({
    isAvailable: false,
    initData: '',
    initDataUnsafe: null
  })

  // Store-backed auth state lets the profile and registration screens
  // use the same data flow that future real screens will rely on.
  const currentUser = computed(() => store.getters[authGetterTypes.user] || null)
  const currentIdentity = computed(() => store.getters[authGetterTypes.identity] || null)
  const isRegistrationSubmitting = computed(() => store.getters[authGetterTypes.isSubmitting])
  const isAuthBootstrapping = computed(() => store.getters[authGetterTypes.isBootstrapping])
  const authErrors = computed(() => store.getters[authGetterTypes.authError] || [])

  const profileInitials = computed(() => {
    const fullName = currentUser.value?.fullName || 'MAX Пользователь'

    return fullName
      .split(' ')
      .filter(Boolean)
      .slice(0, 2)
      .map((item) => item[0])
      .join('')
      .toUpperCase()
  })

  const profileStatusText = computed(() => {
    if (currentUser.value?.registrationPending) {
      return 'Заявка на регистрацию еще на рассмотрении'
    }

    return currentUser.value?.employeeFound
      ? 'Пользователь найден и связан с ITILIUM'
      : 'Не найден, нужна регистрация'
  })

  const profileRegion = computed(() => {
    const department = currentUser.value?.department || ''

    if (!department) {
      return 'Не определено'
    }

    const parts = department.split(',')

    return parts[parts.length - 1].trim()
  })

  const registrationIdentityUserId = computed(() => {
    return currentIdentity.value?.userId || currentUser.value?.userId || ''
  })
  const rawInitData = computed(() => maxBridgeState.value.initData || '')
  const rawInitDataUnsafeUserId = computed(() => {
    const userId = maxBridgeState.value.initDataUnsafe?.user?.id
    return typeof userId === 'string' || typeof userId === 'number' ? String(userId) : ''
  })

  // When the profile changes, the registration form is prefilled from the same source
  // so the user does not re-enter obvious data during the MAX onboarding flow.
  watch([currentUser, currentIdentity], ([user, identity]) => {
    const trustedUserId = identity?.userId || user?.userId || ''
    const hasResolvedEmployeeProfile = Boolean(user?.employeeFound) && !user?.registrationRequired
    registrationForm.value = {
      employeeNumber: registrationForm.value.employeeNumber || trustedUserId,
      fullName: registrationForm.value.fullName || (hasResolvedEmployeeProfile ? user?.fullName || '' : ''),
      organization: registrationForm.value.organization || '',
      department: registrationForm.value.department || (hasResolvedEmployeeProfile ? user?.department || '' : ''),
      position: registrationForm.value.position || '',
      phone: registrationForm.value.phone || '',
      comment: registrationForm.value.comment || defaultRegistrationComment
    }
  }, { immediate: true })

  async function bootstrapAuth() {
    notifyMaxAppReady()

    const launchData = getMaxBridgeLaunchData()
    maxBridgeState.value = launchData
    const bootstrapResponse = await store.dispatch(authActionTypes.bootstrap, {
      initData: launchData.initData
    })

    if (!bootstrapResponse?.data?.success) {
      activeScreen.value = 'profile'
      return bootstrapResponse
    }

    return store.dispatch(authActionTypes.me)
      .then((response) => {
        const user = response?.data?.data || null
        if (!user) {
          return {
            data: {
              success: false,
              stage: 'profile',
              user: null
            }
          }
        }

        if (user.registrationPending) {
          activeScreen.value = 'profile'
          submitBanner.value = user.statusMessage || ''
          return {
            data: {
              success: true,
              stage: 'registration_pending',
              user
            }
          }
        }

        if (user.registrationRequired) {
          activeScreen.value = 'registration'
          return {
            data: {
              success: true,
              stage: 'registration_required',
              user
            }
          }
        }

        activeScreen.value = 'home'
        return {
          data: {
            success: true,
            stage: 'ready',
            user
          }
        }
      })
  }

  // Registration goes through the shared auth module so the UI already uses
  // the same request lifecycle that future componentized screens will reuse.
  async function submitRegistration() {
    const response = await store.dispatch(authActionTypes.register, {
      userId: registrationIdentityUserId.value,
      employeeNumber: registrationForm.value.employeeNumber,
      fullName: registrationForm.value.fullName,
      organization: registrationForm.value.organization,
      department: registrationForm.value.department,
      position: registrationForm.value.position,
      phone: registrationForm.value.phone,
      comment: registrationForm.value.comment
    })

    if (response?.data?.success) {
      submitBanner.value = 'Регистрационная форма отправлена на проверку.'
      activeScreen.value = 'profile'
    }
  }

  return {
    registrationForm,
    currentUser,
    currentIdentity,
    isRegistrationSubmitting,
    isAuthBootstrapping,
    authErrors,
    profileInitials,
    profileStatusText,
    profileRegion,
    registrationIdentityUserId,
    maxBridgeState,
    rawInitData,
    rawInitDataUnsafeUserId,
    bootstrapAuth,
    submitRegistration
  }
}

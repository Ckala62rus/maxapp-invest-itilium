import { computed, ref, watch } from 'vue'

import {
  actionTypes as authActionTypes,
  getterTypes as authGetterTypes
} from '@/store/modules/auth'

// useAuthFlow groups auth-backed profile and registration state so App.vue can
// focus on screen composition while auth interactions stay in one place.
export function useAuthFlow({ store, activeScreen, submitBanner }) {
  // The registration form keeps editable UI state while auth data lives in Vuex.
  const registrationForm = ref({
    employeeNumber: '004512',
    fullName: '',
    department: '',
    phone: '+7 (999) 123-45-67',
    comment: 'Прошу связать мой аккаунт MAX с карточкой сотрудника.'
  })

  // Store-backed auth state lets the profile and registration screens
  // use the same data flow that future real screens will rely on.
  const currentUser = computed(() => store.getters[authGetterTypes.user] || null)
  const isRegistrationSubmitting = computed(() => store.getters[authGetterTypes.isSubmitting])
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

  // When the profile changes, the registration form is prefilled from the same source
  // so the user does not re-enter obvious data during the MAX onboarding flow.
  watch(currentUser, (user) => {
    registrationForm.value = {
      employeeNumber: registrationForm.value.employeeNumber || '004512',
      fullName: user?.fullName || '',
      department: user?.department || '',
      phone: registrationForm.value.phone || '+7 (999) 123-45-67',
      comment: registrationForm.value.comment || 'Прошу связать мой аккаунт MAX с карточкой сотрудника.'
    }
  }, { immediate: true })

  function loadAuthProfile() {
    return store.dispatch(authActionTypes.me)
  }

  // Registration goes through the shared auth module so the UI already uses
  // the same request lifecycle that future componentized screens will reuse.
  async function submitRegistration() {
    const response = await store.dispatch(authActionTypes.register, {
      userId: currentUser.value?.userId || '',
      employeeNumber: registrationForm.value.employeeNumber,
      fullName: registrationForm.value.fullName,
      department: registrationForm.value.department,
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
    isRegistrationSubmitting,
    authErrors,
    profileInitials,
    profileStatusText,
    profileRegion,
    loadAuthProfile,
    submitRegistration
  }
}

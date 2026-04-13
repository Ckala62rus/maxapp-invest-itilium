<script setup>
import { computed, onMounted, ref, watch } from 'vue'
import { useStore } from 'vuex'

import {
  actionTypes as authActionTypes,
  getterTypes as authGetterTypes
} from '@/store/modules/auth'
import HomeScreen from '@/screens/HomeScreen.vue'
import ProfileScreen from '@/screens/ProfileScreen.vue'
import RegistrationScreen from '@/screens/RegistrationScreen.vue'
import CreateTicketScreen from '@/screens/CreateTicketScreen.vue'
import {
  actionTypes as ticketActionTypes,
  getterTypes as ticketGetterTypes
} from '@/store/modules/tickets'
import MyTicketsScreen from '@/screens/MyTicketsScreen.vue'
import ResponsibleTicketsScreen from '@/screens/ResponsibleTicketsScreen.vue'
import SearchTicketScreen from '@/screens/SearchTicketScreen.vue'
import TicketDetailsScreen from '@/screens/TicketDetailsScreen.vue'
import { useTicketFlow } from '@/composables/useTicketFlow'

const store = useStore()

// Screen ids are used by the static prototype navigator.
const screenOptions = [
  { id: 'home', label: 'Главная' },
  { id: 'profile', label: 'Профиль' },
  { id: 'registration', label: 'Регистрация' },
  { id: 'create', label: 'Создать заявку' },
  { id: 'myTickets', label: 'Мои заявки' },
  { id: 'responsible', label: 'В ответственности' },
  { id: 'search', label: 'Поиск' },
  { id: 'details', label: 'Карточка заявки' }
]

// The active screen simulates page transitions for the stakeholder demo.
const activeScreen = ref('home')

// The submission banner imitates the visual result of a completed action.
const submitBanner = ref('')

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

const {
  searchQuery,
  currentTicketsPage,
  commentDraft,
  statusForm,
  selectedResponsibleId,
  createTicketForm,
  isLoadingMyTickets,
  isLoadingResponsibleTickets,
  isCreatingTicket,
  isLoadingTicketDetails,
  isLoadingResponsibleOptions,
  isSubmittingComment,
  isChangingStatus,
  isChangingResponsible,
  listErrors,
  ticketErrors,
  createErrors,
  normalizedResponsibleTickets,
  summaryCards,
  paginatedTickets,
  pageCount,
  selectedTicket,
  detailStatusTone,
  availableStatusOptions,
  availableResponsibleOptions,
  selectedTicketTimeline,
  loadTicketLists,
  openTicketDetails,
  searchTicketByNumber,
  submitCreateTicket,
  submitComment,
  submitStatusChange,
  assignResponsible,
  setTicketsPage,
  setSearchQuery,
  setCommentDraft
} = useTicketFlow({
  store,
  currentUser,
  activeScreen,
  submitBanner
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

onMounted(() => {
  store.dispatch(authActionTypes.me)
  loadTicketLists()
})

// This helper switches screens from the top navigation and action buttons.
function openScreen(screenId) {
  activeScreen.value = screenId
  submitBanner.value = ''
}

// This helper only changes UI feedback for the visual prototype.
function simulateSubmit(message) {
  submitBanner.value = message
}

// Registration now goes through the shared auth module so the UI already uses
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

</script>

<template>
  <div class="prototype-shell">
    <aside class="overview-panel">
      <div class="brand-card">
        <p class="eyebrow">MAX Mini App</p>
        <h1>ITILIUM service desk prototype</h1>
        <p class="supporting-text">
          Статический адаптивный шаблон для показа заказчику. Здесь собраны все ключевые экраны:
          регистрация, создание заявок, списки, поиск, карточка инцидента, пагинация и состояния загрузки.
        </p>
      </div>

      <div class="overview-card">
        <h2>Навигация по экранам</h2>
        <div class="screen-grid">
          <button
            v-for="screen in screenOptions"
            :key="screen.id"
            class="screen-chip"
            :class="{ active: activeScreen === screen.id }"
            @click="openScreen(screen.id)"
          >
            {{ screen.label }}
          </button>
        </div>
      </div>

      <div class="overview-card">
        <h2>Что уже показано в макете</h2>
        <ul class="check-list">
          <li>Приветственный экран mini app</li>
          <li>Экран профиля MAX пользователя</li>
          <li>Регистрация, если пользователя нет в ITILIUM</li>
          <li>Форма создания обычной и маркетинговой заявки</li>
          <li>Мои заявки и заявки в ответственности</li>
          <li>Карточка заявки, комментарии и действия</li>
          <li>Поиск по номеру, пагинация, spinner, empty, error, success</li>
        </ul>
      </div>
    </aside>

    <main class="phone-stage">
      <div class="phone-frame">
        <header class="app-header">
          <div>
            <p class="eyebrow">MAX x ITILIUM</p>
            <strong>Сервисные заявки</strong>
          </div>
          <button class="ghost-button" @click="openScreen('home')">Домой</button>
        </header>

        <div class="tab-strip">
          <button
            v-for="screen in screenOptions"
            :key="screen.id"
            class="tab-button"
            :class="{ active: activeScreen === screen.id }"
            @click="openScreen(screen.id)"
          >
            {{ screen.label }}
          </button>
        </div>

        <HomeScreen
          v-if="activeScreen === 'home'"
          :summary-cards="summaryCards"
          @open-screen="openScreen"
        />

        <ProfileScreen
          v-else-if="activeScreen === 'profile'"
          :current-user="currentUser"
          :profile-initials="profileInitials"
          :profile-status-text="profileStatusText"
          :profile-region="profileRegion"
          @open-screen="openScreen"
        />

        <RegistrationScreen
          v-else-if="activeScreen === 'registration'"
          :registration-form="registrationForm"
          :auth-errors="authErrors"
          :is-registration-submitting="isRegistrationSubmitting"
          @submit-registration="submitRegistration"
        />

        <CreateTicketScreen
          v-else-if="activeScreen === 'create'"
          :create-ticket-form="createTicketForm"
          :create-errors="createErrors"
          :is-creating-ticket="isCreatingTicket"
          @submit-create-ticket="submitCreateTicket"
          @open-screen="openScreen"
        />

        <MyTicketsScreen
          v-else-if="activeScreen === 'myTickets'"
          :is-loading-my-tickets="isLoadingMyTickets"
          :list-errors="listErrors"
          :paginated-tickets="paginatedTickets"
          :page-count="pageCount"
          :current-tickets-page="currentTicketsPage"
          @open-ticket-details="openTicketDetails"
          @set-tickets-page="setTicketsPage"
        />

        <ResponsibleTicketsScreen
          v-else-if="activeScreen === 'responsible'"
          :is-loading-responsible-tickets="isLoadingResponsibleTickets"
          :list-errors="listErrors"
          :normalized-responsible-tickets="normalizedResponsibleTickets"
          @open-ticket-details="openTicketDetails"
        />

        <SearchTicketScreen
          v-else-if="activeScreen === 'search'"
          :search-query="searchQuery"
          :ticket-errors="ticketErrors"
          :is-loading-ticket-details="isLoadingTicketDetails"
          @update:search-query="setSearchQuery"
          @search-ticket="searchTicketByNumber"
          @open-ticket-details="openTicketDetails(searchQuery)"
        />

        <TicketDetailsScreen
          v-else-if="activeScreen === 'details'"
          :selected-ticket="selectedTicket"
          :search-query="searchQuery"
          :detail-status-tone="detailStatusTone"
          :is-loading-ticket-details="isLoadingTicketDetails"
          :ticket-errors="ticketErrors"
          :selected-ticket-timeline="selectedTicketTimeline"
          :comment-draft="commentDraft"
          :is-submitting-comment="isSubmittingComment"
          :status-form="statusForm"
          :available-status-options="availableStatusOptions"
          :is-changing-status="isChangingStatus"
          :is-loading-responsible-options="isLoadingResponsibleOptions"
          :available-responsible-options="availableResponsibleOptions"
          :is-changing-responsible="isChangingResponsible"
          :selected-responsible-id="selectedResponsibleId"
          @open-screen="openScreen"
          @update:comment-draft="setCommentDraft"
          @submit-comment="submitComment"
          @submit-status-change="submitStatusChange"
          @assign-responsible="assignResponsible"
        />

        <footer class="app-footer">
          <p>Прототип адаптирован под mobile webview и готов к дальнейшему переносу в реальные API-сценарии.</p>
        </footer>

        <transition name="banner">
          <div v-if="submitBanner" class="submit-banner">
            {{ submitBanner }}
          </div>
        </transition>
      </div>
    </main>
  </div>
</template>

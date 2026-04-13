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

// The search field persists the current ticket number between list and search flows.
const searchQuery = ref('SC-000245')

// The current list page shows how pagination will look in the future app.
const currentTicketsPage = ref(1)

// The comment draft keeps the prototype form editable until comment actions move to API.
const commentDraft = ref('Нужна срочная проверка после ночного обновления.')

// The status form mirrors the backend transition contract used by the ticket card.
const statusForm = ref({
  state: '',
  comment: '',
  date: ''
})

// The responsible selector keeps the chosen ITILIUM assignee id before submit.
const selectedResponsibleId = ref('')

// The create form mirrors the backend ticket creation contract.
const createTicketForm = ref({
  requestType: 'Заявка в отдел ИТ',
  title: 'Не открывается 1С на кассе',
  description: 'После обновления 1С не запускается на рабочем месте кассира. Нужна диагностика и восстановление работы.',
  department: 'Отдел ИТ',
  executionDate: '2026-04-11',
  attachments: ['screenshot-1.png', 'error-log.pdf']
})

// The registration form keeps editable UI state while auth data lives in Vuex.
const registrationForm = ref({
  employeeNumber: '004512',
  fullName: '',
  department: '',
  phone: '+7 (999) 123-45-67',
  comment: 'Прошу связать мой аккаунт MAX с карточкой сотрудника.'
})

// These cards imitate the current user's tickets.
const myTickets = [
  { number: 'SC-000245', title: 'Не открывается 1С на кассе', state: 'В работе', deadline: '11.04.2026', tone: 'amber' },
  { number: 'SC-000244', title: 'Нужен доступ к отчету по складу', state: 'Зарегистрирована', deadline: '10.04.2026', tone: 'blue' },
  { number: 'SC-000238', title: 'Ошибка печати ценников', state: 'На согласовании', deadline: '09.04.2026', tone: 'purple' },
  { number: 'SC-000229', title: 'Не загружается каталог в терминале', state: 'Закрыта', deadline: '07.04.2026', tone: 'green' },
  { number: 'SC-000220', title: 'Сбой в учете продаж по магазину', state: 'Отложена', deadline: '15.04.2026', tone: 'slate' },
  { number: 'SC-000210', title: 'Нужен макет для акции', state: 'Маркетинг', deadline: '18.04.2026', tone: 'pink' }
]

// These cards imitate tickets where the current user is responsible.
const responsibleTickets = [
  { number: 'SC-000310', title: 'Проблема с авторизацией сотрудника', state: 'Ожидает ответа', owner: 'Магазин 17', tone: 'amber' },
  { number: 'SC-000308', title: 'Нужна смена ответственного', state: 'В работе', owner: 'Офис продаж', tone: 'blue' },
  { number: 'SC-000299', title: 'Добавить комментарий по инциденту', state: 'На согласовании', owner: 'РЦ Казань', tone: 'purple' }
]

// The responsible selector demonstrates a future modal with paginated assignees.
const responsiblePeople = [
  { team: 'Отдел ИТ', person: 'Иван Петров', role: 'Старший инженер', externalId: 'emp-1' },
  { team: 'Отдел ИТ', person: 'Елена Орлова', role: 'Системный аналитик', externalId: 'emp-2' },
  { team: 'Маркетинг', person: 'Мария Соколова', role: 'Маркетолог', externalId: 'emp-3' }
]

// A computed slice is enough to show how page navigation will behave.
const paginatedTickets = computed(() => {
  const start = (currentTicketsPage.value - 1) * 3
  return normalizedMyTickets.value.slice(start, start + 3)
})

// The page count drives the pager buttons in the static demo.
const pageCount = computed(() => Math.ceil(normalizedMyTickets.value.length / 3))

// Store-backed auth state lets the profile and registration screens
// use the same data flow that future real screens will rely on.
const currentUser = computed(() => store.getters[authGetterTypes.user] || null)
const isRegistrationSubmitting = computed(() => store.getters[authGetterTypes.isSubmitting])
const authErrors = computed(() => store.getters[authGetterTypes.authError] || [])
const storeMyTickets = computed(() => store.getters[ticketGetterTypes.myTickets] || [])
const storeResponsibleTickets = computed(() => store.getters[ticketGetterTypes.responsibleTickets] || [])
const selectedTicket = computed(() => store.getters[ticketGetterTypes.selectedTicket] || null)
const responsibleOptions = computed(() => store.getters[ticketGetterTypes.responsibleOptions] || [])
const isLoadingMyTickets = computed(() => store.getters[ticketGetterTypes.isLoadingMyTickets])
const isLoadingResponsibleTickets = computed(() => store.getters[ticketGetterTypes.isLoadingResponsibleTickets])
const isCreatingTicket = computed(() => store.getters[ticketGetterTypes.isCreatingTicket])
const isLoadingTicketDetails = computed(() => store.getters[ticketGetterTypes.isLoadingTicketDetails])
const isLoadingResponsibleOptions = computed(() => store.getters[ticketGetterTypes.isLoadingResponsibleOptions])
const isSubmittingComment = computed(() => store.getters[ticketGetterTypes.isSubmittingComment])
const isChangingStatus = computed(() => store.getters[ticketGetterTypes.isChangingStatus])
const isChangingResponsible = computed(() => store.getters[ticketGetterTypes.isChangingResponsible])
const listErrors = computed(() => store.getters[ticketGetterTypes.listError] || [])
const ticketErrors = computed(() => store.getters[ticketGetterTypes.ticketError] || [])
const createErrors = computed(() => store.getters[ticketGetterTypes.createError] || [])
const normalizedMyTickets = computed(() => {
  const source = storeMyTickets.value.length ? storeMyTickets.value : myTickets

  return source.map((ticket) => ({
    ...ticket,
    tone: ticket.tone || resolveTicketTone(ticket.state)
  }))
})
const normalizedResponsibleTickets = computed(() => {
  const source = storeResponsibleTickets.value.length ? storeResponsibleTickets.value : responsibleTickets

  return source.map((ticket) => ({
    ...ticket,
    tone: ticket.tone || resolveTicketTone(ticket.state),
    owner: ticket.owner || ticket.responsibleTeam || 'Не указано'
  }))
})
// The dashboard now summarizes the same auth and ticket data that power the screens,
// so the home overview reacts when lists load or new tickets are created.
const summaryCards = computed(() => {
  const inProgressCount = normalizedMyTickets.value.filter((ticket) => {
    return ['В работе', 'Ожидает ответа'].includes(ticket.state)
  }).length

  return [
    {
      title: 'Мои заявки',
      value: String(normalizedMyTickets.value.length),
      tone: 'blue'
    },
    {
      title: 'В работе',
      value: String(inProgressCount),
      tone: 'amber'
    },
    {
      title: 'В моей ответственности',
      value: String(normalizedResponsibleTickets.value.length),
      tone: 'purple'
    },
    {
      title: 'Нужна регистрация',
      value: currentUser.value?.registrationRequired ? '1' : '0',
      tone: 'rose'
    }
  ]
})
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
const detailStatusTone = computed(() => {
  return resolveTicketTone(selectedTicket.value?.state)
})
const availableStatusOptions = computed(() => {
  return selectedTicket.value?.availableStates || []
})
const availableResponsibleOptions = computed(() => {
  return responsibleOptions.value.length ? responsibleOptions.value : responsiblePeople
})
const selectedTicketTimeline = computed(() => {
  if (!selectedTicket.value?.timeline?.length) {
    return []
  }

  return selectedTicket.value.timeline.map((item) => ({
    actor: item.author || item.actor || 'Система',
    text: item.message || item.text || '',
    time: formatTimelineTime(item.createdAt || item.time || '')
  }))
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

// When a new detail payload arrives, derived forms stay in sync with the same
// Vuex source of truth instead of preserving stale values from the previous ticket.
watch(selectedTicket, async (ticket) => {
  commentDraft.value = 'Нужна срочная проверка после ночного обновления.'
  selectedResponsibleId.value = ''
  statusForm.value = {
    state: ticket?.availableStates?.[0] || '',
    comment: '',
    date: ticket?.deadline || ''
  }

  if (ticket?.number && ticket.canChangeResponsible) {
    await store.dispatch(ticketActionTypes.loadResponsibleOptions, ticket.number)
  }
}, { immediate: true })

onMounted(() => {
  store.dispatch(authActionTypes.me)
  store.dispatch(ticketActionTypes.loadMyTickets)
  store.dispatch(ticketActionTypes.loadResponsibleTickets)
})

// This helper switches screens from the top navigation and action buttons.
function openScreen(screenId) {
  activeScreen.value = screenId
  submitBanner.value = ''
}

// This helper simulates navigation from list card to detail page.
async function openTicketDetails(ticketNumber) {
  searchQuery.value = ticketNumber
  activeScreen.value = 'details'
  await store.dispatch(ticketActionTypes.loadTicketDetails, ticketNumber)
}

// This helper only changes UI feedback for the visual prototype.
function simulateSubmit(message) {
  submitBanner.value = message
}

// Search now goes through the dedicated backend endpoint so the details screen
// can open from either a manual search or a list click using shared Vuex state.
async function searchTicketByNumber() {
  const number = searchQuery.value.trim()

  if (!number) {
    return
  }

  const response = await store.dispatch(ticketActionTypes.searchTicket, {
    number,
    userId: currentUser.value?.userId || ''
  })

  if (response?.data?.success) {
    activeScreen.value = 'details'
  }
}

async function submitCreateTicket() {
  const response = await store.dispatch(ticketActionTypes.createTicket, {
    userId: currentUser.value?.userId || '',
    requestType: createTicketForm.value.requestType,
    title: createTicketForm.value.title,
    description: createTicketForm.value.description,
    department: createTicketForm.value.department,
    executionDate: createTicketForm.value.executionDate,
    attachments: createTicketForm.value.attachments
  })

  if (response?.data?.success) {
    searchQuery.value = response?.data?.data?.number || ''
    submitBanner.value = 'Заявка создана и открыта в карточке.'
    activeScreen.value = 'details'
  }
}

async function submitComment() {
  if (!selectedTicket.value?.number || !commentDraft.value.trim()) {
    return
  }

  const response = await store.dispatch(ticketActionTypes.addComment, {
    number: selectedTicket.value.number,
    data: {
      userId: currentUser.value?.userId || '',
      message: commentDraft.value.trim(),
      attachments: []
    }
  })

  if (response?.data?.success) {
    commentDraft.value = ''
    submitBanner.value = 'Комментарий отправлен и карточка заявки обновлена.'
  }
}

async function submitStatusChange() {
  if (!selectedTicket.value?.number || !statusForm.value.state) {
    return
  }

  const response = await store.dispatch(ticketActionTypes.changeStatus, {
    number: selectedTicket.value.number,
    data: {
      userId: currentUser.value?.userId || '',
      state: statusForm.value.state,
      comment: statusForm.value.comment,
      date: statusForm.value.date
    }
  })

  if (response?.data?.success) {
    submitBanner.value = 'Статус заявки обновлен.'
  }
}

async function assignResponsible(responsibleId) {
  if (!selectedTicket.value?.number || !responsibleId) {
    return
  }

  selectedResponsibleId.value = responsibleId

  const response = await store.dispatch(ticketActionTypes.changeResponsible, {
    number: selectedTicket.value.number,
    data: {
      userId: currentUser.value?.userId || '',
      responsibleId
    }
  })

  if (response?.data?.success) {
    submitBanner.value = 'Ответственный по заявке обновлен.'
    return
  }

  selectedResponsibleId.value = ''
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

// This helper changes the pager while keeping the prototype deterministic.
function setTicketsPage(page) {
  currentTicketsPage.value = page
}

function setSearchQuery(value) {
  searchQuery.value = value
}

function setCommentDraft(value) {
  commentDraft.value = value
}

function resolveTicketTone(state) {
  switch (state) {
    case 'Зарегистрирована':
      return 'blue'
    case 'В работе':
      return 'amber'
    case 'На согласовании':
      return 'purple'
    case 'Закрыта':
      return 'green'
    case 'Отложена':
      return 'slate'
    case 'Ожидает ответа':
      return 'amber'
    default:
      return 'info'
  }
}

function formatTimelineTime(value) {
  if (!value) {
    return ''
  }

  const parsedDate = new Date(value)

  if (Number.isNaN(parsedDate.getTime())) {
    return value
  }

  return parsedDate.toLocaleTimeString('ru-RU', {
    hour: '2-digit',
    minute: '2-digit'
  })
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

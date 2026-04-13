<script setup>
import { computed, onMounted, ref, watch } from 'vue'
import { useStore } from 'vuex'

import {
  actionTypes as authActionTypes,
  getterTypes as authGetterTypes
} from '@/store/modules/auth'
import {
  actionTypes as ticketActionTypes,
  getterTypes as ticketGetterTypes
} from '@/store/modules/tickets'

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

// The prototype dashboard metrics visually summarize the product scope.
const summaryCards = [
  { title: 'Мои заявки', value: '12', tone: 'blue' },
  { title: 'В работе', value: '4', tone: 'amber' },
  { title: 'В моей ответственности', value: '7', tone: 'purple' },
  { title: 'Нужна регистрация', value: '1', tone: 'rose' }
]

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

        <section v-if="activeScreen === 'home'" class="screen">
          <div class="hero-card">
            <p class="status-pill success">Mini App готов к демонстрации</p>
            <h2>Управляйте заявками ITILIUM прямо внутри MAX</h2>
            <p>
              Пользователь сможет пройти регистрацию, создать обращение, отследить статус,
              оставить комментарий и работать с заявками в своей ответственности.
            </p>
            <div class="hero-actions">
              <button class="primary-button" @click="openScreen('create')">Создать заявку</button>
              <button class="secondary-button" @click="openScreen('myTickets')">Мои заявки</button>
            </div>
          </div>

          <div class="summary-grid">
            <article
              v-for="card in summaryCards"
              :key="card.title"
              class="summary-card"
              :class="card.tone"
            >
              <span>{{ card.title }}</span>
              <strong>{{ card.value }}</strong>
            </article>
          </div>

          <div class="state-grid">
            <article class="state-card">
              <div class="spinner"></div>
              <div>
                <h3>Loading state</h3>
                <p>Используем на экранах поиска, списка и отправки заявки.</p>
              </div>
            </article>
            <article class="state-card">
              <div class="state-icon empty">0</div>
              <div>
                <h3>Empty state</h3>
                <p>Нет заявок в выборке. Предлагаем создать новое обращение.</p>
              </div>
            </article>
            <article class="state-card">
              <div class="state-icon error">!</div>
              <div>
                <h3>Error state</h3>
                <p>Итилиум недоступен или вернул ошибку. Показываем дружелюбный текст.</p>
              </div>
            </article>
          </div>
        </section>

        <section v-else-if="activeScreen === 'profile'" class="screen">
          <div class="section-header">
            <div>
              <p class="eyebrow">Профиль</p>
              <h2>Пользователь MAX</h2>
            </div>
            <span class="status-pill info">Авторизация пройдена</span>
          </div>

          <article class="content-card">
            <div class="profile-row">
              <div class="avatar">{{ profileInitials }}</div>
              <div>
                <h3>{{ currentUser?.fullName || 'Загрузка профиля...' }}</h3>
                <p>@{{ currentUser?.username || 'unknown' }} · user_id {{ currentUser?.userId || '...' }}</p>
              </div>
            </div>
            <div class="details-grid">
              <div>
                <span>Статус в ITILIUM</span>
                <strong>{{ profileStatusText }}</strong>
              </div>
              <div>
                <span>Роль в MAX</span>
                <strong>Пользователь mini app</strong>
              </div>
              <div>
                <span>Регион</span>
                <strong>{{ profileRegion }}</strong>
              </div>
              <div>
                <span>Последний вход</span>
                <strong>09.04.2026 22:30</strong>
              </div>
            </div>
            <button
              class="primary-button wide"
              @click="openScreen(currentUser?.registrationRequired ? 'registration' : 'home')"
            >
              {{ currentUser?.registrationRequired ? 'Перейти к регистрации' : 'Перейти на главную' }}
            </button>
          </article>
        </section>

        <section v-else-if="activeScreen === 'registration'" class="screen">
          <div class="section-header">
            <div>
              <p class="eyebrow">Регистрация</p>
              <h2>Вас не нашли в ITILIUM</h2>
            </div>
            <span class="status-pill warning">Требуется заполнение формы</span>
          </div>

          <article class="content-card form-card">
            <label>
              Табельный номер
              <input v-model="registrationForm.employeeNumber" type="text" />
            </label>
            <label>
              ФИО
              <input v-model="registrationForm.fullName" type="text" />
            </label>
            <label>
              Магазин / подразделение
              <input v-model="registrationForm.department" type="text" />
            </label>
            <label>
              Телефон
              <input v-model="registrationForm.phone" type="text" />
            </label>
            <label>
              Комментарий
              <textarea v-model="registrationForm.comment" rows="4"></textarea>
            </label>
            <p v-if="authErrors.length" class="status-pill rose">{{ authErrors[0] }}</p>
            <button class="primary-button wide" :disabled="isRegistrationSubmitting" @click="submitRegistration">
              {{ isRegistrationSubmitting ? 'Отправка...' : 'Отправить заявку на регистрацию' }}
            </button>
          </article>
        </section>

        <section v-else-if="activeScreen === 'create'" class="screen">
          <div class="section-header">
            <div>
              <p class="eyebrow">Новая заявка</p>
              <h2>Создание обращения</h2>
            </div>
            <span class="status-pill info">Поддерживает файлы и маркетинг</span>
          </div>

          <article class="content-card form-card">
            <label>
              Тип заявки
              <select v-model="createTicketForm.requestType">
                <option>Заявка в отдел ИТ</option>
                <option>Маркетинговая заявка</option>
              </select>
            </label>
            <label>
              Краткая тема
              <input v-model="createTicketForm.title" type="text" />
            </label>
            <label>
              Подробное описание
              <textarea v-model="createTicketForm.description" rows="5"></textarea>
            </label>
            <label>
              Подразделение
              <select v-model="createTicketForm.department">
                <option>Отдел ИТ</option>
                <option>Маркетинг</option>
              </select>
            </label>
            <label>
              Исполнить до
              <input v-model="createTicketForm.executionDate" type="date" />
            </label>

            <p v-if="createErrors.length" class="status-pill rose">{{ createErrors[0] }}</p>

            <div class="upload-box">
              <div>
                <strong>Вложения</strong>
                <p>Скриншоты, фото, документы, голосовые сообщения.</p>
              </div>
              <button class="secondary-button" disabled>Добавить файл</button>
            </div>

            <div class="chip-list">
              <span v-for="fileName in createTicketForm.attachments" :key="fileName" class="file-chip">{{ fileName }}</span>
            </div>

            <div class="hero-actions">
              <button
                class="primary-button"
                :disabled="isCreatingTicket || !createTicketForm.title || !createTicketForm.description"
                @click="submitCreateTicket"
              >
                {{ isCreatingTicket ? 'Отправка...' : 'Отправить заявку' }}
              </button>
              <button class="secondary-button" :disabled="isCreatingTicket" @click="openScreen('home')">Отмена</button>
            </div>
          </article>
        </section>

        <section v-else-if="activeScreen === 'myTickets'" class="screen">
          <div class="section-header">
            <div>
              <p class="eyebrow">Мои заявки</p>
              <h2>История обращений</h2>
            </div>
            <span class="status-pill info">Пагинация готова</span>
          </div>

          <article v-if="isLoadingMyTickets" class="state-card">
            <div class="spinner"></div>
            <div>
              <h3>Загружаем список</h3>
              <p>Получаем ваши заявки из общего Vuex store и backend API.</p>
            </div>
          </article>

          <p v-else-if="listErrors.length" class="status-pill rose">{{ listErrors[0] }}</p>

          <div v-else class="list-stack">
            <article
              v-for="ticket in paginatedTickets"
              :key="ticket.number"
              class="ticket-card"
              @click="openTicketDetails(ticket.number)"
            >
              <div class="ticket-topline">
                <strong>{{ ticket.number }}</strong>
                <span class="status-pill" :class="ticket.tone">{{ ticket.state }}</span>
              </div>
              <h3>{{ ticket.title }}</h3>
              <p>Срок реакции до {{ ticket.deadline }}</p>
            </article>
          </div>

          <div v-if="pageCount > 1" class="pagination">
            <button
              v-for="page in pageCount"
              :key="page"
              class="page-button"
              :class="{ active: page === currentTicketsPage }"
              @click="setTicketsPage(page)"
            >
              {{ page }}
            </button>
          </div>
        </section>

        <section v-else-if="activeScreen === 'responsible'" class="screen">
          <div class="section-header">
            <div>
              <p class="eyebrow">Ответственность</p>
              <h2>Заявки, закрепленные за мной</h2>
            </div>
            <span class="status-pill warning">Требуют реакции</span>
          </div>

          <article v-if="isLoadingResponsibleTickets" class="state-card">
            <div class="spinner"></div>
            <div>
              <h3>Загружаем ответственность</h3>
              <p>Получаем заявки, закрепленные за текущим пользователем.</p>
            </div>
          </article>

          <p v-else-if="listErrors.length" class="status-pill rose">{{ listErrors[0] }}</p>

          <div v-else class="list-stack">
            <article
              v-for="ticket in normalizedResponsibleTickets"
              :key="ticket.number"
              class="ticket-card"
              @click="openTicketDetails(ticket.number)"
            >
              <div class="ticket-topline">
                <strong>{{ ticket.number }}</strong>
                <span class="status-pill" :class="ticket.tone">{{ ticket.state }}</span>
              </div>
              <h3>{{ ticket.title }}</h3>
              <p>Инициатор: {{ ticket.owner }}</p>
            </article>
          </div>
        </section>

        <section v-else-if="activeScreen === 'search'" class="screen">
          <div class="section-header">
            <div>
              <p class="eyebrow">Поиск</p>
              <h2>Поиск заявки по номеру</h2>
            </div>
            <span class="status-pill info">Быстрый доступ к карточке</span>
          </div>

          <article class="content-card form-card">
            <label>
              Номер заявки
              <input v-model="searchQuery" type="text" />
            </label>
            <p v-if="ticketErrors.length" class="status-pill rose">{{ ticketErrors[0] }}</p>
            <div class="hero-actions">
              <button class="primary-button" :disabled="isLoadingTicketDetails" @click="searchTicketByNumber">
                {{ isLoadingTicketDetails ? 'Ищем заявку...' : 'Найти заявку' }}
              </button>
              <button class="secondary-button" :disabled="isLoadingTicketDetails" @click="openTicketDetails(searchQuery)">
                Открыть карточку по номеру
              </button>
            </div>
          </article>
        </section>

        <section v-else-if="activeScreen === 'details'" class="screen">
          <div class="section-header">
            <div>
              <p class="eyebrow">Карточка заявки</p>
              <h2>{{ selectedTicket?.number || searchQuery }}</h2>
            </div>
            <span class="status-pill" :class="detailStatusTone">{{ selectedTicket?.state || 'info' }}</span>
          </div>

          <article v-if="isLoadingTicketDetails" class="state-card">
            <div class="spinner"></div>
            <div>
              <h3>Загружаем карточку</h3>
              <p>Получаем полную информацию о заявке и ее ленту событий из backend API.</p>
            </div>
          </article>

          <article v-else-if="ticketErrors.length" class="content-card">
            <h3>Заявку не удалось открыть</h3>
            <p>{{ ticketErrors[0] }}</p>
            <div class="hero-actions">
              <button class="primary-button" @click="openScreen('search')">Вернуться к поиску</button>
            </div>
          </article>

          <article v-else-if="selectedTicket" class="content-card">
            <div class="details-grid">
              <div>
                <span>Краткая тема</span>
                <strong>{{ selectedTicket.title }}</strong>
              </div>
              <div>
                <span>Ответственная команда</span>
                <strong>{{ selectedTicket.responsibleTeam }}</strong>
              </div>
              <div>
                <span>Срок</span>
                <strong>{{ selectedTicket.deadline }}</strong>
              </div>
              <div>
                <span>Можно сменить ответственного</span>
                <strong>{{ selectedTicket.canChangeResponsible ? 'Да' : 'Нет' }}</strong>
              </div>
            </div>

            <div class="content-card compact">
              <span>Описание</span>
              <p>{{ selectedTicket.description }}</p>
            </div>

            <div class="action-grid">
              <button class="secondary-button" :disabled="isSubmittingComment">Добавить комментарий</button>
              <button class="secondary-button" :disabled="isChangingStatus">Поменять статус</button>
              <button class="secondary-button" :disabled="isChangingResponsible">Сменить ответственного</button>
              <button class="secondary-button" disabled>Оценить решение</button>
            </div>
          </article>

          <article v-else class="content-card">
            <h3>Карточка пока не выбрана</h3>
            <p>Открой заявку из списка или найдите ее по номеру, чтобы загрузить детали из Vuex store.</p>
          </article>

          <article v-if="selectedTicket" class="content-card">
            <h3>Лента событий</h3>
            <div class="timeline">
              <div v-for="item in selectedTicketTimeline" :key="`${item.actor}-${item.time}-${item.text}`" class="timeline-item">
                <strong>{{ item.actor }}</strong>
                <p>{{ item.text }}</p>
                <span>{{ item.time }}</span>
              </div>
            </div>
          </article>

          <article v-if="selectedTicket" class="content-card">
            <h3>Новый комментарий</h3>
            <div class="form-card compact">
              <label>
                Текст комментария
                <textarea v-model="commentDraft" rows="4"></textarea>
              </label>
              <div class="hero-actions">
                <button
                  class="primary-button"
                  :disabled="isSubmittingComment || !commentDraft.trim()"
                  @click="submitComment"
                >
                  {{ isSubmittingComment ? 'Отправляем...' : 'Отправить комментарий' }}
                </button>
                <button class="secondary-button" disabled>Прикрепить файл</button>
              </div>
            </div>
          </article>

          <article v-if="selectedTicket" class="content-card">
            <h3>Смена статуса</h3>
            <div class="form-card compact">
              <label>
                Новый статус
                <select v-model="statusForm.state">
                  <option disabled value="">Выберите статус</option>
                  <option v-for="status in availableStatusOptions" :key="status" :value="status">
                    {{ status }}
                  </option>
                </select>
              </label>
              <label>
                Комментарий к переходу
                <textarea v-model="statusForm.comment" rows="3"></textarea>
              </label>
              <label>
                Дата исполнения
                <input v-model="statusForm.date" type="text" />
              </label>
              <div class="hero-actions">
                <button
                  class="primary-button"
                  :disabled="isChangingStatus || !statusForm.state"
                  @click="submitStatusChange"
                >
                  {{ isChangingStatus ? 'Сохраняем...' : 'Сменить статус' }}
                </button>
              </div>
            </div>
          </article>

          <article v-if="selectedTicket" class="content-card">
            <h3>Выбор нового ответственного</h3>
            <article v-if="isLoadingResponsibleOptions" class="state-card compact">
              <div class="spinner"></div>
              <div>
                <h3>Загружаем список</h3>
                <p>Получаем доступных ответственных для этой заявки.</p>
              </div>
            </article>
            <div v-else class="selector-list">
              <div v-for="person in availableResponsibleOptions" :key="person.externalId || person.person" class="selector-item">
                <div>
                  <strong>{{ person.person }}</strong>
                  <p>{{ person.team }} · {{ person.role }}</p>
                </div>
                <button
                  class="ghost-button"
                  :disabled="isChangingResponsible"
                  @click="assignResponsible(person.externalId)"
                >
                  {{ isChangingResponsible && selectedResponsibleId === person.externalId ? 'Сохраняем...' : 'Выбрать' }}
                </button>
              </div>
            </div>
          </article>
        </section>

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

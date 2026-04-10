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

// The search field is static, but still behaves like a real form input.
const searchQuery = ref('SC-000245')

// The current list page shows how pagination will look in the future app.
const currentTicketsPage = ref(1)

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

// The detail timeline shows comments and audit entries on the ticket page.
const ticketTimeline = [
  { actor: 'Пользователь', text: 'Не могу открыть 1С на кассе после обновления.', time: '09:15' },
  { actor: 'Система', text: 'Заявка зарегистрирована и отправлена в отдел ИТ.', time: '09:18' },
  { actor: 'Ответственный', text: 'Проверяю обновление, вернусь с ответом через 10 минут.', time: '09:34' }
]

// The responsible selector demonstrates a future modal with paginated assignees.
const responsiblePeople = [
  { team: 'Отдел ИТ', person: 'Иван Петров', role: 'Старший инженер' },
  { team: 'Отдел ИТ', person: 'Елена Орлова', role: 'Системный аналитик' },
  { team: 'Маркетинг', person: 'Мария Соколова', role: 'Маркетолог' }
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
const isLoadingMyTickets = computed(() => store.getters[ticketGetterTypes.isLoadingMyTickets])
const isLoadingResponsibleTickets = computed(() => store.getters[ticketGetterTypes.isLoadingResponsibleTickets])
const ticketsErrors = computed(() => store.getters[ticketGetterTypes.ticketsError] || [])
const normalizedMyTickets = computed(() => {
  return storeMyTickets.value.length ? storeMyTickets.value : myTickets
})
const normalizedResponsibleTickets = computed(() => {
  const source = storeResponsibleTickets.value.length ? storeResponsibleTickets.value : responsibleTickets

  return source.map((ticket) => ({
    ...ticket,
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
  store.dispatch(ticketActionTypes.loadMyTickets)
  store.dispatch(ticketActionTypes.loadResponsibleTickets)
})

// This helper switches screens from the top navigation and action buttons.
function openScreen(screenId) {
  activeScreen.value = screenId
  submitBanner.value = ''
}

// This helper simulates navigation from list card to detail page.
function openTicketDetails(ticketNumber) {
  searchQuery.value = ticketNumber
  activeScreen.value = 'details'
}

// This helper only changes UI feedback for the visual prototype.
function simulateSubmit(message) {
  submitBanner.value = message
}

// This helper simulates the search result flow.
function simulateSearch() {
  activeScreen.value = 'details'
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
              <select>
                <option>Заявка в отдел ИТ</option>
                <option>Маркетинговая заявка</option>
              </select>
            </label>
            <label>
              Краткая тема
              <input type="text" value="Не открывается 1С на кассе" />
            </label>
            <label>
              Подробное описание
              <textarea rows="5">После обновления 1С не запускается на рабочем месте кассира. Нужна диагностика и восстановление работы.</textarea>
            </label>
            <label>
              Подразделение
              <select>
                <option>Отдел ИТ</option>
                <option>Маркетинг</option>
              </select>
            </label>
            <label>
              Исполнить до
              <input type="date" value="2026-04-11" />
            </label>

            <div class="upload-box">
              <div>
                <strong>Вложения</strong>
                <p>Скриншоты, фото, документы, голосовые сообщения.</p>
              </div>
              <button class="secondary-button">Добавить файл</button>
            </div>

            <div class="chip-list">
              <span class="file-chip">screenshot-1.png</span>
              <span class="file-chip">error-log.pdf</span>
            </div>

            <div class="hero-actions">
              <button class="primary-button" @click="simulateSubmit('Заявка подготовлена и отправлена в ITILIUM.')">
                Отправить заявку
              </button>
              <button class="secondary-button" @click="openScreen('home')">Отмена</button>
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

          <p v-else-if="ticketsErrors.length" class="status-pill rose">{{ ticketsErrors[0] }}</p>

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

          <p v-else-if="ticketsErrors.length" class="status-pill rose">{{ ticketsErrors[0] }}</p>

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
            <div class="hero-actions">
              <button class="primary-button" @click="simulateSearch">Найти заявку</button>
              <button class="secondary-button" @click="openScreen('details')">Открыть демо-карточку</button>
            </div>
          </article>
        </section>

        <section v-else-if="activeScreen === 'details'" class="screen">
          <div class="section-header">
            <div>
              <p class="eyebrow">Карточка заявки</p>
              <h2>{{ searchQuery }}</h2>
            </div>
            <span class="status-pill amber">В работе</span>
          </div>

          <article class="content-card">
            <div class="details-grid">
              <div>
                <span>Краткая тема</span>
                <strong>Не открывается 1С на кассе</strong>
              </div>
              <div>
                <span>Ответственная команда</span>
                <strong>Отдел ИТ</strong>
              </div>
              <div>
                <span>Срок</span>
                <strong>11.04.2026</strong>
              </div>
              <div>
                <span>Можно сменить ответственного</span>
                <strong>Да</strong>
              </div>
            </div>

            <div class="action-grid">
              <button class="secondary-button">Добавить комментарий</button>
              <button class="secondary-button">Поменять статус</button>
              <button class="secondary-button">Сменить ответственного</button>
              <button class="secondary-button">Оценить решение</button>
            </div>
          </article>

          <article class="content-card">
            <h3>Лента событий</h3>
            <div class="timeline">
              <div v-for="item in ticketTimeline" :key="`${item.actor}-${item.time}`" class="timeline-item">
                <strong>{{ item.actor }}</strong>
                <p>{{ item.text }}</p>
                <span>{{ item.time }}</span>
              </div>
            </div>
          </article>

          <article class="content-card">
            <h3>Новый комментарий</h3>
            <div class="form-card compact">
              <label>
                Текст комментария
                <textarea rows="4">Нужна срочная проверка после ночного обновления.</textarea>
              </label>
              <div class="hero-actions">
                <button class="primary-button" @click="simulateSubmit('Комментарий отправлен, ответ ITILIUM получен.')">
                  Отправить комментарий
                </button>
                <button class="secondary-button">Прикрепить файл</button>
              </div>
            </div>
          </article>

          <article class="content-card">
            <h3>Выбор нового ответственного</h3>
            <div class="selector-list">
              <div v-for="person in responsiblePeople" :key="person.person" class="selector-item">
                <div>
                  <strong>{{ person.person }}</strong>
                  <p>{{ person.team }} · {{ person.role }}</p>
                </div>
                <button class="ghost-button">Выбрать</button>
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

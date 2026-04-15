import { computed, ref, watch } from 'vue'

import {
  actionTypes as ticketActionTypes,
  getterTypes as ticketGetterTypes
} from '@/store/modules/tickets'

// useTicketFlow groups ticket-specific screen state, derived data and actions
// so App.vue can stay focused on auth bootstrap, layout and screen switching.
export function useTicketFlow({ store, currentUser, activeScreen, submitBanner }) {
  const defaultRequestType = 'Заявка в отдел ИТ'

  // Page size for «Мои заявки» when the list is built from ITILIUM `servicecalls`.
  const myTicketsPageSize = 5

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
    requestType: defaultRequestType,
    title: '',
    description: '',
    department: currentUser.value?.department || '',
    executionDate: '',
    attachments: []
  })

  const availableRequestTypes = computed(() => {
    const options = [defaultRequestType]

    if (currentUser.value?.canCreateMarketingRequests) {
      options.push('Маркетинговая заявка')
    }
    if (currentUser.value?.canCreateDaxRequests) {
      options.push('Заявка в DAX')
    }

    return options
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

  const storeMyTickets = computed(() => store.getters[ticketGetterTypes.myTickets] || [])
  const storeResponsibleTickets = computed(() => store.getters[ticketGetterTypes.responsibleTickets] || [])

  // When ITILIUM returns `servicecalls`, the list is driven by those numbers until a dedicated list API exists.
  const ticketsFromServiceCalls = computed(() => {
    const user = currentUser.value
    const ids = user?.servicecalls

    if (!user?.employeeFound || user?.registrationRequired || !Array.isArray(ids) || ids.length === 0) {
      return null
    }

    return ids
      .map((id) => String(id).trim())
      .filter(Boolean)
      .map((number) => ({
        number,
        title: `Заявка ${number}`,
        state: 'Откройте карточку',
        deadline: '—',
        tone: 'info'
      }))
  })
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
    const fromCalls = ticketsFromServiceCalls.value

    if (fromCalls) {
      return fromCalls.map((ticket) => ({
        ...ticket,
        tone: ticket.tone || resolveTicketTone(ticket.state)
      }))
    }

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
      },
      {
        title: 'Мои номера заявок',
        value: String(currentUser.value?.servicecalls?.length || normalizedMyTickets.value.length),
        tone: 'slate'
      }
    ]
  })

  watch(() => currentUser.value?.servicecalls, () => {
    currentTicketsPage.value = 1
  })

  watch([currentUser, availableRequestTypes], ([user, requestTypes]) => {
    if (!requestTypes.includes(createTicketForm.value.requestType)) {
      createTicketForm.value.requestType = requestTypes[0] || defaultRequestType
    }

    createTicketForm.value = {
      ...createTicketForm.value,
      department: createTicketForm.value.department || user?.department || ''
    }
  }, { immediate: true })

  // A computed slice is enough to show how page navigation will behave.
  const paginatedTickets = computed(() => {
    const start = (currentTicketsPage.value - 1) * myTicketsPageSize
    return normalizedMyTickets.value.slice(start, start + myTicketsPageSize)
  })

  // The page count drives the pager buttons in the static demo.
  const pageCount = computed(() => Math.ceil(normalizedMyTickets.value.length / myTicketsPageSize))

  const myTicketsListSource = computed(() => (ticketsFromServiceCalls.value ? 'servicecalls' : 'store'))
  const detailStatusTone = computed(() => resolveTicketTone(selectedTicket.value?.state))
  const availableStatusOptions = computed(() => selectedTicket.value?.availableStates || [])
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

  function loadTicketLists() {
    store.dispatch(ticketActionTypes.loadMyTickets)
    store.dispatch(ticketActionTypes.loadResponsibleTickets)
  }

  // This helper simulates navigation from list card to detail page.
  async function openTicketDetails(ticketNumber) {
    searchQuery.value = ticketNumber
    activeScreen.value = 'details'
    await store.dispatch(ticketActionTypes.loadTicketDetails, ticketNumber)
  }

  // Search now goes through the dedicated backend endpoint so the details screen
  // can open from either a manual search or a list click using shared Vuex state.
  async function searchTicketByNumber() {
    const number = searchQuery.value.trim()

    if (!number) {
      return
    }

    const response = await store.dispatch(ticketActionTypes.searchTicket, {
      number
    })

    if (response?.data?.success) {
      activeScreen.value = 'details'
    }
  }

  async function submitCreateTicket() {
    const response = await store.dispatch(ticketActionTypes.createTicket, {
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

  function setCreateExecutionDate(value) {
    createTicketForm.value.executionDate = value || ''
  }

  function addCreateAttachments(files) {
    const nextFiles = Array.from(files || [])
      .map((file) => file?.name || '')
      .filter(Boolean)

    createTicketForm.value.attachments = [
      ...createTicketForm.value.attachments,
      ...nextFiles
    ]
  }

  function removeCreateAttachment(fileName) {
    createTicketForm.value.attachments = createTicketForm.value.attachments.filter((item) => item !== fileName)
  }

  async function submitComment() {
    if (!selectedTicket.value?.number || !commentDraft.value.trim()) {
      return
    }

    const response = await store.dispatch(ticketActionTypes.addComment, {
      number: selectedTicket.value.number,
      data: {
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
        responsibleId
      }
    })

    if (response?.data?.success) {
      submitBanner.value = 'Ответственный по заявке обновлен.'
      return
    }

    selectedResponsibleId.value = ''
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

  return {
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
    availableRequestTypes,
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
    setCreateExecutionDate,
    addCreateAttachments,
    removeCreateAttachment,
    submitComment,
    submitStatusChange,
    assignResponsible,
    setTicketsPage,
    setSearchQuery,
    setCommentDraft,
    myTicketsListSource
  }
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

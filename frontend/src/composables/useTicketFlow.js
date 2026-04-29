import { computed, ref, watch } from 'vue'
import Swal from 'sweetalert2'

import {
  actionTypes as ticketActionTypes,
  getterTypes as ticketGetterTypes
} from '@/store/modules/tickets'

/**
 * Композабл заявок: списки из API или из servicecalls профиля, карточка, поиск, создание, комментарии.
 * Содержит запасные статические списки для демо, если store ещё пустой.
 */
export function useTicketFlow({ store, currentUser, activeScreen, submitBanner }) {
  const defaultRequestType = 'Заявка в отдел ИТ'

  // Page size for «Мои заявки» when the list is built from ITILIUM `servicecalls`.
  const myTicketsPageSize = 10

  // The search field persists the current ticket number between list and search flows.
  const searchQuery = ref('')

  // The current list page shows how pagination will look in the future app.
  const currentTicketsPage = ref(1)

  // Текст комментария в карточке; вложения — отдельным списком File (multipart на бэкенд).
  const commentDraft = ref('')
  const commentAttachmentFiles = ref([])
  /** Увеличивается после успешной отправки комментария — экран карточки закрывает панель и показывает короткое уведомление. */
  const commentSuccessTick = ref(0)
  /** После успешной отправки оценки — закрыть панель оценки в карточке. */
  const ratingSuccessTick = ref(0)
  const detailsOrigin = ref('search')

  // The status form mirrors the backend transition contract used by the ticket card.
  const statusForm = ref({
    state: '',
    comment: '',
    date: ''
  })

  // The responsible selector keeps the chosen ITILIUM assignee id before submit.
  const selectedResponsibleId = ref('')

  // Форма создания заявки: текстовые поля + реальные объекты File для multipart (см. `tickets.createTicket`).
  const createTicketForm = ref({
    requestType: defaultRequestType,
    title: '',
    description: '',
    department: currentUser.value?.department || '',
    executionDate: '',
    /** @type {File[]} */
    attachmentFiles: []
  })
  // Динамические поля маркетинговой формы приходят из 1С по formNumber.
  const marketingFormData = ref({})
  const selectedMarketingService = ref(null)
  const createValidationErrors = ref({})
  const createValidationStarted = ref(false)
  const createSubmitError = ref('')

  const availableRequestTypes = computed(() => {
    const options = [defaultRequestType]
    const canCreateMarketing = Boolean(
      currentUser.value?.canCreateMarketingRequests ||
      currentUser.value?.canCreateMarketing ||
      currentUser.value?.can_marketing
    )
    const canCreateDax = Boolean(
      currentUser.value?.canCreateDaxRequests ||
      currentUser.value?.canCreateDax ||
      currentUser.value?.can_dax
    )
    const hasMarketingServices = marketingServices.value.length > 0

    // Если профильный флаг нестандартный, но 1С вернул маркетинговые типы — считаем маркетинг доступным.
    if (canCreateMarketing || hasMarketingServices) {
      options.push('Маркетинговая заявка')
    }
    if (canCreateDax) {
      options.push('Заявка в DAX')
    }

    return options
  })

  // The responsible selector demonstrates a future modal with paginated assignees.
  const responsiblePeople = [
    { team: 'Отдел ИТ', person: 'Иван Петров', role: 'Старший инженер', externalId: 'emp-1' },
    { team: 'Отдел ИТ', person: 'Елена Орлова', role: 'Системный аналитик', externalId: 'emp-2' },
    { team: 'Маркетинг', person: 'Мария Соколова', role: 'Маркетолог', externalId: 'emp-3' }
  ]

  const storeMyTickets = computed(() => store.getters[ticketGetterTypes.myTickets] || [])
  const storeResponsibleTickets = computed(() => store.getters[ticketGetterTypes.responsibleTickets] || [])

  // Экран «Мои заявки» строим только из servicecalls профиля.
  const ticketsFromServiceCalls = computed(() => {
    const user = currentUser.value
    const ids = user?.servicecalls

    if (!user?.employeeFound || user?.registrationRequired || !Array.isArray(ids) || ids.length === 0) {
      return []
    }

    const uniqueIds = Array.from(new Set(
      ids
        .map((id) => String(id).trim())
        .filter(Boolean)
    ))

    return uniqueIds
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
  const isSubmittingTicketRating = computed(() => store.getters[ticketGetterTypes.isSubmittingTicketRating])
  const listErrors = computed(() => store.getters[ticketGetterTypes.listError] || [])
  const ticketErrors = computed(() => store.getters[ticketGetterTypes.ticketError] || [])
  const createErrors = computed(() => store.getters[ticketGetterTypes.createError] || [])
  const marketingErrors = computed(() => store.getters[ticketGetterTypes.marketingError] || [])
  const marketingServices = computed(() => store.getters[ticketGetterTypes.marketingServices] || [])
  const marketingSubdivisions = computed(() => store.getters[ticketGetterTypes.marketingSubdivisions] || [])
  const isLoadingMarketingServices = computed(() => store.getters[ticketGetterTypes.isLoadingMarketingServices])
  const isLoadingMarketingSubdivisions = computed(() => store.getters[ticketGetterTypes.isLoadingMarketingSubdivisions])
  const isCreatingMarketingRequest = computed(() => store.getters[ticketGetterTypes.isCreatingMarketingRequest])
  const createErrorMessage = computed(() => {
    if (createSubmitError.value) {
      return createSubmitError.value
    }

    return createErrors.value[0] || ''
  })
  const marketingErrorMessage = computed(() => marketingErrors.value[0] || '')
  const currentMarketingSchema = computed(() => selectedMarketingService.value?.formSchema || null)

  const normalizedMyTickets = computed(() => {
    const source = storeMyTickets.value.length ? storeMyTickets.value : ticketsFromServiceCalls.value

    return source.map((ticket) => {
      const merged = {
        ...ticket
      }

      return {
        ...merged,
        tone: merged.tone || resolveTicketTone(merged.state)
      }
    })
  })

  const normalizedResponsibleTickets = computed(() => {
    const source = storeResponsibleTickets.value

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

  watch(
    () => activeScreen.value,
    (screen) => {
      if (screen === 'myTickets') {
        store.dispatch(ticketActionTypes.loadMyTickets)
      }
      if (screen === 'responsible') {
        store.dispatch(ticketActionTypes.loadResponsibleTickets)
      }
    }
  )

  watch([currentUser, availableRequestTypes], ([user, requestTypes]) => {
    if (!requestTypes.includes(createTicketForm.value.requestType)) {
      createTicketForm.value.requestType = requestTypes[0] || defaultRequestType
    }

    createTicketForm.value = {
      ...createTicketForm.value,
      department: createTicketForm.value.department || user?.department || ''
    }
  }, { immediate: true })

  // При выборе маркетингового типа подгружаем справочники 1С и инициализируем шаги wizard.
  watch(
    () => createTicketForm.value.requestType,
    async (requestType) => {
      if (requestType !== 'Маркетинговая заявка') {
        selectedMarketingService.value = null
        marketingFormData.value = {}
        return
      }

      await loadMarketingReferenceData()
    },
    { immediate: true }
  )

  watch(
    () => activeScreen.value,
    async (screen) => {
      if (screen !== 'create') {
        return
      }
      // Пробуем заранее подтянуть маркетинговые типы: это даёт признак доступа даже при нестандартных флагах профиля.
      await store.dispatch(ticketActionTypes.loadMarketingServices)
    }
  )

  watch(
    () => [
      createTicketForm.value.requestType,
      createTicketForm.value.title,
      createTicketForm.value.description,
      createTicketForm.value.department,
      createTicketForm.value.executionDate
    ],
    () => {
      if (createValidationStarted.value) {
        createValidationErrors.value = validateCreateTicketForm(
          createTicketForm.value,
          selectedMarketingService.value,
          marketingFormData.value
        )
        if (Object.keys(createValidationErrors.value).length === 0) {
          createSubmitError.value = ''
        }
      }
    }
  )

  // A computed slice is enough to show how page navigation will behave.
  const paginatedTickets = computed(() => {
    const start = (currentTicketsPage.value - 1) * myTicketsPageSize
    return normalizedMyTickets.value.slice(start, start + myTicketsPageSize)
  })

  // The page count drives the pager buttons in the static demo.
  const pageCount = computed(() => Math.ceil(normalizedMyTickets.value.length / myTicketsPageSize))

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
  watch(selectedTicket, (ticket) => {
    commentDraft.value = ''
    commentAttachmentFiles.value = []
    selectedResponsibleId.value = ''
    statusForm.value = {
      state: ticket?.availableStates?.[0] || '',
      comment: '',
      date: ticket?.deadline || ''
    }

  }, { immediate: true })

  function loadTicketLists() {
    // Для «Мои заявки» используем backend list_sc (через /api/v1/tickets) и fallback на servicecalls.
    store.dispatch(ticketActionTypes.loadMyTickets)
  }

  // This helper simulates navigation from list card to detail page.
  async function openTicketDetails(ticketNumber, source = 'search') {
    detailsOrigin.value = source
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
      detailsOrigin.value = 'search'
      activeScreen.value = 'details'
    }
  }

  async function loadMarketingReferenceData() {
    await Promise.all([
      store.dispatch(ticketActionTypes.loadMarketingServices),
      store.dispatch(ticketActionTypes.loadMarketingSubdivisions)
    ])

    if (!selectedMarketingService.value && marketingServices.value.length > 0) {
      selectedMarketingService.value = marketingServices.value[0]
    }
    if (!createTicketForm.value.department && marketingSubdivisions.value.length > 0) {
      createTicketForm.value.department = marketingSubdivisions.value[0]?.name || ''
    }
  }

  function setMarketingService(serviceCode) {
    const service = marketingServices.value.find((item) => item.code === serviceCode) || null
    selectedMarketingService.value = service
    marketingFormData.value = {}
  }

  function setMarketingFieldValue(key, value) {
    marketingFormData.value = {
      ...marketingFormData.value,
      [key]: value
    }
  }

  async function submitCreateTicket() {
    createValidationStarted.value = true
    createValidationErrors.value = validateCreateTicketForm(
      createTicketForm.value,
      selectedMarketingService.value,
      marketingFormData.value
    )
    if (Object.keys(createValidationErrors.value).length > 0) {
      createSubmitError.value = 'Заполните обязательные поля формы.'
      return
    }

    createSubmitError.value = ''
    const isMarketingFlow = createTicketForm.value.requestType === 'Маркетинговая заявка'
    const response = isMarketingFlow
      ? await store.dispatch(ticketActionTypes.createMarketingRequest, {
        serviceCode: selectedMarketingService.value?.code || '',
        formNumber: selectedMarketingService.value?.formNumber || '',
        subdivision: createTicketForm.value.department,
        executionDate: createTicketForm.value.executionDate,
        withoutDate: false,
        formData: marketingFormData.value,
        attachmentFiles: createTicketForm.value.attachmentFiles || []
      })
      : await store.dispatch(ticketActionTypes.createTicket, {
        requestType: createTicketForm.value.requestType,
        title: createTicketForm.value.title,
        description: createTicketForm.value.description,
        department: createTicketForm.value.department,
        executionDate: createTicketForm.value.executionDate,
        attachmentFiles: createTicketForm.value.attachmentFiles || []
      })

    if (response?.data?.success) {
      // После успешного создания очищаем вложения, чтобы не тащить File в следующую заявку.
      createTicketForm.value.attachmentFiles = []
      marketingFormData.value = {}
      searchQuery.value = response?.data?.data?.number || ''
      submitBanner.value = 'Заявка создана и открыта в карточке.'
      activeScreen.value = 'details'
    }
  }

  function setCreateExecutionDate(value) {
    createTicketForm.value.executionDate = value || ''
  }

  // Добавляем выбранные в браузере файлы в список перед отправкой (не только имена).
  function addCreateAttachments(files) {
    const nextFiles = Array.from(files || []).filter((f) => f instanceof File)
    createTicketForm.value.attachmentFiles = [...(createTicketForm.value.attachmentFiles || []), ...nextFiles]
  }

  function removeCreateAttachment(index) {
    const list = [...(createTicketForm.value.attachmentFiles || [])]
    list.splice(index, 1)
    createTicketForm.value.attachmentFiles = list
  }

  async function submitComment() {
    if (!selectedTicket.value?.number) {
      return
    }

    const text = commentDraft.value.trim()
    const files = commentAttachmentFiles.value || []
    if (!text && files.length === 0) {
      return
    }

    const response = await store.dispatch(ticketActionTypes.addComment, {
      number: selectedTicket.value.number,
      data: {
        message: text,
        attachments: [],
        attachmentFiles: files
      }
    })

    if (response?.data?.success) {
      commentDraft.value = ''
      commentAttachmentFiles.value = []
      commentSuccessTick.value += 1
    }
  }

  function addCommentAttachments(fileList) {
    const next = Array.from(fileList || []).filter((f) => f instanceof File)
    commentAttachmentFiles.value = [...(commentAttachmentFiles.value || []), ...next]
  }

  function removeCommentAttachment(index) {
    const list = [...(commentAttachmentFiles.value || [])]
    list.splice(index, 1)
    commentAttachmentFiles.value = list
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

    const confirm = await Swal.fire({
      title: 'Смена ответственного',
      text: 'Назначить выбранного сотрудника ответственным по этой заявке?',
      icon: 'question',
      showCancelButton: true,
      confirmButtonText: 'Да, назначить',
      cancelButtonText: 'Отмена',
      reverseButtons: true,
      focusCancel: true
    })
    if (!confirm.isConfirmed) {
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

  async function requestResponsibleOptions(number) {
    const n = String(number || '').trim()
    if (!n) {
      return
    }
    await store.dispatch(ticketActionTypes.loadResponsibleOptions, n)
  }

  async function submitTicketRating(payload) {
    if (!selectedTicket.value?.number || payload?.mark === undefined || payload?.mark === null) {
      return
    }

    const response = await store.dispatch(ticketActionTypes.confirmTicket, {
      number: selectedTicket.value.number,
      data: {
        mark: payload.mark,
        comment: payload.comment || ''
      }
    })

    if (response?.data?.success) {
      ratingSuccessTick.value += 1
      submitBanner.value = 'Оценка отправлена.'
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

  return {
    searchQuery,
    currentTicketsPage,
    commentDraft,
    commentAttachmentFiles,
    commentSuccessTick,
    ratingSuccessTick,
    detailsOrigin,
    statusForm,
    selectedResponsibleId,
    createTicketForm,
    marketingFormData,
    selectedMarketingService,
    isLoadingMyTickets,
    isLoadingResponsibleTickets,
    isCreatingTicket,
    isLoadingTicketDetails,
    isLoadingResponsibleOptions,
    isSubmittingComment,
    isChangingStatus,
    isChangingResponsible,
    isSubmittingTicketRating,
    listErrors,
    ticketErrors,
    createErrors,
    createValidationErrors,
    createValidationStarted,
    createErrorMessage,
    marketingErrorMessage,
    marketingServices,
    marketingSubdivisions,
    isLoadingMarketingServices,
    isLoadingMarketingSubdivisions,
    isCreatingMarketingRequest,
    currentMarketingSchema,
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
    loadMarketingReferenceData,
    setMarketingService,
    setMarketingFieldValue,
    setCreateExecutionDate,
    addCreateAttachments,
    removeCreateAttachment,
    submitComment,
    addCommentAttachments,
    removeCommentAttachment,
    submitStatusChange,
    assignResponsible,
    requestResponsibleOptions,
    submitTicketRating,
    setTicketsPage,
    setSearchQuery,
    setCommentDraft
  }
}

function validateCreateTicketForm(form, selectedMarketingService, marketingFormData) {
  const errors = {}

  if (!String(form.requestType || '').trim()) {
    errors.requestType = 'Выберите тип заявки.'
  }
  if (!String(form.title || '').trim()) {
    errors.title = 'Укажите краткую тему.'
  }
  if (!String(form.description || '').trim()) {
    errors.description = 'Добавьте подробное описание.'
  }
  if (!String(form.department || '').trim()) {
    errors.department = 'Выберите подразделение.'
  }
  if (!String(form.executionDate || '').trim()) {
    errors.executionDate = 'Выберите дату исполнения.'
  }

  if (form.requestType === 'Маркетинговая заявка') {
    if (!selectedMarketingService?.code) {
      errors.serviceCode = 'Выберите тип маркетинговой заявки.'
    }
    if (!/^\d{2}\.\d{2}\.\d{4}$/.test(String(form.executionDate || '').trim())) {
      errors.executionDate = 'Укажите дату в формате ДД.ММ.ГГГГ.'
    }

    const requiredFields = selectedMarketingService?.formSchema?.fields?.filter((field) => field.required) || []
    requiredFields.forEach((field) => {
      if (!String(marketingFormData?.[field.key] || '').trim()) {
        errors[`form_${field.key}`] = `Заполните поле «${field.label}».`
      }
    })
  }

  return errors
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

import { computed, ref, watch } from 'vue'
import { confirmAction } from '@/helpers/confirmDialog'
import { withBusyModal } from '@/helpers/busyModal'
import { validateStatusTransition } from '@/helpers/ticketWorkflow'
import { prepareAttachmentFiles } from '@/utils/prepareAttachmentFiles'

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
  // Пагинация ленты комментариев на карточке (ответ list_comment целиком, режем на клиенте).
  const commentsPageSize = 5
  const currentCommentsPage = ref(1)

  // The search field persists the current ticket number between list and search flows.
  const searchQuery = ref('')

  // The current list page shows how pagination will look in the future app.
  const currentTicketsPage = ref(1)

  // Текст комментария в карточке; вложения — отдельным списком File (multipart на бэкенд).
  const commentDraft = ref('')
  const commentAttachmentFiles = ref([])
  /** Пока сжимаем фото с камеры перед показом в списке вложений. */
  const createAttachmentsPreparing = ref(false)
  /** Увеличивается после успешной отправки комментария — экран карточки закрывает панель и показывает короткое уведомление. */
  const commentSuccessTick = ref(0)
  /** После успешной отправки оценки — закрыть панель оценки в карточке. */
  const ratingSuccessTick = ref(0)
  /** После успешной смены ответственного — закрыть панель выбора и показать уведомление. */
  const responsibleSuccessTick = ref(0)
  const detailsOrigin = ref('search')

  // The status form mirrors the backend transition contract used by the ticket card.
  const statusForm = ref({
    state: '',
    comment: '',
    date: ''
  })

  // The responsible selector keeps the chosen ITILIUM assignee id before submit.
  const selectedResponsibleId = ref('')

  // Форма создания заявки: маркетинговый поток использует department/executionDate как обязательные общие поля.
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

  const storeMyTickets = computed(() => store.getters[ticketGetterTypes.myTickets] || [])
  const storeResponsibleTickets = computed(() => store.getters[ticketGetterTypes.responsibleTickets] || [])

  // Если backend не смог получить краткие карточки из 1С, показываем хотя бы номера из servicecalls профиля.
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
        title: 'Тема загрузится после открытия',
        state: 'Статус не загружен',
        creationDate: '',
        deadline: '—',
        tone: resolveTicketTone('')
      }))
  })
  const selectedTicket = computed(() => store.getters[ticketGetterTypes.selectedTicket] || null)
  const storeTicketComments = computed(() => store.getters[ticketGetterTypes.ticketComments] || [])
  const responsibleOptions = computed(() => store.getters[ticketGetterTypes.responsibleOptions] || [])
  const isLoadingMyTickets = computed(() => store.getters[ticketGetterTypes.isLoadingMyTickets])
  const isLoadingResponsibleTickets = computed(() => store.getters[ticketGetterTypes.isLoadingResponsibleTickets])
  const isCreatingTicket = computed(() => store.getters[ticketGetterTypes.isCreatingTicket])
  const isLoadingTicketDetails = computed(() => store.getters[ticketGetterTypes.isLoadingTicketDetails])
  const isLoadingTicketComments = computed(() => store.getters[ticketGetterTypes.isLoadingTicketComments])
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
        title: normalizeTicketTitle(merged),
        creationDate: normalizeTicketCreationDate(merged),
        state: normalizeTicketState(merged.state),
        tone: merged.tone || resolveTicketTone(merged.state)
      }
    })
  })

  const normalizedResponsibleTickets = computed(() => {
    const source = storeResponsibleTickets.value

    return source.map((ticket) => ({
      ...ticket,
      title: normalizeTicketTitle(ticket),
      creationDate: normalizeTicketCreationDate(ticket),
      state: normalizeTicketState(ticket.state),
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

  // При выборе маркетингового типа подгружаем справочники 1С для динамической части формы.
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
      createTicketForm.value.executionDate,
      selectedMarketingService.value?.code,
      JSON.stringify(marketingFormData.value)
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
  const availableResponsibleOptions = computed(() => responsibleOptions.value)

  const selectedTicketTimeline = computed(() => {
    // Приоритет — отдельный list_comment; fallback — timeline из карточки (демо/после add_comment).
    const source = storeTicketComments.value.length
      ? storeTicketComments.value
      : (selectedTicket.value?.timeline || [])
    const myName = normalizePersonName(currentUser.value?.fullName || '')

    return source.map((item) => {
      const actor = item.author || item.actor || 'Система'
      return {
        actor,
        text: item.message || item.text || '',
        time: formatTimelineTime(item.createdAt || item.time || ''),
        // Свои сообщения — справа в «чате», чужие — слева.
        isMine: Boolean(myName) && normalizePersonName(actor) === myName
      }
    })
  })

  const commentsPageCount = computed(() => Math.ceil(selectedTicketTimeline.value.length / commentsPageSize) || 0)

  const paginatedTicketComments = computed(() => {
    const start = (currentCommentsPage.value - 1) * commentsPageSize
    return selectedTicketTimeline.value.slice(start, start + commentsPageSize)
  })

  // When a new detail payload arrives, derived forms stay in sync with the same
  // Vuex source of truth instead of preserving stale values from the previous ticket.
  watch(selectedTicket, (ticket) => {
    commentDraft.value = ''
    commentAttachmentFiles.value = []
    selectedResponsibleId.value = ''
    currentCommentsPage.value = 1
    statusForm.value = {
      state: ticket?.availableStates?.[0] || '',
      comment: '',
      date: ''
    }

  }, { immediate: true })

  watch(commentsPageCount, (count) => {
    if (count > 0 && currentCommentsPage.value > count) {
      currentCommentsPage.value = count
    }
  })

  function loadTicketLists() {
    // Для «Мои заявки» используем backend list_sc (через /api/v1/tickets) и fallback на servicecalls.
    store.dispatch(ticketActionTypes.loadMyTickets)
  }

  // This helper simulates navigation from list card to detail page.
  async function openTicketDetails(ticketNumber, source = 'search') {
    detailsOrigin.value = source
    searchQuery.value = ticketNumber
    activeScreen.value = 'details'
    currentCommentsPage.value = 1
    await store.dispatch(ticketActionTypes.loadTicketDetails, ticketNumber)
    // Комментарии подгружаем параллельно после карточки — ошибка list_comment не закрывает details.
    store.dispatch(ticketActionTypes.loadTicketComments, ticketNumber)
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
      currentCommentsPage.value = 1
      store.dispatch(ticketActionTypes.loadTicketComments, number)
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
    if (marketingSubdivisions.value.length > 0) {
      const selectedDepartment = String(createTicketForm.value.department || '').trim()
      const hasSelectedDepartment = marketingSubdivisions.value.some((item) => item?.name === selectedDepartment)
      if (!hasSelectedDepartment) {
        createTicketForm.value.department = marketingSubdivisions.value[0]?.name || ''
      }
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
    const payload = isMarketingFlow
      ? {
        serviceCode: selectedMarketingService.value?.code || '',
        formNumber: selectedMarketingService.value?.formNumber || '',
        subdivision: createTicketForm.value.department,
        executionDate: createTicketForm.value.executionDate,
        withoutDate: false,
        formData: marketingFormData.value,
        attachmentFiles: createTicketForm.value.attachmentFiles || []
      }
      : {
        requestType: createTicketForm.value.requestType,
        title: createTicketForm.value.title,
        description: createTicketForm.value.description,
        department: createTicketForm.value.department,
        executionDate: createTicketForm.value.executionDate,
        attachmentFiles: createTicketForm.value.attachmentFiles || []
      }

    try {
      const response = await withBusyModal(
        'Создаём заявку…',
        () => (isMarketingFlow
          ? store.dispatch(ticketActionTypes.createMarketingRequest, payload)
          : store.dispatch(ticketActionTypes.createTicket, payload))
      )

      if (!response?.data?.success) {
        createSubmitError.value = 'Сервер вернул неожиданный ответ. Повторите отправку.'
        return
      }

      // После успешного создания очищаем вложения, чтобы не тащить File в следующую заявку.
      createTicketForm.value.attachmentFiles = []
      marketingFormData.value = {}
      const ticketNumber = String(response?.data?.data?.number || '').trim()
      searchQuery.value = ticketNumber
      submitBanner.value = ticketNumber
        ? 'Заявка создана и открыта в карточке.'
        : 'Заявка передана в ITILIUM. Номер появится в списке «Мои заявки».'
      activeScreen.value = ticketNumber ? 'details' : 'home'
      if (!ticketNumber) {
        await store.dispatch(ticketActionTypes.loadMyTickets)
      }
    } catch {
      // Ошибка уже в createErrors / marketingErrors через Vuex.
    }
  }

  function setCreateExecutionDate(value) {
    createTicketForm.value.executionDate = value || ''
  }

  // Сразу сжимаем фото: с камеры в MAX часто уходит несколько МБ, хотя в галерее видно «400 КБ».
  async function addCreateAttachments(files) {
    const nextFiles = Array.from(files || []).filter((f) => f instanceof File)
    if (!nextFiles.length) {
      return
    }
    createAttachmentsPreparing.value = true
    try {
      const prepared = await prepareAttachmentFiles(nextFiles)
      createTicketForm.value.attachmentFiles = [
        ...(createTicketForm.value.attachmentFiles || []),
        ...prepared
      ]
    } finally {
      createAttachmentsPreparing.value = false
    }
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
      submitBanner.value = 'Комментарий успешно отправлен.'
      commentSuccessTick.value += 1
      // После записи обновляем ленту из list_comment — find_sc timeline не заполняет.
      await store.dispatch(ticketActionTypes.loadTicketComments, selectedTicket.value.number)
      currentCommentsPage.value = 1
    }
  }

  function setCommentsPage(page) {
    const next = Number(page) || 1
    if (next < 1 || (commentsPageCount.value > 0 && next > commentsPageCount.value)) {
      return
    }
    currentCommentsPage.value = next
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

    const validationError = validateStatusTransition(statusForm.value.state, statusForm.value)
    if (validationError) {
      return
    }

    const response = await withBusyModal('Меняем статус заявки…', () => store.dispatch(ticketActionTypes.changeStatus, {
      number: selectedTicket.value.number,
      data: {
        state: statusForm.value.state,
        comment: statusForm.value.comment,
        date: statusForm.value.date
      }
    }))

    if (response?.data?.success) {
      submitBanner.value = 'Статус заявки обновлен.'
    }
    return response
  }

  async function assignResponsible(selection) {
    const responsibleId = typeof selection === 'string' ? selection : selection?.responsibleId
    const teamId = typeof selection === 'object' && selection ? selection.teamId : ''
    if (!selectedTicket.value?.number || !responsibleId) {
      return
    }

    const confirm = await confirmAction({
      title: 'Смена ответственного',
      text: 'Назначить выбранного сотрудника ответственным по этой заявке?',
      confirmButtonText: 'Да, назначить',
      cancelButtonText: 'Отмена'
    })
    if (!confirm.isConfirmed) {
      return
    }

    selectedResponsibleId.value = responsibleId

    const response = await withBusyModal('Назначаем ответственного…', () => store.dispatch(ticketActionTypes.changeResponsible, {
      number: selectedTicket.value.number,
      data: {
        responsibleId,
        ...(teamId ? { teamId } : {})
      }
    }))

    if (response?.data?.success) {
      submitBanner.value = 'Ответственный по заявке обновлен.'
      responsibleSuccessTick.value += 1
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
    responsibleSuccessTick,
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
    isLoadingTicketComments,
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
    paginatedTicketComments,
    commentsPageCount,
    currentCommentsPage,
    loadTicketLists,
    openTicketDetails,
    searchTicketByNumber,
    submitCreateTicket,
    loadMarketingReferenceData,
    setMarketingService,
    setMarketingFieldValue,
    setCreateExecutionDate,
    addCreateAttachments,
    createAttachmentsPreparing,
    removeCreateAttachment,
    submitComment,
    addCommentAttachments,
    removeCommentAttachment,
    submitStatusChange,
    assignResponsible,
    requestResponsibleOptions,
    submitTicketRating,
    setTicketsPage,
    setCommentsPage,
    setSearchQuery,
    setCommentDraft
  }
}

function validateCreateTicketForm(form, selectedMarketingService, marketingFormData) {
  const errors = {}

  if (!String(form.requestType || '').trim()) {
    errors.requestType = 'Выберите тип заявки.'
  }
  if (form.requestType === 'Маркетинговая заявка') {
    if (!selectedMarketingService?.code) {
      errors.serviceCode = 'Выберите тип маркетинговой заявки.'
    }
    if (!String(form.department || '').trim()) {
      errors.department = 'Выберите подразделение.'
    }
    if (!String(form.executionDate || '').trim()) {
      errors.executionDate = 'Укажите дату исполнения.'
    }

    const requiredFields = selectedMarketingService?.formSchema?.fields?.filter((field) => field.required) || []
    requiredFields.forEach((field) => {
      if (!String(marketingFormData?.[field.key] || '').trim()) {
        errors[`form_${field.key}`] = `Заполните поле «${field.label}».`
      }
    })
  } else {
    if (!String(form.title || '').trim()) {
      errors.title = 'Укажите краткую тему.'
    }
    if (!String(form.description || '').trim()) {
      errors.description = 'Добавьте подробное описание.'
    }
  }

  return errors
}

function normalizeTicketTitle(ticket) {
  const title = String(ticket?.title || '').trim()
  const number = String(ticket?.number || '').trim()
  if (!title || title === `Заявка ${number}`) {
    return 'Тема не загружена'
  }
  return title
}

function normalizeTicketCreationDate(ticket) {
  return String(ticket?.creationDate || ticket?.createdAt || ticket?.dateCreate || '').trim()
}

function normalizeTicketState(state) {
  const text = String(state || '').trim()
  if (text === 'Откройте карточку') {
    return 'Статус не загружен'
  }
  return text || 'Статус не загружен'
}

function resolveTicketTone(state) {
  const normalized = String(state || '').toLowerCase()
  // Цвета меняются централизованно через классы status-pill.* в styles.css.
  if (normalized.includes('закрыт') || normalized.includes('выполн')) {
    return 'green'
  }
  if (normalized.includes('работ')) {
    return 'amber'
  }
  if (normalized.includes('соглас')) {
    return 'purple'
  }
  if (normalized.includes('отлож')) {
    return 'slate'
  }
  if (normalized.includes('ожидан') || normalized.includes('ответ')) {
    return 'amber'
  }
  if (normalized.includes('зарегистр')) {
    return 'blue'
  }
  return 'info'
}

function normalizePersonName(value) {
  return String(value || '')
    .trim()
    .toLowerCase()
    .replace(/\s+/g, ' ')
}

function formatTimelineTime(value) {
  if (!value) {
    return ''
  }

  const raw = String(value).trim()
  // Формат list_comment: «20.04.2026 22:24:35» — показываем как есть.
  if (/^\d{2}\.\d{2}\.\d{4}/.test(raw)) {
    return raw
  }

  const parsedDate = new Date(raw)

  if (Number.isNaN(parsedDate.getTime())) {
    return raw
  }

  return parsedDate.toLocaleString('ru-RU', {
    day: '2-digit',
    month: '2-digit',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  })
}

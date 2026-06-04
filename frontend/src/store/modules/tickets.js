/**
 * Vuex-модуль заявок: списки, карточка, поиск, комментарии и смена статуса/ответственного.
 * Данные приходят из ITILIUM через backend; при успешных мутациях списки на экране синхронизируются с карточкой.
 */
import ticketsApi from '@/api/tickets'

function isLongRunningClientFailure(error) {
  const code = String(error?.code || '').toUpperCase()
  const message = String(error?.message || '').toLowerCase()
  return (
    code === 'ECONNABORTED' ||
    code === 'ERR_NETWORK' ||
    /timeout/i.test(message) ||
    /network error/i.test(message) ||
    (!error?.response && message !== '')
  )
}

function normalizeTicketError(error) {
  const backendMessage = error?.response?.data?.message
  const statusCode = error?.response?.status
  const rawMessage = String(backendMessage || error?.message || '').trim()

  if (isLongRunningClientFailure(error)) {
    return 'Создание заявки в ITILIUM может занять до минуты. Если появилась ошибка — откройте «Мои заявки» и обновите список: заявка могла уже создаться.'
  }

  if (statusCode === 413 || /status code 413/i.test(rawMessage)) {
    return 'Вложение слишком большое. Один файл — до 15 МБ, не более 20 файлов за раз. Сожмите фото или отправьте без вложения.'
  }
  if (statusCode === 404 || /status 404/i.test(rawMessage)) {
    return 'Заявка не найдена в ITILIUM.'
  }
  if (statusCode === 400) {
    if (/ticket number is required/i.test(rawMessage)) {
      return 'Проверьте номер заявки и повторите запрос.'
    }
    if (backendMessage && String(backendMessage).trim()) {
      return String(backendMessage).trim()
    }
    return 'Проверьте номер заявки и повторите запрос.'
  }
  if (statusCode === 502 && backendMessage && String(backendMessage).trim()) {
    return String(backendMessage).trim()
  }
  if (/ticket number is required/i.test(rawMessage)) {
    return 'Проверьте номер заявки и повторите запрос.'
  }
  if (/mark must be between 0 and 5/i.test(rawMessage)) {
    return 'Оценка должна быть от 0 до 5.'
  }
  if (/comment is required for ratings 0 through 2/i.test(rawMessage)) {
    return 'Для оценок 0, 1 и 2 укажите комментарий.'
  }
  if (statusCode === 502 || statusCode === 504 || statusCode === 408) {
    return 'Сервер не успел ответить вовремя. Заявка могла уже создаться в ITILIUM — откройте «Мои заявки» и обновите список.'
  }
  if (statusCode >= 500) {
    return 'ITILIUM временно недоступен. Повторите попытку позже.'
  }
  if (!rawMessage) {
    return 'Не удалось выполнить запрос к ITILIUM.'
  }
  if (/itilium request failed with status/i.test(rawMessage)) {
    return 'Не удалось получить данные из ITILIUM. Повторите попытку позже.'
  }

  return rawMessage
}

const state = {
  myTickets: [],
  responsibleTickets: [],
  selectedTicket: null,
  responsibleOptions: [],
  isLoadingMyTickets: false,
  isLoadingResponsibleTickets: false,
  isLoadingTicketDetails: false,
  isLoadingResponsibleOptions: false,
  isCreatingTicket: false,
  isSubmittingComment: false,
  isChangingStatus: false,
  isChangingResponsible: false,
  isSubmittingTicketRating: false,
  isLoadingMarketingServices: false,
  isLoadingMarketingSubdivisions: false,
  isCreatingMarketingRequest: false,
  marketingServices: [],
  marketingSubdivisions: [],
  listError: [],
  ticketError: [],
  createError: [],
  marketingError: []
}

export const mutationTypes = {
  loadMyTicketsStart: '[tickets] loadMyTicketsStart',
  loadMyTicketsSuccess: '[tickets] loadMyTicketsSuccess',
  loadMyTicketsFail: '[tickets] loadMyTicketsFail',
  loadResponsibleTicketsStart: '[tickets] loadResponsibleTicketsStart',
  loadResponsibleTicketsSuccess: '[tickets] loadResponsibleTicketsSuccess',
  loadResponsibleTicketsFail: '[tickets] loadResponsibleTicketsFail',
  createTicketStart: '[tickets] createTicketStart',
  createTicketSuccess: '[tickets] createTicketSuccess',
  createTicketFail: '[tickets] createTicketFail',
  loadTicketDetailsStart: '[tickets] loadTicketDetailsStart',
  loadTicketDetailsSuccess: '[tickets] loadTicketDetailsSuccess',
  loadTicketDetailsFail: '[tickets] loadTicketDetailsFail',
  loadResponsibleOptionsStart: '[tickets] loadResponsibleOptionsStart',
  loadResponsibleOptionsSuccess: '[tickets] loadResponsibleOptionsSuccess',
  loadResponsibleOptionsFail: '[tickets] loadResponsibleOptionsFail',
  addCommentStart: '[tickets] addCommentStart',
  addCommentSuccess: '[tickets] addCommentSuccess',
  addCommentFail: '[tickets] addCommentFail',
  changeStatusStart: '[tickets] changeStatusStart',
  changeStatusSuccess: '[tickets] changeStatusSuccess',
  changeStatusFail: '[tickets] changeStatusFail',
  changeResponsibleStart: '[tickets] changeResponsibleStart',
  changeResponsibleSuccess: '[tickets] changeResponsibleSuccess',
  changeResponsibleFail: '[tickets] changeResponsibleFail',
  confirmTicketStart: '[tickets] confirmTicketStart',
  confirmTicketSuccess: '[tickets] confirmTicketSuccess',
  confirmTicketFail: '[tickets] confirmTicketFail',
  loadMarketingServicesStart: '[tickets] loadMarketingServicesStart',
  loadMarketingServicesSuccess: '[tickets] loadMarketingServicesSuccess',
  loadMarketingServicesFail: '[tickets] loadMarketingServicesFail',
  loadMarketingSubdivisionsStart: '[tickets] loadMarketingSubdivisionsStart',
  loadMarketingSubdivisionsSuccess: '[tickets] loadMarketingSubdivisionsSuccess',
  loadMarketingSubdivisionsFail: '[tickets] loadMarketingSubdivisionsFail',
  createMarketingRequestStart: '[tickets] createMarketingRequestStart',
  createMarketingRequestSuccess: '[tickets] createMarketingRequestSuccess',
  createMarketingRequestFail: '[tickets] createMarketingRequestFail'
}

export const actionTypes = {
  loadMyTickets: '[tickets] loadMyTickets',
  createTicket: '[tickets] createTicket',
  loadResponsibleTickets: '[tickets] loadResponsibleTickets',
  searchTicket: '[tickets] searchTicket',
  loadTicketDetails: '[tickets] loadTicketDetails',
  loadResponsibleOptions: '[tickets] loadResponsibleOptions',
  addComment: '[tickets] addComment',
  changeStatus: '[tickets] changeStatus',
  changeResponsible: '[tickets] changeResponsible',
  confirmTicket: '[tickets] confirmTicket',
  loadMarketingServices: '[tickets] loadMarketingServices',
  loadMarketingSubdivisions: '[tickets] loadMarketingSubdivisions',
  createMarketingRequest: '[tickets] createMarketingRequest'
}

export const getterTypes = {
  myTickets: '[tickets] myTickets',
  responsibleTickets: '[tickets] responsibleTickets',
  selectedTicket: '[tickets] selectedTicket',
  responsibleOptions: '[tickets] responsibleOptions',
  isLoadingMyTickets: '[tickets] isLoadingMyTickets',
  isLoadingResponsibleTickets: '[tickets] isLoadingResponsibleTickets',
  isCreatingTicket: '[tickets] isCreatingTicket',
  isLoadingTicketDetails: '[tickets] isLoadingTicketDetails',
  isLoadingResponsibleOptions: '[tickets] isLoadingResponsibleOptions',
  isSubmittingComment: '[tickets] isSubmittingComment',
  isChangingStatus: '[tickets] isChangingStatus',
  isChangingResponsible: '[tickets] isChangingResponsible',
  isSubmittingTicketRating: '[tickets] isSubmittingTicketRating',
  isLoadingMarketingServices: '[tickets] isLoadingMarketingServices',
  isLoadingMarketingSubdivisions: '[tickets] isLoadingMarketingSubdivisions',
  isCreatingMarketingRequest: '[tickets] isCreatingMarketingRequest',
  marketingServices: '[tickets] marketingServices',
  marketingSubdivisions: '[tickets] marketingSubdivisions',
  listError: '[tickets] listError',
  ticketError: '[tickets] ticketError',
  createError: '[tickets] createError',
  marketingError: '[tickets] marketingError'
}

// После загрузки полной карточки подменяем краткие поля в «Мои» / «Ответственные» без второго запроса списка.
function syncTicketSummaryList(list, ticket) {
  if (!ticket?.number) {
    return list
  }

  return list.map((item) => {
    if (item.number !== ticket.number) {
      return item
    }

    return {
      ...item,
      title: ticket.title,
      state: ticket.state,
      creationDate: ticket.creationDate,
      deadline: ticket.deadline,
      responsibleTeam: ticket.responsibleTeam
    }
  })
}

// Новая заявка после создания — в начало списка «Мои», дубликаты по номеру убираем.
function prependTicketSummary(list, ticket) {
  if (!ticket?.number) {
    return list
  }

  const summary = {
    number: ticket.number,
    title: ticket.title,
    state: ticket.state,
    creationDate: ticket.creationDate,
    deadline: ticket.deadline,
    responsibleTeam: ticket.responsibleTeam
  }

  return [summary, ...list.filter((item) => item.number !== ticket.number)]
}

const getters = {
  [getterTypes.myTickets]: (localState) => localState.myTickets,
  [getterTypes.responsibleTickets]: (localState) => localState.responsibleTickets,
  [getterTypes.selectedTicket]: (localState) => localState.selectedTicket,
  [getterTypes.responsibleOptions]: (localState) => localState.responsibleOptions,
  [getterTypes.isLoadingMyTickets]: (localState) => localState.isLoadingMyTickets,
  [getterTypes.isLoadingResponsibleTickets]: (localState) => localState.isLoadingResponsibleTickets,
  [getterTypes.isCreatingTicket]: (localState) => localState.isCreatingTicket,
  [getterTypes.isLoadingTicketDetails]: (localState) => localState.isLoadingTicketDetails,
  [getterTypes.isLoadingResponsibleOptions]: (localState) => localState.isLoadingResponsibleOptions,
  [getterTypes.isSubmittingComment]: (localState) => localState.isSubmittingComment,
  [getterTypes.isChangingStatus]: (localState) => localState.isChangingStatus,
  [getterTypes.isChangingResponsible]: (localState) => localState.isChangingResponsible,
  [getterTypes.isSubmittingTicketRating]: (localState) => localState.isSubmittingTicketRating,
  [getterTypes.isLoadingMarketingServices]: (localState) => localState.isLoadingMarketingServices,
  [getterTypes.isLoadingMarketingSubdivisions]: (localState) => localState.isLoadingMarketingSubdivisions,
  [getterTypes.isCreatingMarketingRequest]: (localState) => localState.isCreatingMarketingRequest,
  [getterTypes.marketingServices]: (localState) => localState.marketingServices,
  [getterTypes.marketingSubdivisions]: (localState) => localState.marketingSubdivisions,
  [getterTypes.listError]: (localState) => localState.listError,
  [getterTypes.ticketError]: (localState) => localState.ticketError,
  [getterTypes.createError]: (localState) => localState.createError,
  [getterTypes.marketingError]: (localState) => localState.marketingError
}

const mutations = {
  [mutationTypes.loadMyTicketsStart](localState) {
    localState.isLoadingMyTickets = true
    localState.listError = []
  },
  [mutationTypes.loadMyTicketsSuccess](localState, tickets) {
    localState.isLoadingMyTickets = false
    localState.myTickets = tickets
  },
  [mutationTypes.loadMyTicketsFail](localState, errors) {
    localState.isLoadingMyTickets = false
    localState.listError = errors
  },

  [mutationTypes.loadResponsibleTicketsStart](localState) {
    localState.isLoadingResponsibleTickets = true
    localState.listError = []
  },
  [mutationTypes.loadResponsibleTicketsSuccess](localState, tickets) {
    localState.isLoadingResponsibleTickets = false
    localState.responsibleTickets = tickets
  },
  [mutationTypes.loadResponsibleTicketsFail](localState, errors) {
    localState.isLoadingResponsibleTickets = false
    localState.listError = errors
  },

  [mutationTypes.createTicketStart](localState) {
    localState.isCreatingTicket = true
    localState.createError = []
  },
  [mutationTypes.createTicketSuccess](localState, ticket) {
    localState.isCreatingTicket = false
    localState.selectedTicket = ticket
    localState.myTickets = prependTicketSummary(localState.myTickets, ticket)
  },
  [mutationTypes.createTicketFail](localState, errors) {
    localState.isCreatingTicket = false
    localState.createError = errors
  },

  [mutationTypes.loadTicketDetailsStart](localState) {
    localState.isLoadingTicketDetails = true
    localState.selectedTicket = null
    localState.responsibleOptions = []
    localState.ticketError = []
  },
  [mutationTypes.loadTicketDetailsSuccess](localState, ticket) {
    localState.isLoadingTicketDetails = false
    localState.selectedTicket = ticket
    localState.myTickets = syncTicketSummaryList(localState.myTickets, ticket)
    localState.responsibleTickets = syncTicketSummaryList(localState.responsibleTickets, ticket)
  },
  [mutationTypes.loadTicketDetailsFail](localState, errors) {
    localState.isLoadingTicketDetails = false
    localState.selectedTicket = null
    localState.responsibleOptions = []
    localState.ticketError = errors
  },

  [mutationTypes.loadResponsibleOptionsStart](localState) {
    localState.isLoadingResponsibleOptions = true
    localState.responsibleOptions = []
  },
  [mutationTypes.loadResponsibleOptionsSuccess](localState, options) {
    localState.isLoadingResponsibleOptions = false
    localState.responsibleOptions = options
  },
  [mutationTypes.loadResponsibleOptionsFail](localState, errors) {
    localState.isLoadingResponsibleOptions = false
    localState.responsibleOptions = []
    // Ошибка загрузки списка ответственных не должна ломать уже открытую карточку заявки.
    // Эту ошибку показываем локально в блоке выбора ответственного (если нужно), но не как ticketError.
  },

  [mutationTypes.addCommentStart](localState) {
    localState.isSubmittingComment = true
    localState.ticketError = []
  },
  [mutationTypes.addCommentSuccess](localState, ticket) {
    localState.isSubmittingComment = false
    localState.selectedTicket = ticket
    localState.myTickets = syncTicketSummaryList(localState.myTickets, ticket)
    localState.responsibleTickets = syncTicketSummaryList(localState.responsibleTickets, ticket)
  },
  [mutationTypes.addCommentFail](localState, errors) {
    localState.isSubmittingComment = false
    localState.ticketError = errors
  },

  [mutationTypes.changeStatusStart](localState) {
    localState.isChangingStatus = true
    localState.ticketError = []
  },
  [mutationTypes.changeStatusSuccess](localState, ticket) {
    localState.isChangingStatus = false
    localState.selectedTicket = ticket
    localState.myTickets = syncTicketSummaryList(localState.myTickets, ticket)
    localState.responsibleTickets = syncTicketSummaryList(localState.responsibleTickets, ticket)
  },
  [mutationTypes.changeStatusFail](localState, errors) {
    localState.isChangingStatus = false
    localState.ticketError = errors
  },

  [mutationTypes.changeResponsibleStart](localState) {
    localState.isChangingResponsible = true
    localState.ticketError = []
  },
  [mutationTypes.changeResponsibleSuccess](localState, ticket) {
    localState.isChangingResponsible = false
    localState.selectedTicket = ticket
    localState.myTickets = syncTicketSummaryList(localState.myTickets, ticket)
    localState.responsibleTickets = syncTicketSummaryList(localState.responsibleTickets, ticket)
  },
  [mutationTypes.changeResponsibleFail](localState, errors) {
    localState.isChangingResponsible = false
    localState.ticketError = errors
  },

  [mutationTypes.confirmTicketStart](localState) {
    localState.isSubmittingTicketRating = true
    localState.ticketError = []
  },
  [mutationTypes.confirmTicketSuccess](localState, ticket) {
    localState.isSubmittingTicketRating = false
    localState.selectedTicket = ticket
    localState.myTickets = syncTicketSummaryList(localState.myTickets, ticket)
    localState.responsibleTickets = syncTicketSummaryList(localState.responsibleTickets, ticket)
  },
  [mutationTypes.confirmTicketFail](localState, errors) {
    localState.isSubmittingTicketRating = false
    localState.ticketError = errors
  },

  [mutationTypes.loadMarketingServicesStart](localState) {
    localState.isLoadingMarketingServices = true
    localState.marketingError = []
  },
  [mutationTypes.loadMarketingServicesSuccess](localState, services) {
    localState.isLoadingMarketingServices = false
    localState.marketingServices = services
  },
  [mutationTypes.loadMarketingServicesFail](localState, errors) {
    localState.isLoadingMarketingServices = false
    localState.marketingServices = []
    localState.marketingError = errors
  },

  [mutationTypes.loadMarketingSubdivisionsStart](localState) {
    localState.isLoadingMarketingSubdivisions = true
    localState.marketingError = []
  },
  [mutationTypes.loadMarketingSubdivisionsSuccess](localState, subdivisions) {
    localState.isLoadingMarketingSubdivisions = false
    localState.marketingSubdivisions = subdivisions
  },
  [mutationTypes.loadMarketingSubdivisionsFail](localState, errors) {
    localState.isLoadingMarketingSubdivisions = false
    localState.marketingSubdivisions = []
    localState.marketingError = errors
  },

  [mutationTypes.createMarketingRequestStart](localState) {
    localState.isCreatingMarketingRequest = true
    localState.marketingError = []
  },
  [mutationTypes.createMarketingRequestSuccess](localState, ticket) {
    localState.isCreatingMarketingRequest = false
    localState.selectedTicket = ticket
    localState.myTickets = prependTicketSummary(localState.myTickets, ticket)
  },
  [mutationTypes.createMarketingRequestFail](localState, errors) {
    localState.isCreatingMarketingRequest = false
    localState.marketingError = errors
  }
}

const actions = {
  [actionTypes.loadMyTickets](context) {
    return new Promise((resolve) => {
      context.commit(mutationTypes.loadMyTicketsStart)

      ticketsApi.listMyTickets()
        .then((response) => {
          context.commit(mutationTypes.loadMyTicketsSuccess, response?.data?.data || [])
          resolve(response)
        })
        .catch((error) => {
          context.commit(mutationTypes.loadMyTicketsFail, [normalizeTicketError(error)])
          resolve(error)
        })
    })
  },

  [actionTypes.createTicket](context, payload) {
    return new Promise((resolve, reject) => {
      context.commit(mutationTypes.createTicketStart)

      ticketsApi.createTicket(payload)
        .then((response) => {
          context.commit(mutationTypes.createTicketSuccess, response?.data?.data || null)
          resolve(response)
        })
        .catch((error) => {
          context.commit(mutationTypes.createTicketFail, [normalizeTicketError(error)])
          reject(error)
        })
    })
  },

  [actionTypes.loadResponsibleTickets](context) {
    return new Promise((resolve) => {
      context.commit(mutationTypes.loadResponsibleTicketsStart)

      ticketsApi.listResponsibleTickets()
        .then((response) => {
          context.commit(mutationTypes.loadResponsibleTicketsSuccess, response?.data?.data || [])
          resolve(response)
        })
        .catch((error) => {
          context.commit(mutationTypes.loadResponsibleTicketsFail, [normalizeTicketError(error)])
          resolve(error)
        })
    })
  },

  // Поиск по номеру: тот же тип ответа, что и GET карточки — кладём в selectedTicket.
  [actionTypes.searchTicket](context, payload) {
    return new Promise((resolve) => {
      context.commit(mutationTypes.loadTicketDetailsStart)

      ticketsApi.searchTicket(payload)
        .then((response) => {
          context.commit(mutationTypes.loadTicketDetailsSuccess, response?.data?.data || null)
          resolve(response)
        })
        .catch((error) => {
          context.commit(mutationTypes.loadTicketDetailsFail, [normalizeTicketError(error)])
          resolve(error)
        })
    })
  },

  [actionTypes.loadTicketDetails](context, number) {
    return new Promise((resolve) => {
      context.commit(mutationTypes.loadTicketDetailsStart)

      ticketsApi.getTicketDetails(number)
        .then((response) => {
          context.commit(mutationTypes.loadTicketDetailsSuccess, response?.data?.data || null)
          resolve(response)
        })
        .catch((error) => {
          context.commit(mutationTypes.loadTicketDetailsFail, [normalizeTicketError(error)])
          resolve(error)
        })
    })
  },

  [actionTypes.loadResponsibleOptions](context, number) {
    return new Promise((resolve) => {
      context.commit(mutationTypes.loadResponsibleOptionsStart)

      ticketsApi.listResponsibleOptions(number)
        .then((response) => {
          context.commit(mutationTypes.loadResponsibleOptionsSuccess, response?.data?.data || [])
          resolve(response)
        })
        .catch((error) => {
          context.commit(mutationTypes.loadResponsibleOptionsFail, [normalizeTicketError(error)])
          resolve(error)
        })
    })
  },

  [actionTypes.addComment](context, payload) {
    return new Promise((resolve) => {
      context.commit(mutationTypes.addCommentStart)

      ticketsApi.addComment(payload.number, payload.data)
        .then((response) => {
          context.commit(mutationTypes.addCommentSuccess, response?.data?.data || null)
          resolve(response)
        })
        .catch((error) => {
          context.commit(mutationTypes.addCommentFail, [normalizeTicketError(error)])
          resolve(error)
        })
    })
  },

  [actionTypes.changeStatus](context, payload) {
    return new Promise((resolve) => {
      context.commit(mutationTypes.changeStatusStart)

      ticketsApi.changeStatus(payload.number, payload.data)
        .then((response) => {
          context.commit(mutationTypes.changeStatusSuccess, response?.data?.data || null)
          resolve(response)
        })
        .catch((error) => {
          context.commit(mutationTypes.changeStatusFail, [normalizeTicketError(error)])
          resolve(error)
        })
    })
  },

  [actionTypes.changeResponsible](context, payload) {
    return new Promise((resolve) => {
      context.commit(mutationTypes.changeResponsibleStart)

      ticketsApi.changeResponsible(payload.number, payload.data)
        .then((response) => {
          context.commit(mutationTypes.changeResponsibleSuccess, response?.data?.data || null)
          resolve(response)
        })
        .catch((error) => {
          context.commit(mutationTypes.changeResponsibleFail, [normalizeTicketError(error)])
          resolve(error)
        })
    })
  },

  [actionTypes.confirmTicket](context, payload) {
    return new Promise((resolve) => {
      context.commit(mutationTypes.confirmTicketStart)

      ticketsApi.confirmTicket(payload.number, payload.data)
        .then((response) => {
          context.commit(mutationTypes.confirmTicketSuccess, response?.data?.data || null)
          resolve(response)
        })
        .catch((error) => {
          context.commit(mutationTypes.confirmTicketFail, [normalizeTicketError(error)])
          resolve(error)
        })
    })
  },

  [actionTypes.loadMarketingServices](context) {
    return new Promise((resolve) => {
      context.commit(mutationTypes.loadMarketingServicesStart)

      ticketsApi.listMarketingServices()
        .then((response) => {
          context.commit(mutationTypes.loadMarketingServicesSuccess, response?.data?.data || [])
          resolve(response)
        })
        .catch((error) => {
          context.commit(mutationTypes.loadMarketingServicesFail, [normalizeTicketError(error)])
          resolve(error)
        })
    })
  },

  [actionTypes.loadMarketingSubdivisions](context) {
    return new Promise((resolve) => {
      context.commit(mutationTypes.loadMarketingSubdivisionsStart)

      ticketsApi.listMarketingSubdivisions()
        .then((response) => {
          context.commit(mutationTypes.loadMarketingSubdivisionsSuccess, response?.data?.data || [])
          resolve(response)
        })
        .catch((error) => {
          context.commit(mutationTypes.loadMarketingSubdivisionsFail, [normalizeTicketError(error)])
          resolve(error)
        })
    })
  },

  [actionTypes.createMarketingRequest](context, payload) {
    return new Promise((resolve, reject) => {
      context.commit(mutationTypes.createMarketingRequestStart)

      ticketsApi.createMarketingRequest(payload)
        .then((response) => {
          context.commit(mutationTypes.createMarketingRequestSuccess, response?.data?.data || null)
          resolve(response)
        })
        .catch((error) => {
          context.commit(mutationTypes.createMarketingRequestFail, [normalizeTicketError(error)])
          reject(error)
        })
    })
  }
}

export default {
  state,
  getters,
  mutations,
  actions
}

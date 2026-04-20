/**
 * Vuex-модуль заявок: списки, карточка, поиск, комментарии и смена статуса/ответственного.
 * Данные приходят из ITILIUM через backend; при успешных мутациях списки на экране синхронизируются с карточкой.
 */
import ticketsApi from '@/api/tickets'

function normalizeTicketError(error) {
  const backendMessage = error?.response?.data?.message
  const statusCode = error?.response?.status
  const rawMessage = String(backendMessage || error?.message || '').trim()

  if (statusCode === 404 || /status 404/i.test(rawMessage)) {
    return 'Заявка не найдена в ITILIUM.'
  }
  if (statusCode === 400 || /ticket number is required/i.test(rawMessage)) {
    return 'Проверьте номер заявки и повторите запрос.'
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
  listError: [],
  ticketError: [],
  createError: []
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
  changeResponsibleFail: '[tickets] changeResponsibleFail'
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
  changeResponsible: '[tickets] changeResponsible'
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
  listError: '[tickets] listError',
  ticketError: '[tickets] ticketError',
  createError: '[tickets] createError'
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
  [getterTypes.listError]: (localState) => localState.listError,
  [getterTypes.ticketError]: (localState) => localState.ticketError,
  [getterTypes.createError]: (localState) => localState.createError
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
    return new Promise((resolve) => {
      context.commit(mutationTypes.createTicketStart)

      ticketsApi.createTicket(payload)
        .then((response) => {
          context.commit(mutationTypes.createTicketSuccess, response?.data?.data || null)
          resolve(response)
        })
        .catch((error) => {
          context.commit(mutationTypes.createTicketFail, [normalizeTicketError(error)])
          resolve(error)
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
  }
}

export default {
  state,
  getters,
  mutations,
  actions
}

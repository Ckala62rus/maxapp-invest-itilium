import ticketsApi from '@/api/tickets'

const state = {
  myTickets: [],
  responsibleTickets: [],
  selectedTicket: null,
  responsibleOptions: [],
  isLoadingMyTickets: false,
  isLoadingResponsibleTickets: false,
  isLoadingTicketDetails: false,
  isLoadingResponsibleOptions: false,
  isSubmittingComment: false,
  isChangingStatus: false,
  isChangingResponsible: false,
  listError: [],
  ticketError: []
}

export const mutationTypes = {
  loadMyTicketsStart: '[tickets] loadMyTicketsStart',
  loadMyTicketsSuccess: '[tickets] loadMyTicketsSuccess',
  loadMyTicketsFail: '[tickets] loadMyTicketsFail',
  loadResponsibleTicketsStart: '[tickets] loadResponsibleTicketsStart',
  loadResponsibleTicketsSuccess: '[tickets] loadResponsibleTicketsSuccess',
  loadResponsibleTicketsFail: '[tickets] loadResponsibleTicketsFail',
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
  isLoadingTicketDetails: '[tickets] isLoadingTicketDetails',
  isLoadingResponsibleOptions: '[tickets] isLoadingResponsibleOptions',
  isSubmittingComment: '[tickets] isSubmittingComment',
  isChangingStatus: '[tickets] isChangingStatus',
  isChangingResponsible: '[tickets] isChangingResponsible',
  listError: '[tickets] listError',
  ticketError: '[tickets] ticketError'
}

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

const getters = {
  [getterTypes.myTickets]: (localState) => localState.myTickets,
  [getterTypes.responsibleTickets]: (localState) => localState.responsibleTickets,
  [getterTypes.selectedTicket]: (localState) => localState.selectedTicket,
  [getterTypes.responsibleOptions]: (localState) => localState.responsibleOptions,
  [getterTypes.isLoadingMyTickets]: (localState) => localState.isLoadingMyTickets,
  [getterTypes.isLoadingResponsibleTickets]: (localState) => localState.isLoadingResponsibleTickets,
  [getterTypes.isLoadingTicketDetails]: (localState) => localState.isLoadingTicketDetails,
  [getterTypes.isLoadingResponsibleOptions]: (localState) => localState.isLoadingResponsibleOptions,
  [getterTypes.isSubmittingComment]: (localState) => localState.isSubmittingComment,
  [getterTypes.isChangingStatus]: (localState) => localState.isChangingStatus,
  [getterTypes.isChangingResponsible]: (localState) => localState.isChangingResponsible,
  [getterTypes.listError]: (localState) => localState.listError,
  [getterTypes.ticketError]: (localState) => localState.ticketError
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
    localState.ticketError = errors
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
          context.commit(mutationTypes.loadMyTicketsFail, [error?.response?.data?.message || error.message])
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
          context.commit(mutationTypes.loadResponsibleTicketsFail, [error?.response?.data?.message || error.message])
          resolve(error)
        })
    })
  },

  // Search uses the dedicated backend endpoint but stores the same full
  // ticket detail model that the direct details endpoint returns.
  [actionTypes.searchTicket](context, payload) {
    return new Promise((resolve) => {
      context.commit(mutationTypes.loadTicketDetailsStart)

      ticketsApi.searchTicket(payload)
        .then((response) => {
          context.commit(mutationTypes.loadTicketDetailsSuccess, response?.data?.data || null)
          resolve(response)
        })
        .catch((error) => {
          context.commit(mutationTypes.loadTicketDetailsFail, [error?.response?.data?.message || error.message])
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
          context.commit(mutationTypes.loadTicketDetailsFail, [error?.response?.data?.message || error.message])
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
          context.commit(mutationTypes.loadResponsibleOptionsFail, [error?.response?.data?.message || error.message])
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
          context.commit(mutationTypes.addCommentFail, [error?.response?.data?.message || error.message])
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
          context.commit(mutationTypes.changeStatusFail, [error?.response?.data?.message || error.message])
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
          context.commit(mutationTypes.changeResponsibleFail, [error?.response?.data?.message || error.message])
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

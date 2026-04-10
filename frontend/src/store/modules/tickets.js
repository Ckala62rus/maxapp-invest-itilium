import ticketsApi from '@/api/tickets'

const state = {
  myTickets: [],
  responsibleTickets: [],
  selectedTicket: null,
  isLoadingMyTickets: false,
  isLoadingResponsibleTickets: false,
  isLoadingTicketDetails: false,
  ticketsError: []
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
  loadTicketDetailsFail: '[tickets] loadTicketDetailsFail'
}

export const actionTypes = {
  loadMyTickets: '[tickets] loadMyTickets',
  loadResponsibleTickets: '[tickets] loadResponsibleTickets',
  loadTicketDetails: '[tickets] loadTicketDetails'
}

export const getterTypes = {
  myTickets: '[tickets] myTickets',
  responsibleTickets: '[tickets] responsibleTickets',
  selectedTicket: '[tickets] selectedTicket',
  isLoadingMyTickets: '[tickets] isLoadingMyTickets',
  isLoadingResponsibleTickets: '[tickets] isLoadingResponsibleTickets',
  isLoadingTicketDetails: '[tickets] isLoadingTicketDetails',
  ticketsError: '[tickets] ticketsError'
}

const getters = {
  [getterTypes.myTickets]: (localState) => localState.myTickets,
  [getterTypes.responsibleTickets]: (localState) => localState.responsibleTickets,
  [getterTypes.selectedTicket]: (localState) => localState.selectedTicket,
  [getterTypes.isLoadingMyTickets]: (localState) => localState.isLoadingMyTickets,
  [getterTypes.isLoadingResponsibleTickets]: (localState) => localState.isLoadingResponsibleTickets,
  [getterTypes.isLoadingTicketDetails]: (localState) => localState.isLoadingTicketDetails,
  [getterTypes.ticketsError]: (localState) => localState.ticketsError
}

const mutations = {
  [mutationTypes.loadMyTicketsStart](localState) {
    localState.isLoadingMyTickets = true
    localState.ticketsError = []
  },
  [mutationTypes.loadMyTicketsSuccess](localState, tickets) {
    localState.isLoadingMyTickets = false
    localState.myTickets = tickets
  },
  [mutationTypes.loadMyTicketsFail](localState, errors) {
    localState.isLoadingMyTickets = false
    localState.ticketsError = errors
  },

  [mutationTypes.loadResponsibleTicketsStart](localState) {
    localState.isLoadingResponsibleTickets = true
    localState.ticketsError = []
  },
  [mutationTypes.loadResponsibleTicketsSuccess](localState, tickets) {
    localState.isLoadingResponsibleTickets = false
    localState.responsibleTickets = tickets
  },
  [mutationTypes.loadResponsibleTicketsFail](localState, errors) {
    localState.isLoadingResponsibleTickets = false
    localState.ticketsError = errors
  },

  [mutationTypes.loadTicketDetailsStart](localState) {
    localState.isLoadingTicketDetails = true
    localState.ticketsError = []
  },
  [mutationTypes.loadTicketDetailsSuccess](localState, ticket) {
    localState.isLoadingTicketDetails = false
    localState.selectedTicket = ticket
  },
  [mutationTypes.loadTicketDetailsFail](localState, errors) {
    localState.isLoadingTicketDetails = false
    localState.ticketsError = errors
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
  }
}

export default {
  state,
  getters,
  mutations,
  actions
}

import axios from '@/api/axios'
import urls from '@/api/urls'

// listMyTickets loads the current user's own ticket list.
const listMyTickets = () => {
  return axios.get(urls.myTickets)
}

// listResponsibleTickets loads the list where the current user is responsible.
const listResponsibleTickets = () => {
  return axios.get(urls.responsibleTickets)
}

// searchTicket resolves one ticket number into the detailed ticket payload.
const searchTicket = (payload) => {
  return axios.post(urls.searchTicket, payload)
}

// getTicketDetails loads the detailed card for a selected ticket.
const getTicketDetails = (number) => {
  return axios.get(urls.ticketDetails(number))
}

export default {
  listMyTickets,
  listResponsibleTickets,
  searchTicket,
  getTicketDetails
}

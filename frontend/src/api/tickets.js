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

// addComment submits a new comment and returns the refreshed ticket detail.
const addComment = (number, payload) => {
  return axios.post(urls.ticketComments(number), payload)
}

// changeStatus submits a workflow transition and returns the refreshed ticket detail.
const changeStatus = (number, payload) => {
  return axios.post(urls.ticketStatus(number), payload)
}

// listResponsibleOptions loads assignable people for the ticket card.
const listResponsibleOptions = (number) => {
  return axios.get(urls.ticketResponsibles(number))
}

// changeResponsible submits a new assignee and returns the refreshed ticket detail.
const changeResponsible = (number, payload) => {
  return axios.post(urls.ticketResponsible(number), payload)
}

export default {
  listMyTickets,
  listResponsibleTickets,
  searchTicket,
  getTicketDetails,
  addComment,
  changeStatus,
  listResponsibleOptions,
  changeResponsible
}

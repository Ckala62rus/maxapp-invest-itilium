import axios from '@/api/axios'
import urls from '@/api/urls'

/** Список «мои заявки». */
const listMyTickets = () => {
  return axios.get(urls.myTickets)
}

/** Создание заявки, в ответе — полная карточка. */
const createTicket = (payload) => {
  return axios.post(urls.myTickets, payload)
}

/** Заявки, где пользователь в ответственных. */
const listResponsibleTickets = () => {
  return axios.get(urls.responsibleTickets)
}

/** Поиск по номеру заявки. */
const searchTicket = (payload) => {
  return axios.post(urls.searchTicket, payload)
}

/** Карточка заявки по номеру из URL/списка. */
const getTicketDetails = (number) => {
  return axios.get(urls.ticketDetails(number))
}

/** Комментарий к заявке; ответ — обновлённая карточка. */
const addComment = (number, payload) => {
  return axios.post(urls.ticketComments(number), payload)
}

/** Смена статуса (workflow). */
const changeStatus = (number, payload) => {
  return axios.post(urls.ticketStatus(number), payload)
}

/** Кандидаты в ответственные для выпадающего списка. */
const listResponsibleOptions = (number) => {
  return axios.get(urls.ticketResponsibles(number))
}

/** Назначить ответственного. */
const changeResponsible = (number, payload) => {
  return axios.post(urls.ticketResponsible(number), payload)
}

export default {
  listMyTickets,
  createTicket,
  listResponsibleTickets,
  searchTicket,
  getTicketDetails,
  addComment,
  changeStatus,
  listResponsibleOptions,
  changeResponsible
}

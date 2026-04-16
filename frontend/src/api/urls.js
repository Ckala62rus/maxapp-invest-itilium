/** Централизованные пути backend API (совпадают с chi-маршрутами). */
const urls = {
  maxAuthValidate: '/api/v1/auth/max/validate',
  registration: '/api/v1/users/register',
  me: '/api/v1/users/me',
  myTickets: '/api/v1/tickets',
  responsibleTickets: '/api/v1/tickets/responsible',
  searchTicket: '/api/v1/tickets/search',
  ticketDetails: (number) => `/api/v1/tickets/${number}`,
  ticketComments: (number) => `/api/v1/tickets/${number}/comments`,
  ticketStatus: (number) => `/api/v1/tickets/${number}/status`,
  ticketResponsibles: (number) => `/api/v1/tickets/${number}/responsibles`,
  ticketResponsible: (number) => `/api/v1/tickets/${number}/responsible`
}

export default urls

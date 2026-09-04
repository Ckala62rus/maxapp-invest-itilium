import axios, { longRunningRequestTimeoutMs } from '@/api/axios'
import urls from '@/api/urls'
import { prepareAttachmentFiles } from '@/utils/prepareAttachmentFiles'

const longRunningConfig = { timeout: longRunningRequestTimeoutMs }

/** Список «мои заявки». */
const listMyTickets = () => {
  return axios.get(urls.myTickets)
}

/**
 * Создание заявки: без файлов — JSON; с файлами — multipart, как ожидает бэкенд
 * (поле `payload` со строкой JSON + части `attachments`).
 */
const createTicket = async (payload) => {
  const { attachmentFiles, ...fields } = payload
  const raw = Array.isArray(attachmentFiles) ? attachmentFiles.filter((f) => f instanceof File) : []
  const files = await prepareAttachmentFiles(raw)

  if (files.length === 0) {
    return axios.post(
      urls.myTickets,
      {
        requestType: fields.requestType,
        title: fields.title,
        description: fields.description,
        department: fields.department,
        executionDate: fields.executionDate
      },
      longRunningConfig
    )
  }

  const formData = new FormData()
  // Имена вложений на стороне сервера соберутся из загруженных файлов; в JSON оставляем пустой массив.
  formData.append(
    'payload',
    JSON.stringify({
      requestType: fields.requestType,
      title: fields.title,
      description: fields.description,
      department: fields.department,
      executionDate: fields.executionDate,
      attachments: []
    })
  )
  files.forEach((file) => {
    formData.append('attachments', file)
  })

  return axios.post(urls.myTickets, formData, longRunningConfig)
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

/** Лента комментариев заявки (ITILIUM list_comment через backend). */
const listTicketComments = (number) => {
  return axios.get(urls.ticketComments(number))
}

/**
 * Комментарий к заявке; ответ — обновлённая карточка.
 * Без файлов — JSON; с файлами — multipart (`payload` + части `attachments`), как при создании заявки.
 */
const addComment = async (number, payload) => {
  const { attachmentFiles, ...fields } = payload
  const raw = Array.isArray(attachmentFiles) ? attachmentFiles.filter((f) => f instanceof File) : []
  const files = await prepareAttachmentFiles(raw)

  if (files.length === 0) {
    return axios.post(urls.ticketComments(number), {
      message: fields.message,
      attachments: fields.attachments || []
    })
  }

  const formData = new FormData()
  formData.append(
    'payload',
    JSON.stringify({
      message: fields.message || '',
      attachments: []
    })
  )
  files.forEach((file) => {
    formData.append('attachments', file)
  })

  return axios.post(urls.ticketComments(number), formData)
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

/** Оценка решения (confirm_sc). */
const confirmTicket = (number, payload) => {
  return axios.post(urls.ticketConfirm(number), payload)
}

/** Маркетинговые типы из 1С: code/name/formNumber/formSchema. */
const listMarketingServices = () => {
  return axios.get(urls.marketingServices)
}

/** Подразделения маркетинга из 1С для шага 2. */
const listMarketingSubdivisions = () => {
  return axios.get(urls.marketingSubdivisions)
}

/** Создать маркетинговую заявку (wizard 4 шага + динамические поля шага 4). */
const createMarketingRequest = async (payload) => {
  const { attachmentFiles, ...fields } = payload
  const raw = Array.isArray(attachmentFiles) ? attachmentFiles.filter((f) => f instanceof File) : []
  const files = await prepareAttachmentFiles(raw)

  if (files.length === 0) {
    return axios.post(urls.marketingRequests, fields, longRunningConfig)
  }

  const formData = new FormData()
  formData.append(
    'payload',
    JSON.stringify({
      ...fields,
      attachments: []
    })
  )
  files.forEach((file) => {
    formData.append('attachments', file)
  })

  return axios.post(urls.marketingRequests, formData, longRunningConfig)
}

export default {
  listMyTickets,
  createTicket,
  listResponsibleTickets,
  searchTicket,
  getTicketDetails,
  listTicketComments,
  addComment,
  changeStatus,
  listResponsibleOptions,
  changeResponsible,
  confirmTicket,
  listMarketingServices,
  listMarketingSubdivisions,
  createMarketingRequest
}

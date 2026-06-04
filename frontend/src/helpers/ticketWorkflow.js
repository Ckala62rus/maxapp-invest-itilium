/** Статус «В ожидании ответа» — в ITILIUM обязателен комментарий. */
export function isWaitingForResponseStatus(status) {
  return String(status || '').toLowerCase().includes('в ожидании ответа')
}

/** Статус «Отложено» (коды вида 05_Отложено) — нужны комментарий и дата отложения. */
export function isPostponedStatus(status) {
  const normalized = String(status || '').toLowerCase()
  return normalized.includes('отлож')
}

/**
 * Проверка полей перед сменой статуса. Возвращает текст ошибки или пустую строку.
 */
export function validateStatusTransition(status, form) {
  const comment = String(form?.comment || '').trim()
  const date = String(form?.date || '').trim()

  if (isWaitingForResponseStatus(status) && !comment) {
    return 'Для статуса «В ожидании ответа» комментарий обязателен.'
  }
  if (isPostponedStatus(status)) {
    if (!comment) {
      return 'Для статуса «Отложено» укажите комментарий.'
    }
    if (!date) {
      return 'Для статуса «Отложено» укажите дату, до которой откладывается заявка.'
    }
  }
  return ''
}

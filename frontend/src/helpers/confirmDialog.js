import Swal from 'sweetalert2'
import { purgeTouchBlockers } from '@/helpers/purgeTouchBlockers'

/** Общие классы темы SweetAlert2 для confirm-диалогов mini app. */
export const swalConfirmClasses = {
  popup: 'maxapp-swal-popup',
  confirmButton: 'maxapp-swal-confirm',
  cancelButton: 'maxapp-swal-cancel'
}

/**
 * В WebView (MAX/Qt) контейнер Swal иногда остаётся с pointer-events: none — кнопки не кликаются.
 * Принудительно включаем события на контейнере, попапе и кнопках при открытии.
 */
function enableSwalInteraction(popup) {
  document.body.classList.add('swal2-shown')
  document.documentElement.classList.add('swal2-shown')
  const container = Swal.getContainer()
  if (container) {
    container.style.setProperty('display', 'flex', 'important')
    container.style.visibility = 'visible'
    container.style.pointerEvents = 'auto'
    container.removeAttribute('aria-hidden')
  }
  if (popup) {
    popup.style.pointerEvents = 'auto'
    popup.querySelectorAll('button').forEach((button) => {
      button.type = 'button'
      button.style.pointerEvents = 'auto'
      button.style.touchAction = 'manipulation'
    })
  }
}

/**
 * Confirm-диалог «Да / Отмена» с едиными настройками для desktop и mobile WebView.
 */
export function confirmAction({
  title,
  text,
  confirmButtonText = 'Да',
  cancelButtonText = 'Отмена'
}) {
  return Swal.fire({
    title,
    text,
    icon: 'question',
    showCancelButton: true,
    confirmButtonText,
    cancelButtonText,
    reverseButtons: true,
    focusCancel: true,
    heightAuto: false,
    customClass: swalConfirmClasses,
    buttonsStyling: true,
    didOpen: enableSwalInteraction,
    didClose: () => purgeTouchBlockers()
  })
}

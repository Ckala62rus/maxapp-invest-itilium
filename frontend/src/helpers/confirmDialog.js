import Swal from 'sweetalert2'

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
  const container = Swal.getContainer()
  if (container) {
    container.style.pointerEvents = 'auto'
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
    didOpen: enableSwalInteraction
  })
}

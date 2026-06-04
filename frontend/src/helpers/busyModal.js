import Swal from 'sweetalert2'

import { swalConfirmClasses } from '@/helpers/confirmDialog'
import { purgeTouchBlockers } from '@/helpers/purgeTouchBlockers'

const busyPopupClass = {
  popup: `${swalConfirmClasses.popup} maxapp-swal-busy`
}

function enableBusyModalInteraction(popup) {
  document.body.classList.add('swal2-shown')
  document.documentElement.classList.add('swal2-shown')
  const container = Swal.getContainer()
  if (container) {
    container.style.setProperty('display', 'flex', 'important')
    container.style.visibility = 'visible'
    container.style.pointerEvents = 'auto'
    container.removeAttribute('aria-hidden')
  }
  popup?.style.setProperty('pointer-events', 'auto')
}

/**
 * Блокирующее модальное окно со спиннером на время долгого запроса (create_sc, смена статуса и т.д.).
 */
export async function withBusyModal(title, task) {
  void Swal.fire({
    title,
    html: '<div class="spinner swal-busy-spinner" aria-hidden="true"></div>',
    showConfirmButton: false,
    showCancelButton: false,
    allowOutsideClick: false,
    allowEscapeKey: false,
    heightAuto: false,
    customClass: busyPopupClass,
    didOpen: enableBusyModalInteraction,
    didClose: () => purgeTouchBlockers()
  })

  try {
    return await task()
  } finally {
    await Swal.close()
    purgeTouchBlockers()
  }
}

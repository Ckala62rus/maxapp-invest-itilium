import Swal from 'sweetalert2'

const busyPopupClass = {
  popup: 'maxapp-swal-popup maxapp-swal-busy'
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
    customClass: busyPopupClass
  })

  try {
    return await task()
  } finally {
    Swal.close()
  }
}

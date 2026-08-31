import Swal from 'sweetalert2'

/**
 * Убирает «залипшие» оверлеи (SweetAlert2 и т.п.), которые на Android WebView перехватывают tap.
 */
export function purgeTouchBlockers() {
  if (typeof Swal.isVisible === 'function' && Swal.isVisible()) {
    return
  }
  const body = document.body
  const root = document.documentElement
  if (!body || !root) {
    return
  }
  body.classList.remove('swal2-shown', 'swal2-height-auto', 'swal2-no-backdrop')
  root.classList.remove('swal2-shown')

  document.querySelectorAll('.swal2-container').forEach((container) => {
    if (body.classList.contains('swal2-shown')) {
      return
    }
    container.style.display = 'none'
    container.style.visibility = 'hidden'
    container.style.pointerEvents = 'none'
    container.setAttribute('aria-hidden', 'true')
  })

  document.querySelectorAll('.el-overlay').forEach((overlay) => {
    const style = window.getComputedStyle(overlay)
    if (style.display === 'none' || style.visibility === 'hidden') {
      overlay.style.pointerEvents = 'none'
    }
  })
}

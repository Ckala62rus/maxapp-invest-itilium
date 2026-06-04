/**
 * Убирает «залипшие» оверлеи (SweetAlert2 и т.п.), которые на Android WebView перехватывают tap.
 */
export function purgeTouchBlockers() {
  document.body.classList.remove('swal2-shown', 'swal2-height-auto', 'swal2-no-backdrop')
  document.documentElement.classList.remove('swal2-shown')

  document.querySelectorAll('.swal2-container').forEach((container) => {
    if (document.body.classList.contains('swal2-shown')) {
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

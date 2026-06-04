import { getMaxWebAppPlatform } from '@/api/maxBridge'

const INTERACTIVE_SELECTOR = [
  'button',
  'a[href]',
  '[role="button"]',
  '.ticket-card',
  '.burger-menu-item',
  '.menu-toggle',
  '.page-button',
  '.primary-button',
  '.secondary-button',
  '.ghost-button',
  '.file-chip-remove',
  '.responsible-group-header:not(:disabled)'
].join(', ')

const TOUCH_MOVE_THRESHOLD_PX = 12
const TOUCH_CLICK_GUARD_MS = 450

let touchStartX = 0
let touchStartY = 0
let touchStartTarget = null
let lastSyntheticTarget = null
let lastSyntheticClickAt = 0

function isAndroidMaxWebView() {
  const platform = getMaxWebAppPlatform()
  if (platform === 'android') {
    return true
  }
  if (platform) {
    return false
  }
  return /android/i.test(navigator.userAgent || '')
}

function findInteractiveTarget(node) {
  if (!(node instanceof Element)) {
    return null
  }
  return node.closest(INTERACTIVE_SELECTOR)
}

function isDisabledInteractive(element) {
  if (!element) {
    return true
  }
  if (element.matches(':disabled') || element.getAttribute('aria-disabled') === 'true') {
    return true
  }
  if (element.closest('.swal2-container') && !document.body.classList.contains('swal2-shown')) {
    return true
  }
  return false
}

/**
 * На Android WebView MAX синтетический click после touchend часто не доходит до Vue.
 * Для короткого tap без сдвига вызываем element.click() сами.
 */
function installAndroidTapBridge() {
  if (!isAndroidMaxWebView()) {
    return
  }

  document.addEventListener(
    'touchstart',
    (event) => {
      const touch = event.changedTouches?.[0]
      if (!touch) {
        return
      }
      touchStartX = touch.clientX
      touchStartY = touch.clientY
      touchStartTarget = findInteractiveTarget(event.target)
    },
    { passive: true, capture: true }
  )

  document.addEventListener(
    'touchend',
    (event) => {
      const touch = event.changedTouches?.[0]
      if (!touch) {
        return
      }

      const target = findInteractiveTarget(event.target) || touchStartTarget
      touchStartTarget = null
      if (!target || isDisabledInteractive(target)) {
        return
      }

      const deltaX = Math.abs(touch.clientX - touchStartX)
      const deltaY = Math.abs(touch.clientY - touchStartY)
      if (deltaX > TOUCH_MOVE_THRESHOLD_PX || deltaY > TOUCH_MOVE_THRESHOLD_PX) {
        return
      }

      if (event.cancelable) {
        event.preventDefault()
      }

      const now = Date.now()
      if (target === lastSyntheticTarget && now - lastSyntheticClickAt < TOUCH_CLICK_GUARD_MS) {
        return
      }
      lastSyntheticTarget = target
      lastSyntheticClickAt = now
      target.click()
    },
    { passive: false, capture: true }
  )
}

/**
 * В WebView MAX кнопки без type="button" и скрытый SweetAlert2 ломают tap.
 */
export function ensureMobileInteractions() {
  const patchButtons = (root) => {
    if (!root?.querySelectorAll) {
      return
    }
    root.querySelectorAll('button:not([type])').forEach((button) => {
      button.type = 'button'
    })
  }

  patchButtons(document.body)
  installAndroidTapBridge()

  const observer = new MutationObserver((mutations) => {
    mutations.forEach((mutation) => {
      mutation.addedNodes.forEach((node) => {
        if (node.nodeType !== Node.ELEMENT_NODE) {
          return
        }
        if (node.matches?.('button:not([type])')) {
          node.type = 'button'
        }
        patchButtons(node)
      })
    })
  })

  observer.observe(document.body, { childList: true, subtree: true })
}

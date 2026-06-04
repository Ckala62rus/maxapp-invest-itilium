import { purgeTouchBlockers } from '@/helpers/purgeTouchBlockers'

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

const TOUCH_MOVE_THRESHOLD_PX = 14
const TAP_DEBOUNCE_MS = 400

let touchStartX = 0
let touchStartY = 0
let touchStartTarget = null
let lastSyntheticTarget = null
let lastSyntheticClickAt = 0
let tapActivatedAt = 0
let tapActivatedEl = null

function findInteractiveTarget(node) {
  if (!(node instanceof Element)) {
    return null
  }
  return node.closest(INTERACTIVE_SELECTOR)
}

function shouldInstallTapBridge() {
  if (window.__maxappTouchBridgeInstalled) {
    return false
  }
  if (window.WebApp) {
    return true
  }
  if (window.matchMedia('(pointer: coarse)').matches) {
    return true
  }
  return /android|iphone|ipad|mobile/i.test(navigator.userAgent || '')
}

function installGhostClickGuard() {
  if (window.__maxappGhostClickGuardInstalled) {
    return
  }
  window.__maxappGhostClickGuardInstalled = true
  document.addEventListener(
    'click',
    (event) => {
      const target = findInteractiveTarget(event.target)
      if (!target || target !== tapActivatedEl) {
        return
      }
      const dt = Date.now() - tapActivatedAt
      if (dt > 25 && dt < 800) {
        event.preventDefault()
        event.stopImmediatePropagation()
        tapActivatedEl = null
      }
    },
    true
  )
}

function activateTarget(target) {
  const now = Date.now()
  if (target === lastSyntheticTarget && now - lastSyntheticClickAt < TAP_DEBOUNCE_MS) {
    return
  }
  lastSyntheticTarget = target
  lastSyntheticClickAt = now
  tapActivatedAt = Date.now()
  tapActivatedEl = target
  target.click()
  window.setTimeout(() => {
    if (tapActivatedEl === target) {
      tapActivatedEl = null
    }
  }, 800)
}

/**
 * На Android WebView MAX синтетический click часто не доходит до Vue — дублируем tap через touchend.
 */
function installMobileTapBridge() {
  if (!shouldInstallTapBridge()) {
    return
  }
  window.__maxappTouchBridgeInstalled = true
  installGhostClickGuard()

  document.addEventListener(
    'touchstart',
    (event) => {
      const touch = event.touches?.[0] ?? event.changedTouches?.[0]
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
      if (!target || target.disabled || target.getAttribute('aria-disabled') === 'true') {
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
      activateTarget(target)
    },
    { passive: false, capture: true }
  )
}

/**
 * В WebView MAX кнопки без type="button" и скрытые оверлеи ломают tap.
 */
export function ensureMobileInteractions() {
  purgeTouchBlockers()

  const patchButtons = (root) => {
    if (!root?.querySelectorAll) {
      return
    }
    root.querySelectorAll('button:not([type])').forEach((button) => {
      button.type = 'button'
    })
  }

  patchButtons(document.body)
  installMobileTapBridge()

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
        if (node.matches?.('.swal2-container') || node.querySelector?.('.swal2-container')) {
          purgeTouchBlockers()
        }
      })
    })
  })

  observer.observe(document.body, { childList: true, subtree: true })
}

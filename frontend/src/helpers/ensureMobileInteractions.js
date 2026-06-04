/**
 * В WebView MAX (и части мобильных браузеров) кнопки без type="button" и без touch-action
 * часто не получают click. Нормализуем type и подстраховываем динамически добавленные кнопки.
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

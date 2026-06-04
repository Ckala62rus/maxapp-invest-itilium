/**
 * Сжимает фото перед отправкой: nginx перед 1С (itilium_test) режет тело примерно от 1 MiB.
 * С камеры в MAX WebView часто приходит полноразмерный JPEG (несколько МБ), хотя в галерее видно «400 КБ».
 */

/** Запас под лимит nginx 1С (~1 MiB) с учётом multipart-обёртки. */
const ITILIUM_SAFE_MAX_BYTES = 512 * 1024

/** Всё, что больше — пережимаем (в т.ч. «небольшие» снимки с камеры). */
const IMAGE_COMPRESS_MIN_BYTES = 200 * 1024

const IMAGE_MAX_EDGE = 1280
const JPEG_QUALITIES = [0.82, 0.68, 0.52, 0.38, 0.28]

const IMAGE_EXT_RE = /\.(jpe?g|png|gif|webp|heic|heif)$/i

function looksLikeImage(file) {
  if (!file) {
    return false
  }
  const type = String(file.type || '').toLowerCase()
  if (type.startsWith('image/')) {
    return true
  }
  return IMAGE_EXT_RE.test(String(file.name || ''))
}

function loadImageFromFile(file) {
  return new Promise((resolve, reject) => {
    const url = URL.createObjectURL(file)
    const img = new Image()
    img.onload = () => {
      URL.revokeObjectURL(url)
      resolve(img)
    }
    img.onerror = () => {
      URL.revokeObjectURL(url)
      reject(new Error('failed to load image'))
    }
    img.src = url
  })
}

function canvasToJpegBlob(canvas, quality) {
  return new Promise((resolve, reject) => {
    canvas.toBlob(
      (blob) => (blob ? resolve(blob) : reject(new Error('failed to encode jpeg'))),
      'image/jpeg',
      quality
    )
  })
}

function drawScaledImage(img, maxEdge) {
  let { width, height } = img
  const maxSide = Math.max(width, height)
  if (maxSide > maxEdge) {
    const scale = maxEdge / maxSide
    width = Math.round(width * scale)
    height = Math.round(height * scale)
  }
  const canvas = document.createElement('canvas')
  canvas.width = Math.max(1, width)
  canvas.height = Math.max(1, height)
  const ctx = canvas.getContext('2d')
  if (!ctx) {
    return null
  }
  ctx.drawImage(img, 0, 0, canvas.width, canvas.height)
  return canvas
}

async function compressImageFile(file) {
  if (!looksLikeImage(file)) {
    return file
  }
  if (file.size <= IMAGE_COMPRESS_MIN_BYTES && file.size <= ITILIUM_SAFE_MAX_BYTES) {
    return file
  }

  let img
  try {
    img = await loadImageFromFile(file)
  } catch {
    return file
  }

  const canvas = drawScaledImage(img, IMAGE_MAX_EDGE)
  if (!canvas) {
    return file
  }

  let bestBlob = null
  for (const q of JPEG_QUALITIES) {
    const blob = await canvasToJpegBlob(canvas, q)
    if (!bestBlob || blob.size < bestBlob.size) {
      bestBlob = blob
    }
    if (blob.size <= ITILIUM_SAFE_MAX_BYTES) {
      bestBlob = blob
      break
    }
  }

  if (!bestBlob) {
    return file
  }

  const baseName = (file.name || 'photo').replace(/\.[^.]+$/, '') || 'photo'
  const out = new File([bestBlob], `${baseName}.jpg`, { type: 'image/jpeg', lastModified: Date.now() })
  // Всегда берём сжатый вариант, если он меньше исходника (даже если ещё чуть выше целевого лимита — backend дожмёт).
  if (out.size < file.size) {
    return out
  }
  return file
}

/** Подготавливает вложения к multipart: фото сжимаются под лимит ITILIUM. */
export async function prepareAttachmentFiles(files) {
  const list = Array.isArray(files) ? files.filter((f) => f instanceof File) : []
  return Promise.all(list.map((f) => compressImageFile(f)))
}

/** Человекочитаемый размер для подписи в UI. */
export function formatFileSize(bytes) {
  const n = Number(bytes) || 0
  if (n < 1024) {
    return `${n} Б`
  }
  if (n < 1024 * 1024) {
    return `${(n / 1024).toFixed(n >= 100 * 1024 ? 0 : 1)} КБ`
  }
  return `${(n / (1024 * 1024)).toFixed(1)} МБ`
}

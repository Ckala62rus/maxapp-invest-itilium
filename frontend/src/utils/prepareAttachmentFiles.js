/**
 * Сжимает фото перед отправкой: nginx перед 1С (itilium_test) отдаёт 413 на файлы ~>1 MiB.
 */
const ITILIUM_SAFE_MAX_BYTES = 900 * 1024
const IMAGE_MAX_EDGE = 2048
const JPEG_QUALITIES = [0.85, 0.7, 0.55, 0.4, 0.3]

function isImageFile(file) {
  return file && typeof file.type === 'string' && file.type.startsWith('image/')
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
  canvas.width = width
  canvas.height = height
  const ctx = canvas.getContext('2d')
  if (!ctx) {
    return null
  }
  ctx.drawImage(img, 0, 0, width, height)
  return canvas
}

async function compressImageFile(file) {
  if (!isImageFile(file) || file.size <= ITILIUM_SAFE_MAX_BYTES) {
    return file
  }

  let img
  try {
    img = await loadImageFromFile(file)
  } catch {
    return file
  }

  let canvas = drawScaledImage(img, IMAGE_MAX_EDGE)
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

  if (!bestBlob || bestBlob.size >= file.size) {
    return file
  }

  const baseName = (file.name || 'photo').replace(/\.[^.]+$/, '') || 'photo'
  return new File([bestBlob], `${baseName}.jpg`, { type: 'image/jpeg', lastModified: Date.now() })
}

/** Подготавливает вложения к multipart: фото сжимаются под лимит ITILIUM. */
export async function prepareAttachmentFiles(files) {
  const list = Array.isArray(files) ? files.filter((f) => f instanceof File) : []
  return Promise.all(list.map((f) => (isImageFile(f) ? compressImageFile(f) : f)))
}

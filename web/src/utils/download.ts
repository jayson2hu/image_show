export async function downloadImage(url: string, filename = 'image-show.png') {
  if (!url) {
    return
  }
  try {
    const response = await fetch(url)
    if (!response.ok) {
      throw new Error(`download failed: ${response.status}`)
    }
    const blob = await response.blob()
    const objectURL = URL.createObjectURL(blob)
    triggerDownload(objectURL, filenameWithExtension(filename, blob.type))
    window.setTimeout(() => URL.revokeObjectURL(objectURL), 1000)
  } catch {
    triggerDownload(url, filenameWithExtension(filename, contentTypeFromURL(url)))
  }
}

function triggerDownload(url: string, filename: string) {
  const anchor = document.createElement('a')
  anchor.href = url
  anchor.download = filename
  anchor.rel = 'noopener'
  anchor.style.display = 'none'
  document.body.appendChild(anchor)
  anchor.click()
  anchor.remove()
}

function filenameWithExtension(filename: string, contentType = '') {
  if (/\.(png|jpe?g|webp|gif)$/i.test(filename)) {
    return filename
  }
  const ext = extensionFromContentType(contentType)
  return `${filename}.${ext}`
}

function extensionFromContentType(contentType = '') {
  if (contentType.includes('image/jpeg')) {
    return 'jpg'
  }
  if (contentType.includes('image/webp')) {
    return 'webp'
  }
  if (contentType.includes('image/gif')) {
    return 'gif'
  }
  return 'png'
}

function contentTypeFromURL(url: string) {
  if (url.startsWith('data:image/jpeg') || /\.(jpe?g)(\?|#|$)/i.test(url)) {
    return 'image/jpeg'
  }
  if (url.startsWith('data:image/webp') || /\.webp(\?|#|$)/i.test(url)) {
    return 'image/webp'
  }
  if (url.startsWith('data:image/gif') || /\.gif(\?|#|$)/i.test(url)) {
    return 'image/gif'
  }
  return 'image/png'
}

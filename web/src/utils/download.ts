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
    triggerDownload(objectURL, filename)
    window.setTimeout(() => URL.revokeObjectURL(objectURL), 1000)
  } catch {
    openDownloadFallback(url)
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

function openDownloadFallback(url: string) {
  const opened = window.open(url, '_blank', 'noopener,noreferrer')
  if (!opened) {
    window.location.href = url
  }
}

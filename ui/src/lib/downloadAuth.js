/** Download a URL with the stored JWT (Authorization header). */
export async function downloadWithAuth(url, filename) {
  const token = localStorage.getItem('gr33n_token') ?? ''
  const resp = await fetch(url, {
    headers: token ? { Authorization: `Bearer ${token}` } : {},
  })
  if (!resp.ok) {
    throw new Error(`download failed (${resp.status})`)
  }
  const blob = await resp.blob()
  const objectUrl = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = objectUrl
  a.download = filename
  a.click()
  URL.revokeObjectURL(objectUrl)
}

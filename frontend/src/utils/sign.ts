function normalizeSignValue(value: unknown): string {
  if (value === undefined || value === null) return ''
  if (typeof value === 'number') {
    if (Number.isFinite(value)) return String(value)
    return ''
  }
  return String(value)
}

function toHex(bytes: Uint8Array): string {
  return Array.from(bytes)
    .map((b) => b.toString(16).padStart(2, '0'))
    .join('')
}

async function hmacSHA256Hex(key: string, message: string): Promise<string> {
  const encoder = new TextEncoder()
  const keyData = encoder.encode(key)
  const msgData = encoder.encode(message)
  const cryptoKey = await crypto.subtle.importKey(
    'raw',
    keyData,
    { name: 'HMAC', hash: 'SHA-256' },
    false,
    ['sign']
  )
  const sig = await crypto.subtle.sign('HMAC', cryptoKey, msgData)
  return toHex(new Uint8Array(sig)).toLowerCase()
}

export async function makeOpenAPISign(params: Record<string, unknown>, apiKey: string): Promise<string> {
  const keys = Object.keys(params).sort()
  const pairs: string[] = []
  for (const key of keys) {
    if (key === 'sign' || key === 'sign_type') continue
    const value = normalizeSignValue(params[key])
    if (value === '') continue
    pairs.push(`${key}=${value}`)
  }
  pairs.push(`key=${apiKey}`)
  const payload = pairs.join('&')
  return hmacSHA256Hex(apiKey, payload)
}

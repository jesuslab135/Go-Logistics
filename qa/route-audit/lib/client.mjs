export function buildUrl(baseUrl, path, query) {
  const base = baseUrl.replace(/\/+$/, '')
  let url = `${base}${path}`
  if (query) {
    const params = new URLSearchParams()
    for (const [k, v] of Object.entries(query)) {
      if (v === undefined || v === null) continue
      params.append(k, String(v))
    }
    const qs = params.toString()
    if (qs) url += `?${qs}`
  }
  return url
}

export class ApiClient {
  constructor({ baseUrl, recorder }) {
    this.baseUrl = baseUrl
    this.recorder = recorder
    this.accessToken = null
    this.refreshToken = null
  }

  get token() { return this.accessToken }

  setToken(token) { this.accessToken = token }

  async login(email, password) {
    const res = await this.request('POST', '/auth/login', {
      body: { email, password },
      anonymous: true,
      opKey: 'POST /auth/login',
    })
    if (res.status !== 200) {
      throw new Error(`login failed: ${res.status} ${JSON.stringify(res.body)}`)
    }
    this.accessToken = res.body.access_token
    this.refreshToken = res.body.refresh_token
  }

  async refresh() {
    const res = await this.request('POST', '/auth/refresh', {
      body: { refresh_token: this.refreshToken },
      anonymous: true,
      opKey: 'POST /auth/refresh',
    })
    if (res.status !== 200) {
      throw new Error(`refresh failed: ${res.status} ${JSON.stringify(res.body)}`)
    }
    this.accessToken = res.body.access_token
    if (res.body.refresh_token) this.refreshToken = res.body.refresh_token
  }

  async request(method, path, opts = {}) {
    const { body, query, opKey, anonymous = false, retryOn401 = true } = opts
    const res = await this.#send(method, path, { body, query, anonymous })

    // The access token lives 900s. An unexpected 401 on an authenticated call
    // is far more often expiry than a real authorization defect, so refresh
    // once and retry before letting it be recorded as a finding.
    if (res.status === 401 && !anonymous && retryOn401 && this.refreshToken) {
      await this.refresh()
      const retried = await this.#send(method, path, { body, query, anonymous })
      this.#log(opKey, method, path, body, retried)
      return retried
    }

    this.#log(opKey, method, path, body, res)
    return res
  }

  async #send(method, path, { body, query, anonymous }) {
    const url = buildUrl(this.baseUrl, path, query)
    const headers = {}
    if (body !== undefined) headers['Content-Type'] = 'application/json'
    if (!anonymous && this.accessToken) {
      headers.Authorization = `Bearer ${this.accessToken}`
    }

    const started = process.hrtime.bigint()
    const response = await fetch(url, {
      method,
      headers,
      body: body === undefined ? undefined : JSON.stringify(body),
    })
    const durationMs = Number(process.hrtime.bigint() - started) / 1e6

    const text = await response.text()
    let parsed = null
    if (text) {
      try { parsed = JSON.parse(text) } catch { parsed = text }
    }

    return {
      status: response.status,
      body: parsed,
      headers: Object.fromEntries(response.headers.entries()),
      durationMs,
    }
  }

  #log(opKey, method, path, requestBody, res) {
    this.recorder.record({
      opKey: opKey ?? `${method} ${path}`,
      method,
      url: path,
      requestBody: requestBody ?? null,
      status: res.status,
      responseBody: res.body,
      durationMs: res.durationMs,
    })
  }
}

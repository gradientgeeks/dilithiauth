import axios from 'axios'

const apiBaseURL = import.meta.env.VITE_API_BASE_URL || 'https://dilithiauthapis.gradientgeeks.tech/api/v1/developer'

const api = axios.create({
  baseURL: apiBaseURL,
  withCredentials: true,
})

export interface Developer {
  id: string
  email: string
  name: string
  csrf_token?: string
}

// CSRF token from the most recent /me (or /bff/me) response. State-changing
// requests (create/rotate/delete app) must echo it back via X-CSRF-Token —
// the server rejects those requests without it, even with a valid session
// cookie, since SameSite=Lax alone is not treated as sufficient CSRF defense.
let csrfToken: string | undefined

api.interceptors.request.use(config => {
  if (csrfToken && ['post', 'put', 'delete', 'patch'].includes((config.method || '').toLowerCase())) {
    config.headers = config.headers || {}
    config.headers['X-CSRF-Token'] = csrfToken
  }
  return config
})

export interface OAuthApp {
  ID: string
  DeveloperID: string
  Name: string
  Description: string
  HomepageURL: string
  ClientID: string
  RedirectURIs: string[]
  IsPublic: boolean
  CreatedAt: string
  SecretRotatedAt: string | null
}

export const auth = {
  register: (email: string, name: string, password: string) =>
    api.post<Developer>('/register', { email, name, password }),
  verifyEmail: async (email: string, code: string) => {
    const res = await api.post<{ status: string; message: string; csrf_token?: string }>('/verify-email', { email, code })
    if (res.data.csrf_token) csrfToken = res.data.csrf_token
    return res
  },
  resendCode: (email: string) =>
    api.post<{ status: string }>('/resend-code', { email }),
  login: async (email: string, password: string) => {
    const res = await api.post<Developer>('/login', { email, password })
    if (res.data.csrf_token) csrfToken = res.data.csrf_token
    return res
  },
  logout: () => api.post('/logout'),
  me: async () => {
    const res = await axios.get<Developer>(
      (import.meta.env.VITE_API_BASE_URL || 'https://dilithiauthapis.gradientgeeks.tech/api/v1/developer').replace('/api/v1/developer', '') + '/api/v1/bff/me',
      { withCredentials: true }
    )
    csrfToken = res.data.csrf_token
    return res
  },
}

export const apps = {
  list: () => api.get<OAuthApp[]>('/apps'),
  get: (id: string) => api.get<OAuthApp>(`/apps/${id}`),
  create: (data: {
    name: string
    description: string
    homepage_url: string
    redirect_uris: string[]
    is_public: boolean
  }) => api.post<{ app: OAuthApp; client_secret?: string }>('/apps', data),
  rotateSecret: (id: string) =>
    api.post<{ app: OAuthApp; client_secret: string }>(`/apps/${id}/rotate-secret`),
  delete: (id: string) => api.delete(`/apps/${id}`),
}

import axios from 'axios'

const api = axios.create({
  baseURL: '/api/v1/developer',
  withCredentials: true,
})

export interface Developer {
  id: string
  email: string
  name: string
}

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
  login: (email: string, password: string) =>
    api.post<Developer>('/login', { email, password }),
  logout: () => api.post('/logout'),
  me: () => api.get<Developer>('/me'),
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

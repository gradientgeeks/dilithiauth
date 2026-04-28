import { createContext, useContext, useEffect, useState } from 'react'
import type { ReactNode } from 'react'
import { auth } from './api'
import type { Developer } from './api'

interface AuthCtx {
  developer: Developer | null
  loading: boolean
  login: (email: string, password: string) => Promise<void>
  register: (email: string, name: string, password: string) => Promise<void>
  logout: () => Promise<void>
}

const AuthContext = createContext<AuthCtx | null>(null)

export function AuthProvider({ children }: { children: ReactNode }) {
  const [developer, setDeveloper] = useState<Developer | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    auth.me()
      .then(r => setDeveloper(r.data))
      .catch(() => setDeveloper(null))
      .finally(() => setLoading(false))
  }, [])

  const login = async (email: string, password: string) => {
    const r = await auth.login(email, password)
    setDeveloper(r.data)
  }

  const register = async (email: string, name: string, password: string) => {
    const r = await auth.register(email, name, password)
    setDeveloper(r.data)
  }

  const logout = async () => {
    await auth.logout()
    setDeveloper(null)
  }

  return (
    <AuthContext.Provider value={{ developer, loading, login, register, logout }}>
      {children}
    </AuthContext.Provider>
  )
}

export function useAuth() {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error('useAuth must be used within AuthProvider')
  return ctx
}

import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { ApiClient } from '../api/client.js'
import { AuthContext } from './auth-context.js'

export function AuthProvider({ children }) {
  const [accessToken, setAccessToken] = useState('')
  const [user, setUser] = useState(null)
  const [isReady, setIsReady] = useState(false)
  const accessTokenRef = useRef('')
  const clientRef = useRef(null)

  accessTokenRef.current = accessToken

  if (!clientRef.current) {
    const sharedOptions = {
      getAccessToken: () => accessTokenRef.current,
      setAccessToken: (token) => setAccessToken(token || ''),
      clearSession: () => {
        setAccessToken('')
        setUser(null)
      },
    }

    clientRef.current = new ApiClient(sharedOptions)
  }

  const bootstrapSession = useCallback(async () => {
    try {
      await clientRef.current.refresh()
      const profile = await clientRef.current.getMe()
      setUser(profile)
    } catch {
      setAccessToken('')
      setUser(null)
    } finally {
      setIsReady(true)
    }
  }, [])

  useEffect(() => {
    bootstrapSession()
  }, [bootstrapSession])

  const login = useCallback(async (credentials) => {
    const data = await clientRef.current.login(credentials)
    const profile = data.user || (await clientRef.current.getMe())
    setUser(profile)
    return { ...data, user: profile }
  }, [])

  const register = useCallback(
    async (payload) => {
      await clientRef.current.register(payload)
      return login({ email: payload.email, password: payload.password })
    },
    [login],
  )

  const refreshProfile = useCallback(async () => {
    const profile = await clientRef.current.getMe()
    setUser(profile)
    return profile
  }, [])

  const logout = useCallback(async (allDevices = false) => {
    try {
      if (allDevices) {
        await clientRef.current.logoutAll()
      } else {
        await clientRef.current.logout()
      }
    } finally {
      setAccessToken('')
      setUser(null)
    }
  }, [])

  const value = useMemo(
    () => ({
      api: clientRef.current,
      accessToken,
      user,
      isReady,
      isAuthenticated: Boolean(user),
      login,
      register,
      logout,
      refreshProfile,
      setUser,
    }),
    [accessToken, isReady, login, logout, refreshProfile, register, user],
  )

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

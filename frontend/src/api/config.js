export const API_BASE_URL = (import.meta.env.VITE_API_URL || '/api/v1').replace(/\/$/, '')

export function getApiBaseUrl() {
  return API_BASE_URL
}

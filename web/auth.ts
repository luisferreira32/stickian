// Utility function for making authenticated API requests
export const apiRequest = async (url: string, options: RequestInit = {}) => {
  const token = localStorage.getItem('accessToken')

  const headers = {
    'Content-Type': 'application/json',
    ...options.headers,
    ...(token && { Authorization: `Bearer ${token}` }),
  }

  const response = await fetch(url, {
    ...options,
    headers,
  })

  // If unauthorized, clear the token and redirect to login
  if (response.status === 401) {
    localStorage.removeItem('accessToken')
    window.location.href = '/login'
    throw new Error('Authentication failed')
  }

  return response
}

// Check if user is authenticated
export const isAuthenticated = (): boolean => {
  return !!localStorage.getItem('accessToken')
}

// Get the stored token
export const getToken = (): string | null => {
  return localStorage.getItem('accessToken')
}

// Clear authentication
export const logout = () => {
  localStorage.removeItem('accessToken')
  window.location.href = '/login'
}

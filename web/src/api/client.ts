import type { Transaction, TransactionInput, User } from '../types'

const API_URL = import.meta.env.VITE_API_URL ?? '/api'

function getToken(): string | null {
  return localStorage.getItem('token')
}

async function request<T>(
  path: string,
  options: RequestInit = {},
): Promise<T> {
  const headers = new Headers(options.headers)
  if (options.body && !headers.has('Content-Type')) {
    headers.set('Content-Type', 'application/json')
  }

  const token = getToken()
  if (token) {
    headers.set('Authorization', `Bearer ${token}`)
  }

  const res = await fetch(`${API_URL}${path}`, { ...options, headers })

  if (!res.ok) {
    const text = await res.text()
    throw new Error(text || res.statusText)
  }

  if (res.status === 204) {
    return undefined as T
  }

  return res.json() as Promise<T>
}

export const api = {
  register: (data: { name: string; email: string; password: string }) =>
    request<User>('/register', { method: 'POST', body: JSON.stringify(data) }),

  login: (data: { email: string; password: string }) =>
    request<{ token: string }>('/login', {
      method: 'POST',
      body: JSON.stringify(data),
    }),

  getTransactions: () => request<Transaction[]>('/transactions'),

  createTransaction: (data: TransactionInput) =>
    request<Transaction>('/transactions', {
      method: 'POST',
      body: JSON.stringify(data),
    }),

  updateTransaction: (id: number, data: TransactionInput) =>
    request<Transaction>(`/transactions/${id}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    }),

  deleteTransaction: (id: number) =>
    request<void>(`/transactions/${id}`, { method: 'DELETE' }),

  createLinkToken: () =>
    request<{ link_token: string }>('/banking/link-token', { method: 'POST' }),

  exchangePublicToken: (public_token: string) =>
    request<{ status: string; item_id: string }>('/banking/exchange-token', {
      method: 'POST',
      body: JSON.stringify({ public_token }),
    }),

  syncBankTransactions: () =>
    request<{ imported: number; skipped: number }>('/banking/sync', {
      method: 'POST',
    }),

  getBankingStatus: () =>
    request<{ connected: boolean }>('/banking/status'),
}

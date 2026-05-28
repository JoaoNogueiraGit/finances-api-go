export type TransactionType = 'income' | 'expense'

export interface User {
  id: number
  name: string
  email: string
}

export interface Transaction {
  id: number
  user_id: number
  amount: number
  currency: string
  type: TransactionType
  category: string
  description: string
  date: string
  plaid_transaction_id?: string | null
}

export interface TransactionInput {
  amount: number
  currency: string
  type: TransactionType
  category: string
  description: string
}

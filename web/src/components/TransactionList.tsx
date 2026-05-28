import { useState } from 'react'
import { api } from '../api/client'
import type { Transaction, TransactionInput } from '../types'
import { formatDate, formatMoney } from '../utils/format'
import { TransactionForm } from './TransactionForm'

interface TransactionListProps {
  transactions: Transaction[]
  onChanged: () => void
}

export function TransactionList({ transactions, onChanged }: TransactionListProps) {
  const [editingId, setEditingId] = useState<number | null>(null)
  const [deletingId, setDeletingId] = useState<number | null>(null)

  async function handleUpdate(id: number, data: TransactionInput) {
    await api.updateTransaction(id, data)
    setEditingId(null)
    onChanged()
  }

  async function handleDelete(id: number) {
    if (!confirm('Apagar esta transação?')) return
    setDeletingId(id)
    try {
      await api.deleteTransaction(id)
      onChanged()
    } finally {
      setDeletingId(null)
    }
  }

  if (transactions.length === 0) {
    return (
      <div className="empty-state">
        <strong>Sem transações</strong>
        <p>Adiciona a primeira receita ou despesa acima.</p>
      </div>
    )
  }

  return (
    <ul className="tx-list">
      {transactions.map((tx) => (
        <li key={tx.id} className="tx-item">
          {editingId === tx.id ? (
            <div style={{ width: '100%' }}>
              <TransactionForm
                initial={tx}
                submitLabel="Guardar alterações"
                onSubmit={(data) => handleUpdate(tx.id, data)}
                onCancel={() => setEditingId(null)}
              />
            </div>
          ) : (
            <>
              <div className="tx-main">
                <div className="tx-category">
                  {tx.category}
                  <span className={`badge ${tx.type}`}>
                    {tx.type === 'income' ? 'Receita' : 'Despesa'}
                  </span>
                  {tx.plaid_transaction_id && (
                    <span className="badge bank">Banco</span>
                  )}
                </div>
                {tx.description && <div className="tx-desc">{tx.description}</div>}
                <div className="tx-meta">{formatDate(tx.date)}</div>
              </div>
              <div className="tx-right">
                <div className={`tx-amount ${tx.type}`}>
                  {tx.type === 'expense' ? '−' : '+'}
                  {formatMoney(tx.amount, tx.currency)}
                </div>
                {!tx.plaid_transaction_id && (
                  <div className="tx-actions">
                    <button
                      type="button"
                      className="btn btn-ghost btn-sm"
                      onClick={() => setEditingId(tx.id)}
                    >
                      Editar
                    </button>
                    <button
                      type="button"
                      className="btn btn-danger btn-sm"
                      disabled={deletingId === tx.id}
                      onClick={() => handleDelete(tx.id)}
                    >
                      {deletingId === tx.id ? '…' : 'Apagar'}
                    </button>
                  </div>
                )}
              </div>
            </>
          )}
        </li>
      ))}
    </ul>
  )
}

import { useState, type FormEvent } from 'react'
import type { Transaction, TransactionInput, TransactionType } from '../types'

const emptyForm: TransactionInput = {
  amount: 0,
  currency: 'EUR',
  type: 'expense',
  category: '',
  description: '',
}

interface TransactionFormProps {
  initial?: Transaction
  onSubmit: (data: TransactionInput) => Promise<void>
  onCancel?: () => void
  submitLabel?: string
}

export function TransactionForm({
  initial,
  onSubmit,
  onCancel,
  submitLabel = 'Adicionar',
}: TransactionFormProps) {
  const [form, setForm] = useState<TransactionInput>(
    initial
      ? {
          amount: initial.amount,
          currency: initial.currency,
          type: initial.type,
          category: initial.category,
          description: initial.description,
        }
      : emptyForm,
  )
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  async function handleSubmit(e: FormEvent) {
    e.preventDefault()
    setError('')
    if (form.amount <= 0) {
      setError('O valor deve ser maior que zero.')
      return
    }
    if (!form.category.trim()) {
      setError('Indica uma categoria.')
      return
    }
    setLoading(true)
    try {
      await onSubmit(form)
      if (!initial) setForm(emptyForm)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Erro ao guardar')
    } finally {
      setLoading(false)
    }
  }

  return (
    <form onSubmit={handleSubmit}>
      {error && <div className="form-error">{error}</div>}

      <div className="form-row">
        <div className="form-group">
          <label htmlFor="type">Tipo</label>
          <select
            id="type"
            value={form.type}
            onChange={(e) =>
              setForm((f) => ({ ...f, type: e.target.value as TransactionType }))
            }
          >
            <option value="expense">Despesa</option>
            <option value="income">Receita</option>
          </select>
        </div>
        <div className="form-group">
          <label htmlFor="currency">Moeda</label>
          <select
            id="currency"
            value={form.currency}
            onChange={(e) => setForm((f) => ({ ...f, currency: e.target.value }))}
          >
            <option value="EUR">EUR</option>
            <option value="USD">USD</option>
            <option value="GBP">GBP</option>
          </select>
        </div>
      </div>

      <div className="form-row">
        <div className="form-group">
          <label htmlFor="amount">Valor</label>
          <input
            id="amount"
            type="number"
            min="0.01"
            step="0.01"
            value={form.amount || ''}
            onChange={(e) =>
              setForm((f) => ({ ...f, amount: parseFloat(e.target.value) || 0 }))
            }
            required
          />
        </div>
        <div className="form-group">
          <label htmlFor="category">Categoria</label>
          <input
            id="category"
            type="text"
            placeholder="ex: Alimentação"
            value={form.category}
            onChange={(e) => setForm((f) => ({ ...f, category: e.target.value }))}
            required
          />
        </div>
      </div>

      <div className="form-group">
        <label htmlFor="description">Descrição</label>
        <textarea
          id="description"
          rows={2}
          placeholder="Notas opcionais..."
          value={form.description}
          onChange={(e) => setForm((f) => ({ ...f, description: e.target.value }))}
        />
      </div>

      <div style={{ display: 'flex', gap: '0.5rem' }}>
        <button type="submit" className="btn btn-primary" disabled={loading}>
          {loading ? 'A guardar…' : submitLabel}
        </button>
        {onCancel && (
          <button type="button" className="btn btn-ghost" onClick={onCancel}>
            Cancelar
          </button>
        )}
      </div>
    </form>
  )
}

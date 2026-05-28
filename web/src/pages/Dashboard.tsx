import { useCallback, useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { api } from '../api/client'
import { BankConnect } from '../components/BankConnect'
import { SummaryCards } from '../components/SummaryCards'
import { TransactionForm } from '../components/TransactionForm'
import { TransactionList } from '../components/TransactionList'
import { useAuth } from '../context/AuthContext'
import type { Transaction, TransactionInput } from '../types'

export function Dashboard() {
  const { logout } = useAuth()
  const navigate = useNavigate()
  const [transactions, setTransactions] = useState<Transaction[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  const load = useCallback(async () => {
    setError('')
    try {
      const data = await api.getTransactions()
      setTransactions(data)
    } catch (err) {
      const msg = err instanceof Error ? err.message : 'Erro ao carregar'
      if (msg.includes('Token') || msg.includes('autorizado') || msg.includes('negado')) {
        logout()
        navigate('/login', { replace: true })
        return
      }
      setError(msg)
    } finally {
      setLoading(false)
    }
  }, [logout, navigate])

  useEffect(() => {
    load()
  }, [load])

  async function handleCreate(data: TransactionInput) {
    await api.createTransaction(data)
    await load()
  }

  function handleLogout() {
    logout()
    navigate('/login', { replace: true })
  }

  return (
    <div className="app-shell">
      <header className="app-header">
        <div className="logo">
          Fin<span>ances</span>
        </div>
        <button type="button" className="btn btn-ghost btn-sm" onClick={handleLogout}>
          Sair
        </button>
      </header>

      {error && <div className="form-error">{error}</div>}

      {loading ? (
        <div className="loading">A carregar transações…</div>
      ) : (
        <>
          <SummaryCards transactions={transactions} />

          <BankConnect onSynced={load} />

          <section className="panel">
            <h2>Nova transação</h2>
            <TransactionForm onSubmit={handleCreate} />
          </section>

          <section className="panel">
            <h2>Histórico</h2>
            <TransactionList transactions={transactions} onChanged={load} />
          </section>
        </>
      )}
    </div>
  )
}

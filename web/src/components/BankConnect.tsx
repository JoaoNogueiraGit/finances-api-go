import { useCallback, useEffect, useState } from 'react'
import { usePlaidLink } from 'react-plaid-link'
import { api } from '../api/client'

interface BankConnectProps {
  onSynced: () => void
}

export function BankConnect({ onSynced }: BankConnectProps) {
  const [connected, setConnected] = useState(false)
  const [linkToken, setLinkToken] = useState<string | null>(null)
  const [loading, setLoading] = useState(true)
  const [busy, setBusy] = useState(false)
  const [error, setError] = useState('')
  const [syncMessage, setSyncMessage] = useState('')

  const loadStatus = useCallback(async () => {
    setError('')
    try {
      const status = await api.getBankingStatus()
      setConnected(status.connected)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Erro ao verificar banco')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    loadStatus()
  }, [loadStatus])

  const fetchLinkToken = useCallback(async () => {
    setError('')
    setBusy(true)
    try {
      const { link_token } = await api.createLinkToken()
      setLinkToken(link_token)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Erro ao preparar ligação')
    } finally {
      setBusy(false)
    }
  }, [])

  const handleSuccess = useCallback(
    async (publicToken: string) => {
      setBusy(true)
      setError('')
      setSyncMessage('')
      try {
        await api.exchangePublicToken(publicToken)
        setConnected(true)
        const result = await api.syncBankTransactions()
        setSyncMessage(
          `${result.imported} transação(ões) importada(s)` +
            (result.skipped > 0 ? `, ${result.skipped} ignorada(s)` : ''),
        )
        onSynced()
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Erro ao ligar banco')
      } finally {
        setBusy(false)
        setLinkToken(null)
      }
    },
    [onSynced],
  )

  const { open, ready } = usePlaidLink({
    token: linkToken,
    onSuccess: (publicToken) => {
      void handleSuccess(publicToken)
    },
    onExit: () => {
      setLinkToken(null)
    },
  })

  useEffect(() => {
    if (linkToken && ready) {
      open()
    }
  }, [linkToken, ready, open])

  async function handleSync() {
    setBusy(true)
    setError('')
    setSyncMessage('')
    try {
      const result = await api.syncBankTransactions()
      setSyncMessage(
        `${result.imported} nova(s), ${result.skipped} já existente(s)/ignorada(s)`,
      )
      onSynced()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Erro ao sincronizar')
    } finally {
      setBusy(false)
    }
  }

  if (loading) {
    return <div className="bank-connect loading-inline">A verificar ligação bancária…</div>
  }

  return (
    <section className="panel bank-connect">
      <div className="bank-connect-header">
        <div>
          <h2>Conta bancária</h2>
          <p className="bank-connect-desc">
            Liga o teu banco via Plaid para importar transações reais automaticamente.
          </p>
        </div>
        <span className={`bank-status ${connected ? 'connected' : ''}`}>
          {connected ? 'Ligado' : 'Não ligado'}
        </span>
      </div>

      {error && <div className="form-error">{error}</div>}
      {syncMessage && <div className="form-success">{syncMessage}</div>}

      <div className="bank-connect-actions">
        {!connected ? (
          <button
            type="button"
            className="btn btn-primary"
            disabled={busy}
            onClick={() => void fetchLinkToken()}
          >
            {busy ? 'A preparar…' : 'Ligar banco'}
          </button>
        ) : (
          <>
            <button
              type="button"
              className="btn btn-primary"
              disabled={busy}
              onClick={() => void fetchLinkToken()}
            >
              Ligar outro banco
            </button>
            <button
              type="button"
              className="btn btn-ghost"
              disabled={busy}
              onClick={() => void handleSync()}
            >
              {busy ? 'A sincronizar…' : 'Sincronizar transações'}
            </button>
          </>
        )}
      </div>
    </section>
  )
}

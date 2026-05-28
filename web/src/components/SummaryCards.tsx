import type { Transaction } from '../types'
import { formatMoney } from '../utils/format'

interface SummaryCardsProps {
  transactions: Transaction[]
}

export function SummaryCards({ transactions }: SummaryCardsProps) {
  const totals = transactions.reduce(
    (acc, tx) => {
      if (tx.type === 'income') acc.income += tx.amount
      else acc.expense += tx.amount
      return acc
    },
    { income: 0, expense: 0 },
  )

  const balance = totals.income - totals.expense
  const currency = transactions[0]?.currency ?? 'EUR'

  return (
    <div className="summary-grid">
      <div className="summary-card income">
        <div className="label">Receitas</div>
        <div className="value">{formatMoney(totals.income, currency)}</div>
      </div>
      <div className="summary-card expense">
        <div className="label">Despesas</div>
        <div className="value">{formatMoney(totals.expense, currency)}</div>
      </div>
      <div className="summary-card balance">
        <div className="label">Saldo</div>
        <div className="value">{formatMoney(balance, currency)}</div>
      </div>
    </div>
  )
}

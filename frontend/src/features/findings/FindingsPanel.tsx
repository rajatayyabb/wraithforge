import { useEffect, useState } from 'react'
import { Finding, Severity, fetchFindings } from '../../providers/api'

type Props = {
  flowId: string
}

const severityStyle: Record<Severity, string> = {
  critical: 'bg-red-500/20 text-red-300 border-red-500/40',
  high: 'bg-orange-500/20 text-orange-300 border-orange-500/40',
  medium: 'bg-yellow-500/20 text-yellow-200 border-yellow-500/40',
  low: 'bg-sky-500/20 text-sky-300 border-sky-500/40',
  info: 'bg-slate-500/20 text-slate-300 border-slate-500/40',
}

const severityOrder: Severity[] = ['critical', 'high', 'medium', 'low', 'info']

export default function FindingsPanel({ flowId }: Props) {
  const [findings, setFindings] = useState<Finding[]>([])
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    setError(null)
    fetchFindings(flowId)
      .then(setFindings)
      .catch((e) => {
        setFindings([])
        setError(String(e))
      })
  }, [flowId])

  const counts = severityOrder
    .map((sev) => ({ sev, n: findings.filter((f) => f.severity === sev).length }))
    .filter((c) => c.n > 0)

  return (
    <div className="h-full overflow-y-auto border border-slate-800 rounded-lg">
      <div className="bg-slate-900 px-3 py-2 text-sm font-medium flex items-center justify-between">
        <span>Findings</span>
        <div className="flex gap-1">
          {counts.map((c) => (
            <span
              key={c.sev}
              className={`rounded border px-1.5 py-0.5 text-[10px] uppercase ${severityStyle[c.sev]}`}
            >
              {c.sev} {c.n}
            </span>
          ))}
        </div>
      </div>
      <ul className="p-2 space-y-2">
        {findings.map((f) => (
          <FindingRow key={f.id} finding={f} />
        ))}
        {findings.length === 0 && !error && (
          <li className="px-2 text-sm text-slate-500">no findings — run the triage CLI to ingest tool output</li>
        )}
        {error && <li className="px-2 text-xs text-red-400">{error}</li>}
      </ul>
    </div>
  )
}

function FindingRow({ finding }: { finding: Finding }) {
  const [open, setOpen] = useState(false)
  return (
    <li className="rounded bg-slate-900">
      <button
        onClick={() => setOpen((v) => !v)}
        className="w-full text-left px-3 py-2 text-sm flex items-start gap-2"
      >
        <span
          className={`mt-0.5 shrink-0 rounded border px-1.5 py-0.5 text-[10px] uppercase ${severityStyle[finding.severity]}`}
        >
          {finding.severity}
        </span>
        <span className="flex-1">
          <span className="block">{finding.title}</span>
          <span className="block text-xs text-slate-400">
            {finding.asset} · {finding.sourceTool}
          </span>
        </span>
      </button>
      {open && (
        <div className="px-3 pb-2 text-xs text-slate-300 space-y-2">
          {finding.description && <p>{finding.description}</p>}
          {finding.evidence && (
            <pre className="whitespace-pre-wrap rounded bg-slate-800/60 p-2 text-[11px] text-slate-200">
              {finding.evidence}
            </pre>
          )}
        </div>
      )}
    </li>
  )
}

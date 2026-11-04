import { useEffect, useRef, useState } from 'react'
import { AgentLog, streamAgentLogs } from '../../providers/api'

type Props = {
  flowId: string
}

export default function FlowTerminal({ flowId }: Props) {
  const [logs, setLogs] = useState<AgentLog[]>([])
  const [status, setStatus] = useState<'connecting' | 'open' | 'reconnecting'>('connecting')
  const scrollRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    setLogs([])
    const stop = streamAgentLogs(
      flowId,
      (log) => setLogs((prev) => [...prev, log]),
      setStatus
    )
    return stop
  }, [flowId])

  useEffect(() => {
    scrollRef.current?.scrollTo({ top: scrollRef.current.scrollHeight })
  }, [logs])

  return (
    <div className="flex h-full flex-col border border-slate-800 rounded-lg overflow-hidden">
      <div className="flex items-center justify-between bg-slate-900 px-3 py-2 text-sm">
        <span className="font-medium">Flow Terminal</span>
        <StatusPill status={status} />
      </div>
      <div ref={scrollRef} className="flex-1 overflow-y-auto bg-black p-3 font-mono text-xs space-y-1">
        {logs.length === 0 && <div className="text-slate-500">waiting for logs…</div>}
        {logs.map((log) => (
          <div key={log.id}>
            <span className="text-slate-500">{new Date(log.createdAt).toLocaleTimeString()}</span>{' '}
            <span className="text-emerald-400">[{log.agentRole}]</span> {log.message}
          </div>
        ))}
      </div>
    </div>
  )
}

function StatusPill({ status }: { status: 'connecting' | 'open' | 'reconnecting' }) {
  const styles = {
    connecting: 'bg-amber-500/20 text-amber-300',
    open: 'bg-emerald-500/20 text-emerald-300',
    reconnecting: 'bg-red-500/20 text-red-300'
  }[status]

  const label = {
    connecting: 'connecting…',
    open: 'live',
    reconnecting: 'reconnecting…'
  }[status]

  return <span className={`rounded px-2 py-0.5 text-xs ${styles}`}>{label}</span>
}

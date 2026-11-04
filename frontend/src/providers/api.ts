const API_TOKEN = import.meta.env.VITE_API_TOKEN ?? ''

export type Flow = {
  id: string
  title: string
  goal: string
  status: string
  createdAt: string
  updatedAt: string
}

export type Task = {
  id: string
  flowId: string
  title: string
  status: string
  position: number
}

export type Subtask = {
  id: string
  taskId: string
  title: string
  status: string
  position: number
}

export type AgentLog = {
  id: string
  flowId: string
  agentRole: string
  message: string
  createdAt: string
}

function headers() {
  return { Authorization: `Bearer ${API_TOKEN}` }
}

export async function fetchFlows(): Promise<Flow[]> {
  const res = await fetch('/api/flows', { headers: headers() })
  if (!res.ok) throw new Error(`fetchFlows: ${res.status}`)
  return res.json()
}

export async function fetchTasks(flowId: string): Promise<Task[]> {
  const res = await fetch(`/api/flows/${flowId}/tasks`, { headers: headers() })
  if (!res.ok) throw new Error(`fetchTasks: ${res.status}`)
  return res.json()
}

export async function fetchSubtasks(taskId: string): Promise<Subtask[]> {
  const res = await fetch(`/api/tasks/${taskId}/subtasks`, { headers: headers() })
  if (!res.ok) throw new Error(`fetchSubtasks: ${res.status}`)
  return res.json()
}

export function streamAgentLogs(
  flowId: string,
  onLog: (log: AgentLog) => void,
  onStatusChange: (status: 'connecting' | 'open' | 'reconnecting') => void
): () => void {
  let source: EventSource | null = null
  let retryTimer: number | null = null
  let closed = false

  const connect = () => {
    if (closed) return
    onStatusChange('connecting')
    source = new EventSource(
      `/api/flows/${flowId}/agent-logs/stream?token=${encodeURIComponent(API_TOKEN)}`
    )

    source.onopen = () => onStatusChange('open')

    source.addEventListener('agentLog', (event) => {
      const log = JSON.parse((event as MessageEvent).data) as AgentLog
      onLog(log)
    })

    source.onerror = () => {
      source?.close()
      onStatusChange('reconnecting')
      retryTimer = window.setTimeout(connect, 1500)
    }
  }

  connect()

  return () => {
    closed = true
    source?.close()
    if (retryTimer) window.clearTimeout(retryTimer)
  }
}

import { useEffect, useState } from 'react'
import { Flow, Task, fetchFlows, fetchTasks } from './providers/api'
import FlowTerminal from './features/flows/FlowTerminal'
import TaskTree from './features/tasks/TaskTree'

export default function App() {
  const [flows, setFlows] = useState<Flow[]>([])
  const [selectedFlowId, setSelectedFlowId] = useState<string | null>(null)
  const [tasks, setTasks] = useState<Task[]>([])
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    fetchFlows()
      .then((fs) => {
        setFlows(fs)
        if (fs.length > 0) setSelectedFlowId(fs[0].id)
      })
      .catch((e) => setError(String(e)))
  }, [])

  useEffect(() => {
    if (!selectedFlowId) return
    fetchTasks(selectedFlowId).then(setTasks).catch(() => setTasks([]))
  }, [selectedFlowId])

  return (
    <div className="h-screen flex flex-col">
      <header className="flex items-center gap-4 border-b border-slate-800 px-4 py-3">
        <h1 className="text-lg font-semibold">Wraithforge Console</h1>
        <select
          className="bg-slate-900 border border-slate-800 rounded px-2 py-1 text-sm"
          value={selectedFlowId ?? ''}
          onChange={(e) => setSelectedFlowId(e.target.value)}
        >
          {flows.map((f) => (
            <option key={f.id} value={f.id}>
              {f.title}
            </option>
          ))}
        </select>
        {error && <span className="text-xs text-red-400">{error}</span>}
      </header>
      <main className="flex-1 grid grid-cols-2 gap-3 p-3 min-h-0">
        {selectedFlowId ? <FlowTerminal flowId={selectedFlowId} /> : <div />}
        <TaskTree tasks={tasks} />
      </main>
    </div>
  )
}

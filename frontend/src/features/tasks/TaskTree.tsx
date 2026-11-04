import { useEffect, useState } from 'react'
import { Subtask, Task, fetchSubtasks } from '../../providers/api'

type Props = {
  tasks: Task[]
}

export default function TaskTree({ tasks }: Props) {
  return (
    <div className="h-full overflow-y-auto border border-slate-800 rounded-lg">
      <div className="bg-slate-900 px-3 py-2 text-sm font-medium">Tasks</div>
      <ul className="p-2 space-y-2">
        {tasks.map((task) => (
          <TaskNode key={task.id} task={task} />
        ))}
        {tasks.length === 0 && <li className="px-2 text-sm text-slate-500">no tasks yet</li>}
      </ul>
    </div>
  )
}

function TaskNode({ task }: { task: Task }) {
  const [open, setOpen] = useState(false)
  const [subtasks, setSubtasks] = useState<Subtask[]>([])

  useEffect(() => {
    if (open && subtasks.length === 0) {
      fetchSubtasks(task.id).then(setSubtasks).catch(() => {})
    }
  }, [open, task.id, subtasks.length])

  return (
    <li>
      <button
        onClick={() => setOpen((v) => !v)}
        className="w-full text-left rounded bg-slate-900 px-3 py-2 text-sm flex items-center justify-between"
      >
        <span>{task.title}</span>
        <span className="text-xs text-slate-400">{task.status}</span>
      </button>
      {open && (
        <ul className="ml-4 mt-1 space-y-1">
          {subtasks.map((st) => (
            <li key={st.id} className="rounded bg-slate-800/60 px-3 py-1.5 text-xs flex items-center justify-between">
              <span>{st.title}</span>
              <span className="text-slate-400">{st.status}</span>
            </li>
          ))}
        </ul>
      )}
    </li>
  )
}

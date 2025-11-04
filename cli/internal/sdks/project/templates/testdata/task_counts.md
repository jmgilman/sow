{{$impl := phase . "implementation"}}
Completed: {{countTasksByStatus $impl "completed"}}
In Progress: {{countTasksByStatus $impl "in_progress"}}
Pending: {{countTasksByStatus $impl "pending"}}

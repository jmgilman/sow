{{$impl := phase . "implementation"}}
Completed: {{countTasksByStatus $impl "implementation" "completed"}}
In Progress: {{countTasksByStatus $impl "implementation" "in_progress"}}
Pending: {{countTasksByStatus $impl "implementation" "pending"}}

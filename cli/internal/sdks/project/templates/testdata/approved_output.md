{{$planning := phase . "planning"}}
{{if hasApprovedOutput $planning "task_list"}}Task list is approved{{else}}Task list not approved{{end}}

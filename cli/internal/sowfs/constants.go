package sowfs

// Version constants.
const (
	// CurrentVersion is the current .sow structure version.
	CurrentVersion = "1.0.0"
)

// Directory constants (relative to repository root).
const (
	// DirSow is the root .sow directory.
	DirSow = ".sow"

	// DirKnowledge is the knowledge directory.
	DirKnowledge = ".sow/knowledge"

	// DirRefs is the refs directory.
	DirRefs = ".sow/refs"

	// DirProject is the project directory.
	DirProject = ".sow/project"

	// DirProjectPhases is the phases directory within project.
	DirProjectPhases = ".sow/project/phases"

	// DirImplementationTasks is the tasks directory within implementation phase.
	DirImplementationTasks = ".sow/project/phases/implementation/tasks"
)

// File constants (relative to .sow/ directory, used by SowFS which is chrooted).
const (
	// FileVersion is the version file.
	FileVersion = ".version"

	// FileRefsCommittedIndex is the committed refs index.
	FileRefsCommittedIndex = "refs/index.json"

	// FileRefsLocalIndex is the local refs index.
	FileRefsLocalIndex = "refs/index.local.json"

	// FileRefsGitignore is the .gitignore file in refs directory.
	FileRefsGitignore = "refs/.gitignore"

	// FileProjectState is the project state file.
	FileProjectState = "project/state.yaml"

	// FileProjectLog is the project log file.
	FileProjectLog = "project/log.md"

	// FileTaskState is the task state filename (without directory path).
	FileTaskState = "state.yaml"

	// FileTaskDescription is the task description filename.
	FileTaskDescription = "description.md"

	// FileTaskLog is the task log filename.
	FileTaskLog = "log.md"
)

// Directory constants (relative to .sow/, used by SowFS which is chrooted).
const (
	// PathRefs is the refs directory path.
	PathRefs = "refs"

	// PathKnowledge is the knowledge directory path.
	PathKnowledge = "knowledge"

	// PathProject is the project directory path.
	PathProject = "project"

	// PathProjectContext is the project context directory path.
	PathProjectContext = "project/context"

	// PathProjectPhases is the project phases directory path.
	PathProjectPhases = "project/phases"

	// PathProjectTasksDir is the tasks directory path.
	PathProjectTasksDir = "project/phases/implementation/tasks"
)

// Initial file content templates.
const (
	// VersionFileContent is the content for .sow/.version file.
	VersionFileContent = CurrentVersion + "\n"

	// RefsIndexContent is the initial content for refs/index.json.
	RefsIndexContent = `{
  "version": "1.0.0",
  "refs": []
}
`

	// RefsGitignoreContent is the content for refs/.gitignore.
	RefsGitignoreContent = `# Ignore all symlinks to cached repositories
*

# But keep the indexes
!index.json
!index.local.json
!.gitignore
`
)

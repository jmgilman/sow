package exploration

import "errors"

var (
	// ErrNoExploration indicates no exploration exists in the current context.
	ErrNoExploration = errors.New("no active exploration")

	// ErrExplorationExists indicates an exploration already exists.
	ErrExplorationExists = errors.New("exploration already exists")

	// ErrFileNotFound indicates a file is not in the exploration index.
	ErrFileNotFound = errors.New("file not found in exploration index")

	// ErrFileExists indicates a file already exists in the exploration index.
	ErrFileExists = errors.New("file already exists in exploration index")
)

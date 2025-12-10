module github.com/jmgilman/sow/libs/project

go 1.25.3

require github.com/stretchr/testify v1.11.1

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace (
	github.com/jmgilman/go/fs/billy => /Users/josh/code/go/fs/billy
	github.com/jmgilman/go/fs/core => /Users/josh/code/go/fs/core
	github.com/jmgilman/sow/libs/schemas => ../schemas
)

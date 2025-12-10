module github.com/jmgilman/sow/libs/config

go 1.25.3

require (
	github.com/jmgilman/go/fs/billy v0.1.1
	github.com/jmgilman/go/fs/core v0.2.0
	github.com/jmgilman/sow/libs/schemas v0.0.0
	github.com/stretchr/testify v1.11.1
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/cyphar/filepath-securejoin v0.4.1 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/go-git/go-billy/v5 v5.6.2 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	golang.org/x/sys v0.36.0 // indirect
)

replace (
	github.com/jmgilman/go/fs/billy => /Users/josh/code/go/fs/billy
	github.com/jmgilman/go/fs/core => /Users/josh/code/go/fs/core
	github.com/jmgilman/sow/libs/schemas => ../schemas
)

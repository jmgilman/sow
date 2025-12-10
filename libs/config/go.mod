module github.com/jmgilman/sow/libs/config

go 1.25.3

require github.com/jmgilman/sow/libs/schemas v0.0.0

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/stretchr/testify v1.11.1 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/jmgilman/sow/libs/schemas => ../schemas

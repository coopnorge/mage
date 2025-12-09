module github.com/coopnorge/mage

go 1.25.0

require (
	github.com/bmatcuk/doublestar/v4 v4.9.1
	github.com/magefile/mage v1.15.0
	github.com/stretchr/testify v1.11.1
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

tool github.com/magefile/mage

retract [v0.1.0, v0.16.3] // Retracted due to critical bug in earlier versions

module github.com/coopnorge/mage

go 1.26.0

require (
	github.com/bmatcuk/doublestar/v4 v4.10.0
	github.com/hashicorp/go-version v1.9.0
	github.com/magefile/mage v1.17.2
	github.com/stretchr/testify v1.12.0
)

require gopkg.in/yaml.v3 v3.0.1 // indirect

tool github.com/magefile/mage

retract [v0.1.0, v0.16.3] // Retracted due to critical bug in earlier versions

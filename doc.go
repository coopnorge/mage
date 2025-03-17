/*
Package mage implements Goop Norge SA's opinionated CI. [Mage] is used to
implement the [targets].

Go and Docker are the minimum requirements to run the targets provided by this
Go module. The targets are designed to be [imported] in local [magefiles] in
repositories. Unless specifically noted targets should work both macOS, Linux
and GitHub Actions, on Windows your mileage may wary.

# Provided target packages

  - [github.com/coopnorge/mage/targets/goapp]
  - [github.com/coopnorge/mage/targets/golib]

# Setup

In the root of the repository initialize a new Go modules and import this
module and mage as as tool.

	go mod init
	go get github.com/coopnorge/mage@latest
	go get -tool github.com/magefile/mage

Configure mage in magefiles/magefile.go (goapp is used for the example).

	import (
		//mage:import
		_ "github.com/coopnorge/mage/targets/goapp"
	)

In the repository add a GitHub Actions Workflow

	on:
	  pull_request: {}
	  push:
	    branches:
	      - main
	jobs:
	  cicd:
	    uses: coopnorge/mage/.github/workflows/mage.yaml@v0 # Use the latest available version
	    permissions:
	      contents: read
	      id-token: write
	      packages: read
	    secrets: inherit

# Run the CI targets locally

List the available targets

	go tool mage -l

Run targets with

	go tool mage <target>

# Extend the pre-defined target packages

Exported packages with targets can be used to compose targets for repositories
that contain applications written in multiple technologies.

[Mage]: https://magefile.org/
[targets]: https://magefile.org/targets/
[imported]: https://magefile.org/importing/
[magefiles]: https://magefile.org/magefiles/
*/
package mage

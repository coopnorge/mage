# Mage CI targets

[Mage](https://magefile.org/) CI targets

Package Documentation: <https://pkg.go.dev/github.com/coopnorge/mage>

## Usage

```shell
go mod init # At the root of the repo
go get github.com/coopnorge/mage
go get -tool github.com/magefile/mage
```

If the repo is using Go 1.23 or lower set the Go toolchain to 1.24.0

```gomod title="go.mod"
module github.com/example/example

go 1.23.0

toolchain go1.24.0

require github.com/coopnorge/mage v0.0.1

require github.com/magefile/mage v1.15.0 // indirect

tool github.com/magefile/mage
```

Create a `magefiles/magefile.go` and import the relevant target for the tech
stack.

### Go app

Ensure that the `main` packages for the commands in the Go modules are in
`cmd/<command>/main.go`. See [Organizing a Go module: Multiple
commands](https://go.dev/doc/modules/layout#multiple-commands) for more
information on the topic.

```go title="magefiles/magefile.go"
package main

import (
	//mage:import
	_ "github.com/coopnorge/mage/targets/goapp"
)
```

#### Targets for Go apps

- [X] Go run
- [X] Go build
- [X] Go tests
- [X] Go linting
- [ ] Go mock generation
- [ ] Docker image build
- [ ] Docker image push
- [ ] Techdocs CI
- [ ] Kubernetes CI
- [ ] Security Scanning

### Go module

```go title="magefiles/magefile.go"
package main

import (
	//mage:import
	_ "github.com/coopnorge/mage/targets/golib"
)
```

#### Targets for Go modules

- [X] Go tests
- [X] Go linting
- [X] Go code generation
- [ ] Go mock generation
- [ ] Techdocs CI
- [ ] Security Scanning

## Run CI

```console
go tool mage <target>
```

## List targets

```console
go tool mage -l
```

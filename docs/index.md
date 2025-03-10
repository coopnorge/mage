# Mage CI targets

[Mage](https://magefile.org/) CI targets

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

### Go module

```go title="magefiles/magefile.go"
package main

import (
	//mage:import
	_ "github.com/coopnorge/mage/targets/golib"
)
```

### Includes

- [ X ] Go tests
- [ X ] Go linting
- [ X ] Go code generation
- [   ] Go mock generation
- [   ] Techdocs CI
- [   ] Security Scanning

## Run CI

```shell
go tool mage <target>
```

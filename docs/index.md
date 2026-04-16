# Mage CI targets

[Mage](https://magefile.org/) CI targets

Package Documentation: <https://pkg.go.dev/github.com/coopnorge/mage>

## Usage

```shell
go mod init # At the root of the repo
go get github.com/coopnorge/mage
go get -tool github.com/magefile/mage
```

If the repo is using Go 1.23 or lower set the Go `toolchain` to 1.24.0

```gomod title="go.mod"
module github.com/example/example

go 1.23.0

toolchain go1.24.0

require github.com/coopnorge/mage v0.16.7

require (
 github.com/bmatcuk/doublestar/v4 v4.9.1 // indirect
 github.com/magefile/mage v1.15.0 // indirect
)

tool github.com/magefile/mage
```

### Go app

Ensure the following project structure is used. See
[helloworld repo](https://github.com/coopnorge/helloworld) for reference and
[Organizing a Go module: Multiple commands](https://go.dev/doc/modules/layout#multiple-commands)
for more information on the topic.

```title="project structure"
helloworld/
├── cmd/
│   ├── helloworld/
│   │   └── main.go
│   └── data-sync/
│       └── main.go
├── internal/
├── go.mod
├── go.sum
magefiles/
└── magefile.go
go.mod
go.sum
...
```

#### Create `magefiles/magefile.go`

Create a `magefiles/magefile.go` file and import the shared mage module and
other relevant targets for the tech stack.

```go title="magefiles/magefile.go"
package main

import (
    //mage:import
    _ "github.com/coopnorge/mage/targets/goapp"
)
```

#### Targets for Go apps

- [ ] Go run
- [x] Go build
- [x] Go tests
- [x] Go linting
- [ ] Go mock generation
- [x] Docker image build
- [x] Docker image push
- [x] Terraform CI
- [x] Policy-bot config validation
- [x] Catalog-info validation
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

- [x] Go tests
- [x] Go linting
- [x] Go code generation
- [ ] Go mock generation
- [x] Policy-bot config validation
- [x] Catalog-info validation
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

## Build Go binaries

Builds binaries for all commands in the `cmd` directory.

```console
go tool mage go:build
```

## Run in GitHub Actions

Add this job to your GitHub actions workflow

### When using `goapp` as target

```yaml
go-app:
  uses: coopnorge/mage/.github/workflows/goapp.yaml@main
  permissions:
    contents: read
    id-token: write
    packages: write
    pull-requests: write
    checks: read
  secrets: inherit
  with:
    oci-image-base: ${{ vars.PALLET_REGISTRY_URL }}
    push-oci-image: ${{ github.ref == 'refs/heads/main' }}
    workload-identity-provider: ${{ vars.PALLET_WORKLOAD_IDENTITY_PROVIDER }}
    service-account: ${{ vars.PALLET_SERVICE_ACCOUNT }}
    tag-based-diff: true
```

If you did not create a system through inventory you have to hard-code the
inputs.

```yaml
oci-image-base: europe-docker.pkg.dev/helloworld-shared-0918
push-oci-image: ${{ github.ref == 'refs/heads/main' }}
workload-identity-provider: projects/889992792607/locations/global/workloadIdentityPools/github-actions/providers/github-actions-provider
service-account: gh-ap-helloworld@helloworld-shared-0918.iam.gserviceaccount.com
```

### When using `golib` as target

```yaml
mage:
  uses: coopnorge/mage/.github/workflows/mage.yaml@main
  permissions:
    contents: read
    id-token: write
    packages: read
  secrets: inherit
```

## Updating OCI tags after build

You can use renovate to create pull request that update your infrastructure. You
need a renovate config. Save this in `.github/renovate.json5`

```json5
{
  extends: ["github>coopnorge/github-workflow-renovate"],
  packageRules: [
    {
      matchDatasources: ["helm"],
      automerge: true,
      rebaseWhen: "behind-base-branch",
      minimumReleaseAge: "0 days",
    },
    {
      matchManagers: ["helm-values"],
      matchDatasources: ["docker"],
      rebaseWhen: "behind-base-branch",
      prHourlyLimit: 0,
      minimumReleaseAge: "0 days",
    },
    {
      branchPrefix: "helm/dev/",
      matchManagers: ["helm-values"],
      matchDatasources: ["docker"],
      automerge: true,
      matchFileNames: ["/(^|/)values-dev.yaml$/"],
      addLabels: ["development"],
      prHourlyLimit: 0,
    },
    {
      branchPrefix: "helm/staging/",
      matchManagers: ["helm-values"],
      matchDatasources: ["docker"],
      automerge: true,
      matchFileNames: ["/(^|/)values-staging.yaml$/"],
      addLabels: ["staging"],
      prHourlyLimit: 0,
    },
    {
      branchPrefix: "helm/production/",
      matchManagers: ["helm-values"],
      matchDatasources: ["docker"],
      automerge: true,
      matchFileNames: ["/(^|/)values-production.yaml$/"],
      addLabels: ["production"],
      prHourlyLimit: 0,
    },
  ],
  "helm-values": {
    managerFilePatterns: ["/(^|/)values(-\\w+)?\\.ya?ml$/"],
  },
}
```

> This renovate config is also valid for updating your helm chart dependencies.

For now you also need a GitHub action job for running renovate. In the future
this might be done by a actual renovate server.

Save this to `.github/workflows/renovate.yaml`

> make sure update the values to your own repo settings

```yaml
on:
  push:
    tags:
      - "v*"
  workflow_dispatch:
    inputs:
      log-level:
        description: "Override default log level"
        required: false
        default: "debug"
        type: string
  schedule:
    # At 07:30 AM and 12:30 PM, every day
    - cron: "30 7,12 * * *"

jobs:
  renovate:
    permissions:
      contents: write
      pull-requests: write
      id-token: write
      issues: write
    uses: coopnorge/github-workflow-renovate/.github/workflows/renovate.yaml@v0
    secrets: inherit
    with:
      config-file: ".github/renovate.json5"
      log-level: ${{ inputs.log-level }}
      gcp-workload-identity-provider: projects/889992792607/locations/global/workloadIdentityPools/github-actions/providers/github-actions-provider
      gcp-service-account: gh-ap-helloworld@helloworld-shared-0918.iam.gserviceaccount.com
```

## Troubleshooting

- During build the command `git status --porcelain` returns the error message
  `fatal: detected dubious ownership in repository at '/src'`

  Solution: Add this lines to `.gitconfig`

```shell
  [safe]
    directory = *
```

- During the build you get the error message

```shell
  ERROR: failed to build: OCI exporter is not supported for the docker
  driver. Switch to a different driver, or turn on the containerd image store,
  and try again. Learn more at https://docs.docker.com/go/build-exporters/
```

Solution:

```shell
  DOCKER_BUILDKIT=1 docker buildx create --use --driver docker-container
```

# github.com/coopnorge/mage

[Mage](https://magefile.org/) CI targets

## Usage

See [Inventory](https://inventory.internal.coop/docs/default/component/mage)
or [docs](./docs).

## Development workflow

### Validation

```console
$ go tool mage validate
```

### List other targets

```console
$ go tool mage -l
```

### Go module documentation preview

```console
$ go install golang.org/x/pkgsite/cmd/pkgsite@latest
$ pkgsite
```

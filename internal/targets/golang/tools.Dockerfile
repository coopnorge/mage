FROM docker.io/library/golang:1.24.0@sha256:3f7444391c51a11a039bf0359ee81cc64e663c17d787ad0e637a4de1a3f62a71 AS golang
FROM golangci/golangci-lint:v1.64.5@sha256:9faef4dda4304c4790a14c5b8c8cd8c2715a8cb754e13f61d8ceaa358f5a454a AS golangci-lint

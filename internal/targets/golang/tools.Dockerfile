FROM docker.io/library/golang:1.24.2@sha256:991aa6a6e4431f2f01e869a812934bd60fbc87fb939e4a1ea54b8494ab9d2fc6 AS golang
FROM golangci/golangci-lint:v2.0.2@sha256:d55581f7797e7a0877a7c3aaa399b01bdc57d2874d6412601a046cc4062cb62e AS golangci-lint

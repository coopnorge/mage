FROM docker.io/library/golang:1.24.4@sha256:db5d0afbfb4ab648af2393b92e87eaae9ad5e01132803d80caef91b5752d289c AS golang
FROM golangci/golangci-lint:v2.1.6@sha256:568ee1c1c53493575fa9494e280e579ac9ca865787bafe4df3023ae59ecf299b AS golangci-lint

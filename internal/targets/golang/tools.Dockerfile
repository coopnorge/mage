FROM docker.io/library/golang:1.24.1@sha256:8678013a2add364dc3d5df2acc2b36893fbbd60ebafa5d5149bc22158512f021 AS golang
FROM golangci/golangci-lint:v1.64.7@sha256:c2f5e6aaa7f89e7ab49f6bd45d8ce4ee5a030b132a5fbcac68b7959914a5a890 AS golangci-lint

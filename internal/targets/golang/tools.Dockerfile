FROM docker.io/library/golang:1.24.1@sha256:ceb568d0de81fbef40ce4fee77eab524a0a0a8536065c51866ad8c59b7a912cf AS golang
FROM golangci/golangci-lint:v1.64.7@sha256:c2f5e6aaa7f89e7ab49f6bd45d8ce4ee5a030b132a5fbcac68b7959914a5a890 AS golangci-lint

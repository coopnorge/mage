FROM docker.io/library/golang:1.24.2@sha256:d9db32125db0c3a680cfb7a1afcaefb89c898a075ec148fdc2f0f646cc2ed509 AS golang
FROM golangci/golangci-lint:v2.1.2@sha256:86f65772316ad8baa4bd5bb1363640fa4054a9df0ae8150b1eef893c4751533c AS golangci-lint

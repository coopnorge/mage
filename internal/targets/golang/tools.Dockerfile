FROM docker.io/library/golang:1.24.2@sha256:1ecc479bc712a6bdb56df3e346e33edcc141f469f82840bab9f4bc2bc41bf91d AS golang-universal
FROM golangci/golangci-lint:v2.1.2@sha256:86f65772316ad8baa4bd5bb1363640fa4054a9df0ae8150b1eef893c4751533c AS golangci-lint-universal


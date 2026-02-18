FROM docker.io/library/golang:1.26.0@sha256:c83e68f3ebb6943a2904fa66348867d108119890a2c6a2e6f07b38d0eb6c25c5 AS golang
FROM golangci/golangci-lint:v2.10.1@sha256:ea84d14c2fef724411be7dc45e09e6ef721d748315252b02df19a7e3113ee763 AS golangci-lint
FROM ghcr.io/yannh/kubeconform:v0.7.0@sha256:85dbef6b4b312b99133decc9c6fc9495e9fc5f92293d4ff3b7e1b30f5611823c AS kubeconform
FROM docker.io/palantirtechnologies/policy-bot:1.41.0@sha256:7efd58e7005071d295c698a7132bee9e305720d2d5204fac80f77019097067fa AS policy-bot-version-tracker
FROM docker.io/hashicorp/terraform:1.14.5@sha256:96d2bc440714bf2b2f2998ac730fd4612f30746df43fca6f0892b2e2035b11bc AS terraform
FROM ghcr.io/terraform-linters/tflint:v0.61.0@sha256:b835d64d66abfdbc146694b918eb3cd733ec772465ad511464d4e8bebbdd6732 AS tflint
FROM docker.io/aquasec/trivy:0.69.1@sha256:1c78ed1ef824ab8bb05b04359d186e4c1229d0b3e67005faacb54a7d71974f73 AS trivy
FROM quay.io/terraform-docs/terraform-docs:0.20.0@sha256:37329e2dc2518e7f719a986a3954b10771c3fe000f50f83fd4d98d489df2eae2 AS terraform-docs

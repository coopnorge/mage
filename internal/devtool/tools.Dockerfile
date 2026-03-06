FROM docker.io/library/golang:1.26.1@sha256:e2ddb153f786ee6210bf8c40f7f35490b3ff7d38be70d1a0d358ba64225f6428 AS golang
FROM golangci/golangci-lint:v2.10.1@sha256:ea84d14c2fef724411be7dc45e09e6ef721d748315252b02df19a7e3113ee763 AS golangci-lint
FROM ghcr.io/yannh/kubeconform:v0.7.0@sha256:85dbef6b4b312b99133decc9c6fc9495e9fc5f92293d4ff3b7e1b30f5611823c AS kubeconform
FROM docker.io/palantirtechnologies/policy-bot:1.41.1@sha256:1babe4d474db1ed7184146fbd2d332138c30121b2f860451347d9dbb825f5a20 AS policy-bot-version-tracker
FROM docker.io/hashicorp/terraform:1.5.7@sha256:9fc0d70fb0f858b0af1fadfcf8b7510b1b61e8b35e7a4bb9ff39f7f6568c321d AS terraform
FROM ghcr.io/terraform-linters/tflint:v0.61.0@sha256:b835d64d66abfdbc146694b918eb3cd733ec772465ad511464d4e8bebbdd6732 AS tflint
FROM docker.io/aquasec/trivy:0.69.3@sha256:bcc376de8d77cfe086a917230e818dc9f8528e3c852f7b1aff648949b6258d1c AS trivy
FROM quay.io/terraform-docs/terraform-docs:0.20.0@sha256:37329e2dc2518e7f719a986a3954b10771c3fe000f50f83fd4d98d489df2eae2 AS terraform-docs

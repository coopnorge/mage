FROM docker.io/library/golang:1.26.0@sha256:fb612b7831d53a89cbc0aaa7855b69ad7b0caf603715860cf538df854d047b84 AS golang
FROM golangci/golangci-lint:v2.10.1@sha256:ea84d14c2fef724411be7dc45e09e6ef721d748315252b02df19a7e3113ee763 AS golangci-lint
FROM ghcr.io/yannh/kubeconform:v0.7.0@sha256:85dbef6b4b312b99133decc9c6fc9495e9fc5f92293d4ff3b7e1b30f5611823c AS kubeconform
FROM docker.io/palantirtechnologies/policy-bot:1.41.0@sha256:7efd58e7005071d295c698a7132bee9e305720d2d5204fac80f77019097067fa AS policy-bot-version-tracker
FROM docker.io/hashicorp/terraform:1.5.7@sha256:9fc0d70fb0f858b0af1fadfcf8b7510b1b61e8b35e7a4bb9ff39f7f6568c321d AS terraform
FROM ghcr.io/terraform-linters/tflint:v0.61.0@sha256:b835d64d66abfdbc146694b918eb3cd733ec772465ad511464d4e8bebbdd6732 AS tflint
FROM docker.io/aquasec/trivy:0.69.2@sha256:3d1f862cb6c4fe13c1506f96f816096030d8d5ccdb2380a3069f7bf07daa86aa AS trivy
FROM quay.io/terraform-docs/terraform-docs:0.20.0@sha256:37329e2dc2518e7f719a986a3954b10771c3fe000f50f83fd4d98d489df2eae2 AS terraform-docs

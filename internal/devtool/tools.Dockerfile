FROM docker.io/library/golang:1.25.6@sha256:4c973c7cf9e94ad40236df4a9a762b44f6680560a7fb8a4e69513b5957df7217 AS golang
FROM golangci/golangci-lint:v2.8.0@sha256:bebcfa63db7df53e417845ed61e4540519cf74fcba22793cdd174b3415a9e4e2 AS golangci-lint
FROM ghcr.io/yannh/kubeconform:v0.7.0@sha256:85dbef6b4b312b99133decc9c6fc9495e9fc5f92293d4ff3b7e1b30f5611823c AS kubeconform
FROM docker.io/palantirtechnologies/policy-bot:1.41.0@sha256:7efd58e7005071d295c698a7132bee9e305720d2d5204fac80f77019097067fa AS policy-bot-version-tracker
FROM docker.io/hashicorp/terraform:1.5.7@sha256:9fc0d70fb0f858b0af1fadfcf8b7510b1b61e8b35e7a4bb9ff39f7f6568c321d AS terraform
FROM ghcr.io/terraform-linters/tflint:v0.60.0@sha256:cef181224b4a9cea521d8f785d50957ea3215b449e2d97e7793f222e2808d188 AS tflint
FROM docker.io/aquasec/trivy:0.68.2@sha256:05d0126976bdedcd0782a0336f77832dbea1c81b9cc5e4b3a5ea5d2ec863aca7 AS trivy
FROM quay.io/terraform-docs/terraform-docs:0.20.0@sha256:37329e2dc2518e7f719a986a3954b10771c3fe000f50f83fd4d98d489df2eae2 AS terraform-docs

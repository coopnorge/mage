FROM ghcr.io/terraform-linters/tflint:v0.57.0 AS tflint
FROM docker.io/hashicorp/terraform:1.5.7@sha256:9fc0d70fb0f858b0af1fadfcf8b7510b1b61e8b35e7a4bb9ff39f7f6568c321d AS terraform
FROM docker.io/aquasec/trivy:0.62.1@sha256:fc10faf341a1d8fa8256c5ff1a6662ef74dd38b65034c8ce42346cf958a02d5d AS trivy

FROM ghcr.io/terraform-linters/tflint:v0.59.1 AS tflint
FROM docker.io/hashicorp/terraform:1.5.7@sha256:9fc0d70fb0f858b0af1fadfcf8b7510b1b61e8b35e7a4bb9ff39f7f6568c321d AS terraform
FROM docker.io/aquasec/trivy:0.67.0@sha256:94711c60051c6cab848a292e3a67f62623fcee361b2bb661f43b17184f4afdac AS trivy

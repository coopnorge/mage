FROM ghcr.io/terraform-linters/tflint:v0.59.1 AS tflint
FROM docker.io/hashicorp/terraform:1.5.7@sha256:9fc0d70fb0f858b0af1fadfcf8b7510b1b61e8b35e7a4bb9ff39f7f6568c321d AS terraform
FROM docker.io/aquasec/trivy:0.67.1@sha256:c0ed528623baf6e250e2225010e5fbb4b91f6983595dafc1beb81ff686ba4734 AS trivy

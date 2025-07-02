FROM ghcr.io/terraform-linters/tflint:v0.58.0 AS tflint
FROM docker.io/hashicorp/terraform:1.5.7@sha256:9fc0d70fb0f858b0af1fadfcf8b7510b1b61e8b35e7a4bb9ff39f7f6568c321d AS terraform
FROM docker.io/aquasec/trivy:0.64.0@sha256:ec9b6eb27be1eb0c97355e34183d3217d3f4f7566c53da5fd4f5bfbfcf87de60 AS trivy
FROM quay.io/terraform-docs/terraform-docs:0.20.0@sha256:37329e2dc2518e7f719a986a3954b10771c3fe000f50f83fd4d98d489df2eae2 as terraform-docs

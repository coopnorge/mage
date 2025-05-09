FROM ghcr.io/terraform-linters/tflint:v0.57.0 AS tflint
FROM docker.io/hashicorp/terraform:1.5.7@sha256:9fc0d70fb0f858b0af1fadfcf8b7510b1b61e8b35e7a4bb9ff39f7f6568c321d AS terraform
FROM docker.io/aquasec/trivy:0.62.0@sha256:cb84170aa1fb6942c11b638969dce0845735c35f40d3bd48851c7f2d83a3c1ae AS trivy

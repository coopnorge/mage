FROM ghcr.io/terraform-linters/tflint:v0.57.0 AS tflint-universal
FROM docker.io/hashicorp/terraform:1.5.7@sha256:c3bc74e7a2a8fab8216cbbedf12a9637db09288806a6aa537b6f397cba04dd93 AS terraform-amd64
FROM docker.io/hashicorp/terraform:1.5.7@sha256:bbc41b888559ece9b3fce0ab834ed7120d898e1a808e4716c37b2d2efb4f2782 AS terraform-arm64
FROM docker.io/aquasec/trivy:0.62.0@sha256:cb84170aa1fb6942c11b638969dce0845735c35f40d3bd48851c7f2d83a3c1ae AS trivy-universal

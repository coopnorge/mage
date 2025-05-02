FROM ghcr.io/terraform-linters/tflint-bundle:v0.48.0.0@sha256:523a85b4bf7af415cbb5f53ac1b56cd7501c8c8125f96c5b5cc19c895d894513 AS tflint-universal
FROM docker.io/hashicorp/terraform:1.5.7@sha256:c3bc74e7a2a8fab8216cbbedf12a9637db09288806a6aa537b6f397cba04dd93 AS terraform-amd64
FROM docker.io/hashicorp/terraform:1.5.7@sha256:bbc41b888559ece9b3fce0ab834ed7120d898e1a808e4716c37b2d2efb4f2782 AS terraform-arm64

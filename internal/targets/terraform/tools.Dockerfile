FROM ghcr.io/terraform-linters/tflint-bundle:v0.48.0.0@sha256:523a85b4bf7af415cbb5f53ac1b56cd7501c8c8125f96c5b5cc19c895d894513 AS tflint
FROM docker.io/hashicorp/terraform:1.5.7 AS terraform

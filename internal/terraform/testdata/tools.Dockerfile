FROM ghcr.io/terraform-linters/tflint:v0.61.0 AS tflint
FROM docker.io/hashicorp/terraform:1.14.5@sha256:96d2bc440714bf2b2f2998ac730fd4612f30746df43fca6f0892b2e2035b11bc AS terraform
FROM docker.io/aquasec/trivy:0.69.1@sha256:1c78ed1ef824ab8bb05b04359d186e4c1229d0b3e67005faacb54a7d71974f73 AS trivy

FROM docker.io/bufbuild/buf:1.50.0@sha256:c34c81ac26044490a10fb5009eb618640834b9048f38d4717538421c6a25e4d7 AS bufsource
FROM docker.io/library/golang:1.24.1@sha256:ceb568d0de81fbef40ce4fee77eab524a0a0a8536065c51866ad8c59b7a912cf AS buf

COPY --from=bufsource /usr/local/bin/buf /usr/local/bin/

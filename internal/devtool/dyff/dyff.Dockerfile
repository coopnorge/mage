FROM alpine:3.24.2@sha256:294b683cb724975bec92580e1e685676bd4b50bda910ddb8c51d4cabeaec77e6 AS downloader

RUN apk add --no-cache curl tar

ARG DYFF_VERSION
ARG TARGETARCH
ARG RELEASE_URL="https://github.com/homeport/dyff/releases/download/v${DYFF_VERSION}/dyff_${DYFF_VERSION}_linux_${TARGETARCH}.tar.gz"

WORKDIR /tmp
RUN curl -L ${RELEASE_URL} | tar -xz

FROM alpine:3.24.2@sha256:294b683cb724975bec92580e1e685676bd4b50bda910ddb8c51d4cabeaec77e6
COPY --from=downloader /tmp/dyff /usr/local/bin/dyff
ENTRYPOINT ["dyff"]





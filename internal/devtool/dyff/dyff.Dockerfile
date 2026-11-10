FROM alpine:3.24.3@sha256:be2656061b536c1883c4034790e21d0a09c5aaaf51968b89368a3c0371ea2c9a AS downloader

RUN apk add --no-cache curl tar

ARG DYFF_VERSION
ARG TARGETARCH
ARG RELEASE_URL="https://github.com/homeport/dyff/releases/download/v${DYFF_VERSION}/dyff_${DYFF_VERSION}_linux_${TARGETARCH}.tar.gz"

WORKDIR /tmp
RUN curl -L ${RELEASE_URL} | tar -xz

FROM alpine:3.24.3@sha256:be2656061b536c1883c4034790e21d0a09c5aaaf51968b89368a3c0371ea2c9a
COPY --from=downloader /tmp/dyff /usr/local/bin/dyff
ENTRYPOINT ["dyff"]





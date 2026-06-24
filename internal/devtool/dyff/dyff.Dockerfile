FROM alpine:3.24.0@sha256:a2d49ea686c2adfe3c992e47dc3b5e7fa6e6b5055609400dc2acaeb241c829f4 AS downloader

RUN apk add --no-cache curl tar

ARG DYFF_VERSION
ARG TARGETARCH
ARG RELEASE_URL="https://github.com/homeport/dyff/releases/download/v${DYFF_VERSION}/dyff_${DYFF_VERSION}_linux_${TARGETARCH}.tar.gz"

WORKDIR /tmp
RUN curl -L ${RELEASE_URL} | tar -xz

FROM alpine:3.24.0@sha256:a2d49ea686c2adfe3c992e47dc3b5e7fa6e6b5055609400dc2acaeb241c829f4
COPY --from=downloader /tmp/dyff /usr/local/bin/dyff
ENTRYPOINT ["dyff"]





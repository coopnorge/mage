FROM alpine:3.23.3@sha256:25109184c71bdad752c8312a8623239686a9a2071e8825f20acb8f2198c3f659 AS downloader

RUN apk add --no-cache curl tar

ARG DYFF_VERSION
ARG TARGETARCH
ARG RELEASE_URL="https://github.com/homeport/dyff/releases/download/v${DYFF_VERSION}/dyff_${DYFF_VERSION}_linux_${TARGETARCH}.tar.gz"

WORKDIR /tmp
RUN curl -L ${RELEASE_URL} | tar -xz

FROM alpine:3.23.3@sha256:25109184c71bdad752c8312a8623239686a9a2071e8825f20acb8f2198c3f659
COPY --from=downloader /tmp/dyff /usr/local/bin/dyff
ENTRYPOINT ["dyff"]





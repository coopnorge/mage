FROM alpine:3.24.4@sha256:f59abe656b692d901e184be3445e084786602108599d2a51d229c0f5a48c71bf AS downloader

RUN apk add --no-cache curl tar

ARG DYFF_VERSION
ARG TARGETARCH
ARG RELEASE_URL="https://github.com/homeport/dyff/releases/download/v${DYFF_VERSION}/dyff_${DYFF_VERSION}_linux_${TARGETARCH}.tar.gz"

WORKDIR /tmp
RUN curl -L ${RELEASE_URL} | tar -xz

FROM alpine:3.24.4@sha256:f59abe656b692d901e184be3445e084786602108599d2a51d229c0f5a48c71bf
COPY --from=downloader /tmp/dyff /usr/local/bin/dyff
ENTRYPOINT ["dyff"]





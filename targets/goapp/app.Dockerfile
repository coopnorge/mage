FROM docker.io/kvij/scuttle:1.1.12@sha256:a74554fa3a918c86e1ac1a0e992c5c7e764d3e60055bc859457592158702b91b AS scuttle
FROM docker.io/library/alpine:3.23.0@sha256:51183f2cfa6320055da30872f211093f9ff1d3cf06f39a0bdb212314c5dc7375 AS alpine

FROM alpine AS runtime

ARG APP
ARG TARGETPLATFORM
ARG BINARY

ARG group_name=${APP}
ARG user_name=${BINARY}
ARG workdir=/var/opt/${APP}

RUN \
    addgroup -S ${group_name} && \
    adduser -S -H -D -h ${workdir} -G ${group_name} ${user_name} && \
    mkdir -vp ${workdir} && \
    chown -R ${user_name}:${group_name} ${workdir} && \
    true

RUN mkdir /root/tmp
WORKDIR /root/tmp
COPY ./var/${APP}/bin/${TARGETPLATFORM}/${BINARY} /usr/local/bin/
RUN chmod +x /usr/local/bin/${BINARY}

COPY --from=scuttle /scuttle /usr/local/bin/scuttle

RUN \
    apk --no-cache update &&\
    apk --no-cache add \
        tzdata \
    && \
    ln -vfs /usr/share/zoneinfo/UTC /etc/localtime && \
    true

ARG GIT_REPOSITORY_URL
ARG GIT_COMMIT_SHA

LABEL org.opencontainers.image.source=${GIT_REPOSITORY_URL}
LABEL org.opencontainers.image.revision=${GIT_COMMIT_SHA}

USER ${user_name}:${group_name}
WORKDIR ${workdir}

ENV DD_GIT_REPOSITORY_URL=${GIT_REPOSITORY_URL}
ENV DD_GIT_COMMIT_SHA=${GIT_COMMIT_SHA}

ENV APP_BIN=/usr/local/bin/${BINARY}

CMD ["sh", "-c", "exec ${APP_BIN}"]

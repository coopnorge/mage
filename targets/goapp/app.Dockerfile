FROM docker.io/kvij/scuttle:1.1.17@sha256:95778d965791ff21fae43c00dc92ecf6ff61e58835d1d9fbc0bb2d4bd66ecec4 AS scuttle
FROM docker.io/library/alpine:3.24.2@sha256:294b683cb724975bec92580e1e685676bd4b50bda910ddb8c51d4cabeaec77e6 AS alpine

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

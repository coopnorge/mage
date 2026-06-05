FROM docker.io/kvij/scuttle:1.1.15@sha256:1b3fb35aa13dd80b30ecceee9f913156eace6cf7656129a99a2dfad5010f68a4 AS scuttle
FROM docker.io/library/alpine:3.23.4@sha256:5b10f432ef3da1b8d4c7eb6c487f2f5a8f096bc91145e68878dd4a5019afde11 AS alpine

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

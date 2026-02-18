FROM docker.io/kvij/scuttle:1.1.14@sha256:385f44bddd506fbff1256f6895a768e1bfd767620e71ab90b55cae6f48fe7706 AS scuttle
FROM docker.io/library/alpine:3.23.3@sha256:25109184c71bdad752c8312a8623239686a9a2071e8825f20acb8f2198c3f659 AS alpine

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

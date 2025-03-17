FROM docker.io/kvij/scuttle:1.1.10@sha256:924a0504f0a8574dd53edc5b6a58f17c27ef44351e0c92d4ee257836ddd38483 AS scuttle
FROM docker.io/library/alpine:3.21.3@sha256:a8560b36e8b8210634f77d9f7f9efd7ffa463e380b75e2e74aff4511df3ef88c AS alpine

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

ENV __dba_app_executable=${BINARY}
ENV __dba_app_name=${BINARY}

HEALTHCHECK CMD ["sh", "-c", "exec /usr/local/bin/${__dba_app_executable} healthcheck"]
ENTRYPOINT ["sh", "-c", "exec /usr/local/bin/${__dba_app_executable} \"${@}\"", "/usr/local/bin/${__dba_app_executable}"]

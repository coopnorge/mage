FROM node:lts-slim@sha256:fe64023c6490eb001c7a28e9f92ef8deb6e40e1b7fc5352d695dcaef59e1652d AS builder

ARG BUILD_SCRIPT=build

WORKDIR /app
COPY . .

RUN --mount=type=secret,id=github_token \
    GITHUB_TOKEN=$(cat /run/secrets/github_token) npm install

RUN npm run ${BUILD_SCRIPT}

FROM builder AS runner
WORKDIR /app

# Uncomment the following line in case you want to disable telemetry during runtime.
ENV NEXT_TELEMETRY_DISABLED=1

ARG GROUP=nodejs
ARG USER=nextjs
ARG DISTFOLDER=.next

RUN addgroup --system --gid 1001 ${GROUP}
RUN adduser --system --uid 1001 ${USER}

COPY --from=builder --chown=${USER}:${GROUP} /app/${DISTFOLDER}/standalone ./
COPY --from=builder --chown=${USER}:${GROUP} /app/${DISTFOLDER}/static ./static
COPY --from=builder --chown=${USER}:${GROUP} /app/public ./public


ARG GIT_REPOSITORY_URL
ARG GIT_COMMIT_SHA


LABEL org.opencontainers.image.source=${GIT_REPOSITORY_URL}
LABEL org.opencontainers.image.revision=${GIT_COMMIT_SHA}

USER ${USER}:${GROUP}

ENV DD_GIT_REPOSITORY_URL=${GIT_REPOSITORY_URL}
ENV DD_GIT_COMMIT_SHA=${GIT_COMMIT_SHA}
EXPOSE 3000
ENV PORT=3000

# server.js is created by next build from the standalone output
# https://nextjs.org/docs/pages/api-reference/next-config-js/output
ENV HOSTNAME="0.0.0.0"

CMD ["node", "server.js"]

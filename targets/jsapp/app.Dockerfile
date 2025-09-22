FROM node:lts-slim@sha256:fe64023c6490eb001c7a28e9f92ef8deb6e40e1b7fc5352d695dcaef59e1652d AS builder

ARG ENVIRONMENT=production
ARG BUILD_SCRIPT=build
ENV BUILD_ENV=$ENVIRONMENT
RUN apt-get update && apt-get install -y make && apt-get clean && rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY . .

RUN --mount=type=secret,id=github_token \
    GITHUB_TOKEN=$(cat /run/secrets/github_token) npm install

RUN npm run build

# Production image, copy all the files and run next
FROM builder AS runner
WORKDIR /app

# Uncomment the following line in case you want to disable telemetry during runtime.
ENV NEXT_TELEMETRY_DISABLED=1

RUN addgroup --system --gid 1001 nodejs
RUN adduser --system --uid 1001 nextjs

# Automatically leverage output traces to reduce image size
# https://nextjs.org/docs/advanced-features/output-file-tracing
COPY --from=builder --chown=nextjs:nodejs /app/.next/standalone ./
COPY --from=builder --chown=nextjs:nodejs /app/.next/static ./static
COPY --from=builder /app/public ./public

ARG APP=nodejs
# ARG TARGETPLATFORM
ARG BINARY=nextjs

ARG GIT_REPOSITORY_URL
ARG GIT_COMMIT_SHA

ARG group_name=${APP}
ARG user_name=${BINARY}

LABEL org.opencontainers.image.source=${GIT_REPOSITORY_URL}
LABEL org.opencontainers.image.revision=${GIT_COMMIT_SHA}

USER ${user_name}:${group_name}

ENV DD_GIT_REPOSITORY_URL=${GIT_REPOSITORY_URL}
ENV DD_GIT_COMMIT_SHA=${GIT_COMMIT_SHA}
EXPOSE 3000
ENV PORT=3000

# server.js is created by next build from the standalone output
# https://nextjs.org/docs/pages/api-reference/next-config-js/output
ENV HOSTNAME="0.0.0.0"

ENV APP_ENV=$BUILD_ENV
CMD ["node", "server.js"]

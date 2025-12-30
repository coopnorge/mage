FROM alpine AS download

ARG POLICY_BOT_VERSION

RUN apk add --no-cache curl \
    && curl -L "https://github.com/palantir/policy-bot/releases/download/v$POLICY_BOT_VERSION/policy-bot-$POLICY_BOT_VERSION.tgz" -o "/tmp/policy-bot.tgz" \
    && mkdir "/app" \
    && tar xzvf "/tmp/policy-bot.tgz" --strip-components=1 -C /app

FROM scratch AS policy-bot
# this is modeled like the upstream dockerfile for policy-bot
# ref: https://github.com/palantir/policy-bot/blob/develop/docker/Dockerfile
ARG TARGETARCH

WORKDIR /policy-bot

COPY --from=download /app/ /policy-bot

ENTRYPOINT ["bin/linux-$TARGETARCH/policy-bot"]

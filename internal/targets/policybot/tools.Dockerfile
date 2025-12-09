FROM alpine AS download

ENV POLICY_BOT_VERSION=1.40.0

RUN apk add --no-cache curl \
  && curl -L "https://github.com/palantir/policy-bot/releases/download/v$POLICY_BOT_VERSION/policy-bot-$POLICY_BOT_VERSION.tgz" -o "/tmp/policy-bot.tgz" \
  && mkdir "/app" \
  && tar xzvf "/tmp/policy-bot.tgz" --strip-components=1 -C /app

FROM scratch AS policy-bot

ARG TARGETARCH

WORKDIR /app

COPY --from=download /app/ /app

ENTRYPOINT ["bin/linux-$TARGETARCH/policy-bot"]

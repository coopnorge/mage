FROM alpine AS download

ENV POLICY_BOT_VERSION=1.40.0

RUN apk add --no-cache curl \
  && curl -L "https://github.com/palantir/policy-bot/releases/download/v$POLICY_BOT_VERSION/policy-bot-$POLICY_BOT_VERSION.tgz" -o "/tmp/policy-bot.tgz" \
  && mkdir "/app" \
  && tar xzvf "/tmp/policy-bot.tgz" --strip-components=1 -C /app

FROM scratch AS policy-bot

WORKDIR /app

COPY --from=download /app/ /app

FROM policy-bot AS policy-bot-amd64

ENTRYPOINT ["bin/linux-amd64/policy-bot"]

FROM policy-bot AS policy-bot-arm64

ENTRYPOINT ["bin/linux-arm64/policy-bot"]

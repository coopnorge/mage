FROM golang:1.25.4@sha256:f60eaa87c79e604967c84d18fd3b151b3ee3f033bcdade4f3494e38411e60963 AS build

WORKDIR /app

RUN git clone -b add-validate https://github.com/AtzeDeVries/policy-bot.git .

RUN CGO_ENABLED=0 go build -o policy-bot .

FROM alpine:3.22.2@sha256:4b7ce07002c69e8f3d704a9c5d6fd3053be500b7f1c69fc0d80990c2ad8dd412 AS policy-bot

RUN apk add --no-cache curl tar

ENV POLICYBOT_VERSION=1.39.3
ENV POLICYBOT_URL="https://github.com/palantir/policy-bot/releases/download/v${POLICYBOT_VERSION}/policy-bot-${POLICYBOT_VERSION}.tgz"

# Download and extract binaries
RUN curl -sSL "$POLICYBOT_URL" -o /tmp/policy-bot.tgz \
 && mkdir -p /tmp/pb \
 && tar -xzf /tmp/policy-bot.tgz -C /tmp/pb \
 && cp -r "/tmp/pb/policy-bot-$POLICYBOT_VERSION/" "/policy-bot-$POLICYBOT_VERSION" \
 && rm -rf /tmp/pb /tmp/policy-bot.tgz

# Copy your checker + config
COPY --from=build /app/policy-bot /policy-bot-$POLICYBOT_VERSION/bin/linux/policy-bot

COPY policy-bot.yml /secrets/policy-bot.yml

WORKDIR /policy-bot-$POLICYBOT_VERSION

ENTRYPOINT ["bin/linux/policy-bot"]

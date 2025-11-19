FROM golang:1.25.4@sha256:f60eaa87c79e604967c84d18fd3b151b3ee3f033bcdade4f3494e38411e60963 AS build

COPY policy-bot-config-check/ /app/

WORKDIR /app

RUN CGO_ENABLED=0 go build -o policy-bot-config-check .

FROM docker.io/palantirtechnologies/policy-bot:1.39.3@sha256:a96dbd467736b37b3fef99819b7571655c5bbdcd3641aa0bf34afd0ea49d161a

COPY --from=build /app/policy-bot-config-check /usr/local/bin/policy-bot-config-check
COPY --from=build /app/policy-bot.yml /secrets/policy-bot.yml

ENTRYPOINT ["/usr/local/bin/policy-bot-config-check"]

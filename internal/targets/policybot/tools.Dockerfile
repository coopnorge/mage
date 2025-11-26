# When validate is available in the official release, we can remove the build stage
FROM golang:1.25.4@sha256:f60eaa87c79e604967c84d18fd3b151b3ee3f033bcdade4f3494e38411e60963 AS build

WORKDIR /app

RUN git clone -b add-validate https://github.com/AtzeDeVries/policy-bot.git .

RUN CGO_ENABLED=0 go build -o policy-bot .

FROM palantirtechnologies/policy-bot:1.39.3@sha256:a96dbd467736b37b3fef99819b7571655c5bbdcd3641aa0bf34afd0ea49d161a AS policy-bot

COPY --from=build /app/policy-bot bin/linux-arm64/policy-bot
COPY --from=build /app/policy-bot bin/linux-amd64/policy-bot

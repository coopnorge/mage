FROM golang:1.25.4@sha256:f60eaa87c79e604967c84d18fd3b151b3ee3f033bcdade4f3494e38411e60963 AS build

WORKDIR /app

RUN git clone -b add-validate https://github.com/AtzeDeVries/policy-bot.git .

RUN CGO_ENABLED=0 go build -o policy-bot .

FROM alpine:3.22.2@sha256:4b7ce07002c69e8f3d704a9c5d6fd3053be500b7f1c69fc0d80990c2ad8dd412 AS policy-bot

COPY --from=build /app/policy-bot /usr/local/bin/policy-bot

ENTRYPOINT ["/usr/local/bin/policy-bot"]

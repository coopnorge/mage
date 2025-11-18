FROM golang:1.25.4@sha256:f60eaa87c79e604967c84d18fd3b151b3ee3f033bcdade4f3494e38411e60963 AS mock-server

COPY mock-server/ /app/

WORKDIR /app

# Build static binary
RUN CGO_ENABLED=0 go build -o mock-server .

FROM golang:1.25.4@sha256:f60eaa87c79e604967c84d18fd3b151b3ee3f033bcdade4f3494e38411e60963 AS supervisor

COPY supervisor/ /app/

WORKDIR /app

# Build static binary
RUN CGO_ENABLED=0 go build -o supervisor .

FROM docker.io/palantirtechnologies/policy-bot:1.39.3@sha256:a96dbd467736b37b3fef99819b7571655c5bbdcd3641aa0bf34afd0ea49d161a

COPY --from=mock-server /app/mock-server /usr/local/bin/mock-server
COPY --from=supervisor /app/supervisor /usr/local/bin/supervisor

ADD policy-bot.yml /secrets/

ENTRYPOINT ["/usr/local/bin/supervisor"]

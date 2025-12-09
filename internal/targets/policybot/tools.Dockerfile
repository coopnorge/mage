# When validate is available in the official release, we can remove the build stage
FROM golang:1.25.5@sha256:20b91eda7a9627c127c0225b0d4e8ec927b476fa4130c6760928b849d769c149 AS build

WORKDIR /app

RUN git clone -b add-validate https://github.com/AtzeDeVries/policy-bot.git .

RUN CGO_ENABLED=0 go build -o policy-bot .

FROM palantirtechnologies/policy-bot:1.40.0@sha256:5663e52393d080ab26f9059d81d2fad3eeb4da876719ed00d496acc5b55a510f AS policy-bot

COPY --from=build /app/policy-bot bin/linux-arm64/policy-bot
COPY --from=build /app/policy-bot bin/linux-amd64/policy-bot

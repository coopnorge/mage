FROM node:25-alpine@sha256:26ded7f450a0ad37241d2ae97ea521a59cb551a1785c8a950f74b0a291ad3aea AS backstage-entity-validator-arm64

ENV BACKSTAGE_ENTITY_VALIDATOR_VERSION=0.5.0

RUN npm install -g npm && npm install --global @roadiehq/backstage-entity-validator@$BACKSTAGE_ENTITY_VALIDATOR_VERSION

RUN addgroup -S toolgroup && adduser -S tooluser -G toolgroup

USER tooluser

ENTRYPOINT ["validate-entity"]

CMD ["validate-entity"]

WORKDIR /src

FROM backstage-entity-validator-arm64 AS backstage-entity-validator-amd64

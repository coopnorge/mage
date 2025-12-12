FROM node:24-slim@sha256:04d9cbb7297edb843581b9bb9bbed6d7efb459447d5b6ade8d8ef988e6737804 AS backstage-entity-validator

ENV BACKSTAGE_ENTITY_VALIDATOR_VERSION=0.5.1

RUN npm install -g npm && npm install --global @roadiehq/backstage-entity-validator@$BACKSTAGE_ENTITY_VALIDATOR_VERSION

ENTRYPOINT ["validate-entity"]

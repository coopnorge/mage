FROM node:25-slim@sha256:9d346b36433145de8bde85fb11f37820ae7b3fcf0b0771d0fbcfa01c79607909 AS backstage-entity-validator

ENV BACKSTAGE_ENTITY_VALIDATOR_VERSION=0.5.0

RUN npm install -g npm && npm install --global @roadiehq/backstage-entity-validator@$BACKSTAGE_ENTITY_VALIDATOR_VERSION

ENTRYPOINT ["validate-entity"]

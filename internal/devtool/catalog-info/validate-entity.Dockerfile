FROM node:24-slim@sha256:61bf992754b4ab288d41cb92c25392195d0035871b4723a0abd30a49dcba356c AS backstage-entity-validator

ARG BACKSTAGE_ENTITY_VALIDATOR_VERSION

RUN npm install -g npm && npm install --global @roadiehq/backstage-entity-validator@$BACKSTAGE_ENTITY_VALIDATOR_VERSION

ENTRYPOINT ["validate-entity"]

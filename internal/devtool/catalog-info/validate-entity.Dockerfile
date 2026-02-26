FROM node:24-slim@sha256:e8e2e91b1378f83c5b2dd15f0247f34110e2fe895f6ca7719dbb780f929368eb AS backstage-entity-validator

ARG BACKSTAGE_ENTITY_VALIDATOR_VERSION

RUN npm install -g npm && npm install --global @roadiehq/backstage-entity-validator@$BACKSTAGE_ENTITY_VALIDATOR_VERSION

ENTRYPOINT ["validate-entity"]

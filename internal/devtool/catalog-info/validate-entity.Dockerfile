FROM node:24-slim@sha256:879b21aec4a1ad820c27ccd565e7c7ed955f24b92e6694556154f251e4bdb240 AS backstage-entity-validator

ARG BACKSTAGE_ENTITY_VALIDATOR_VERSION

RUN npm install -g npm && npm install --global @roadiehq/backstage-entity-validator@$BACKSTAGE_ENTITY_VALIDATOR_VERSION

ENTRYPOINT ["validate-entity"]

FROM node:24-slim@sha256:732887b2c52d7251840a19d203afc68a61532b1ff8b5bf3b76efd49ded39e908 AS backstage-entity-validator

ARG BACKSTAGE_ENTITY_VALIDATOR_VERSION

RUN npm install -g npm && npm install --global @roadiehq/backstage-entity-validator@$BACKSTAGE_ENTITY_VALIDATOR_VERSION

ENTRYPOINT ["validate-entity"]

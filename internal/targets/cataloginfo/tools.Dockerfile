FROM node:24-slim@sha256:b83af04d005d8e3716f542469a28ad2947ba382f6b4a76ddca0827a21446a540 AS backstage-entity-validator

ENV BACKSTAGE_ENTITY_VALIDATOR_VERSION=0.5.1

RUN npm install -g npm && npm install --global @roadiehq/backstage-entity-validator@$BACKSTAGE_ENTITY_VALIDATOR_VERSION

ENTRYPOINT ["validate-entity"]


# renovate: datasource=github-releases depName=hashicorp/terraform versioning=docker
FROM hashicorp/terraform:1.5.7 as tf

# renovate: datasource=github-releases depName=weaveworks/tf-runner versioning=docker
COPY --from=ghcr.io/weaveworks/tf-runner:v0.16.0-rc.2 --chown=65532:65532 /usr/local/bin/tf-runner /usr/local/bin/tf-runner

# renovate: datasource=npm depName=@bitwarden/cli versioning=npm
ARG BW_CLI_VERSION=2023.9.0

USER root

RUN apk add --no-cache npm && npm i -g @bitwarden/cli@${BW_CLI_VERSION}

COPY --from=tf /bin/terraform /usr/local/bin/terraform

RUN chmod +x /usr/local/bin/terraform

USER 65532:65532
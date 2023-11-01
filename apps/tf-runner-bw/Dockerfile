# renovate: datasource=github-releases depName=hashicorp/terraform versioning=docker
FROM hashicorp/terraform:1.6.2 as terraform

# renovate: datasource=github-releases depName=weaveworks/tf-runner versioning=docker
FROM ghcr.io/weaveworks/tf-runner:v0.16.0-rc.2 as runner
USER root

# renovate: datasource=npm depName=@bitwarden/cli versioning=npm
ARG BW_CLI_VERSION=2023.10.0

RUN apk add --no-cache npm && npm i -g @bitwarden/cli@${BW_CLI_VERSION}

COPY --from=terraform --chown=65532:65532 /bin/terraform /usr/local/bin/terraform

USER 65532:65532
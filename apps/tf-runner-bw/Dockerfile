ARG VERSION=0.16.0-rc.3
ARG TARGETPLATFORM

# renovate: datasource=github-releases depName=hashicorp/terraform versioning=docker
FROM hashicorp/terraform:1.7.5 as terraform

# renovate: datasource=github-releases depName=weaveworks/tf-runner versioning=docker
FROM ghcr.io/weaveworks/tf-runner:v${VERSION} as runner
USER root

# renovate: datasource=npm depName=@bitwarden/cli versioning=npm
ARG BW_VERSION=2024.2.1

RUN apk add --no-cache npm && npm i -g @bitwarden/cli@${BW_VERSION}

COPY --from=terraform --chown=65532:65532 /bin/terraform /usr/local/bin/terraform

USER 65532:65532
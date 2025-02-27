ARG VERSION=0.16.0-rc.4
ARG TARGETPLATFORM

# renovate: datasource=github-releases depName=hashicorp/terraform versioning=docker
FROM hashicorp/terraform:1.11.0 AS terraform

# renovate: datasource=github-releases depName=flux-iac/tf-runner versioning=docker
FROM ghcr.io/flux-iac/tf-runner:v${VERSION} AS runner
USER root

COPY --from=terraform --chown=65532:65532 /bin/terraform /usr/local/bin/terraform

USER 65532:65532
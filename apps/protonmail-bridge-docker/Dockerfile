FROM golang:1.23 AS build

ARG TARGETPLATFORM
# renovate: datasource=github-releases depName=ProtonMail/proton-bridge versioning=docker
ARG VERSION=3.6.1

# Install dependencies
RUN apt-get update && apt-get install -y --no-install-recommends git build-essential libsecret-1-dev

# Build
WORKDIR /build/
COPY build.sh /build/
RUN bash build.sh ${VERSION} ${TARGETPLATFORM}

FROM ubuntu:noble

EXPOSE 25/tcp
EXPOSE 143/tcp

# Install dependencies and protonmail bridge
RUN apt-get update \
    && apt-get install -y --no-install-recommends socat pass libsecret-1-0 ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Copy bash scripts
COPY gpgparams entrypoint.sh /protonmail/

# Copy protonmail
COPY --from=build /build/proton-bridge/bridge /protonmail/
COPY --from=build /build/proton-bridge/proton-bridge /protonmail/

ENTRYPOINT ["bash", "/protonmail/entrypoint.sh"]

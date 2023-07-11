FROM alpine:3.18 AS download

WORKDIR /usr/local/bin
RUN apk update && apk add curl unzip jq
RUN export VER=$(curl -H "Accept: application/vnd.github+json" https://api.github.com/repos/bitwarden/clients/releases | jq  -r 'sort_by(.published_at) | reverse | .[].name | select( index("CLI") )' | sed 's:.*CLI v::' | head -n 1) && \
curl -LO "https://github.com/bitwarden/clients/releases/download/cli-v{$VER}/bw-linux-{$VER}.zip" && echo $VER \
&& unzip *.zip && chmod +x ./bw

RUN rm *.zip

FROM ghcr.io/weaveworks/tf-runner:v0.15.1
COPY --from=download /usr/local/bin/bw /usr/local/bin/bw

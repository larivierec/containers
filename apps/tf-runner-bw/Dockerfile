FROM ubuntu

RUN apt update -y \
  && apt install -y \
    ca-certificates \
    git \
    gnupg \
    libssl-dev \
    openssh-client \
    tini \
    unzip \
    curl \
    jq

RUN echo 'hosts: files dns' > /etc/nsswitch.conf
RUN addgroup --gid 65532 --system runner \
  && adduser --uid 65532 --system runner --group

RUN export VER=$(curl -H "Accept: application/vnd.github+json" https://api.github.com/repos/bitwarden/clients/releases | jq  -r 'sort_by(.published_at) | reverse | .[].name | select( index("CLI") )' | sed 's:.*CLI v::' | head -n 1) && \
curl -LO "https://github.com/bitwarden/clients/releases/download/cli-v{$VER}/bw-linux-{$VER}.zip" \
&& unzip *.zip && chmod +x ./bw && rm *.zip && mv ./bw /usr/local/bin

LABEL org.opencontainers.image.source="https://github.com/weaveworks/tf-controller"

COPY --from=ghcr.io/weaveworks/tf-runner:v0.16.0-rc.2 --chown=65532:65532 /usr/local/bin/tf-runner /usr/local/bin/tf-runner 
COPY --from=public.ecr.aws/hashicorp/terraform:1.5 --chown=65532:65532 /bin/terraform /usr/local/bin/terraform 

RUN chmod +x /usr/local/bin/terraform \
  && chmod +x /usr/local/bin/tf-runner \
  && chmod +x /usr/local/bin/bw \
  && chown 65532:65532 /usr/local/bin/bw

WORKDIR /home/runner
USER 65532:65532
ENV GNUPGHOME=/tmp
ENTRYPOINT [ "/bin/tini", "--", "tf-runner" ]
ARG ALPINE_GIT_VERSION=latest
FROM alpine/git:${ALPINE_GIT_VERSION}

ARG TARGETOS
ARG TARGETARCH

ARG MASKCMD_VERSION=v0.0.9

RUN wget -O /usr/local/bin/maskcmd https://github.com/caseycs/maskcmd/releases/download/$MASKCMD_VERSION/maskcmd-$TARGETOS-$TARGETARCH \
    && chmod +x /usr/local/bin/maskcmd

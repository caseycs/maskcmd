ARG TARGETOS
ARG TARGETARCH

ARG ALPINE_GIT_VERSION
FROM alpine/git:$ALPINE_GIT_VERSION

ARG MASKCMD_VERSION

RUN wget -o /usr/local/bin/maskcmd https://github.com/caseycs/maskcmd/releases/download/$MASKCMD_VERSION/maskcmd-$TARGETOS-$TARGETARCH \
    && chmod +x /usr/local/bin/maskcmd

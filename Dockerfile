ARG TARGETOS
ARG TARGETARCH

ARG ALPINE_GIT_VERSION=latest
FROM alpine/git:$ALPINE_GIT_VERSION

ARG MASKCMD_VERSION=0.0.4

RUN echo $ALPINE_GIT_VERSION $MASKCMD_VERSION

RUN wget -o /usr/local/bin/maskcmd https://github.com/caseycs/maskcmd/releases/download/$MASKCMD_VERSION/maskcmd-$TARGETOS-$TARGETARCH \
    && chmod +x /usr/local/bin/maskcmd

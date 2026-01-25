ARG GO_VERSION=1.24
ARG ALPINE_GIT_VERSION=latest

FROM golang:${GO_VERSION}-alpine AS builder
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY *.go ./
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o maskcmd .

FROM alpine/git:${ALPINE_GIT_VERSION}
COPY --from=builder /build/maskcmd /usr/local/bin/maskcmd

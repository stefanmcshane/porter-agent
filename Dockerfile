# syntax=docker/dockerfile:1.1.7-experimental

# Base Go environment
# -------------------
FROM golang:1.18-alpine as base
WORKDIR /porter

RUN apk update && apk add --no-cache gcc musl-dev git protoc

COPY go.mod go.sum ./
RUN go mod download

COPY main.go ./
COPY /api ./api
COPY /cli ./cli
COPY /internal ./internal
COPY /pkg ./pkg


# Go build environment
# --------------------
FROM base AS build-go

# build proto files
ARG version=production

RUN go build -a -o ./bin/agent .
RUN go build -a -o ./bin/agent-cli ./cli

# Deployment environment
# ----------------------
FROM alpine
RUN apk update

COPY --from=build-go /porter/bin/agent /porter/
COPY --from=build-go /porter/bin/agent-cli /porter/

ENV SERVER_PORT=8080
ENV SERVER_TIMEOUT_READ=5s
ENV SERVER_TIMEOUT_WRITE=10s
ENV SERVER_TIMEOUT_IDLE=15s

EXPOSE 8080
CMD /porter/agent

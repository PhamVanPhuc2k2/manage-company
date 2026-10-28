FROM golang:1.27-alpine@sha256:f92b6ef800e499660581efdabdf25d9d817a9d124eaf900924f0504e7e27e12d AS base

WORKDIR /app

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

FROM base AS dev

ARG BINARY
ENV BINARY=${BINARY}

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go install github.com/air-verse/air@v1.61.7

EXPOSE 8080

CMD ["sh", "-c", "air -c .air.${BINARY}.toml"]

FROM base AS builder

ARG BINARY
ARG VERSION=dev
ARG GIT_SHA=unknown
ARG BUILD_TIME=unknown

COPY . .

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build \
        --trimpath \
        -ldflags="-s -w \
            -X main.version=${VERSION} \
            -X main.gitSHA=${GIT_SHA} \
            -X main.buildTime=${BUID_TIME}" \
        -o /out/server \
        ./cmd/${BINARY}

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    mkdir -p /out/extra && \
    if ["${BINARY}" = "api"]; then \
        CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" \
            -o /out/extra/seed ./cmd/seed; \
    fi

FROM alpine:3.24@sha256:294b683cb724975bec92580e1e685676bd4b50bda910ddb8c51d4cabeaec77e6 AS pord

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

RUN addgroup -g 1001 -S app && \
    adduser -u 1001 -S app -G app -H -D

COPY --from=builder /out/server /usr/local/bin/server
COPY --from=builder /out/extra/ /usr/local/bin/

USER app

ENV TZ=Aisa/Ho_Chi_Minh

ENTRYPOINT ["/usr/local/bin/server"]
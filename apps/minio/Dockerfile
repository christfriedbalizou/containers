ARG VERSION

FROM --platform=$BUILDPLATFORM golang:1.25-alpine AS builder

ARG VERSION
ARG TARGETOS
ARG TARGETARCH

WORKDIR /src

RUN apk add --no-cache git wget

RUN git clone --depth 1 --branch "${VERSION}" https://github.com/minio/minio.git .

ENV CGO_ENABLED=0
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH go build -trimpath -ldflags "-s -w" -o /minio

RUN wget -O /mc https://dl.min.io/client/mc/release/linux-${TARGETARCH}/mc && \
    chmod +x /mc

FROM alpine:3.22
RUN apk add --no-cache ca-certificates

RUN adduser -D -H -u 1000 minio

COPY --from=builder /minio /usr/local/bin/minio
COPY --from=builder /mc /usr/local/bin/mc
COPY --chmod=0755 entrypoint.sh /usr/local/bin/entrypoint.sh

USER minio
VOLUME ["/data"]
EXPOSE 9000 9001

ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
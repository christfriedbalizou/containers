ARG VERSION

FROM --platform=$BUILDPLATFORM golang:1.25-alpine AS builder

ARG VERSION
ARG TARGETOS
ARG TARGETARCH

WORKDIR /src

RUN apk add --no-cache git

RUN git clone --depth 1 --branch "${VERSION}" https://github.com/minio/minio.git .

ENV CGO_ENABLED=0
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH go build -trimpath -ldflags "-s -w" -o /minio


FROM alpine:3.19
RUN apk add --no-cache ca-certificates

RUN adduser -D -H -u 1000 minio

COPY --from=builder /minio /usr/local/bin/minio
COPY --chmod=0755 entrypoint.sh /usr/local/bin/entrypoint.sh

USER minio
VOLUME ["/data"]
EXPOSE 9000 9001

ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
CMD ["minio","server","/data","--console-address=:9001"]
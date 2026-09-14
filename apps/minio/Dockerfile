# syntax=docker/dockerfile:1@sha256:ecfaec9ed6d810b56388c508f4121597bfbba70d41a6dfeee4d8cad5f295fc32

FROM --platform=$BUILDPLATFORM golang:1.27-alpine@sha256:cf6fca6641884b8433441b2b0652976f975e1d0fdd26d177eaaf8596087f3125 AS builder

ARG VERSION
ARG SILO_REVISION
ARG TARGETOS
ARG TARGETARCH

ENV CGO_ENABLED=0 GOTOOLCHAIN=local
WORKDIR /src/silo
RUN --mount=type=secret,id=build_ca \
    if [ -f /run/secrets/build_ca ]; then cat /run/secrets/build_ca >> /etc/ssl/certs/ca-certificates.crt; fi \
    && apk add --no-cache git

# Fetch immutable source revisions, never a moving branch at build time.
RUN test -n "$VERSION" && test "${#SILO_REVISION}" -eq 40 \
    && git init . \
    && git remote add origin https://github.com/pgsty/silo.git \
    && git fetch --depth=1 origin "$SILO_REVISION" \
    && git checkout --detach FETCH_HEAD \
    && test "$(git rev-parse HEAD)" = "$SILO_REVISION"

RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download && go mod verify
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    GOOS="$TARGETOS" GOARCH="$TARGETARCH" go build -mod=readonly -trimpath \
    -ldflags="$(MINIO_RELEASE=RELEASE go run buildscripts/gen-ldflags.go)" -o /out/minio . \
    && mkdir -p /out/licenses /out/build \
    && cp LICENSE /out/licenses/silo.LICENSE \
    && cp NOTICE /out/licenses/silo.NOTICE \
    && cp CREDITS /out/licenses/silo.CREDITS \
    && cp go.mod go.sum /out/build/ \
    && printf 'image=%s\nsilo=%s\n' "$VERSION" "$SILO_REVISION" > /out/build/revisions

FROM alpine:3.22@sha256:14358309a308569c32bdc37e2e0e9694be33a9d99e68afb0f5ff33cc1f695dce
RUN --mount=type=secret,id=build_ca \
    cat /etc/ssl/certs/ca-certificates.crt > /tmp/build-ca.pem \
    && if [ -f /run/secrets/build_ca ]; then cat /run/secrets/build_ca >> /tmp/build-ca.pem; fi \
    && SSL_CERT_FILE=/tmp/build-ca.pem apk add --no-cache ca-certificates tzdata \
    && rm /tmp/build-ca.pem \
    && addgroup -g 1000 minio \
    && adduser -D -H -u 1000 -G minio minio \
    && mkdir -p /data /tmp/minio \
    && chown 1000:1000 /data /tmp/minio

ENV HOME=/tmp/minio MINIO_UPDATE=off
COPY --from=builder /out/minio /usr/local/bin/minio
COPY --from=builder /out/licenses/ /usr/share/licenses/minio/
COPY --from=builder /out/build/ /usr/share/minio/build/
COPY --chmod=0755 entrypoint.sh /usr/local/bin/entrypoint.sh

USER 1000:1000
WORKDIR /data
VOLUME ["/data"]
EXPOSE 9000 9001
HEALTHCHECK --interval=30s --timeout=10s --start-period=30s --retries=3 CMD ["/usr/local/bin/minio", "healthcheck", "ready"]
ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
CMD ["server", "/data", "--console-address", ":9001"]

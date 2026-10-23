# syntax=docker/dockerfile:1@sha256:4edf897a3ffa55b89f906fc8cc78afdb3f1834cc9c7083565e611a8a7d5fe99e

FROM node:24.22.1-bookworm-slim@sha256:55fb5f5a527ea2c5c4a95f963e48ef90aa38c7b346054fdc8ee9957a2b55d205 AS frontend
WORKDIR /src
COPY --from=translator-source package.json package-lock.json ./
RUN --mount=type=cache,target=/root/.npm npm ci
COPY --from=translator-source . .
RUN npm run build

FROM ghcr.io/astral-sh/uv:0.13.5@sha256:32a7bb39b496babbd997b0a39a4d66c126bdaf5661b302bf6f06abda8deca6ff AS uv

FROM python:3.12.15-slim-bookworm@sha256:6a0e06f5ba0b373d051aabd3fffb9b873f27f9a2a3a14bcbe37dee6f0e3bbc70 AS python-base

FROM python-base AS backend
COPY --from=uv /uv /usr/local/bin/uv
ENV UV_PYTHON_DOWNLOADS=never UV_LINK_MODE=copy
WORKDIR /opt/translator
COPY --from=translator-source . .
RUN --mount=type=cache,target=/root/.cache/uv \
    uv sync --locked --no-dev --no-editable

FROM backend AS encrypted
RUN apt-get update && apt-get install -y --no-install-recommends gnupg \
    && rm -rf /var/lib/apt/lists/*
# Archive before encrypting: failures cannot be hidden by a shell pipeline.
RUN --mount=type=secret,id=translator-public-key,required=true \
    --mount=type=tmpfs,target=/root/.gnupg \
    --mount=type=tmpfs,target=/payload \
    tar -czf /payload/app.tar.gz .venv alembic.ini migrations \
    && chmod 0700 /root/.gnupg && gpg --batch --no-options --homedir /root/.gnupg \
        --recipient-file /run/secrets/translator-public-key \
        --compress-algo none --output /payload.tar.gz.gpg --encrypt /payload/app.tar.gz

FROM python-base AS runtime
RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates gnupg libgomp1 libseccomp2 libstdc++6 \
    && rm -rf /var/lib/apt/lists/* \
    && groupadd --gid 1000 translator \
    && useradd --uid 1000 --gid 1000 --no-create-home --home-dir /tmp/translator translator \
    && install -d -o 1000 -g 1000 /data /tmp/translator

ENV PATH="/opt/translator/.venv/bin:${PATH}" \
    PYTHONDONTWRITEBYTECODE=1 \
    PYTHONUNBUFFERED=1 \
    HOME=/tmp/translator
WORKDIR /opt/translator
COPY --from=encrypted /payload.tar.gz.gpg /usr/share/translator/payload.tar.gz.gpg
COPY --from=frontend /src/dist /usr/share/translator/dist
COPY --from=translator-source LICENSE /usr/share/translator/LICENSE
COPY decrypt.py /usr/local/lib/translator-decrypt.py
COPY --from=translator-source .source-revision /usr/share/translator/source-revision
COPY --chmod=0755 entrypoint.sh /usr/local/bin/entrypoint.sh

ARG VERSION
RUN test -n "$VERSION" && printf '%s\n' "$VERSION" > /usr/share/translator/source-tag
USER 1000:1000
EXPOSE 8000
ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
CMD ["api"]

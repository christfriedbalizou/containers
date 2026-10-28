# syntax=docker/dockerfile:1@sha256:4edf897a3ffa55b89f906fc8cc78afdb3f1834cc9c7083565e611a8a7d5fe99e

FROM node:24.22.1-bookworm-slim@sha256:be8204d4bd327363357f532187b752ff62759c31bedf98e9e18fc08c9cbd605c AS frontend
WORKDIR /src
COPY --from=translator-source package.json package-lock.json ./
RUN --mount=type=cache,target=/root/.npm npm ci
COPY --from=translator-source . .
RUN npm run build

FROM ghcr.io/astral-sh/uv:0.13.6@sha256:af22a52f5eafef3d78e551599d405c02df61919d18028f0cea8eda93bc9986d0 AS uv

FROM python:3.12.16-slim-bookworm@sha256:6dd8b6dab6b3adc7812bab69eca5dbdc1a5cc690ad28306d234ffb4ef8ac9753 AS python-base

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

# syntax=docker/dockerfile:1@sha256:4edf897a3ffa55b89f906fc8cc78afdb3f1834cc9c7083565e611a8a7d5fe99e
FROM ghcr.io/astral-sh/uv:0.13.6@sha256:af22a52f5eafef3d78e551599d405c02df61919d18028f0cea8eda93bc9986d0 AS uv
FROM python:3.14.5-alpine3.22@sha256:6b91e66ab2a880ce9ca5a1b91c70f45963ff71ff68268df056336e1a657d5efd AS python-base

FROM python-base AS build
COPY --from=uv /uv /usr/local/bin/uv
ENV UV_PYTHON_DOWNLOADS=never UV_LINK_MODE=copy
WORKDIR /opt/tv-guides
COPY --from=tv-guides-source . .
RUN --mount=type=cache,target=/root/.cache/uv uv sync --python /usr/local/bin/python --locked --no-dev --no-editable
ARG VERSION
RUN VERSION="$VERSION" .venv/bin/python -c 'import importlib.metadata, os; assert importlib.metadata.version("tv-guides") == os.environ["VERSION"].removeprefix("v")'

FROM python-base AS runtime
RUN addgroup -g 1000 tvguides && adduser -D -u 1000 -G tvguides tvguides \
    && mkdir -p /data && chown 1000:1000 /data && chmod 0700 /data
ENV PATH="/opt/tv-guides/.venv/bin:${PATH}" \
    PYTHONDONTWRITEBYTECODE=1 PYTHONUNBUFFERED=1 \
    TV_GUIDES_WORK_DIR=/data
WORKDIR /opt/tv-guides
COPY --from=build /opt/tv-guides/.venv .venv
COPY --from=tv-guides-source LICENSE /usr/share/tv-guides/LICENSE
COPY --from=tv-guides-source .source-revision /usr/share/tv-guides/source-revision
COPY --chmod=0755 entrypoint.sh /usr/local/bin/entrypoint.sh
ARG VERSION
RUN test -n "$VERSION" && printf '%s\n' "$VERSION" > /usr/share/tv-guides/source-tag
USER 1000:1000
ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
CMD ["run"]

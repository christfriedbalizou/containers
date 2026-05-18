ARG CODE_SERVER_VERSION=4.118.0

FROM ghcr.io/coder/code-server:${CODE_SERVER_VERSION}

USER root

RUN mkdir -p /config/data/User /config/extensions /opt/code-server-defaults /workspace \
    && chown -R nobody:nogroup /config /opt/code-server-defaults /workspace

COPY --chmod=0755 entrypoint.sh /usr/local/bin/entrypoint.sh

ENV HOME=/config \
    XDG_DATA_HOME=/config/data \
    XDG_CONFIG_HOME=/config \
    SERVICE_URL=https://open-vsx.org/vscode/gallery \
    ITEM_URL=https://open-vsx.org/vscode/item \
    CODE_SERVER_BIND_ADDR=0.0.0.0:8080 \
    CODE_SERVER_AUTH=password \
    CODE_SERVER_DEFAULTS_DIR=/opt/code-server-defaults \
    CODE_SERVER_USER_DATA_DIR=/config/data \
    CODE_SERVER_EXTENSIONS_DIR=/config/extensions

USER 65534:65534
WORKDIR /workspace
VOLUME ["/config", "/workspace"]
EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]

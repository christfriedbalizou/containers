ARG CODE_SERVER_VERSION

FROM ghcr.io/coder/code-server:${CODE_SERVER_VERSION}

ARG CODEX_VERSION
ARG MISE_VERSION

USER root

RUN apt-get update \
    && apt-get install -y --no-install-recommends \
      bash \
      bubblewrap \
      ca-certificates \
      curl \
      git \
      nodejs \
      npm \
      openssh-client \
      passwd \
      python3 \
      python3-pip \
      python3-venv \
      unzip \
      xz-utils \
    && rm -rf /var/lib/apt/lists/* \
    && curl -fsSL https://mise.run -o /tmp/mise-install.sh \
    && MISE_VERSION=${MISE_VERSION} MISE_INSTALL_PATH=/usr/local/bin/mise sh /tmp/mise-install.sh \
    && rm -f /tmp/mise-install.sh \
    && npm install -g @openai/codex@${CODEX_VERSION} \
    && usermod -o -u 1024 -g 101 coder \
    && mkdir -p /config/data/User /config/extensions /home/coder/src /opt/code-server-defaults \
    && chown -R 1024:101 /config /home/coder /opt/code-server-defaults

COPY --chmod=0755 entrypoint.sh /usr/local/bin/entrypoint.sh

ENV HOME=/home/coder \
    SHELL=/bin/bash \
    XDG_CACHE_HOME=/home/coder/.cache \
    XDG_CONFIG_HOME=/home/coder/.config \
    XDG_DATA_HOME=/home/coder/.local/share \
    CODEX_HOME=/home/coder/.codex \
    CODE_SERVER_WORKSPACE_DIR=/home/coder/src \
    MISE_TRUSTED_CONFIG_PATHS=/home/coder/src:/home/coder \
    PATH=/home/coder/.local/share/mise/shims:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin \
    SERVICE_URL=https://open-vsx.org/vscode/gallery \
    ITEM_URL=https://open-vsx.org/vscode/item \
    CODE_SERVER_BIND_ADDR=0.0.0.0:8080 \
    CODE_SERVER_AUTH=password \
    CODE_SERVER_DEFAULTS_DIR=/opt/code-server-defaults \
    CODE_SERVER_USER_DATA_DIR=/config/data \
    CODE_SERVER_EXTENSIONS_DIR=/config/extensions

USER 1024:101
WORKDIR /home/coder/src
VOLUME ["/config", "/home/coder"]
EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]

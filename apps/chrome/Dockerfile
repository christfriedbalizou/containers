ARG VERSION

FROM lscr.io/linuxserver/chrome:${VERSION}

RUN apt-get update \
    && apt-get install -y --no-install-recommends \
      fontconfig \
      fonts-cascadia-code \
      fonts-powerline \
    && fc-cache -f \
    && rm -rf /var/lib/apt/lists/*

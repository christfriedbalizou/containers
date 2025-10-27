ARG VERSION

FROM louislam/uptime-kuma:${VERSION}

WORKDIR /app

USER root

COPY --chmod=0755 entrypoint.sh /usr/local/bin/entrypoint.sh

ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
CMD ["node", "server/server.js"]
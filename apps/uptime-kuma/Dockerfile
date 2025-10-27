ARG VERSION

FROM louislam/uptime-kuma:${VERSION}

WORKDIR /app

COPY --chmod=0755 entrypoint.sh /usr/local/bin/entrypoint.sh

ENTRYPOINT ["/usr/bin/dumb-init", "--", "entrypoint.sh"]
CMD ["node", "server/server.js"]
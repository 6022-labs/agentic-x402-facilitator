FROM grafana/loki:3.7.6

COPY loki-config.yaml /etc/loki/loki.yaml
USER root

ENTRYPOINT ["/usr/bin/loki"]
CMD ["-config.file=/etc/loki/loki.yaml"]

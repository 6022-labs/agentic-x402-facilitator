FROM grafana/alloy:v1.21.2

COPY alloy-config.alloy.template /etc/alloy/config.alloy.template
COPY --chmod=755 docker-entrypoint.sh /docker-entrypoint.sh

ENTRYPOINT ["/docker-entrypoint.sh"]

FROM grafana/grafana:13.2.0

COPY provisioning /etc/grafana/provisioning
COPY dashboards /var/lib/grafana/dashboards
COPY --chmod=755 docker-entrypoint.sh /usr/local/bin/docker-entrypoint.sh

ENTRYPOINT ["/usr/local/bin/docker-entrypoint.sh"]

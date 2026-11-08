FROM grafana/mimir:3.2.0

COPY mimir-config.yaml /etc/mimir/mimir.yml

CMD ["--config.file=/etc/mimir/mimir.yml", "--target=all"]

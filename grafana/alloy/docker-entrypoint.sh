#!/usr/bin/env sh
set -e

TEMPLATE="/etc/alloy/config.alloy.template"
# /tmp so the container never needs write access to /etc/alloy.
CONFIG="/tmp/config.alloy"

FRAGMENTS="$(mktemp -d)"
EXTENSION="${FRAGMENTS}/extension"
AUTH_ATTR="${FRAGMENTS}/auth_attr"
TLS_BLOCK="${FRAGMENTS}/tls_block"
: > "$EXTENSION"
: > "$AUTH_ATTR"
: > "$TLS_BLOCK"

# Secrets go through sys.env(), never substituted, so none is written into the rendered config.
if [ -n "$OTLP_AUTH_HTPASSWD" ]; then
  cat > "$EXTENSION" <<'EOF'
otelcol.auth.basic "otlp_auth" {
  htpasswd {
    inline = sys.env("OTLP_AUTH_HTPASSWD")
  }
}
EOF
  echo '    auth = otelcol.auth.basic.otlp_auth.handler' > "$AUTH_ATTR"
  echo "OTLP auth ENABLED (basic — one credential per node)"
elif [ -n "$OTLP_AUTH_TOKEN" ]; then
  cat > "$EXTENSION" <<'EOF'
otelcol.auth.bearer "otlp_auth" {
  token = sys.env("OTLP_AUTH_TOKEN")
}
EOF
  echo '    auth = otelcol.auth.bearer.otlp_auth.handler' > "$AUTH_ATTR"
  echo "OTLP auth ENABLED (bearer — shared token)"
else
  echo "WARNING: no OTLP auth configured — the OTLP ingest (:4317, :4318) is UNAUTHENTICATED. Do not expose publicly."
fi

if [ -n "$OTLP_TLS_CERT_FILE" ] && [ -n "$OTLP_TLS_KEY_FILE" ]; then
  {
    echo '    tls {'
    echo "      cert_file = \"${OTLP_TLS_CERT_FILE}\""
    echo "      key_file  = \"${OTLP_TLS_KEY_FILE}\""
    if [ -n "$OTLP_TLS_CLIENT_CA_FILE" ]; then
      echo "      client_ca_file = \"${OTLP_TLS_CLIENT_CA_FILE}\""
    fi
    echo '    }'
  } > "$TLS_BLOCK"
  echo "OTLP TLS ENABLED (client_ca_file: ${OTLP_TLS_CLIENT_CA_FILE:-none})"
else
  echo "WARNING: no OTLP TLS configured — traffic, and any credential it carries, travels in cleartext."
fi

# `r` queues the fragment after the marker, `d` drops the marker itself.
sed \
  -e "/__OTLP_AUTH_EXTENSION__/r ${EXTENSION}" -e "/__OTLP_AUTH_EXTENSION__/d" \
  -e "/__OTLP_AUTH_ATTR__/r ${AUTH_ATTR}" -e "/__OTLP_AUTH_ATTR__/d" \
  -e "/__OTLP_TLS_BLOCK__/r ${TLS_BLOCK}" -e "/__OTLP_TLS_BLOCK__/d" \
  "$TEMPLATE" > "$CONFIG"

rm -rf "$FRAGMENTS"

exec /bin/alloy run --stability.level=experimental "$CONFIG"

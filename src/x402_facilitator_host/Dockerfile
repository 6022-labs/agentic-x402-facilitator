FROM golang:1.27-alpine AS builder

WORKDIR /usr/local/go/src/github.com/6022-labs/agentic-x402-facilitator

COPY ./go.mod ./
COPY ./go.sum ./

RUN go mod download

COPY ./ ./

WORKDIR /usr/local/go/src/github.com/6022-labs/agentic-x402-facilitator/src/x402_facilitator_host

# Build the executable to `/app`. Mark the build as statically linked.
RUN CGO_ENABLED=0 go build \
    -installsuffix 'static' \
    -o /app .

FROM scratch AS final

# Import the compiled executable from the first stage.
COPY --from=builder /app /app
COPY --from=builder /etc/ssl/certs /etc/ssl/certs

# Client IP for the rate limiter: trust loopback, private and shared-range edges; keep 127.0.0.1 when overriding.
ENV PROXY__HEADER=X-Forwarded-For \
    PROXY__TRUSTED_PROXIES=127.0.0.1,10.0.0.0/8,172.16.0.0/12,192.168.0.0/16,100.64.0.0/10,fc00::/7

EXPOSE 5000

# Run the compiled binary at container start.
ENTRYPOINT ["/app"]

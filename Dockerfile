FROM golang:1 AS builder

COPY . /app
WORKDIR /app
RUN make build

FROM gcr.io/distroless/static-debian12:latest
COPY --from=builder /app/bin/chia-healthcheck /usr/bin/chia-healthcheck
CMD ["/usr/bin/chia-healthcheck", "serve"]

# ------------------------------------------------------------------------
#  Dependencies
# ------------------------------------------------------------------------
FROM golang:1.22.2-alpine as dependencies

RUN apk add --no-cache upx

# only copy the go.mod and go.sum for better docker caching in first stage
COPY go.mod /app/go.mod
#COPY go.sum /app/go.sum

# download dependencies
WORKDIR /app
RUN go mod download

# ------------------------------------------------------------------------
#  Builder
# ------------------------------------------------------------------------
FROM dependencies AS builder
COPY . /app
WORKDIR /app/cmd

## Build static binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
        -ldflags='-w -s -extldflags "-static"' -a \
        -o main .
RUN upx --brute main

## Add Non Root User with uid > 1000 to prevent collisions
RUN adduser \
    --disabled-password \
    --gecos "" \
    --home "/nonexistent" \
    --shell "/sbin/nologin" \
    --no-create-home \
    --uid "10000" \
    "appuser"

## Limit permissions on settings and binary
RUN chown -R appuser:appuser /app
RUN chmod 0440 /app/settings.yaml
RUN chmod 0550 /app/cmd/main

# ------------------------------------------------------------------------
#  Runner
# ------------------------------------------------------------------------
FROM scratch

# Import the user and group files from the builder.
#COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group

# Copy Runtime files
COPY --from=builder /app/cmd/main main
COPY --from=builder /app/settings.yaml settings.yaml

USER appuser:appuser
ENTRYPOINT ["/main"]

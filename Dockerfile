FROM golang:1.21 AS builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
  -trimpath \
  -ldflags "-s -w -X main.version=docker -X main.commit=container -X main.buildDate=unknown" \
  -o /out/mysqlrouter_exporter .

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /out/mysqlrouter_exporter /usr/local/bin/mysqlrouter_exporter
EXPOSE 9165
ENTRYPOINT ["/usr/local/bin/mysqlrouter_exporter"]


FROM golang:1.14-alpine as builder

RUN mkdir -p /go/src/github.com/jfardello/tlsrproxy &&\
    apk update && apk add --no-cache git ca-certificates tzdata &&\
    update-ca-certificates && adduser -D -g '' appuser

WORKDIR /go/src/github.com/jfardello/tlsrproxy
COPY go.mod .
COPY go.sum .
RUN go mod download

#get deps before code to ensure external modules ge cached by podman.
COPY . .
RUN GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /usr/local/bin/tlsrproxy

FROM alpine

COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
ENV DSC_HTTP_ADDR :8888
ENV DSC_HTTP_DRAIN_INTERVAL 1s
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /usr/local/bin/tlsrproxy /usr/local/bin/tlsrproxy
USER appuser

EXPOSE 8888

ENTRYPOINT ["/usr/local/bin/tlsrproxy"]

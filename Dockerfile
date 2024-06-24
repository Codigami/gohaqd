FROM golang:1.15.8-alpine3.13 AS buildimage

RUN apk update && apk add ca-certificates bash
COPY . /go/src/github.com/Codigami/gohaqd/
WORKDIR /go/src/github.com/Codigami/gohaqd
RUN CGO_ENABLED=0 GOOS=linux /bin/bash -c "bash check.sh && go build -a -v -ldflags '-w'"

FROM scratch

COPY --from=buildimage /go/src/github.com/Codigami/gohaqd/gohaqd /
COPY --from=buildimage /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

# Add nsswitch config to resolve DNS using /etc/hosts before calling the DNS server.
COPY --from=buildimage /etc/nsswitch.conf /etc/nsswitch.conf

ENTRYPOINT ["/gohaqd"]


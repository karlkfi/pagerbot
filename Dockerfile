# build image
FROM golang:1.11.2-alpine3.8 as builder
RUN apk update && apk add git && apk add ca-certificates

# create non-root user
RUN adduser -D -g '' appuser
COPY . $GOPATH/src/karlkfi/pagerbot/
WORKDIR $GOPATH/src/karlkfi/pagerbot/

# download dependencies
ENV GO111MODULE=on
RUN go get -d -v

# compile static binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags="-w -s" -o /go/bin/pagerbot

# distributable image
FROM alpine:3.8

# copy dependencies
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd

# copy static executable
COPY --from=builder /go/bin/pagerbot /go/bin/pagerbot

# copy default config
COPY config.yml /

# use non-root user
USER appuser

ENTRYPOINT ["/go/bin/pagerbot"]

FROM golang:1.16.4-alpine3.13 as builder
RUN apk update && apk add --no-cache alpine-sdk ca-certificates
RUN update-ca-certificates
ENV GO111MODULE=on \
    GOOS=linux \
    GOARCH=amd64 \ 
    CGO_ENABLED=1
WORKDIR /build
COPY . .
RUN go mod download 
RUN go build -a -ldflags '-w -extldflags "-static"' -o tweet-beat .

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ 
COPY --from=builder /build/tweet-beat /
COPY --from=builder /build/config.json / 
COPY --from=builder /build/subscribers.sql /
CMD ["/tweet-beat"]

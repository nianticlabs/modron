ARG GOVERSION=1.23

FROM alpine:latest AS ca-certificates_builder
RUN apk add --no-cache ca-certificates

FROM golang:${GOVERSION} AS modron_builder
ENV GOPATH=/go
WORKDIR /app/src
COPY src/go.* /app/src/
COPY src/proto/generated /app/src/proto/generated/
RUN go mod download
COPY ./src/ /app/src/
RUN CGO_ENABLED=0 go build -v -o modron

FROM scratch
WORKDIR /app
COPY --from=ca-certificates_builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=modron_builder /app/src/modron /app/modron
USER 101:101
ENTRYPOINT ["/app/modron"]

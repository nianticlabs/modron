ARG GOVERSION=1.18

FROM alpine:latest as ca-certificates_builder
RUN apk add --no-cache ca-certificates
RUN update-ca-certificates

FROM golang:${GOVERSION} as modron_builder
ENV GOPATH /go
WORKDIR /app
COPY go.* ./
COPY proto/go.* ./proto/
RUN go mod download
COPY . ./
RUN CGO_ENABLED=0 go build -v -o modron

FROM scratch
WORKDIR /app
COPY --from=ca-certificates_builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=modron_builder /app/modron /app/modron
USER 101:101
ENTRYPOINT ["/app/modron", "--logtostderr"]

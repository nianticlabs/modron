ARG GOVERSION=1.18

FROM golang:${GOVERSION} as nagatha_builder
ENV GOPATH /go
WORKDIR /app
COPY go.* ./
COPY proto/go.* ./proto/
RUN go mod download
COPY . ./
RUN CGO_ENABLED=0 go build -v -o nagatha

FROM alpine:latest as ca-certificates_builder
RUN apk update && apk upgrade && apk add --no-cache ca-certificates

FROM scratch
WORKDIR /app
COPY --from=nagatha_builder /app/nagatha /app/nagatha
COPY --from=ca-certificates_builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY emailTemplate.html /app/
USER 101:101
ENTRYPOINT ["/app/nagatha", "--logtostderr"]

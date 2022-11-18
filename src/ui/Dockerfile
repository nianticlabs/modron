ARG GOVERSION=1.19

FROM node:16.16-alpine AS ui_builder
WORKDIR /app
COPY ./client/ .
RUN npm install
RUN npm run build

FROM golang:${GOVERSION} as server_builder
ENV GOPATH /go
WORKDIR /app
COPY go.* ./
RUN go mod download
COPY . ./
RUN CGO_ENABLED=0 go build -v -o modron-ui-server

FROM alpine:latest as ca-certificates_builder
RUN apk add --no-cache ca-certificates
RUN update-ca-certificates

FROM scratch
WORKDIR /app
COPY --from=ca-certificates_builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=ui_builder /app/dist/ .
COPY --from=server_builder /app/modron-ui-server .
USER 101:101
EXPOSE 8080
ENTRYPOINT ["/app/modron-ui-server", "--logtostderr"]

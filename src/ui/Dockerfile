ARG GOVERSION=1.23
ARG NODE_VERSION=20

FROM node:${NODE_VERSION}-alpine AS ui_builder
WORKDIR /app
COPY ./src/ui/client/ .
RUN npm install
RUN npm run build

FROM golang:${GOVERSION} AS server_builder
ENV GOPATH=/go
WORKDIR /app
COPY ./src/ui/go.* ./
RUN go mod download
COPY ./src/ui/ ./
RUN CGO_ENABLED=0 go build -v -o modron-ui-server

FROM alpine:latest AS ca-certificates_builder
RUN apk add --no-cache ca-certificates

# FROM scratch
WORKDIR /app
# COPY --from=ca-certificates_builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=ui_builder /app/dist/ .
COPY --from=server_builder /app/modron-ui-server .
USER 101:101
EXPOSE 8080
ENTRYPOINT ["/app/modron-ui-server", "-logtostderr"]

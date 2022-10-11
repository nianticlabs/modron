#!/usr/bin/env sh
# Generate Modron protos
PROTOC_GEN_GO_PATH="/go/bin/protoc-gen-go"
protoc -I=./proto \
       --experimental_allow_proto3_optional \
       --plugin="protoc-gen-go=${PROTOC_GEN_GO_PATH}" \
       --plugin="protoc-gen-go-grpc=${PROTOC_GEN_GO_PATH}-grpc" \
       --go_out="./proto" \
       --go_opt=paths=source_relative \
       --go-grpc_out="./proto" \
       --go-grpc_opt=paths=source_relative \
       modron.proto notification.proto

# Generate UI protos
PROTOC_GEN_TS_PATH="/usr/local/lib/node_modules/ts-protoc-gen/bin/protoc-gen-ts"
PROTOC_OUT_DIR="./ui/client/src/proto/"
mkdir -p ${PROTOC_OUT_DIR}
protoc -I=./proto \
       --experimental_allow_proto3_optional \
       --plugin="protoc-gen-ts=${PROTOC_GEN_TS_PATH}" \
       --js_out="import_style=commonjs,binary:${PROTOC_OUT_DIR}" \
       --ts_out="service=grpc-web,mode=grpc-js:${PROTOC_OUT_DIR}" \
       modron.proto notification.proto
PROTOC_OUT_DIR="./ui/mock-grpc-server/proto/"
mkdir -p ${PROTOC_OUT_DIR}
protoc -I=./proto \
       --experimental_allow_proto3_optional \
       --plugin="protoc-gen-ts=${PROTOC_GEN_TS_PATH}" \
       --js_out="import_style=commonjs,binary:${PROTOC_OUT_DIR}" \
       --ts_out="mode=grpc-js:${PROTOC_OUT_DIR}" \
       modron.proto notification.proto
module github.com/nianticlabs/modron/src/e2e_test

go 1.21

replace github.com/nianticlabs/modron/src/pb => ../proto/

require (
	github.com/google/go-cmp v0.5.9
	google.golang.org/grpc v1.57.0
	google.golang.org/protobuf v1.31.0
	github.com/nianticlabs/modron/src/pb v0.0.0-00010101000000-000000000000
)

require (
	github.com/golang/protobuf v1.5.3 // indirect
	golang.org/x/net v0.14.0 // indirect
	golang.org/x/sys v0.12.0 // indirect
	golang.org/x/text v0.13.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20230822172742-b8732ec3820d // indirect
)

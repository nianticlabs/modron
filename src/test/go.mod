module github.com/nianticlabs/modron/src/e2e_test

go 1.19

replace github.com/nianticlabs/modron/src/pb => ../proto/

require (
	github.com/google/go-cmp v0.5.8
	google.golang.org/grpc v1.50.1
	google.golang.org/protobuf v1.28.1
	github.com/nianticlabs/modron/src/pb v0.0.0-00010101000000-000000000000
)

require (
	github.com/golang/protobuf v1.5.2 // indirect
	golang.org/x/net v0.2.0 // indirect
	golang.org/x/sys v0.2.0 // indirect
	golang.org/x/text v0.4.0 // indirect
	google.golang.org/genproto v0.0.0-20221107162902-2d387536bcdd // indirect
)

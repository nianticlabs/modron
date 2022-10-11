module github.com/nianticlabs/modron/src/e2e_test

go 1.18

replace github.com/nianticlabs/modron/src/pb => ../proto/

require (
	github.com/google/go-cmp v0.5.8
	google.golang.org/grpc v1.48.0
	google.golang.org/protobuf v1.28.0
	github.com/nianticlabs/modron/src/pb v0.0.0-00010101000000-000000000000
)

require (
	github.com/golang/protobuf v1.5.2 // indirect
	golang.org/x/net v0.0.0-20220706163947-c90051bbdb60 // indirect
	golang.org/x/sys v0.0.0-20220624220833-87e55d714810 // indirect
	golang.org/x/text v0.3.7 // indirect
	google.golang.org/genproto v0.0.0-20220714211235-042d03aeabc9 // indirect
)

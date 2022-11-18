module github.com/nianticlabs/modron/src

go 1.19

replace github.com/nianticlabs/modron/src/pb => ./proto/

require github.com/improbable-eng/grpc-web v0.15.0

require (
	cloud.google.com/go/bigquery v1.43.0
	github.com/golang/glog v1.0.0
	github.com/google/go-cmp v0.5.9
	github.com/google/uuid v1.3.0
	golang.org/x/exp v0.0.0-20221108223516-5d533826c662
	golang.org/x/net v0.2.0
	google.golang.org/api v0.103.0
	google.golang.org/grpc v1.50.1
	google.golang.org/protobuf v1.28.1
	github.com/nianticlabs/modron/src/pb v0.0.0-00010101000000-000000000000
)

require (
	cloud.google.com/go v0.105.0 // indirect
	cloud.google.com/go/compute v1.12.1 // indirect
	cloud.google.com/go/compute/metadata v0.2.1 // indirect
	cloud.google.com/go/iam v0.7.0 // indirect
	github.com/cenkalti/backoff/v4 v4.1.3 // indirect
	github.com/desertbit/timer v0.0.0-20180107155436-c41aec40b27f // indirect
	github.com/gin-gonic/gin v1.8.1 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.2.0 // indirect
	github.com/googleapis/gax-go/v2 v2.7.0 // indirect
	github.com/klauspost/compress v1.15.12 // indirect
	github.com/rs/cors v1.8.2 // indirect
	go.opencensus.io v0.24.0 // indirect
	golang.org/x/crypto v0.0.0-20210921155107-089bfa567519 // indirect
	golang.org/x/oauth2 v0.1.0 // indirect
	golang.org/x/sys v0.2.0 // indirect
	golang.org/x/text v0.4.0 // indirect
	golang.org/x/xerrors v0.0.0-20220907171357-04be3eba64a2 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto v0.0.0-20221107162902-2d387536bcdd // indirect
	nhooyr.io/websocket v1.8.7 // indirect
)

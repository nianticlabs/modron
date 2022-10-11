module github.com/nianticlabs/modron/src

go 1.18

replace github.com/nianticlabs/modron/src/pb => ./proto/

require github.com/improbable-eng/grpc-web v0.15.0

require (
	cloud.google.com/go/bigquery v1.36.0
	github.com/golang/glog v1.0.0
	github.com/google/go-cmp v0.5.8
	github.com/google/uuid v1.3.0
	golang.org/x/exp v0.0.0-20220713135740-79cabaa25d75
	golang.org/x/net v0.0.0-20220706163947-c90051bbdb60
	google.golang.org/api v0.91.0
	google.golang.org/grpc v1.48.0
	google.golang.org/protobuf v1.28.0
	github.com/nianticlabs/modron/src/pb v0.0.0-00010101000000-000000000000
)

require (
	cloud.google.com/go v0.102.1 // indirect
	cloud.google.com/go/compute v1.7.0 // indirect
	cloud.google.com/go/iam v0.3.0 // indirect
	github.com/cenkalti/backoff/v4 v4.1.1 // indirect
	github.com/desertbit/timer v0.0.0-20180107155436-c41aec40b27f // indirect
	github.com/gin-gonic/gin v1.8.1 // indirect
	github.com/golang/groupcache v0.0.0-20200121045136-8c9f03a8e57e // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.1.0 // indirect
	github.com/googleapis/gax-go/v2 v2.4.0 // indirect
	github.com/klauspost/compress v1.11.7 // indirect
	github.com/rs/cors v1.7.0 // indirect
	github.com/stretchr/testify v1.7.1 // indirect
	go.opencensus.io v0.23.0 // indirect
	golang.org/x/crypto v0.0.0-20210921155107-089bfa567519 // indirect
	golang.org/x/oauth2 v0.0.0-20220622183110-fd043fe589d2 // indirect
	golang.org/x/sys v0.0.0-20220624220833-87e55d714810 // indirect
	golang.org/x/text v0.3.7 // indirect
	golang.org/x/xerrors v0.0.0-20220609144429-65e65417b02f // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto v0.0.0-20220714211235-042d03aeabc9 // indirect
	nhooyr.io/websocket v1.8.6 // indirect
)

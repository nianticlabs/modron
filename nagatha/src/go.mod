module github.com/nianticlabs/modron/nagatha/src

go 1.18

replace github.com/nianticlabs/modron/nagatha/src/pb => ./proto/

require (
	cloud.google.com/go/bigquery v1.8.0
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/gomarkdown/markdown v0.0.0-20220830015526-01a3c37d6f50
	github.com/google/go-cmp v0.5.8
	github.com/sendgrid/sendgrid-go v3.11.1+incompatible
	google.golang.org/api v0.89.0
	google.golang.org/grpc v1.47.0
	google.golang.org/protobuf v1.28.0
	github.com/nianticlabs/modron/nagatha/src/pb v0.0.0-00010101000000-000000000000
)

require (
	cloud.google.com/go v0.102.0 // indirect
	cloud.google.com/go/compute v1.7.0 // indirect
	github.com/golang/groupcache v0.0.0-20200121045136-8c9f03a8e57e // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/uuid v1.3.0
	github.com/googleapis/enterprise-certificate-proxy v0.1.0 // indirect
	github.com/googleapis/gax-go/v2 v2.4.0 // indirect
	github.com/sendgrid/rest v2.6.9+incompatible // indirect
	go.opencensus.io v0.23.0 // indirect
	golang.org/x/net v0.0.0-20220706163947-c90051bbdb60 // indirect
	golang.org/x/oauth2 v0.0.0-20220622183110-fd043fe589d2 // indirect
	golang.org/x/sys v0.0.0-20220624220833-87e55d714810 // indirect
	golang.org/x/text v0.3.7 // indirect
	golang.org/x/xerrors v0.0.0-20220609144429-65e65417b02f // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto v0.0.0-20220624142145-8cd45d7dbd1f // indirect
)

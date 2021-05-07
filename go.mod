module github.com/greatfilter/martian

go 1.11

replace github.com/google/martian/v3 => ./

require (
	github.com/golang/protobuf v1.5.2
	github.com/golang/snappy v0.0.3 // indirect
	github.com/google/martian/v3 v3.0.0-00010101000000-000000000000
	golang.org/x/net v0.0.0-20210505214959-0714010a04ed
	google.golang.org/grpc v1.37.0
)

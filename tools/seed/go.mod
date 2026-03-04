module tiny-ils/tools/seed

go 1.25

require (
	google.golang.org/grpc v1.73.0
	tiny-ils/gen v0.0.0
)

require (
	golang.org/x/net v0.38.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	golang.org/x/text v0.26.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250324211829-b45e905df463 // indirect
	google.golang.org/protobuf v1.36.6 // indirect
)

replace tiny-ils/gen => ../../gen

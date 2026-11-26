// Package api contains the generated gRPC API between CLI and service.
package api

//go:generate sh -c "protoc --plugin=protoc-gen-go=$(go tool -n protoc-gen-go) --plugin=protoc-gen-go-grpc=$(go tool -n protoc-gen-go-grpc) --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative messaging.proto"
//go:generate go tool mockgen -source=messaging_grpc.pb.go -destination=mocks/messaging_grpc_pb_mock.go -package=mocks

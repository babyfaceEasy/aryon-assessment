run-connectors:
	@go run go-server/cmd/server/*.go
gen:
	@protoc \
		--proto_path=protobuf "protobuf/connectors.proto" \
		--go_out=go-server/connectors/genproto --go_opt=paths=source_relative \
  	--go-grpc_out=go-server/connectors/genproto --go-grpc_opt=paths=source_relative
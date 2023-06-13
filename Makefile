gen:
	# rpc-server
	protoc -I ./proto \
	--go_out rpc-server/proto/ \
	--go_opt paths=source_relative \
	--go-grpc_out rpc-server/proto/ \
	--go-grpc_opt paths=source_relative \
	 proto/*.proto

	# http-server
	 protoc -I ./proto \
	--go_out http-server/proto/ \
	--go_opt paths=source_relative \
	--go-grpc_out http-server/proto/ \
	--go-grpc_opt paths=source_relative \
	 proto/*.proto
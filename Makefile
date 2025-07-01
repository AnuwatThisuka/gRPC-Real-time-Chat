PROTO_SRC=proto/chat.proto
PROTO_DST=pb

generate:
	protoc --go_out=$(PROTO_DST) --go-grpc_out=$(PROTO_DST) --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative $(PROTO_SRC)

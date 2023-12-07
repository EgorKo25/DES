.PHONY: generate-grpc server-start build clean

PROTO_FILE = server/extension-service-gen/proto/des.proto
GENERATED_DIR = internal/service
BIN_NAME = des

generate-grpc:
	mkdir -p $(GENERATED_DIR)
	PATH="${PATH}:${HOME}/go/bin" protoc --go_out=$(GENERATED_DIR) --go_opt=paths=import \
		--go-grpc_out=$(GENERATED_DIR) --go-grpc_opt=paths=import $(PROTO_FILE)
	mv $(GENERATED_DIR)/github.com/EgorKo25/DES/$(GENERATED_DIR)/* $(GENERATED_DIR)
	rm -rf $(GENERATED_DIR)/github.com

build:
	go build -o $(BIN_NAME) cmd/des/des.go

server-start:
	./$(BIN_NAME)

clean:
	rm -f $(BIN_NAME)
	rm -rf $(GENERATED_DIR)

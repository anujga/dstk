GO_FILES := $(shell find . -type f -name '*.go' -print)
PROTO_FILES := $(shell find . -type f -name '*.proto' -print)
PROTO_PATH := api/protobuf-spec/
PROTO_OUT_DIR := pkg/api/proto/

.PHONY: test
test: $(GO_FILES) | protobuf
	go test ./...

.PHONY: racetest
racetest: $(GO_FILES) | protobuf
	go test -race ./...
.PHONY: protobuf
protobuf: $(PROTO_FILES)
	protoc $(PROTO_FILES) --proto_path=$(PROTO_PATH) --go_out=plugins=grpc:$(PROTO_OUT_DIR)

.PHONY: clean
clean:
	rm -rf $(PROTO_OUT_DIR)/*

.PHONY: fmt
fmt:
	go fmt ./...

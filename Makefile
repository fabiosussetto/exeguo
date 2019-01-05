.PHONY: install generate-protobuf build-dispatcher build-agent clean build-dispatcher-linux build-agent-linux build

DISPATCHER_BINARY_NAME=exeguo-dispatcher
AGENT_BINARY_NAME=exeguo-agent


install:
		dep ensure
		go install ./vendor/github.com/golang/protobuf/protoc-gen-go
		generate-protobuf

generate-protobuf:
		cd exeguo-dispatcher && protoc -I rpc/ rpc/job_service.proto --go_out=plugins=grpc:rpc

build-dispatcher: 
		@echo "Building Dispatcher..."
		go build -o ./release/${DISPATCHER_BINARY_NAME} github.com/fabiosussetto/exeguo/exeguo-dispatcher

build-agent:
		@echo "Building Agent..."
		go build -o ./release/${AGENT_BINARY_NAME} github.com/fabiosussetto/exeguo/exeguo-agent

build: clean build-dispatcher build-agent


clean: 
		go clean
		rm -r release/*

# Cross compilation
build-dispatcher-linux:
		CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GIN_MODE=release go build -o ./release/${DISPATCHER_BINARY_NAME}_linux_amd64 github.com/fabiosussetto/exeguo/exeguo-dispatcher

build-agent-linux: 
		CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./release/${AGENT_BINARY_NAME} github.com/fabiosussetto/exeguo/exeguo-agent
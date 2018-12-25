Regenerate protoc interfaces:

$ protoc -I rpc/ rpc/job_service.proto --go_out=plugins=grpc:rpc
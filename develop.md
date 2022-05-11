## Generate Go code for protocol buffer and grpc 

1. download protoc version 3.14.0 
2. download protoc plugin for protobuff grpc (https://grpc.io/docs/languages/go/quickstart/) 
3. execute 
```
protoc -I . --go_out=./http --go-grpc_out=./http ./api/httpforkjoin.proto
```
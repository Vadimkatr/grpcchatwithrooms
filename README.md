# grpcchatwithrooms

## Protobuf
```
protoc -I ${GOPATH}/src/grpcchatwithrooms internal/proto/service.proto  --go_out=plugins=grpc:./
```

# grpcchatwithrooms

## Protobuf
```
protoc -I ${GOPATH}/src/grpcchatwithrooms internal/proto/service.proto  --go_out=plugins=grpc:./
```

## Features

After running client, you will get main menu:

```
Menu:
1) Create room;
2) List room;
3) Connect to existing rooom;
4) Delete room;
5) Exit
```

You can:
1) create a room;
2) connect to existing room;
3) disconnect from room typing `/exit` or `/menu`;
4) delete room (only user that create room can delete them).

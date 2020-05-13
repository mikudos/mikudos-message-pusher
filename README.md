![Mikudos Plugin](https://img.shields.io/badge/MIKUDOS-Plugin-orange?style=for-the-badge&logo=appveyor)

# mikudos-message-pusher

mikudos-message-pusher is one mikudos plugin working as GRPC server. This service provide off-line message as well as on-line message.

this service provide several GRPC service method as following

```protobuf
syntax = "proto3";
package message_pusher;

service Message_pusher {
  rpc GetConfig(ConfigRequest) returns (ConfigResponse) {}
  rpc PushToChannel(PushMessage) returns (Response) {}
  rpc PushToChannelWithStatus(PushMessage) returns (Response) {}
  rpc GateStream(stream Response) returns (stream Message) {
  } // Gate communicate
}
```

This service can be called within your micro services cluster with any type of GRPC client.

Mikudos-message-pusher is compatible with mikudos-gate, which implement mikudos-socketio-app in typescript.

Every message will be send automaticaly to the mikudos-gate server, when the corresponding user is logined in on any mikudos-gate server.

This service support 3 different working mode. They are "unify", "group" and "every". This setting can be modified in config directory with yaml file.

```yaml
# Copyright (c) All contributors. All rights reserved.
# Licensed under the MIT license. See LICENSE file in the project root for full license information.
# message pusher config file
# port is this grpc serivce listened
port: 50051
####################
# PUSH MODE
# support values are: "unify", "group", "every"
# At unify mode, message pusher will send one time randomly per message, this mode passt to the unify adapted mikudos gate
# At group mode, message pusher will use the groupId value keyed with group in META_DATA, which is send with from mikudos gate or other grpc message pusher client.
# At every mode, message pusher will send message to every mikudos gate, which is connected to the message pusher service.
# default value is "unify"
####################
mode: 'group'
####################
# storageType set the storage engine for message.
# supported values are: "redis", "mysql"
# strong suggest to use redis as storage engine
####################
storageType: 'redis'
####################
# redis storage node config
####################
redisSource:
    'node1:1': 'tcp@localhost:6379'
####################
# mysql storage node config
####################
mySQLSource:
    'node1:1': 'root:MikudosMessagePusherServerPassword@(127.0.0.1:3306)/mikudos_message_pusher?parseTime=true&loc=Local&charset=utf8'
```


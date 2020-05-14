![Mikudos Plugin](https://img.shields.io/badge/MIKUDOS-Plugin-orange?style=for-the-badge&logo=appveyor)

# [![Mikudos](https://raw.githubusercontent.com/mikudos/doc/master/mikudos-logo.png)](https://mikudos.github.io/doc)

# mikudos-message-pusher

mikudos-message-pusher is one mikudos plugin working as GRPC server. This service provide off-line message as well as on-line message.

this service provide several GRPC service method as following

```protobuf
syntax = "proto3";
package message_pusher;

service Message_pusher {
  rpc GetConfig(ConfigRequest) returns (ConfigResponse) {}
  rpc StateInfo(InfoRequest) returns (InfoResponse) {}
  rpc PushToChannel(PushMessage) returns (Response) {}
  rpc PushToChannelWithStatus(PushMessage) returns (Response) {}
  rpc GateStream(stream Response) returns (stream Message) {
  } // Gate communicate
}

message ConfigRequest { repeated string Keys = 1; }

message ConfigResponse {}

message InfoRequest {

}

message InfoResponse {

}

message PushMessage {
  string msg = 2;
  string channelId = 3;
  int32 expire = 5;
}

message Message {
  string msg = 2;
  string channelId = 3;
  int64 msgId = 4;
  int32 expire = 5;
  MessageType messageType = 7;
}

message Request { string name = 1; }

message Response {
  uint32 msgId = 1;
  string channelId = 2;
  string msg = 3;
  int32 expire = 4;
  MessageType messageType = 7;
}

enum MessageType { // message type
  REQUEST = 0;
  RESPONSE = 1;
  RECEIVED = 2;
  UNRECEIVED = 3;
}

```

This service can be called within your micro services cluster with any type of GRPC client.

Mikudos-message-pusher is compatible with mikudos-gate, which implement mikudos-socketio-app in typescript.

Every message will be send automaticaly to the mikudos-gate server, when the corresponding user is logined in on any mikudos-gate server.

This service support 3 different working mode. They are "unify", "group" and "every". This setting can be modified in config directory with yaml file.

Currently only support unify mode.

Mikudos-message-pusher delivery message after receive push message call immediately, then it will choose to save this messge depends on online user resaved or not.

IMPORTENT: Mikudos-message-pusher service will receive response from mikudos-gate server after the message has delivered to it. If the corresponding user not on that mikudos-gate server or mikudos-gate server group(grouped with redisAdapter), the response of unreceived will be send back. The unreceived responce will emit one saveMessage event, which cause message saved to the Storage.

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
# At group mode, message pusher will send message to randomly gate-node in every group use the groupId value keyed with group in META_DATA, which is send with from mikudos gate.
# At every mode, message pusher will send message to every mikudos gate, which is connected to the message pusher service.
# default value is "unify"
####################
mode: 'unify'
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

package handler

import (
	"context"

	pb "github.com/mikudos/mikudos-message-pusher/proto/message-pusher"
)

// Server 事件驱动服务间流程控制方法，提供基本的数据库操作方法
type Server struct {
	pb.MessagePusherServer
}

func (s *Server) Call(ctx context.Context, req *pb.Message) (*pb.Message, error) {
	return req, nil
}

package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"strconv"

	"github.com/mikudos/mikudos-message-pusher/handler"
	pb "github.com/mikudos/mikudos-message-pusher/proto/message-pusher"

	"github.com/mikudos/mikudos-message-pusher/config"

	"google.golang.org/grpc"
)

func main() {
	if os.Getenv("SERVICE_PORT") != "" {
		p, err := strconv.Atoi(os.Getenv("SERVICE_PORT"))
		if err == nil {
			config.RuntimeViper.Set("port", p)
		}
	}
	port := config.RuntimeViper.Get("port")
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port)) //1.指定监听地址:端口号
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()                                  //新建gRPC实例
	pb.RegisterScheduleServiceServer(s, &handler.Server{}) //在gRPC服务器注册服务实现
	log.Println(fmt.Sprintf("server start at port: %d", port))
	if err := s.Serve(lis); err != nil { //Serve()阻塞等待
		log.Fatalf("failed to serve: %v", err)
	}
}

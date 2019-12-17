package handler

import (
	"context"
	"fmt"
	"io"
	"log"

	pb "github.com/mikudos/mikudos-message-pusher/proto/message-pusher"
)

type Server struct {
	Recv chan *pb.Message
	Resp chan *pb.Response
	Done chan bool
}

func (s *Server) PushToChannel(ctx context.Context, req *pb.Message) (*pb.Response, error) {
	s.Recv <- req
	msg := &pb.Response{}
	return msg, nil
}

func (s *Server) PushToChannelWithStatus(ctx context.Context, req *pb.Message) (*pb.Response, error) {
	s.Recv <- req
	for {
		resp := <-s.Resp
		if resp.Msg == "" {
			return resp, nil
		}
	}
}

func (s *Server) GateStream(stream pb.MessagePusher_GateStreamServer) (err error) {
	fmt.Printf("Received GateStream request\n")

	go func() {
		select {
		case <-s.Done:
			return
		case msg := <-s.Recv:
			stream.Send(msg)
		}
	}()

	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Error while reading client stream: %v", err)
		}
		s.Resp <- resp
	}

	return err
}

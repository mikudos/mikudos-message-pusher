package handler

import (
	"context"
	"fmt"
	"io"
	"log"

	pb "github.com/mikudos/mikudos-message-pusher/proto/message-pusher"
)

type Server struct {
	pb.MessagePusherServer
}

var MsgChan *MessageChan

func init() {
	MsgChan = &MessageChan{recv: make(chan *pb.Message), resp: make(chan *pb.Response)}
}

type MessageChan struct {
	recv chan *pb.Message
	resp chan *pb.Response
}

func (s *Server) PushToChannel(ctx context.Context, req *pb.Message) (*pb.Response, error) {
	MsgChan.recv <- req
	msg := &pb.Response{}
	return msg, nil
}

func (s *Server) PushToChannelWithStatus(ctx context.Context, req *pb.Message) (*pb.Response, error) {
	MsgChan.recv <- req
	for {
		resp := <-MsgChan.resp
		if resp.Msg == "" {
			return resp, nil
		}
	}
}

func (s *Server) GateStream(stream pb.MessagePusher_GateStreamServer) (err error) {
	fmt.Printf("Received ComputeAverage RPC\n")
	// s.recv = make(chan *pb.Message)

	go func(stream pb.MessagePusher_GateStreamServer) {
		for {
			resp, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatalf("Error while reading client stream: %v", err)
			}
			MsgChan.resp <- resp
		}
	}(stream)

	select {
	case msg := <-MsgChan.recv:
		stream.Send(msg)
	}

	return err
}

package handler

import (
	"context"
	"fmt"
	"io"
	"log"

	pb "github.com/mikudos/mikudos-message-pusher/proto/message-pusher"
)

// Server implement Message-Pusher Server
type Server struct {
	Recv     chan *pb.Message
	Resp     chan *pb.Response
	Returned map[string]map[int64]chan *pb.Response
	Done     chan bool
}

// PushToChannel push message to the message Gate
func (s *Server) PushToChannel(ctx context.Context, req *pb.Message) (*pb.Response, error) {
	s.Recv <- req
	res := &pb.Response{MsgId: req.MsgId, ChannelId: req.ChannelId}
	return res, nil
}

// PushToChannelWithStatus push message to the message Gate and wait for result
func (s *Server) PushToChannelWithStatus(ctx context.Context, req *pb.Message) (*pb.Response, error) {
	s.Recv <- req
	mid := req.GetMsgId()
	channelID := req.GetChannelId()
	if s.Returned[channelID] == nil {
		s.Returned[channelID] = map[int64]chan *pb.Response{mid: make(chan *pb.Response, 1)}
	} else if s.Returned[channelID][mid] == nil {
		s.Returned[channelID][mid] = make(chan *pb.Response, 1)
	}
	for {
		ret := <-s.Returned[channelID][mid]
		delete(s.Returned[channelID], mid)
		if len(s.Returned[channelID]) == 0 {
			delete(s.Returned, channelID)
		}
		return ret, nil
	}
}

// GateStream gate stream communication
func (s *Server) GateStream(stream pb.MessagePusher_GateStreamServer) (err error) {
	fmt.Printf("Received GateStream request\n")

	go func() {
		defer fmt.Printf("GateStream break\n")
		for {
			select {
			case <-s.Done:
				return
			case msg := <-s.Recv:
				stream.Send(msg)
				break
			case resp := <-s.Resp:
				fmt.Printf("receive Gate response: %v\n", resp)
				break
			}
		}
	}()

	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			s.Done <- true
			break
		}
		if err != nil {
			s.Done <- true
			break
			log.Fatalf("Error while reading client stream: %v", err)
		}
		s.Resp <- resp
	}
	fmt.Printf("GateStream break\n")
	return err
}

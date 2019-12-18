package handler

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"

	pb "github.com/mikudos/mikudos-message-pusher/proto/message-pusher"
	"google.golang.org/grpc/metadata"
)

// Server implement Message-Pusher Server
type Server struct {
	streamId  int
	Mode      string
	Recv      chan *pb.Message
	Returned  map[string]map[int64]chan *pb.Response
	GroupRecv map[string]chan *pb.Message
	EveryRecv map[int]chan *pb.Message
}

func (s *Server) pushToModeChannel(req *pb.Message) {
	switch s.Mode {
	case "every":
		for _, Ch := range s.EveryRecv {
			Ch <- req
		}
		break
	case "group":
		for _, Ch := range s.GroupRecv {
			Ch <- req
		}
		break
	case "unify":
		s.Recv <- req
		break
	}
}

// PushToChannel push message to the message Gate
func (s *Server) PushToChannel(ctx context.Context, req *pb.Message) (*pb.Response, error) {
	s.pushToModeChannel(req)
	res := &pb.Response{MsgId: req.MsgId, ChannelId: req.ChannelId}
	return res, nil
}

// PushToChannelWithStatus push message to the message Gate and wait for result
func (s *Server) PushToChannelWithStatus(ctx context.Context, req *pb.Message) (*pb.Response, error) {
	s.pushToModeChannel(req)
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
	var (
		GateId  int
		GroupId string
	)
	if md, ok := metadata.FromIncomingContext(stream.Context()); ok {
		if len(md["group"]) > 0 {
			GroupId = md["group"][0]
		}
	}
	switch s.Mode {
	case "every":
		s.streamId += 1
		GateId = s.streamId
		s.EveryRecv[GateId] = make(chan *pb.Message)
		break
	case "group":
		if GroupId == "" {
			err = errors.New("METADATA of group cannot be empty")
			return err
		}
		s.GroupRecv[GroupId] = make(chan *pb.Message)
		break
	}
	go func() {
		defer fmt.Printf("GateStream break\n")
		for {
			select {
			case <-stream.Context().Done():
				delete(s.EveryRecv, GateId)
				return
			case msg := <-s.Recv:
				stream.Send(msg)
				break
			case msg := <-s.GroupRecv[GroupId]:
				stream.Send(msg)
				break
			case msg := <-s.EveryRecv[GateId]:
				stream.Send(msg)
				break
			}
		}
	}()

	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}
		channelID := resp.GetChannelId()
		msgId := resp.GetMsgId()
		if !resp.GetReceived() { // message not received
			// msg := resp.GetMsg()
		}
		if s.Returned[channelID] != nil && s.Returned[channelID][msgId] != nil {
			s.Returned[channelID][msgId] <- resp
		} else {
			log.Printf("channelID: %v\n", channelID)
		}
	}
	fmt.Printf("GateStream break\n")
	return err
}

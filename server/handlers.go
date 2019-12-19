package server

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	pb "github.com/mikudos/mikudos-message-pusher/proto/message-pusher"
	"google.golang.org/grpc/metadata"
)

// PushToChannel push message to the message Gate
func (s *Server) PushToChannel(ctx context.Context, req *pb.PushMessage) (*pb.Response, error) {
	s.pushToModeChannel(&pb.Message{Msg: req.GetMsg(), ChannelId: req.GetChannelId(), MsgId: req.GetMsgId(), Expire: req.GetExpire()})
	res := &pb.Response{MsgId: req.MsgId, ChannelId: req.ChannelId}
	return res, nil
}

// PushToChannelWithStatus push message to the message Gate and wait for result
func (s *Server) PushToChannelWithStatus(ctx context.Context, req *pb.PushMessage) (*pb.Response, error) {
	s.pushToModeChannel(&pb.Message{Msg: req.GetMsg(), ChannelId: req.GetChannelId(), MsgId: req.GetMsgId(), Expire: req.GetExpire()})
	mid := req.GetMsgId()
	channelID := req.GetChannelId()
	if s.Returned[channelID] == nil {
		s.Returned[channelID] = map[int64]chan *pb.Response{mid: make(chan *pb.Response, 1)}
	} else if s.Returned[channelID][mid] == nil {
		s.Returned[channelID][mid] = make(chan *pb.Response, 1)
	}
	timeOut := make(chan bool, 1)
	go func() {
		time.Sleep(5 * time.Second)
		if timeOut != nil {
			timeOut <- true
		}
	}()
	select {
	case ret := <-s.Returned[channelID][mid]:
		timeOut = nil
		delete(s.Returned[channelID], mid)
		if len(s.Returned[channelID]) == 0 {
			delete(s.Returned, channelID)
		}
		return ret, nil
	case <-timeOut:
		return nil, errors.New("time out")
	}
}

// GateStream gate stream communication
func (s *Server) GateStream(stream pb.MessagePusher_GateStreamServer) (err error) {
	var (
		GateID  int
		GroupID string
	)
	if md, ok := metadata.FromIncomingContext(stream.Context()); ok {
		if len(md["group"]) > 0 {
			GroupID = md["group"][0]
		}
	}
	switch s.Mode {
	case "every":
		s.streamID++
		GateID = s.streamID
		s.EveryRecv[GateID] = make(chan *pb.Message)
		break
	case "group":
		if GroupID == "" {
			err = errors.New("METADATA of group cannot be empty")
			return err
		}
		s.GroupRecv[GroupID] = make(chan *pb.Message)
		break
	}
	go func() {
		defer fmt.Printf("GateStream break\n")
		for {
			select {
			case <-stream.Context().Done():
				delete(s.EveryRecv, GateID)
				return
			case msg := <-s.Recv:
				stream.Send(msg)
				break
			case msg := <-s.GroupRecv[GroupID]:
				stream.Send(msg)
				break
			case msg := <-s.EveryRecv[GateID]:
				stream.Send(msg)
				break
			}
		}
	}()

	for {
		resp, err := stream.Recv()
		if err != nil {
			break
		}
		channelID := resp.GetChannelId()
		msgID := resp.GetMsgId()
		if resp.GetRequested() { // request channel message
			msgs, err := s.Storage.GetChannel(channelID, msgID)
			if err != nil {
			}
			for _, m := range msgs {
				stream.Send(m)
			}
			fmt.Printf("msgs: %v\n", msgs)
		} else if !resp.GetReceived() && !resp.GetRequested() { // message not received
			msg := pb.Message{MsgId: msgID, ChannelId: channelID, Msg: resp.GetMsg(), Expire: resp.GetExpire()}
			s.SaveMsg <- &msg
		} else if resp.GetReceived() { // message received
			s.Storage.PushDel(channelID, msgID)
		}
		if s.Returned[channelID] != nil && s.Returned[channelID][msgID] != nil {
			s.Returned[channelID][msgID] <- resp
		} else {
			log.Printf("channelID: %v\n", channelID)
		}
	}
	fmt.Printf("GateStream break\n")
	return err
}

// GetConfig GetConfig
func (s *Server) GetConfig(ctx context.Context, req *pb.ConfigRequest) (*pb.ConfigResponse, error) {
	return &pb.ConfigResponse{}, nil
}

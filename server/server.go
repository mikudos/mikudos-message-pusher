package server

import (
	"fmt"

	"github.com/mikudos/mikudos-message-pusher/db"
	pb "github.com/mikudos/mikudos-message-pusher/proto/message-pusher"
)

// Server implement Message-Pusher Server
type Server struct {
	streamID  int
	Mode      string
	Recv      chan *pb.Message
	Returned  map[string]map[int64]chan *pb.Response
	GroupRecv map[string]chan *pb.Message
	EveryRecv map[int]chan *pb.Message
	SaveMsg   chan *pb.Message
	Storage   db.Storage
}

func (s *Server) pushToModeChannel(req *pb.Message) {
	fmt.Printf("push msg: %v\n", req)
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

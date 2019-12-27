package server

import (
	"github.com/mikudos/mikudos-message-pusher/db"
	pb "github.com/mikudos/mikudos-message-pusher/proto/message-pusher"
)

// Server implement Message-Pusher Server
type Server struct {
	AID       uint32
	streamID  uint32
	Mode      string
	Recv      chan *pb.Message
	Returned  map[string]map[uint32]chan *pb.Response
	GroupRecv map[string]chan *pb.Message
	EveryRecv map[uint32]chan *pb.Message
	SaveMsg   chan *pb.Message
	Storage   db.Storage
}

func (s *Server) increment() {
	s.AID++
}

func (s *Server) pushToModeChannel(req *pb.Message) {
	s.increment()
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

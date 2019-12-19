package handler

import (
	"encoding/json"

	"github.com/mikudos/mikudos-message-pusher/db"
	pb "github.com/mikudos/mikudos-message-pusher/proto/message-pusher"
)

// Handler Server instance
var Handler Server

func init() {
	Handler = Server{Mode: "group", Recv: make(chan *pb.Message), Returned: make(map[string]map[int64]chan *pb.Response), GroupRecv: make(map[string]chan *pb.Message), EveryRecv: make(map[int]chan *pb.Message), SaveMsg: make(chan *pb.Message)}

	db.InitStorage()

	initReadRoutine()
}

func initReadRoutine() {
	for index := 0; index < 3; index++ {
		go ReadSaveMsg(&Handler)
	}
}

// ReadSaveMsg ReadSaveMsg method
func ReadSaveMsg(h *Server) {
	for {
		msg := <-h.SaveMsg
		db.UseStorage.SaveChannel("", json.RawMessage(msg.GetMsg()), msg.GetMsgId(), 1)
	}
}

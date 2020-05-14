package server

import (
	"fmt"
	"log"

	"github.com/chenhg5/collection"
	"github.com/mikudos/mikudos-message-pusher/config"
	"github.com/mikudos/mikudos-message-pusher/db"
	pb "github.com/mikudos/mikudos-message-pusher/proto/message-pusher"
)

// Handler Server instance
var Handler Server

func init() {
	mode := config.RuntimeViper.GetString("mode")
	defaults := []string{"unify", "group", "every"}

	fmt.Println(collection.Collect(defaults).Contains(mode))
	if !collection.Collect(defaults).Contains(mode) {
		mode = "unify"
	}
	Handler = Server{Mode: mode, Recv: make(chan *pb.Message), Returned: make(map[string]map[uint32]chan *pb.Response), GroupRecv: make(map[string]chan *pb.Message), EveryRecv: make(map[uint32]chan *pb.Message), SaveMsg: make(chan *pb.Message)}

	db.InitConfig()
	storage, err := db.InitStorage()
	if err != nil {
		log.Fatalf("InitStorage err: %v\n", err)
	}
	Handler.Storage = *storage

	initReadRoutine()
}

package server

import "encoding/json"

func initReadRoutine() {
	for index := 0; index < 1; index++ {
		go ReadSaveMsg(&Handler)
	}
}

// ReadSaveMsg ReadSaveMsg method
func ReadSaveMsg(h *Server) {
	for {
		msg := <-h.SaveMsg
		Handler.Storage.SaveChannel(msg.GetChannelId(), json.RawMessage(msg.GetMsg()), msg.GetMsgId(), uint(msg.GetExpire()))
	}
}

package handler

import (
	"encoding/json"
	"fmt"

	"github.com/robfig/cron/v3"
	"github.com/mikudos/mikudos-schedule/broker"
	"github.com/mikudos/mikudos-schedule/clients"
	pb "github.com/mikudos/mikudos-schedule/proto/schedule"
	"github.com/mikudos/mikudos-schedule/schedule"
)

// AddGrpcCron add cron to emit grpc-call events
func AddGrpcCron(scheduleStr string, grpc *pb.GrpcCall, scs *pb.Schedule, isOneTime bool) (jobID cron.EntryID, err error) {
	// 解析并校验grpc信息是否正确
	switch grpc.GetClientName() {
	case "ai":
		client := clients.AiService{}
		switch grpc.GetMethodName() {
		case "SayHello":
			json.Unmarshal([]byte(grpc.PayloadStr), &client.HelloRequest)
			// cV := reflect.ValueOf(&client)
			// log.Println("ai method:", cV.MethodByName(grpc.MethodName).Call(nil))
			jobID, err = schedule.Cron.AddFunc(scheduleStr, func() {
				client.ClientFunc()
				checkCancelList(jobID)
			})
		case "SayHi":
			json.Unmarshal([]byte(grpc.PayloadStr), &client.HelloRequest)
			// cV := reflect.ValueOf(&client)
			// log.Println("ai method:", cV.MethodByName(grpc.MethodName).Call(nil))
			jobID, err = schedule.Cron.AddFunc(scheduleStr, func() {
				client.ClientFunc()
				checkCancelList(jobID)
			})
		default:
			return 0, fmt.Errorf("set cron task fail: %s", "没有对应grpc method")
		}
	case "learn":
	case "messages":
	case "users":
	default:
		return 0, fmt.Errorf("set cron task fail: %s", "没有对应grpc client")
	}
	scs.Id = int32(jobID)
	scs.ScheduleName += "(GRPC)"
	if isOneTime {
		scs.ScheduleName += "(ONETIME)"
	}
	b, jsonErr := json.Marshal(scs)
	var str string
	if jsonErr != nil {
		str = fmt.Sprintf("{\"ScheduleName\":\"%s\", \"ScheduleComment\": \"%s\"}", scs.GetScheduleName(), scs.GetScheduleComment())
	} else {
		str = string(b)
	}
	if isOneTime {
		schedule.OneTimeJobs[jobID] = str
	} else {
		schedule.CronJobs[jobID] = str
	}
	return jobID, err
}

// AddBrokerCron add cron to emit broker-call events
func AddBrokerCron(scheduleStr string, brokerEvent *pb.BrokerEvent, scs *pb.Schedule, isOneTime bool) (jobID cron.EntryID, err error) {
	jobID, err = schedule.Cron.AddFunc(scheduleStr, func() {
		fmt.Println("run oneTime cron job", schedule.CronJobs)
		broker.BrokerInstance.Send(broker.Msg{Topic: brokerEvent.GetTopic(), Key: brokerEvent.GetKey(), Message: brokerEvent.GetMessage()})
		checkCancelList(jobID)
	})
	scs.Id = int32(jobID)
	scs.ScheduleName += "(BROKER)"
	if isOneTime {
		scs.ScheduleName += "(ONETIME)"
	}
	b, jsonErr := json.Marshal(scs)
	var str string
	if jsonErr != nil {
		str = fmt.Sprintf("{\"ScheduleName\":\"%s\", \"ScheduleComment\": \"%s\"}", scs.GetScheduleName(), scs.GetScheduleComment())
	} else {
		str = string(b)
	}
	if isOneTime {
		schedule.OneTimeJobs[jobID] = str
	} else {
		schedule.CronJobs[jobID] = str
	}
	return jobID, err
}

// RemoveCron aa
func RemoveCron(jobID cron.EntryID, isOneTime bool) {
	schedule.Cron.Remove(jobID)
	if isOneTime {
		delete(schedule.OneTimeJobs, jobID)
	} else {
		delete(schedule.CronJobs, jobID)
	}
}

func checkCancelList(jobID cron.EntryID) {
	if _, ok := schedule.OneTimeJobs[jobID]; ok {
		schedule.Cron.Remove(jobID)
		delete(schedule.OneTimeJobs, jobID)
	}
}

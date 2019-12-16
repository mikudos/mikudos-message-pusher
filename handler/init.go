package handler

import (
	"log"

	pb "github.com/mikudos/mikudos-schedule/proto/schedule"
)

func init() {
	id, err := AddGrpcCron("@every 5s", &pb.GrpcCall{
		ClientName: "ai",
		MethodName: "SayHello",
		PayloadStr: `
		{
			"name":"Yue Guanyu",
			"age":12
		}
		`,
	}, &pb.Schedule{
		ScheduleName:    "测试 ai.SayHello 任务",
		ScheduleComment: "每隔5秒钟调用一次ai.SayHello",
	}, false)
	if err != nil {
		log.Println(err)
	} else {
		log.Println("cron id:", id)
	}
}

package handler

import (
	"context"
	"encoding/json"
	"log"

	"github.com/robfig/cron/v3"

	pb "github.com/mikudos/mikudos-schedule/proto/schedule"
	"github.com/mikudos/mikudos-schedule/schedule"
)

// Server 事件驱动服务间流程控制方法，提供基本的数据库操作方法
type Server struct {
	pb.ScheduleServiceServer
}

func sendAllCrons(server pb.ScheduleService_ListScheduleServer, jobs map[cron.EntryID]string) (err error) {
	for id, str := range jobs {
		sc := pb.Schedule{}
		err = json.Unmarshal([]byte(str), &sc)
		if err != nil {
			log.Printf("error by unmarshal JSON of cron Info, jobID: %d", id)
		}
		server.Send(&sc)
	}
	return err
}

// ListSchedule aaaa
func (s *Server) ListSchedule(req *pb.ListScheduleRequest, server pb.ScheduleService_ListScheduleServer) (err error) {
	// session := db.Session.Copy()
	// defer session.Close()
	// broker.BrokerInstance.Send(broker.Msg{Topic: "", Key: "", Message: ""})
	// return &pb.Schedule{ScheduleName: "Hi, " + req.GetName()}, nil
	switch req.GetType() {
	case 1:
		err = sendAllCrons(server, schedule.CronJobs)
	case 2:
		err = sendAllCrons(server, schedule.OneTimeJobs)
	default:
		for _, jobs := range []map[cron.EntryID]string{schedule.CronJobs, schedule.OneTimeJobs} {
			if err = sendAllCrons(server, jobs); err != nil {
				break
			}
		}
	}
	return err
}

// CreateOneTimeGrpcSchedule CreateOneTimeGrpcSchedule
func (s *Server) CreateOneTimeGrpcSchedule(ctx context.Context, req *pb.CreateGrpcScheduleRequest) (*pb.Schedule, error) {
	sd := req.GetSchedule()
	id, err := AddGrpcCron(req.GetPeriod(), req.GetGrpcCall(), sd, true)
	if err != nil {
		log.Println("grpc cron job register fail")
		return nil, err
	}
	sd.Id = int32(id)
	return sd, nil
}

// CreateGrpcSchedule aaa
func (s *Server) CreateGrpcSchedule(ctx context.Context, req *pb.CreateGrpcScheduleRequest) (*pb.Schedule, error) {
	sd := req.GetSchedule()
	id, err := AddGrpcCron(req.GetPeriod(), req.GetGrpcCall(), sd, false)
	if err != nil {
		log.Println("grpc cron job register fail")
		return nil, err
	}
	sd.Id = int32(id)
	return sd, nil
}

// UpdateGrpcSchedule UpdateGrpcSchedule
func (s *Server) UpdateGrpcSchedule(ctx context.Context, req *pb.UpdateGrpcScheduleRequest) (sd *pb.Schedule, err error) {
	sd = req.GetSchedule()
	id := cron.EntryID(sd.GetId())
	RemoveCron(id, false)
	id, err = AddGrpcCron(req.GetPeriod(), req.GetGrpcCall(), sd, false)
	if err != nil {
		log.Println("grpc cron job register fail")
		return nil, err
	}
	sd.Id = int32(id)
	return sd, nil
}

// CreateOneTimeBrokerSchedule CreateOneTimeBrokerSchedule
func (s *Server) CreateOneTimeBrokerSchedule(ctx context.Context, req *pb.CreateBrokerScheduleRequest) (*pb.Schedule, error) {
	sd := req.GetSchedule()
	id, err := AddBrokerCron(req.GetPeriod(), req.GetBrokerEvent(), sd, true)
	if err != nil {
		log.Println("grpc cron job register fail")
		return nil, err
	}
	sd.Id = int32(id)
	return sd, nil
}

// CreateBrokerSchedule CreateBrokerSchedule
func (s *Server) CreateBrokerSchedule(ctx context.Context, req *pb.CreateBrokerScheduleRequest) (*pb.Schedule, error) {
	sd := req.GetSchedule()
	id, err := AddBrokerCron(req.GetPeriod(), req.GetBrokerEvent(), sd, false)
	if err != nil {
		log.Println("broker cron job register fail")
		return nil, err
	}
	sd.Id = int32(id)
	return sd, nil
}

// UpdateBrokerSchedule UpdateBrokerSchedule
func (s *Server) UpdateBrokerSchedule(ctx context.Context, req *pb.UpdateBrokerScheduleRequest) (sd *pb.Schedule, err error) {
	sd = req.GetSchedule()
	id := cron.EntryID(sd.GetId())
	RemoveCron(id, false)
	id, err = AddBrokerCron(req.GetPeriod(), req.GetBrokerEvent(), sd, false)
	if err != nil {
		log.Println("grpc cron job register fail")
		return nil, err
	}
	sd.Id = int32(id)
	return sd, nil
}

// CancelSchedule CancelSchedule
func (s *Server) CancelSchedule(ctx context.Context, req *pb.Schedule) (*pb.Schedule, error) {
	jobID := cron.EntryID(req.GetId())
	if _, ok := schedule.OneTimeJobs[jobID]; ok {
		RemoveCron(jobID, true)
	} else if _, ok := schedule.CronJobs[jobID]; ok {
		RemoveCron(jobID, false)
	}
	return req, nil
}

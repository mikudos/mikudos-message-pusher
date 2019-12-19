package db

import (
	"encoding/json"
	"errors"
	"fmt"

	log "github.com/alecthomas/log4go"
	pb "github.com/mikudos/mikudos-message-pusher/proto/message-pusher"
)

const (
	RedisStorageType = "redis"
	MySQLStorageType = "mysql"
	ketamaBase       = 255
	saveBatchNum     = 1000
)

var (
	UseStorage     Storage
	ErrStorageType = errors.New("unknown storage type")
)

// Stored messages interface
type Storage interface {
	// GetChannel get channel msgs.
	GetChannel(key string, mid int64) ([]*pb.Message, error)
	// SaveChannel Save single channel msg.
	SaveChannel(key string, msg json.RawMessage, mid int64, expire uint) error
	// SaveChannels save channel msgs return failed keys.
	SaveChannels(keys []string, msg json.RawMessage, mid int64, expire uint) ([]string, error)
	//
	PushDel(key string, mid int64) error
	// DelChannel delete channel msgs.
	DelChannel(key string) error
}

// InitStorage init the storage type(mysql or redis).
func InitStorage() (*Storage, error) {
	fmt.Printf("conf: %v\n", Conf)
	if Conf.StorageType == RedisStorageType {
		UseStorage = NewRedisStorage()
	} else if Conf.StorageType == MySQLStorageType {
		UseStorage = NewMySQLStorage()
	} else {
		log.Error("unknown storage type: \"%s\"", Conf.StorageType)
		return nil, ErrStorageType
	}
	return &UseStorage, nil
}

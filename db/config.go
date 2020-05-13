package db

import (
	"runtime"
	"time"

	"github.com/mikudos/mikudos-message-pusher/config"
)

var (
	Conf *Config
)

// Config struct
type Config struct {
	RPCBind          []string          `goconf:"base:rpc.bind:,"`
	NodeWeight       int               `goconf:"base:node.weight"`
	User             string            `goconf:"base:user"`
	PidFile          string            `goconf:"base:pidfile"`
	Dir              string            `goconf:"base:dir"`
	Log              string            `goconf:"base:log"`
	MaxProc          int               `goconf:"base:maxproc"`
	PprofBind        []string          `goconf:"base:pprof.bind:,"`
	StorageType      string            `goconf:"storage:type"`
	RedisIdleTimeout time.Duration     `goconf:"redis:timeout:time"`
	RedisMaxIdle     int               `goconf:"redis:idle"`
	RedisMaxActive   int               `goconf:"redis:active"`
	RedisMaxStore    int               `goconf:"redis:store"`
	MySQLClean       time.Duration     `goconf:"mysql:clean:time"`
	RedisSource      map[string]string `goconf:"-"`
	MySQLSource      map[string]string `goconf:"-"`
	// zookeeper
	ZookeeperAddr    []string      `goconf:"zookeeper:addr:,"`
	ZookeeperTimeout time.Duration `goconf:"zookeeper:timeout:time"`
	ZookeeperPath    string        `goconf:"zookeeper:path"`
}

// NewConfig parse config file into Config.
func InitConfig() error {
	Conf = &Config{
		// base
		RPCBind:    []string{"localhost:6379"},
		NodeWeight: 1,
		User:       "nobody nobody",
		PidFile:    "/tmp/gopush-cluster-message.pid",
		Dir:        "./",
		Log:        "./log/xml",
		MaxProc:    runtime.NumCPU(),
		PprofBind:  []string{"localhost:6379"},
		// storage
		StorageType: config.RuntimeViper.Get("storageType").(string),
		// redis
		RedisIdleTimeout: 28800 * time.Second,
		RedisMaxIdle:     50,
		RedisMaxActive:   1000,
		RedisMaxStore:    20,
		RedisSource:      make(map[string]string),
		// mysql
		MySQLSource: make(map[string]string),
		MySQLClean:  1 * time.Hour,
	}
	// redis section
	redisAddrsSec := config.RuntimeViper.Get("redisSource").(map[string]interface{})
	if redisAddrsSec != nil {
		for k, v := range redisAddrsSec {
			Conf.RedisSource[k] = v.(string)
		}
	}
	// mysql section
	dbSource := config.RuntimeViper.Get("mySQLSource").(map[string]interface{})
	if dbSource != nil {
		for k, v := range dbSource {
			Conf.MySQLSource[k] = v.(string)
		}
	}
	return nil
}

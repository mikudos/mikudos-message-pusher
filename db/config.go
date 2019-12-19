package db

import (
	"flag"
	"runtime"
	"time"
)

var (
	Conf     *Config
	confFile string
)

func init() {
	flag.StringVar(&confFile, "c", "./message.conf", " set message config file path")
}

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
		StorageType: "redis",
		// redis
		RedisIdleTimeout: 28800 * time.Second,
		RedisMaxIdle:     50,
		RedisMaxActive:   1000,
		RedisMaxStore:    20,
		RedisSource:      make(map[string]string),
		// mysql
		MySQLSource: make(map[string]string),
		MySQLClean:  1 * time.Hour,
		// zookeeper
		ZookeeperAddr:    []string{"localhost:2181"},
		ZookeeperTimeout: 30 * time.Second,
		ZookeeperPath:    "/gopush-cluster-message",
	}
	return nil
}

package db

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	log "github.com/alecthomas/log4go"
	"github.com/garyburd/redigo/redis"
	"github.com/mikudos/mikudos-message-pusher/ketama"
	pb "github.com/mikudos/mikudos-message-pusher/proto/message-pusher"
)

var (
	RedisNoConnErr       = errors.New("can't get a redis conn")
	redisProtocolSpliter = "@"
)

// RedisChannelMessage RedisMessage struct encoding the composite info.
type RedisChannelMessage struct {
	Msg    json.RawMessage `json:"msg"`    // message content
	Expire int64           `json:"expire"` // expire second
}

// RedisDelMessage Struct for delele message
type RedisDelMessage struct {
	Key  string
	MIds []int64
}

// RedisStorage RedisStorage
type RedisStorage struct {
	pool  map[string]*redis.Pool
	ring  *ketama.HashRing
	delCH chan *RedisDelMessage
}

// NewRedisStorage NewRedis initialize the redis pool and consistency hash ring.
func NewRedisStorage() *RedisStorage {
	redisPool := map[string]*redis.Pool{}
	ring := ketama.NewRing(ketamaBase)
	reg := regexp.MustCompile("(.+)@(.+)#(.+)|(.+)@(.+)")
	for n, addr := range Conf.RedisSource {
		nw := strings.Split(n, ":")
		if len(nw) != 2 {
			err := errors.New("node config error, it's nodeN:W")
			log.Error("strings.Split(\"%s\", :) failed (%v)", n, err)
			panic(err)
		}
		w, err := strconv.Atoi(nw[1])
		if err != nil {
			log.Error("strconv.Atoi(\"%s\") failed (%v)", nw[1], err)
			panic(err)
		}
		// get protocol and addr
		pw := reg.FindStringSubmatch(addr)
		if len(pw) < 6 {
			log.Error("strings.regexp(\"%s\", \"%s\") failed (%v)", addr, pw)
			panic(fmt.Sprintf("config redis.source node:\"%s\" format error", addr))
		}
		tmpProto := ""
		tmpAddr := ""
		if pw[1] != "" {
			tmpProto = pw[1]
		} else {
			tmpProto = pw[4]
		}
		if pw[2] != "" {
			tmpAddr = pw[2]
		} else {
			tmpAddr = pw[5]
		}

		// WARN: closures use
		redisPool[nw[0]] = &redis.Pool{
			MaxIdle:     Conf.RedisMaxIdle,
			MaxActive:   Conf.RedisMaxActive,
			IdleTimeout: Conf.RedisIdleTimeout,
			Dial: func() (redis.Conn, error) {
				conn, err := redis.Dial(tmpProto, tmpAddr)
				if err != nil {
					log.Error("redis.Dial(\"%s\", \"%s\") error(%v)", tmpProto, tmpAddr, err)
					return nil, err
				}
				if pw[3] != "" {
					conn.Do("AUTH", pw[3])
				}
				return conn, err
			},
		}
		// add node to ketama hash
		ring.AddNode(nw[0], w)
	}
	ring.Bake()
	s := &RedisStorage{pool: redisPool, ring: ring, delCH: make(chan *RedisDelMessage, 10240)}
	go s.clean()
	return s
}

// SaveChannel implements the Storage SaveChannel method.
func (s *RedisStorage) SaveChannel(key string, msg json.RawMessage, mid int64, expire uint) error {
	rm := &RedisChannelMessage{Msg: msg, Expire: int64(expire) + time.Now().Unix()}
	m, err := json.Marshal(rm)
	if err != nil {
		log.Error("json.Marshal() key:\"%s\" error(%v)", key, err)
		return err
	}
	conn := s.getConn(key)
	if conn == nil {
		log.Error("redis connection err")
		return RedisNoConnErr
	}
	defer conn.Close()
	// check if mid with key exists, if exists then return
	if PushMode == "group" || PushMode == "every" {
		err = conn.Send("ZCOUNT", key, mid, mid)
		err = conn.Flush()
		v, err := conn.Receive()
		if err != nil || v.(int64) > 0 {
			return err
		}
	}
	if err = conn.Send("ZADD", key, mid, m); err != nil {
		log.Error("conn.Send(\"ZADD\", \"%s\", %d, \"%s\") error(%v)", key, mid, string(m), err)
		return err
	}
	if err = conn.Send("ZREMRANGEBYRANK", key, 0, -1*(Conf.RedisMaxStore+1)); err != nil {
		log.Error("conn.Send(\"ZREMRANGEBYRANK\", \"%s\", 0, %d) error(%v)", key, -1*(Conf.RedisMaxStore+1), err)
		return err
	}
	if err = conn.Flush(); err != nil {
		log.Error("conn.Flush() error(%v)", err)
		return err
	}
	if _, err = conn.Receive(); err != nil {
		log.Error("conn.Receive() error(%v)", err)
		return err
	}
	if _, err = conn.Receive(); err != nil {
		log.Error("conn.Receive() error(%v)", err)
		return err
	}
	return nil
}

// SaveChannels implements the Storage SaveChannels method.
func (s *RedisStorage) SaveChannels(keys []string, msg json.RawMessage, mid int64, expire uint) (fkeys []string, err error) {
	// !TODO: check if mid with key exists, if exists then return
	// split as node
	nodes := map[string][]string{}
	fkeysMap := make(map[string]bool, len(keys))
	for _, k := range keys {
		node := s.ring.Hash(k)
		d, ok := nodes[node]
		if !ok {
			d = []string{k}
		} else {
			d = append(d, k)
		}
		nodes[node] = d
		fkeysMap[k] = true
	}
	// append return value
	defer func() {
		for k, _ := range fkeysMap {
			fkeys = append(fkeys, k)
		}
	}()
	// raw msg
	rm := &RedisChannelMessage{Msg: msg, Expire: int64(expire) + time.Now().Unix()}
	m, err := json.Marshal(rm)
	if err != nil {
		log.Error("json.Marshal() key:\"%s\" error(%v)", keys, err)
		return
	}
	// batch
	for n, k := range nodes {
		conn := s.getConnByNode(n)
		if conn == nil {
			log.Error("cann`t get redis connection by node:%s", n)
			err = RedisNoConnErr
			return
		}
		// pipeline batch msgs
		for _, key := range k {
			if err = conn.Send("ZADD", key, mid, m); err != nil {
				conn.Close()
				log.Error("conn.Send(\"ZADD\", \"%s\", %d, \"%s\") error(%v)", key, mid, string(m), err)
				return
			}
			if err = conn.Send("ZREMRANGEBYRANK", key, 0, -1*(Conf.RedisMaxStore+1)); err != nil {
				conn.Close()
				log.Error("conn.Send(\"ZREMRANGEBYRANK\", \"%s\", 0, %d) error(%v)", key, -1*(Conf.RedisMaxStore+1), err)
				return
			}
		}
		// flush commands
		if err = conn.Flush(); err != nil {
			conn.Close()
			log.Error("conn.Flush() error(%v)", err)
			return
		}
		// receive
		for j := 0; j < len(k); j++ {
			if _, err = conn.Receive(); err != nil {
				conn.Close()
				log.Error("conn.Receive() error(%v)", err)
				return
			}
			// delete succeed key
			delete(fkeysMap, k[j])
			if _, err = conn.Receive(); err != nil {
				conn.Close()
				log.Error("conn.Receive() error(%v)", err)
				return
			}
		}
		conn.Close()
	}
	return
}

// GetChannel implements the Storage GetChannel method.
func (s *RedisStorage) GetChannel(key string, mid int64) ([]*pb.Message, error) {
	conn := s.getConn(key)
	if conn == nil {
		return nil, RedisNoConnErr
	}
	defer conn.Close()
	values, err := redis.Values(conn.Do("ZRANGEBYSCORE", key, fmt.Sprintf("(%d", mid), "+inf", "WITHSCORES"))
	if err != nil {
		log.Error("conn.Do(\"ZRANGEBYSCORE\", \"%s\", \"%d\", \"+inf\", \"WITHSCORES\") error(%v)", key, mid, err)
		return nil, err
	}
	msgs := make([]*pb.Message, 0, len(values))
	delMsgs := []int64{}
	now := time.Now().Unix()
	for len(values) > 0 {
		cmid := int64(0)
		b := []byte{}
		values, err = redis.Scan(values, &b, &cmid)
		if err != nil {
			log.Error("redis.Scan() error(%v)", err)
			return nil, err
		}
		rm := &RedisChannelMessage{}
		if err = json.Unmarshal(b, rm); err != nil {
			log.Error("json.Unmarshal(\"%s\", rm) error(%v)", string(b), err)
			delMsgs = append(delMsgs, cmid)
			continue
		}
		// check expire
		if rm.Expire < now {
			log.Warn("user_key: \"%s\" msg: %d expired", key, cmid)
			delMsgs = append(delMsgs, cmid)
			continue
		}
		m := &pb.Message{MsgId: cmid, Msg: string(rm.Msg), ChannelId: ""}
		msgs = append(msgs, m)
	}
	// delete unmarshal failed and expired message
	if len(delMsgs) > 0 {
		select {
		case s.delCH <- &RedisDelMessage{Key: key, MIds: delMsgs}:
		default:
			log.Warn("user_key: \"%s\" send del messages failed, channel full", key)
		}
	}
	return msgs, nil
}

// PushDel PushDel
func (s *RedisStorage) PushDel(key string, mid int64) error {
	s.delCH <- &RedisDelMessage{Key: key, MIds: []int64{mid}}
	return nil
}

// DelChannel implements the Storage DelChannel method.
func (s *RedisStorage) DelChannel(key string) error {
	conn := s.getConn(key)
	if conn == nil {
		return RedisNoConnErr
	}
	defer conn.Close()
	if _, err := conn.Do("DEL", key); err != nil {
		log.Error("conn.Do(\"DEL\", \"%s\") error(%v)", key, err)
		return err
	}
	return nil
}

// DelMulti implements the Storage DelMulti method.
func (s *RedisStorage) clean() {
	for {
		info := <-s.delCH
		conn := s.getConn(info.Key)
		if conn == nil {
			log.Warn("get redis connection nil")
			continue
		}
		for _, mid := range info.MIds {
			if err := conn.Send("ZREMRANGEBYSCORE", info.Key, mid, mid); err != nil {
				log.Error("conn.Send(\"ZREMRANGEBYSCORE\", \"%s\", %d, %d) error(%v)", info.Key, mid, mid, err)
				conn.Close()
				continue
			}
		}
		if err := conn.Flush(); err != nil {
			log.Error("conn.Flush() error(%v)", err)
			conn.Close()
			continue
		}
		for _, _ = range info.MIds {
			_, err := conn.Receive()
			if err != nil {
				log.Error("conn.Receive() error(%v)", err)
				conn.Close()
				continue
			}
		}
		conn.Close()
	}
}

// getConn get the connection of matching with key using ketama hashing.
func (s *RedisStorage) getConn(key string) redis.Conn {
	if len(s.pool) == 0 {
		return nil
	}
	node := s.ring.Hash(key)
	log.Debug("user_key: \"%s\" hit redis node: \"%s\"", key, node)
	return s.getConnByNode(node)
}

func (s *RedisStorage) getConnByNode(node string) redis.Conn {
	p, ok := s.pool[node]
	if !ok {
		log.Warn("no node: \"%s\" in redis pool", node)
		return nil
	}

	return p.Get()
}

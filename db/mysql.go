package db

import (
	"database/sql"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	log "github.com/alecthomas/log4go"
	_ "github.com/go-sql-driver/mysql"
	"github.com/mikudos/mikudos-message-pusher/ketama"
	pb "github.com/mikudos/mikudos-message-pusher/proto/message-pusher"
)

const (
	savePrivateMsgSQL = "INSERT INTO private_msg(skey,mid,ttl,msg,ctime,mtime) VALUES(?,?,?,?,?,?)"
	// TODO limit
	getPrivateMsgSQL        = "SELECT mid, ttl, msg FROM private_msg WHERE skey=? AND mid>? ORDER BY mid"
	delExpiredPrivateMsgSQL = "DELETE FROM private_msg WHERE ttl<=?"
	delPrivateMsgSQL        = "DELETE FROM private_msg WHERE skey=?"
)

var (
	ErrNoMySQLConn     = errors.New("can't get a mysql db")
	mysqlSourceSpliter = ":"
)

// MySQL Storage struct
type MySQLStorage struct {
	pool map[string]*sql.DB
	ring *ketama.HashRing
}

// NewMySQLStorage initialize mysql pool and consistency hash ring.
func NewMySQLStorage() *MySQLStorage {
	dbPool := make(map[string]*sql.DB)
	ring := ketama.NewRing(ketamaBase)
	for n, source := range Conf.MySQLSource {
		nw := strings.Split(n, mysqlSourceSpliter)
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
		db, err := sql.Open("mysql", source)
		if err != nil {
			log.Error("sql.Open(\"mysql\", %s) failed (%v)", source, err)
			panic(err)
		}
		dbPool[nw[0]] = db
		ring.AddNode(nw[0], w)
	}
	ring.Bake()
	s := &MySQLStorage{pool: dbPool, ring: ring}
	go s.clean()
	return s
}

// SaveChannel implements the Storage SaveChannel method.
func (s *MySQLStorage) SaveChannel(key string, msg json.RawMessage, mid int64, expire uint) error {
	db := s.getConn(key)
	if db == nil {
		return ErrNoMySQLConn
	}
	now := time.Now()
	_, err := db.Exec(savePrivateMsgSQL, key, mid, now.Unix()+int64(expire), []byte(msg), now, now)
	if err != nil {
		log.Error("db.Exec(\"%s\",\"%s\",%d,%d,%d,\"%s\",now,now) failed (%v)", savePrivateMsgSQL, key, mid, expire, string(msg), err)
		return err
	}
	return nil
}

// SaveChannels implements the Storage SaveChannels method.
func (s *MySQLStorage) SaveChannels(keys []string, msg json.RawMessage, mid int64, expire uint) ([]string, error) {
	// TODO
	return nil, nil
}

// GetChannel implements the Storage GetChannel method.
func (s *MySQLStorage) GetChannel(key string, mid int64) ([]*pb.Message, error) {
	db := s.getConn(key)
	if db == nil {
		return nil, ErrNoMySQLConn
	}
	now := time.Now().Unix()
	rows, err := db.Query(getPrivateMsgSQL, key, mid)
	if err != nil {
		log.Error("db.Query(\"%s\",\"%s\",%d,now) failed (%v)", getPrivateMsgSQL, key, mid, err)
		return nil, err
	}
	msgs := []*pb.Message{}
	for rows.Next() {
		expire := int64(0)
		cmid := int64(0)
		msg := []byte{}
		if err := rows.Scan(&cmid, &expire, &msg); err != nil {
			log.Error("rows.Scan() failed (%v)", err)
			return nil, err
		}
		if now > expire {
			log.Warn("user_key: \"%s\" mid: %d expired", key, cmid)
			continue
		}
		msgs = append(msgs, &pb.Message{MsgId: cmid, ChannelId: "", Msg: json.RawMessage(msg)})
	}
	return msgs, nil
}

// PushDel PushDel
func (s *MySQLStorage) PushDel(key string, mid int64) error {
	return nil
}

// DelChannel implements the Storage DelChannel method.
func (s *MySQLStorage) DelChannel(key string) error {
	db := s.getConn(key)
	if db == nil {
		return ErrNoMySQLConn
	}
	res, err := db.Exec(delPrivateMsgSQL, key)
	if err != nil {
		log.Error("db.Exec(\"%s\", \"%s\") error(%v)", delPrivateMsgSQL, key, err)
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		log.Error("res.RowsAffected() error(%v)", err)
		return err
	}
	log.Info("user_key: \"%s\" clean message num: %d", rows)
	return nil
}

// clean delete expired messages peroridly.
func (s *MySQLStorage) clean() {
	for {
		log.Info("clean mysql expired message start")
		now := time.Now().Unix()
		affect := int64(0)
		for _, db := range s.pool {
			res, err := db.Exec(delExpiredPrivateMsgSQL, now)
			if err != nil {
				log.Error("db.Exec(\"%s\", %d) failed (%v)", delExpiredPrivateMsgSQL, now, err)
				continue
			}
			aff, err := res.RowsAffected()
			if err != nil {
				log.Error("res.RowsAffected() error(%v)", err)
				continue
			}
			affect += aff
		}
		log.Info("clean mysql expired message finish, num: %d", affect)
		time.Sleep(Conf.MySQLClean)
	}
}

// getConn get the connection of matching with key using ketama hash
func (s *MySQLStorage) getConn(key string) *sql.DB {
	if len(s.pool) == 0 {
		return nil
	}
	node := s.ring.Hash(key)
	p, ok := s.pool[node]
	if !ok {
		log.Warn("user_key: \"%s\" hit mysql node: \"%s\" not in pool", key, node)
		return nil
	}
	log.Info("user_key: \"%s\" hit mysql node: \"%s\"", key, node)
	return p
}

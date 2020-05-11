package outputer

import (
	"github.com/Acey9/apacket/logp"
	"github.com/go-redis/redis"
)

type Redis struct {
	client   *redis.Client
	key      string
	msgQueue chan string
}

func NewRedis(host, password string, db int, key string) Outputer {
	r := &Redis{
		msgQueue: make(chan string, 1024),
		key:      key,
		client: redis.NewClient(&redis.Options{
			Addr:       host,     // use default Addr
			Password:   password, // no password set
			DB:         db,       // use default DB
			MaxRetries: 10,
		}),
	}
	go r.Start()
	return r
}

func (rdb *Redis) Output(msg string) {
	logging(msg)
	rdb.msgQueue <- msg
}
func (rdb *Redis) RPush(msg string) {
	res := rdb.client.RPush(rdb.key, msg)
	if res.Err() != nil {
		logp.Err("redis.cmd.%v %v, failed msg:%v", res.Name(), res.Err(), msg)
	}
}

func (rdb *Redis) Start() {
	defer close(rdb.msgQueue)
	for {
		select {
		case msg := <-rdb.msgQueue:
			rdb.RPush(msg)
		}
	}
}

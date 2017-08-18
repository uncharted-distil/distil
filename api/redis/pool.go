package redis

import (
	"time"

	"github.com/garyburd/redigo/redis"
)

const (
	maxIdle     = 8
	idleTimeout = 10 * time.Second
)

// Pool represents a redis connection pool.
type Pool struct {
	pool   *redis.Pool
	expiry int
}

// NewPool instantiates and returns a new redis pool.
func NewPool(endpoint string, expiry int) *Pool {
	return &Pool{
		pool: &redis.Pool{
			MaxIdle:     maxIdle,
			IdleTimeout: idleTimeout,
			Dial: func() (redis.Conn, error) {
				conn, err := redis.Dial("tcp", endpoint)
				if err != nil {
					return nil, err
				}
				return conn, err
			},
			TestOnBorrow: func(conn redis.Conn, t time.Time) error {
				_, err := conn.Do("PING")
				return err
			},
		},
		expiry: expiry,
	}
}

// NewConn instantiates and returns a new redis connection.
func (p *Pool) NewConn() *Conn {
	return &Conn{
		conn:   p.pool.Get(),
		expiry: p.expiry,
	}
}

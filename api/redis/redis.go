package redis

import (
	"github.com/garyburd/redigo/redis"
)

// Conn represents a single connection to a redis server.
type Conn struct {
	conn   redis.Conn
	expiry int
}

// Get when given a string key will return a byte slice of data from redis.
func (c *Conn) Get(key string) ([]byte, error) {
	return redis.Bytes(c.conn.Do("GET", key))
}

// Set will store a byte slice under a given key in redis.
func (c *Conn) Set(key string, value []byte) error {
	var err error
	if c.expiry > 0 {
		_, err = c.conn.Do("SET", key, value, "NX", "EX", c.expiry)
	} else {
		_, err = c.conn.Do("SET", key, value)
	}
	return err
}

// Exists returns whether or not a key exists in redis.
func (c *Conn) Exists(key string) (bool, error) {
	return redis.Bool(c.conn.Do("Exists", key))
}

// Close closes the redis connection.
func (c *Conn) Close() error {
	return c.conn.Close()
}

package postgres

import (
	"strconv"
	"sync"
	"time"

	"github.com/jackc/pgx"
	"github.com/pkg/errors"
	"github.com/unchartedsoftware/plog"
)

const (
	defaultTimeout = time.Second * 30
)

var (
	mu      = &sync.Mutex{}
	clients map[string]*pgx.ConnPool
)

func init() {
	clients = make(map[string]*pgx.ConnPool)
}

// DatabaseDriver defines the behaviour of the querying engine.
type DatabaseDriver interface {
	Query(string, ...interface{}) (*pgx.Rows, error)
	QueryRow(string, ...interface{}) *pgx.Row
	Exec(string, ...interface{}) (pgx.CommandTag, error)
}

// ClientCtor repressents a client constructor to instantiate a postgres client.
type ClientCtor func() (DatabaseDriver, error)

// NewClient instantiates and returns a new postgres client constructor.
func NewClient(host, port, user, password, database string) ClientCtor {
	return func() (DatabaseDriver, error) {
		endpoint := host + ":" + port
		portInt, err := strconv.Atoi(port)
		if err != nil {
			return nil, errors.Wrap(err, "Unable to connect to Postgres endpoint")
		}

		mu.Lock()
		defer mu.Unlock()

		// see if we have an existing connection
		client, ok := clients[endpoint]
		if !ok {
			log.Infof("Creating new Postgres connection to endpoint %s", endpoint)
			dbConfig := pgx.ConnConfig{
				Host:     host,
				Port:     uint16(portInt),
				User:     user,
				Password: password,
				Database: database,
			}

			poolConfig := pgx.ConnPoolConfig{
				ConnConfig:     dbConfig,
				MaxConnections: 16,
			}
			//TODO: Need to close the pool eventually. Not sure how to hook that in.
			client, err := pgx.NewConnPool(poolConfig)

			if err != nil {
				return nil, errors.Wrap(err, "Postgres client init failed")
			}
			log.Infof("Postgres connection established to endpoint %s", endpoint)
			clients[endpoint] = client
		}
		log.Infof("Obtained Postgres connection to endpoint %s", endpoint)
		return client, nil
	}
}

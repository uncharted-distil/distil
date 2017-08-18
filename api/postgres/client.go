package postgres

import (
	"strconv"
	"sync"
	"time"

	"github.com/jackc/pgx"
	"github.com/pkg/errors"
	"github.com/unchartedsoftware/distil/api/model/filter"
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

// NewClient instantiates and returns a new postgres client constructor.
func NewClient(host, port, user, password, database string) filter.ClientCtor {
	return func() (filter.DatabaseDriver, error) {
		endpoint := host + ":" + port
		portInt, err := strconv.Atoi(port)
		if err != nil {
			return nil, err
		}

		mu.Lock()
		defer mu.Unlock()

		// see if we have an existing connection
		client, ok := clients[endpoint]
		if !ok {
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
		return client, nil
	}
}

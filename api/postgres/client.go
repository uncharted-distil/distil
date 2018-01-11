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

// Adapter for pgx logging
type pgxLogAdapter struct {
}

func (p pgxLogAdapter) Log(level pgx.LogLevel, msg string, data map[string]interface{}) {
	switch level {
	case pgx.LogLevelDebug:
		p.Debug(msg, data)
		break
	case pgx.LogLevelInfo:
		p.Info(msg, data)
		break
	case pgx.LogLevelWarn:
		p.Warn(msg, data)
		break
	case pgx.LogLevelError:
		p.Error(msg, data)
		break
	}
}

func (pgxLogAdapter) Debug(msg string, ctx ...interface{}) {
	log.Debugf("%s - %v", msg, ctx)
}

func (pgxLogAdapter) Info(msg string, ctx ...interface{}) {
	log.Infof("%s - %v", msg, ctx)
}

func (pgxLogAdapter) Warn(msg string, ctx ...interface{}) {
	log.Warnf("%s - %v", msg, ctx)
}

func (pgxLogAdapter) Error(msg string, ctx ...interface{}) {
	log.Errorf("%s - %v", msg, ctx)
}

// NewClient instantiates and returns a new postgres client constructor.  Log level is one
// of none, info, warn, error, debug.
func NewClient(host, port, user, password, database string, logLevel string) ClientCtor {
	return func() (DatabaseDriver, error) {
		endpoint := host + ":" + port
		portInt, err := strconv.Atoi(port)
		if err != nil {
			return nil, errors.Wrap(err, "Unable to connect to Postgres endpoint")
		}

		// Default logs to disabled - note that just setting level to 'none' is insufficient
		// as internally pgx defaults that to Debug.
		var level pgx.LogLevel
		level = pgx.LogLevelNone
		var logAdapter pgxLogAdapter
		if logLevel != "" {
			level, err = pgx.LogLevelFromString(logLevel)
			if err != nil {
				log.Warnf("Failed to parse log level [%s] with error [%s] - Disabling postgres logging", logLevel, err)
				level = pgx.LogLevelNone
			}
			if level != pgx.LogLevelNone {
				logAdapter = pgxLogAdapter{}
			}
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
				Logger:   logAdapter,
				LogLevel: int(level),
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

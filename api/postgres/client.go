//
//   Copyright © 2019 Uncharted Software Inc.
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.

package postgres

import (
	"fmt"
	"sync"
	"time"

	"github.com/go-pg/pg"
	"github.com/jackc/pgx"
	"github.com/pkg/errors"
	"github.com/unchartedsoftware/plog"
)

const (
	defaultTimeout = time.Second * 30
)

var (
	mu      = &sync.Mutex{}
	clients map[string]*IntegratedClient
)

func init() {
	clients = make(map[string]*IntegratedClient)
}

// DatabaseDriver defines the behaviour of the querying engine.
type DatabaseDriver interface {
	Query(string, ...interface{}) (*pgx.Rows, error)
	QueryRow(string, ...interface{}) *pgx.Row
	Exec(string, ...interface{}) (pgx.CommandTag, error)
	GetBatchClient() *pg.DB
}

// ClientCtor repressents a client constructor to instantiate a postgres client.
type ClientCtor func() (DatabaseDriver, error)

// Adapter for pgx logging
type pgxLogAdapter struct {
}

// IntegratedClient is a postgres client that can be used to
// query a postgres database.
type IntegratedClient struct {
	pgxClient *pgx.ConnPool
	host      string
	user      string
	password  string
	database  string
}

// GetBatchClient returns the client to use for updates.
func (ic IntegratedClient) GetBatchClient() *pg.DB {
	return pg.Connect(&pg.Options{
		Addr:     ic.host,
		User:     ic.user,
		Password: ic.password,
		Database: ic.database,
	})
}

// Query queries the database and returns the matching rows.
func (ic IntegratedClient) Query(sql string, params ...interface{}) (*pgx.Rows, error) {
	return ic.pgxClient.Query(sql, params...)
}

// QueryRow returns the first row from the query execution.
func (ic IntegratedClient) QueryRow(sql string, params ...interface{}) *pgx.Row {
	return ic.pgxClient.QueryRow(sql, params...)
}

// Exec executes the sql command.
func (ic IntegratedClient) Exec(sql string, params ...interface{}) (pgx.CommandTag, error) {
	return ic.pgxClient.Exec(sql, params...)
}

func (p pgxLogAdapter) Log(level pgx.LogLevel, msg string, data map[string]interface{}) {
	switch level {
	case pgx.LogLevelDebug:
		p.Debug(msg, data)
	case pgx.LogLevelInfo:
		p.Info(msg, data)
	case pgx.LogLevelWarn:
		p.Warn(msg, data)
	case pgx.LogLevelError:
		p.Error(msg, data)
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
func NewClient(host string, port int, user string, password string, database string, logLevel string) ClientCtor {
	return func() (DatabaseDriver, error) {
		endpoint := fmt.Sprintf("%s:%d", host, port)

		// Default logs to disabled - note that just setting level to 'none' is insufficient
		// as internally pgx defaults that to Debug.
		var level pgx.LogLevel
		var err error
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
				Port:     uint16(port),
				User:     user,
				Password: password,
				Database: database,
				Logger:   logAdapter,
				LogLevel: int(level),
			}

			poolConfig := pgx.ConnPoolConfig{
				ConnConfig:     dbConfig,
				MaxConnections: 64,
			}
			//TODO: Need to close the pool eventually. Not sure how to hook that in.
			pgxClient, err := pgx.NewConnPool(poolConfig)
			client = &IntegratedClient{
				pgxClient: pgxClient,
				host:      fmt.Sprintf("%s:%d", host, port),
				user:      user,
				password:  password,
				database:  database,
			}

			if err != nil {
				return nil, errors.Wrap(err, "Postgres client init failed")
			}
			log.Infof("Postgres connection established to endpoint %s", endpoint)
			clients[endpoint] = client
		}
		return client, nil
	}
}

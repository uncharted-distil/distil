//
//   Copyright © 2021 Uncharted Software Inc.
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
	"context"
	"fmt"
	"sync"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	pool "github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
	log "github.com/unchartedsoftware/plog"
)

var (
	mu           = &sync.Mutex{}
	clients      map[string]*IntegratedClient
	clientsBatch map[string]*IntegratedClient
)

func init() {
	clients = make(map[string]*IntegratedClient)
	clientsBatch = make(map[string]*IntegratedClient)
}

// DatabaseDriver defines the behaviour of the querying engine.
type DatabaseDriver interface {
	Begin() (pgx.Tx, error)
	Query(string, ...interface{}) (pgx.Rows, error)
	QueryRow(string, ...interface{}) pgx.Row
	Exec(string, ...interface{}) (pgconn.CommandTag, error)
	SendBatch(batch *pgx.Batch) pgx.BatchResults
	CopyFrom(string, []string, [][]interface{}) (int64, error)
}

// ClientCtor repressents a client constructor to instantiate a postgres client.
type ClientCtor func() (DatabaseDriver, error)

// Adapter for pgx logging
type pgxLogAdapter struct {
}

// IntegratedClient is a postgres client that can be used to
// query a postgres database.
type IntegratedClient struct {
	pgxClient *pool.Pool
	host      string
	user      string
	password  string
	database  string
}

// Query queries the database and returns the matching rows.
func (ic IntegratedClient) Query(sql string, params ...interface{}) (pgx.Rows, error) {
	return ic.pgxClient.Query(context.Background(), sql, params...)
}

// QueryRow returns the first row from the query execution.
func (ic IntegratedClient) QueryRow(sql string, params ...interface{}) pgx.Row {
	return ic.pgxClient.QueryRow(context.Background(), sql, params...)
}

// Exec executes the sql command.
func (ic IntegratedClient) Exec(sql string, params ...interface{}) (pgconn.CommandTag, error) {
	return ic.pgxClient.Exec(context.Background(), sql, params...)
}

// Begin creates a new transaction.
func (ic IntegratedClient) Begin() (pgx.Tx, error) {
	return ic.pgxClient.Begin(context.Background())
}

// SendBatch submits a batch.
func (ic IntegratedClient) SendBatch(batch *pgx.Batch) pgx.BatchResults {
	return ic.pgxClient.SendBatch(context.Background(), batch)
}

// CopyFrom copies data using the Postgres copy protocol for bulk data insertion.
func (ic IntegratedClient) CopyFrom(storageName string, columns []string, rows [][]interface{}) (int64, error) {
	sourceValues := pgx.CopyFromRows(rows)
	return ic.pgxClient.CopyFrom(context.Background(), pgx.Identifier{storageName}, columns, sourceValues)
}

func (p pgxLogAdapter) Log(ctx context.Context, level pgx.LogLevel, msg string, data map[string]interface{}) {
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
func NewClient(host string, port int, user string, password string, database string, logLevel string, batch bool) ClientCtor {
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
		var clientsMap map[string]*IntegratedClient
		if batch {
			clientsMap = clientsBatch
		} else {
			clientsMap = clients
		}
		client, ok := clientsMap[endpoint]
		if !ok {
			log.Infof("Creating new Postgres connection to endpoint %s", endpoint)
			connString := fmt.Sprintf("user=%s host=%s port=%d dbname=%s pool_max_conns=%d",
				user, host, port, database, 64)
			poolConfig, err := pool.ParseConfig(connString)
			if err != nil {
				return nil, errors.Wrap(err, "unable to parse postgres config")
			}
			poolConfig.LazyConnect = false
			poolConfig.ConnConfig.Logger = logAdapter
			poolConfig.ConnConfig.LogLevel = level
			// BuildStatementCache set to nil prevents the caching of queries
			// This does slow down performance when multiple of the same query is ran
			// However, this also causes issues when types are changing and the caches are not updated
			// One solution would be to reset all pool connection every time a type is changed (but for now this seems to be the best way)
			poolConfig.ConnConfig.BuildStatementCache = nil
			//TODO: Need to close the pool eventually. Not sure how to hook that in.
			pgxClient, err := pool.ConnectConfig(context.Background(), poolConfig)
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
			clientsMap[endpoint] = client
		}
		return client, nil
	}
}

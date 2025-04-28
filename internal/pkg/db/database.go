//go:generate mockgen -source ./database.go -destination=./mocks/database.go -package=mock_database
package db

import (
	"context"
	"github.com/georgysavva/scany/pgxscan"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type DB interface {
	Get(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	Select(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	Exec(ctx context.Context, query string, args ...interface{}) (pgconn.CommandTag, error)
	ExecQueryRow(ctx context.Context, query string, args ...interface{}) pgx.Row
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	GetPool() *pgxpool.Pool
	Close()
}

type PostgresDatabase struct {
	cluster *pgxpool.Pool
}

func newPostgresDatabase(cluster *pgxpool.Pool) *PostgresDatabase {
	return &PostgresDatabase{cluster: cluster}
}
func (db PostgresDatabase) GetPool() *pgxpool.Pool {
	return db.cluster
}

func (db PostgresDatabase) Get(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return pgxscan.Get(ctx, db.cluster, dest, query, args...)
}

func (db PostgresDatabase) Select(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return pgxscan.Select(ctx, db.cluster, dest, query, args...)
}

func (db PostgresDatabase) Exec(ctx context.Context, query string, args ...interface{}) (pgconn.CommandTag, error) {
	return db.cluster.Exec(ctx, query, args...)
}

func (db PostgresDatabase) ExecQueryRow(ctx context.Context, query string, args ...interface{}) pgx.Row {
	return db.cluster.QueryRow(ctx, query, args...)
}

func (db PostgresDatabase) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	return db.cluster.Query(ctx, sql, args)
}

func (db PostgresDatabase) Close() {
	db.cluster.Close()
}

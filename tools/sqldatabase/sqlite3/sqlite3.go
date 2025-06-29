package sqlite3

import (
	"context"
	"database/sql"

	"github.com/sayerxofficial/langchaingo/tools/sqldatabase"

	_ "github.com/mattn/go-sqlite3" // sqlite3 driver
)

const EngineName = "sqlite3"

//nolint:gochecknoinits
func init() {
	sqldatabase.RegisterEngine(EngineName, NewSQLite3)
}

var _ sqldatabase.Engine = SQLite3{}

// SQLite3 is a SQLite3 engine.
type SQLite3 struct {
	db *sql.DB
}

// NewSQLite3 creates a new SQLite3 engine.
// The dsn is the data source name.(e.g. file:locked.sqlite?cache=shared).
func NewSQLite3(dsn string) (sqldatabase.Engine, error) { //nolint:ireturn
	db, err := sql.Open(EngineName, dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)

	return &SQLite3{
		db: db,
	}, nil
}

func (m SQLite3) Dialect() string {
	return EngineName
}

func (m SQLite3) Query(ctx context.Context, query string, args ...any) ([]string, [][]string, error) {
	rows, err := m.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	cols, err := rows.Columns()
	if err != nil {
		return nil, nil, err
	}
	results := make([][]string, 0)
	for rows.Next() {
		row := make([]string, len(cols))
		rowNullable := make([]sql.NullString, len(cols))
		rowPtrs := make([]interface{}, len(cols))
		for i := range row {
			rowPtrs[i] = &rowNullable[i]
		}
		err = rows.Scan(rowPtrs...)
		if err != nil {
			return nil, nil, err
		}
		for i := range rowNullable {
			row[i] = rowNullable[i].String
		}
		results = append(results, row)
	}
	return cols, results, nil
}

func (m SQLite3) TableNames(ctx context.Context) ([]string, error) {
	_, result, err := m.Query(ctx, "SELECT name FROM sqlite_master WHERE type='table';")
	if err != nil {
		return nil, err
	}
	ret := make([]string, 0, len(result))
	for _, row := range result {
		ret = append(ret, row[0])
	}
	return ret, nil
}

func (m SQLite3) TableInfo(ctx context.Context, table string) (string, error) {
	_, result, err := m.Query(ctx, "SELECT sql FROM sqlite_master WHERE type='table' AND name=?;", table)
	if err != nil {
		return "", err
	}
	if len(result) == 0 {
		return "", sqldatabase.ErrTableNotFound
	}
	if len(result[0]) < 1 {
		return "", sqldatabase.ErrInvalidResult
	}

	return result[0][0], nil
}

func (m SQLite3) Close() error {
	return m.db.Close()
}

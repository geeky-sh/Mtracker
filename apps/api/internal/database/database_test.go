package database

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"sync"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/aash/mtracker/apps/api/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
)

// ── failingDialector ─────────────────────────────────────────────────────────

type failingDialector struct{}

func (failingDialector) Name() string                                    { return "failing" }
func (failingDialector) Initialize(*gorm.DB) error                       { return errors.New("open failed") }
func (failingDialector) Migrator(*gorm.DB) gorm.Migrator                 { return nil }
func (failingDialector) DataTypeOf(*schema.Field) string                 { return "" }
func (failingDialector) DefaultValueOf(*schema.Field) clause.Expression  { return nil }
func (failingDialector) BindVarTo(clause.Writer, *gorm.Statement, interface{}) {}
func (failingDialector) QuoteTo(clause.Writer, string)                   {}
func (failingDialector) Explain(string, ...interface{}) string           { return "" }

// ── fake permissive SQL driver ────────────────────────────────────────────────

var registerOnce sync.Once

type fakeDriver struct{}
type fakeConn struct{}
type fakeStmt struct{}
type fakeRows struct{ done bool }
type fakeResult struct{}
type fakeTx struct{}

func (fakeDriver) Open(string) (driver.Conn, error)      { return fakeConn{}, nil }

func (fakeConn) Prepare(string) (driver.Stmt, error)     { return fakeStmt{}, nil }
func (fakeConn) Close() error                            { return nil }
func (fakeConn) Begin() (driver.Tx, error)               { return fakeTx{}, nil }

func (fakeStmt) Close() error                            { return nil }
func (fakeStmt) NumInput() int                           { return -1 }
func (fakeStmt) Exec([]driver.Value) (driver.Result, error) { return fakeResult{}, nil }
func (fakeStmt) Query([]driver.Value) (driver.Rows, error)  { return &fakeRows{}, nil }

func (*fakeRows) Columns() []string              { return nil }
func (*fakeRows) Close() error                   { return nil }
func (r *fakeRows) Next([]driver.Value) error    { return io.EOF }

func (fakeResult) LastInsertId() (int64, error)  { return 0, nil }
func (fakeResult) RowsAffected() (int64, error)  { return 0, nil }

func (fakeTx) Commit() error                     { return nil }
func (fakeTx) Rollback() error                   { return nil }

func newFakeGormDB(t *testing.T) *gorm.DB {
	t.Helper()
	registerOnce.Do(func() {
		sql.Register("fake_postgres", fakeDriver{})
	})
	sqlDB, err := sql.Open("fake_postgres", "")
	require.NoError(t, err)
	db, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{})
	require.NoError(t, err)
	return db
}

// ── Tests ─────────────────────────────────────────────────────────────────────

func TestOpenDialector_Default(t *testing.T) {
	d := openDialector("host=localhost dbname=test")
	assert.NotNil(t, d)
	assert.Equal(t, "postgres", d.Name())
}

func TestConnect_OpenError(t *testing.T) {
	orig := openDialector
	openDialector = func(string) gorm.Dialector { return failingDialector{} }
	t.Cleanup(func() { openDialector = orig })

	_, err := Connect(&config.Config{DatabaseURL: ""})
	assert.Error(t, err)
}

func TestConnect_PgcryptoError(t *testing.T) {
	orig := openDialector
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	openDialector = func(string) gorm.Dialector {
		return postgres.New(postgres.Config{Conn: sqlDB})
	}
	t.Cleanup(func() { openDialector = orig })

	mock.ExpectExec(`CREATE EXTENSION`).WillReturnError(errors.New("pgcrypto denied"))

	_, err = Connect(&config.Config{DatabaseURL: ""})
	assert.Error(t, err)
}

func TestConnect_AutoMigrateError(t *testing.T) {
	orig := openDialector
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	openDialector = func(string) gorm.Dialector {
		return postgres.New(postgres.Config{Conn: sqlDB})
	}
	t.Cleanup(func() { openDialector = orig })

	mock.ExpectExec(`CREATE EXTENSION`).WillReturnResult(sqlmock.NewResult(0, 0))
	// No further expectations — any AutoMigrate query fails as unexpected.

	_, err = Connect(&config.Config{DatabaseURL: ""})
	assert.Error(t, err)
}

func TestConnect_Success(t *testing.T) {
	orig := openDialector
	openDialector = func(string) gorm.Dialector {
		_ = newFakeGormDB(t) // ensure driver registered
		sqlDB, _ := sql.Open("fake_postgres", "")
		return postgres.New(postgres.Config{Conn: sqlDB})
	}
	t.Cleanup(func() { openDialector = orig })

	db, err := Connect(&config.Config{DatabaseURL: ""})
	require.NoError(t, err)
	assert.NotNil(t, db)
}

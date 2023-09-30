package testutil

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
)

type TestUtil struct {
	resource *dockertest.Resource
	pool     *dockertest.Pool
}

type TestDBOptions struct {
	User     string
	Password string
	DBName   string
}
type SeedOpts struct {
	Dev   bool
	Local bool
	DB    *sql.DB
}

func New() *TestUtil {
	return &TestUtil{}
}

func (tu *TestUtil) SetupPostgresContainer(t *testing.T, o TestDBOptions) error {
	t.Helper()

	pool, err := dockertest.NewPool("")
	if err != nil {
		return fmt.Errorf("could not construct pool: %w", err)
	}

	if err := pool.Client.Ping(); err != nil {
		return fmt.Errorf("could not connect to Docker: %w", err)
	}

	tu.pool = pool

	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "14",
		Env: []string{
			"POSTGRES_PASSWORD=" + o.Password,
			"POSTGRES_USER=" + o.User,
			"POSTGRES_DB=" + o.DBName,
			"listen_addresses='*'",
		},
	}, func(config *docker.HostConfig) {
		config.AutoRemove = true                                // resource の Purge 後にコンテナを削除する
		config.RestartPolicy = docker.RestartPolicy{Name: "no"} // コンテナの再起動を行わない
	})
	if err != nil {
		return fmt.Errorf("could not start resource: %w", err)
	}

	tu.resource = resource

	return nil
}

func (tu *TestUtil) ConnectDB(o TestDBOptions) (*sql.DB, error) {
	if tu.pool == nil || tu.resource == nil {
		return nil, fmt.Errorf("container is not initialized")
	}

	databaseURI := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable",
		o.User,
		o.Password,
		tu.resource.GetHostPort("5432/tcp"),
		o.DBName,
	)

	// NOTE: コンテナが立ち上がっても内部のアプリケーションが接続を受け付ける準備ができていない可能性がある
	var db *sql.DB
	tu.pool.MaxWait = 20 * time.Second
	if err := tu.pool.Retry(func() error {
		innerDB, err := sql.Open("pgx", databaseURI)
		if err != nil {
			return err
		}

		if err := innerDB.Ping(); err != nil {
			return fmt.Errorf("could not ping database: %w", err)
		}

		db = innerDB

		return nil
	}); err != nil {
		return nil, fmt.Errorf("could not connect to docker: %w", err)
	}

	return db, nil
}

func (tu *TestUtil) MigrateUp(db *sql.DB, migrateSQLPaths []string) error {
	if tu.pool == nil || tu.resource == nil {
		return fmt.Errorf("container is not initialized")
	}

	for _, migrateSQLPath := range migrateSQLPaths {
		migrateSQL, err := os.ReadFile(migrateSQLPath)
		if err != nil {
			return fmt.Errorf("could not read migrate SQL file: %w", err)
		}

		if _, err := db.Exec(string(migrateSQL)); err != nil {
			return fmt.Errorf("could not execute migrate SQL: %w", err)
		}
	}

	return nil
}

func (tu *TestUtil) Purge() {
	if tu.pool == nil || tu.resource == nil {
		return
	}

	if err := tu.pool.Purge(tu.resource); err != nil {
		log.Panicf("could not purge resource: %s", err)
	}
}

func mustExec(db *sql.DB, stmt string) {
	if _, err := db.Exec(stmt); err != nil {
		panic(fmt.Errorf("stmt=%s: %w", stmt, err))
	}
}

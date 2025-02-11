package db

import (
	"context"
	"crypto/tls"
	"database/sql"
	"fmt"
	"log"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	"connector-recruitment/go-server/connectors/config"
	"connector-recruitment/go-server/connectors/internal"

	"github.com/golang-migrate/migrate"
	"github.com/golang-migrate/migrate/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	_ "github.com/joho/godotenv/autoload"
)

// migrateDB syncs the `dbname` with the migrations specified in cwd()/sql
func migrateDB(dbname string, db *sql.DB) error {
	// log.Println("Migrating database...")
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Println(err)
		return err
	}

	log.Println("driver gotten")

	dir := internal.GetPackagePath()
	dir = filepath.Join(dir, "sql/migrations")
	migrationsDir := fmt.Sprintf("file:///%s", dir)

	log.Println("Migrations dir", dir)

	m, err := migrate.NewWithDatabaseInstance(
		migrationsDir,
		dbname, driver)
	if err != nil {
		log.Println(err)
		return err
	}

	log.Println("NewWithDatabaseInstance")

	if err := m.Up(); err != migrate.ErrNoChange {
		log.Println(err)
		return err
	}

	return nil
}

type Service struct {
	DBPool *pgxpool.Pool // TODO: depreciate this and make use of the DB field only
	DB     *sql.DB
}

var (
	database   string
	dbInstance *Service
)

func NewDB(env config.Env) *Service {
	// Reuse Connection
	if dbInstance != nil {
		return dbInstance
	}
	database = env.PostgresDatabase
	connStr := dsnFromEnv(env)

	// attempts to create db pool
	sqldb, postgresdb, err := newStdDB(connStr, env)
	if err != nil {
		panic(err)
	}

	maxOpenConns := 4 * runtime.GOMAXPROCS(0)
	sqldb.SetMaxOpenConns(maxOpenConns)
	sqldb.SetMaxIdleConns(maxOpenConns)

	/*
		if err := migrateDB(env.PostgresDatabase, sqldb); err != nil {
			fmt.Printf("failed to migrate db: %v\n", err)
			log.Fatalf("failed to migrate db: %v", err)
		}
	*/

	dbInstance = &Service{
		DBPool: postgresdb,
		DB:     sqldb,
	}
	return dbInstance
}

// Health checks the health of the database connection by pinging the database.
func (s *Service) Health() map[string]string {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	stats := make(map[string]string)

	err := s.DB.PingContext(ctx)
	if err != nil {
		stats["status"] = "down"
		stats["error"] = fmt.Sprintf("db down: %v", err)
		return stats
	}

	// Database is up, add more statistics
	stats["status"] = "up"
	stats["message"] = "It's healthy"

	// Get database stats
	dbStats := s.DB.Stats()
	stats["open_connections"] = strconv.Itoa(dbStats.OpenConnections)
	stats["in_use"] = strconv.Itoa(dbStats.InUse)
	stats["idle"] = strconv.Itoa(dbStats.Idle)
	stats["wait_count"] = strconv.FormatInt(dbStats.WaitCount, 10)
	stats["wait_duration"] = dbStats.WaitDuration.String()
	stats["max_idle_closed"] = strconv.FormatInt(dbStats.MaxIdleClosed, 10)
	stats["max_lifetime_closed"] = strconv.FormatInt(dbStats.MaxLifetimeClosed, 10)

	if dbStats.OpenConnections > 40 {
		stats["message"] = "The database is experiencing heavy load."
	}

	if dbStats.WaitCount > 1000 {
		stats["message"] = "The database has a high number of wait events, indicating potential bottlenecks."
	}

	if dbStats.MaxIdleClosed > int64(dbStats.OpenConnections)/2 {
		stats["message"] = "Many idle connections are being closed, consider revising the connection pool settings."
	}

	if dbStats.MaxLifetimeClosed > int64(dbStats.OpenConnections)/2 {
		stats["message"] = "Many connections are being closed due to max lifetime, consider increasing max lifetime or revising the connection usage pattern."
	}

	return stats
}

// Close closes the database connection.
func (s *Service) Close() error {
	log.Printf("Disconnected from database: %s", database)
	return s.DB.Close()
}

func dsnFromEnv(env config.Env) string {
	_, err := strconv.Atoi(env.PostgresPort)
	if err != nil {
		panic(fmt.Sprintf("Invalid Postgres port: %s", env.PostgresPort))
	}

	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", env.PostgresUser, env.PostgresPassword, env.PostgresHost, env.PostgresPort, env.PostgresDatabase)
}

func newStdDB(connStr string, env config.Env) (*sql.DB, *pgxpool.Pool, error) {
	dbpool, err := pgxpool.New(context.Background(), connStr)
	if err != nil {
		return nil, nil, err
	}

	// create std db from pgxpool for the stats section
	sqldb := stdlib.OpenDBFromPool(dbpool, stdlib.OptionBeforeConnect(func(ctx context.Context, cc *pgx.ConnConfig) error {
		if !env.PostgresSecureMode {
			cc.TLSConfig = &tls.Config{InsecureSkipVerify: true}
		}
		return nil
	}))

	return sqldb, dbpool, nil
}

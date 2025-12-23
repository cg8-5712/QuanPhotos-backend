package database

import (
	"fmt"
	"os"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// readFile reads the content of a file
func readFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

// Config holds database configuration
type Config struct {
	Host            string
	Port            string
	Name            string
	User            string
	Password        string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// DB is the global database instance
var DB *sqlx.DB

// Connect establishes a connection to the database
func Connect(cfg Config) (*sqlx.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name, cfg.SSLMode,
	)

	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	// Verify connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	DB = db
	return db, nil
}

// Close closes the database connection
func Close() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}

// GetDB returns the global database instance
func GetDB() *sqlx.DB {
	return DB
}

// NeedsMigration checks if the database needs schema migration
// Returns true if the users table doesn't exist (as a proxy for empty database)
func NeedsMigration(db *sqlx.DB) (bool, error) {
	var exists bool
	query := `
		SELECT EXISTS (
			SELECT FROM information_schema.tables
			WHERE table_schema = 'public'
			AND table_name = 'users'
		)
	`
	if err := db.Get(&exists, query); err != nil {
		return false, fmt.Errorf("failed to check if tables exist: %w", err)
	}
	return !exists, nil
}

// AutoMigrate runs database migrations
// migrationsPath should be the path to migrations directory (e.g., "file://migrations")
func AutoMigrate(db *sqlx.DB, migrationsPath string) error {
	driver, err := postgres.WithInstance(db.DB, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create migration driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		migrationsPath,
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

// RunSeed executes seed data from the specified file
func RunSeed(db *sqlx.DB, seedPath string) error {
	// Read seed file
	content, err := readFile(seedPath)
	if err != nil {
		return fmt.Errorf("failed to read seed file: %w", err)
	}

	// Execute seed SQL
	if _, err := db.Exec(string(content)); err != nil {
		return fmt.Errorf("failed to execute seed: %w", err)
	}

	return nil
}

// IsSeedNeeded checks if seed data is needed (no users exist)
func IsSeedNeeded(db *sqlx.DB) (bool, error) {
	var count int
	if err := db.Get(&count, "SELECT COUNT(*) FROM users"); err != nil {
		return false, fmt.Errorf("failed to check users count: %w", err)
	}
	return count == 0, nil
}

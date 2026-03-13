package postgres

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"

	"practice3/pkg/modules"
)

type Dialect struct {
	DB *sqlx.DB
}

func NewPGXDialect(cfg modules.PostgreConfig) (*Dialect, error) {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.DBName, cfg.SSLMode,
	)

	db, err := sqlx.Connect("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("sqlx connect: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), cfg.ExecTimeout)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("db ping: %w", err)
	}

	if err := AutoMigrate(db, cfg); err != nil {
		return nil, fmt.Errorf("auto migrate: %w", err)
	}

	return &Dialect{DB: db}, nil
}

func AutoMigrate(db *sqlx.DB, cfg modules.PostgreConfig) error {
	sourceURL := "file://database/migrations"

	databaseURL := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.DBName, cfg.SSLMode,
	)

	driver, err := postgres.WithInstance(db.DB, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("migrate driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(sourceURL, databaseURL, driver)
	if err != nil {
		return fmt.Errorf("migrate.New: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("m.Up: %w", err)
	}

	return nil
}

func env(key, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}

func DefaultPostgresConfig() modules.PostgreConfig {
	return modules.PostgreConfig{
		Host:        env("DB_HOST", "localhost"),
		Port:        env("DB_PORT", "5432"),
		Username:    env("DB_USER", "postgres"),
		Password:    env("DB_PASSWORD", "postgres"),
		DBName:      env("DB_NAME", "mydb"),
		SSLMode:     env("DB_SSLMODE", "disable"),
		ExecTimeout: 5 * time.Second,
	}
}

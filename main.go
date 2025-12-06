package main

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"

	"github.com/danielgtaylor/huma/v2/humacli"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/tern/v2/migrate"
	"golang-api/api"
	"golang-api/config"
	"golang-api/db"
)

//go:embed db/migrations/*.sql
var migrationFiles embed.FS

// checkTablesExist проверяет существование необходимых таблиц в базе данных
func checkTablesExist(ctx context.Context, conn *pgxpool.Pool) (bool, error) {
	// Проверяем существование таблицы users
	var usersExists bool
	err := conn.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 
			FROM information_schema.tables 
			WHERE table_name = 'users' AND table_schema = 'public'
		)
	`).Scan(&usersExists)
	if err != nil {
		return false, err
	}

	// Проверяем существование таблицы subscriptions
	var subscriptionsExists bool
	err = conn.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 
			FROM information_schema.tables 
			WHERE table_name = 'subscriptions' AND table_schema = 'public'
		)
	`).Scan(&subscriptionsExists)
	if err != nil {
		return false, err
	}

	// Возвращаем true только если обе таблицы существуют
	return usersExists && subscriptionsExists, nil
}

func main() {
	logFile, err := os.OpenFile("application.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		slog.Error("Failed to open log file", "error", err)
		return
	}
	defer logFile.Close() // Ensure the log file is closed when main exits

	// Create a new slog logger that writes to the file
	logger := slog.New(slog.NewTextHandler(logFile, nil))

	// Set the default logger to use our file-based logger
	slog.SetDefault(logger)

	appConfig, err := config.New()
	if err != nil {
		slog.Error("config error: %v", err)
		panic(err)
	}

	connectionString := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		appConfig.Database.PGHost, appConfig.Database.PGPort, appConfig.Database.PGUser, appConfig.Database.PGPassword, appConfig.Database.PGName, appConfig.Database.SSLMode)

	ctx := context.Background()

	dbpool, err := pgxpool.New(ctx, connectionString)
	if err != nil {
		slog.Error("Unable to create connection pool: ", "error", err)
		panic(err)
	}
	
	conn, err := dbpool.Acquire(ctx)
	if err != nil {
		slog.Error("Unable to acquire connection from pool: ", "error", err)
		panic(err)
	}

	// Проверяем существование таблиц перед запуском миграций
	tablesExist, err := checkTablesExist(ctx, dbpool)
	if err != nil {
		slog.Error("Failed to check table existence", "error", err)
		panic(err)
	}

	if tablesExist {
		slog.Info("Tables already exist, skipping migrations")
	} else {
		slog.Info("Tables do not exist, applying migrations")

		migrator, err := migrate.NewMigrator(ctx, conn.Conn(), "migrations")
		if err != nil {
			slog.Error("Failed to create migrator", "error", err)
			panic(err)
		}

		migrationRoot, _ := fs.Sub(migrationFiles, "db/migrations")

		err = migrator.LoadMigrations(migrationRoot)
		if err != nil {
			panic(err)
		}

		if err := migrator.Migrate(ctx); err != nil {
			slog.Error("Failed to apply migrations", "error", err)
			panic(err)
		}

		slog.Info("Migrations applied successfully")
	}

	cli.Run()
}
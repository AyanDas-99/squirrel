package main

import (
	"context"
	"database/sql"
	"flag"
	"os"
	"time"

	_ "github.com/lib/pq"
	"test.com/internal/data"
	"test.com/internal/jsonlog"
)

type config struct {
	port int
	env  string
	db   struct {
		dsn string
	}
}

type application struct {
	config      config
	logger      *jsonlog.Logger
	items       *data.ItemModel
	issues      *data.IssueModel
	removals    *data.RemovalModel
	additions   *data.AdditionModel
	users       *data.UserModel
	tags        *data.TagModel
	tokens      *data.TokenModel
	permissions *data.PermissionModel
	org         *data.OrganizationsModel
}

func main() {
	// Initialize routes
	config := config{port: 8080}

	flag.StringVar(&config.db.dsn, "dsn", os.Getenv("TEST_DB_DSN"), "PostgreSQL DSN")
	flag.StringVar(&config.env, "env", "development", "Environment (development|staging|production)")
	flag.Parse()
	logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)

	db, err := openDB(config.db.dsn)
	if err != nil {
		logger.PrintFatal(err, map[string]string{})
	}
	logger.PrintInfo("Database connection setup", map[string]string{})
	defer db.Close()

	app := &application{
		items:       &data.ItemModel{DB: db},
		issues:      &data.IssueModel{DB: db},
		removals:    &data.RemovalModel{DB: db},
		additions:   &data.AdditionModel{DB: db},
		tags:        &data.TagModel{DB: db},
		org:         &data.OrganizationsModel{DB: db},
		users:       &data.UserModel{DB: db},
		tokens:      &data.TokenModel{DB: db},
		permissions: &data.PermissionModel{DB: db},
		logger:      logger,
		config:      config,
	}

	err = app.serve()
	if err != nil {
		logger.PrintFatal(err, map[string]string{})
	}
}

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}

	return db, nil
}

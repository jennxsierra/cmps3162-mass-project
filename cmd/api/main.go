package main

import (
	"context"
	"database/sql"
	"flag"
	"log/slog"
	"os"
	"time"

	"github.com/jennxsierra/mass-project/internal/data"
	_ "github.com/lib/pq"
)

const appVersion = "1.0.0"

type serverConfig struct {
	port        int
	environment string
	db          struct {
		dsn string
	}
	limiter struct {
		rps     float64 // requests per second
		burst   int     // initial requests possible
		enabled bool    // enable or disable rate limiter
	}
}

type applicationDependencies struct {
	config serverConfig
	logger *slog.Logger
	models data.Models
}

func main() {
	var settings serverConfig

	flag.IntVar(&settings.port, "port", 4000, "Server port")

	// read in the environment (development|staging|production)
	flag.StringVar(&settings.environment, "env", "development",
		"Environment(development|staging|production)")

	// read in the dsn
	flag.StringVar(&settings.db.dsn, "db-dsn", "", "PostgreSQL DSN")

	// read in the rate limiter settings
	flag.Float64Var(&settings.limiter.rps, "limiter-rps", 2,
		"Rate Limiter maximum requests per second")

	flag.IntVar(&settings.limiter.burst, "limiter-burst", 5,
		"Rate Limiter maximum burst")

	flag.BoolVar(&settings.limiter.enabled, "limiter-enabled", true,
		"Enable rate limiter")

	// parse the command-line flags
	flag.Parse()

	// initialize logger with custom handler for readable time format
	logHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Replace the default timestamp with a more readable format
			if a.Key == slog.TimeKey {
				return slog.String("time", time.Now().Format("2006-01-02 15:04:05"))
			}
			return a
		},
	})
	logger := slog.New(logHandler)

	// the call to openDB() sets up our connection pool
	db, err := openDB(settings)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	// release the database resources before exiting
	defer db.Close()

	logger.Info("database connection pool established")
	appInstance := &applicationDependencies{
		config: settings,
		logger: logger,
		models: data.NewModels(db),
	}

	err = appInstance.serve()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}

func openDB(settings serverConfig) (*sql.DB, error) {
	// open a connection pool
	db, err := sql.Open("postgres", settings.db.dsn)
	if err != nil {
		return nil, err
	}

	// create a context with a 5-second timeout for the ping operation
	ctx, cancel := context.WithTimeout(context.Background(),
		5*time.Second)
	defer cancel()
	// ping the database to check if it's alive
	err = db.PingContext(ctx)
	if err != nil {
		db.Close()
		return nil, err
	}

	// return the connection pool (sql.DB)
	return db, nil

}

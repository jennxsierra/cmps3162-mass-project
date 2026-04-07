package main

import (
	"context"
	"database/sql"
	"expvar"
	"flag"
	"log/slog"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/jennxsierra/mass-project/internal/data"
	"github.com/jennxsierra/mass-project/internal/mailer"
	_ "github.com/lib/pq"
)

const appVersion = "1.0.0"

type serverConfig struct {
	port        int
	environment string
	db          struct {
		dsn string
	}
	shutdown struct {
		timeout time.Duration
	}
	limiter struct {
		rps     float64 // requests per second
		burst   int     // initial requests possible
		enabled bool    // enable or disable rate limiter
	}
	smtp struct {
		host     string
		port     int
		username string
		password string
		sender   string
	}
	cors struct {
		trustedOrigins []string
	}
}

type applicationDependencies struct {
	config serverConfig
	logger *slog.Logger
	models data.Models
	mailer mailer.Mailer
	wg     sync.WaitGroup
}

func main() {
	var settings serverConfig

	// Server Flags
	flag.IntVar(&settings.port, "port", 4000, "Server port")
	flag.StringVar(&settings.environment, "env", "development",
		"Environment(development|staging|production)")
	flag.DurationVar(&settings.shutdown.timeout, "shutdown-timeout", 30*time.Second,
		"Graceful shutdown timeout")

	// Database Flags
	flag.StringVar(&settings.db.dsn, "db-dsn", "", "PostgreSQL DSN")

	// Rate Limiter Flags
	flag.Float64Var(&settings.limiter.rps, "limiter-rps", 2,
		"Rate Limiter maximum requests per second")
	flag.IntVar(&settings.limiter.burst, "limiter-burst", 5,
		"Rate Limiter maximum burst")
	flag.BoolVar(&settings.limiter.enabled, "limiter-enabled", true,
		"Enable rate limiter")

	// SMTP email server flags
	flag.StringVar(&settings.smtp.host, "smtp-host", "sandbox.smtp.mailtrap.io", "SMTP host")
	flag.IntVar(&settings.smtp.port, "smtp-port", 25, "SMTP port")
	flag.StringVar(&settings.smtp.username, "smtp-username", "", "SMTP username")
	flag.StringVar(&settings.smtp.password, "smtp-password", "", "SMTP password")
	flag.StringVar(&settings.smtp.sender, "smtp-sender", "Medical Appointment Scheduling System <no-reply@mass.jsierra.com>", "SMTP sender")

	// CORS Trusted Origins Flags
	flag.Func("cors-trusted-origins", "Trusted CORS origins (space separated)", func(val string) error {
		settings.cors.trustedOrigins = strings.Fields(val)
		return nil
	})

	// parse the command-line flags
	flag.Parse()

	// initialize structured logger in JSON format
	logHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
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

	// publish application metrics using expvar
	expvar.NewString("version").Set(appVersion)

	expvar.Publish("goroutines", expvar.Func(func() any {
		return runtime.NumGoroutine()
	}))

	expvar.Publish("database", expvar.Func(func() any {
		stats := db.Stats()
		return map[string]any{
			"open_connections":     stats.OpenConnections,
			"in_use":               stats.InUse,
			"idle":                 stats.Idle,
			"wait_count":           stats.WaitCount,
			"wait_duration":        stats.WaitDuration.String(),
			"max_idle_closed":      stats.MaxIdleClosed,
			"max_idle_time_closed": stats.MaxIdleTimeClosed,
			"max_lifetime_closed":  stats.MaxLifetimeClosed,
		}
	}))

	expvar.Publish("timestamp", expvar.Func(func() any {
		return time.Now().Unix()
	}))

	// create an instance of our application struct containing the dependencies
	appInstance := &applicationDependencies{
		config: settings,
		logger: logger,
		models: data.NewModels(db),
		mailer: mailer.New(
			settings.smtp.host,
			settings.smtp.port,
			settings.smtp.username,
			settings.smtp.password,
			settings.smtp.sender,
		),
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

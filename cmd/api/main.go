package main

import (
	"context"
	"log"

	"github.com/joho/godotenv"
	"github.com/mightyfzeus/rbac/internal/db"
	"github.com/mightyfzeus/rbac/internal/env"
	"github.com/mightyfzeus/rbac/internal/store"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("⚠️ Could not load .env file, falling back to defaults")
	}

	cfg := config{
		addr:   env.GetString("ADDR", ":8080"),
		apiUrl: env.GetString("API_URL", "localhost:8000"),
		db: dbConfig{
			dbAddr:       env.GetString("DB_ADDR", "postgres://admin:adminpassword@localhost:5433/nara?sslmode=disable"),
			maxOpenConns: env.GetInt("DB_MAX_OPEN_CONNS", 25),
			maxIdleConns: env.GetInt("DB_MAX_IDLE_CONNS", 25),
			maxIdleTime:  env.GetString("DB_MAX_IDLE_TIME", "15m"),
		},
		env:        env.GetString("ENV", "development"),
		mailDomain: env.GetString("MAILGUN_DOMAIN_NAME", "https://needbank.ng"),
		mailApiKey: env.GetString("MAILGUN_API_KEY", "key-3d7e0a1f2b4c5e6f8a9b0c1d2e3f4g5h"),

		payStackSK: env.GetString("PAYSTACK_SECRET_KEY", "pay_stack_secret_key"),
	}

	// logger
	logger := zap.Must(zap.NewProduction()).Sugar()
	defer logger.Sync()

	// db
	gormDB, err := db.New(cfg.db.dbAddr, cfg.db.maxOpenConns, cfg.db.maxIdleConns, cfg.db.maxIdleTime)
	if err != nil {
		logger.Fatal("failed to connect to database", zap.Error(err))
	}

	sqlDB, err := gormDB.DB()
	if err != nil {
		logger.Fatal("error getting sqlDb from gormDB", zap.Error(err))
	}
	defer sqlDB.Close()

	if err := store.AutoMigrate(gormDB); err != nil {
		logger.Fatal("error running migrations", zap.Error(err))
	}
	defer sqlDB.Close()
	logger.Info("db conncetion pool established")

	// Start the application
	// store
	store := store.NewStorage(gormDB)

	app := &application{
		config: cfg,
		logger: logger,
		middleWare: middleWareConfig{
			rateLimiters: make(map[string]*rate.Limiter),
		},

		store: store,
		ctx:   context.Background(),
	}

	mux := app.mount()
	logger.Fatal(app.run(mux))
}

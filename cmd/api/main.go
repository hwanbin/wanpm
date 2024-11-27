package main

import (
	"context"
	"database/sql"
	"flag"
	"log/slog"
	"os"
	"sync"
	"time"

	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/hwanbin/wanpm-api/internal/data"
	_ "github.com/lib/pq"
)

const version = "1.0.0"

type config struct {
	port int
	env  string
	db   struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  time.Duration
	}
	limiter struct {
		rps     float64
		burst   int
		enabled bool
	}
	s3 struct {
		profile string
		bucket  string
	}
}

type s3Actor struct {
	client        *s3.Client
	presignClient *s3.PresignClient
	uploader      *manager.Uploader
}

type application struct {
	config  config
	logger  *slog.Logger
	models  data.Models
	s3actor s3Actor
	wg      sync.WaitGroup
}

func main() {
	var cfg config

	flag.IntVar(&cfg.port, "port", 9000, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")

	flag.StringVar(&cfg.db.dsn, "db-dsn", os.Getenv("WANTONI_DB_DSN"), "PostgreSQL DSN")

	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max idle connections")
	flag.DurationVar(&cfg.db.maxIdleTime, "db-max-idle-time", 15*time.Minute, "PostgreSQL max connection idle time")

	flag.Float64Var(&cfg.limiter.rps, "limter-rps", 2, "Rate limiter maximum requests per second")
	flag.IntVar(&cfg.limiter.burst, "limiter-burst", 4, "Rate limiter maximum burst")
	flag.BoolVar(&cfg.limiter.enabled, "limiter-enabled", true, "Enable rate limiter")

	flag.StringVar(&cfg.s3.profile, "s3-profile", "s3_profile", "S3 profile")
	flag.StringVar(&cfg.s3.bucket, "s3-bucket", os.Getenv("S3_BUCKET_NAME"), "S3 bucket name")

	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	db, err := openDB(cfg)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	defer db.Close()

	logger.Info("database connection pool established")

	s3actor, err := initS3(cfg)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	logger.Info("s3 actor initialized")

	app := &application{
		config:  cfg,
		logger:  logger,
		models:  data.NewModels(db),
		s3actor: s3actor,
	}

	err = app.serve()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}

func openDB(cfg config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(cfg.db.maxOpenConns)
	db.SetMaxIdleConns(cfg.db.maxIdleConns)
	db.SetConnMaxIdleTime(cfg.db.maxIdleTime)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

func initS3(cfg config) (s3Actor, error) {
	s3Cfg, err := awsConfig.LoadDefaultConfig(
		context.Background(),
		awsConfig.WithSharedConfigProfile(cfg.s3.profile),
	)
	if err != nil {
		return s3Actor{}, err
	}

	client := s3.NewFromConfig(s3Cfg)
	uploader := manager.NewUploader(client)
	presignClient := s3.NewPresignClient(client)

	return s3Actor{
		client:        client,
		uploader:      uploader,
		presignClient: presignClient,
	}, nil
}

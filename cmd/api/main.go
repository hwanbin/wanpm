package main

import (
	"context"
	"database/sql"
	"flag"
	"github.com/hwanbin/wanpm/internal/mailer"
	"log/slog"
	"os"
	"sync"
	"time"

	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/hwanbin/wanpm/internal/data"
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
	smtp struct {
		host     string
		port     int
		username string
		password string
		sender   string
	}
	auth Auth
}

type jwtPayload struct {
	issuer       string
	audience     string
	cookieDomain string
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
	mailer  mailer.Mailer
	wg      sync.WaitGroup
}

func main() {
	var cfg config

	flag.IntVar(&cfg.port, "port", 9000, "API server port")
	flag.StringVar(&cfg.env, "env", "remote-dev", "Environment (local-dev|remote-dev|staging|production)")

	flag.StringVar(&cfg.db.dsn, "db-dsn", os.Getenv("WANTONI_DB_DSN"), "PostgreSQL DSN")

	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max idle connections")
	flag.DurationVar(&cfg.db.maxIdleTime, "db-max-idle-time", 15*time.Minute, "PostgreSQL max connection idle time")

	flag.Float64Var(&cfg.limiter.rps, "limter-rps", 20, "Rate limiter maximum requests per second")
	flag.IntVar(&cfg.limiter.burst, "limiter-burst", 40, "Rate limiter maximum burst")
	flag.BoolVar(&cfg.limiter.enabled, "limiter-enabled", true, "Enable rate limiter")

	flag.StringVar(&cfg.s3.profile, "s3-profile", "s3_profile", "S3 profile")
	flag.StringVar(&cfg.s3.bucket, "s3-bucket", os.Getenv("S3_BUCKET_NAME"), "S3 bucket name")

	flag.StringVar(&cfg.smtp.host, "smtp-host", os.Getenv("SMTP_HOST"), "SMTP host")
	flag.StringVar(&cfg.smtp.username, "smtp-username", os.Getenv("SMTP_USERNAME"), "SMTP username")
	flag.StringVar(&cfg.smtp.password, "smtp-password", os.Getenv("SMTP_PASSWORD"), "SMTP password")
	flag.IntVar(&cfg.smtp.port, "smtp-port", 587, "SMTP port")
	flag.StringVar(&cfg.smtp.sender, "smtp-sender", os.Getenv("SMTP_SENDER"), "SMTP sender")

	var jwtPayload jwtPayload
	flag.StringVar(&jwtPayload.issuer, "jwt-issuer", "wanton.app", "signing issuer")
	flag.StringVar(&jwtPayload.audience, "jwt-audience", "wanton.app", "signing audience")
	flag.StringVar(&jwtPayload.cookieDomain, "cookie-domain", ".wanton.app", "cookie domain")

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
		mailer:  mailer.New(cfg.smtp.host, cfg.smtp.port, cfg.smtp.username, cfg.smtp.password, cfg.smtp.sender),
	}

	app.config.auth = Auth{
		Issuer:        jwtPayload.issuer,
		Audience:      jwtPayload.audience,
		TokenExpiry:   time.Minute * 15,
		RefreshExpiry: time.Hour * 24,
		CookiePath:    "/",
		CookieName:    "refresh_token",
		CookieDomain:  jwtPayload.cookieDomain,
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

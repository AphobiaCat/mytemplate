package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	logc "mytemplate/pkg/log"

	"github.com/prometheus/client_golang/prometheus"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/plugin/dbresolver"
)

var (
	loggerIns logger.Interface

	DatabaseRequestDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "web3",
		Subsystem: "database",
		Name:      "request_duration",
		Help:      "Request duration",
	}, []string{
		"method",
	})
)

func init() {
	loggerIns = logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Second, // Slow SQL threshold
			LogLevel:                  logger.Warn, // Log level
			IgnoreRecordNotFoundError: true,        // Ignore ErrRecordNotFound error for logger
			ParameterizedQueries:      false,       // Don't include params in the SQL log
			Colorful:                  false,       // Disable color
		},
	)
}

func InitMySQLEngine(ctx context.Context, dsn, replicasDsn string, models ...interface{}) (*gorm.DB, error) {
	fmt.Println("dsn:", dsn)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: loggerIns,
	})
	if err != nil {
		logc.DebugError(ctx, "init mysql failed", err)
		return nil, err
	}
	if replicasDsn != "" {
		db.Use(dbresolver.Register(dbresolver.Config{
			Replicas: []gorm.Dialector{mysql.Open(replicasDsn)},
			Policy:   dbresolver.RandomPolicy{},
		}))
	}

	if err := db.AutoMigrate(models...); err != nil {
		panic(err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		logc.DebugError(ctx, "init mysql failed", err)
		return nil, err
	}
	sqlDB.SetMaxOpenConns(600)
	sqlDB.SetMaxIdleConns(300)
	sqlDB.SetConnMaxLifetime(3 * time.Minute)

	return db, sqlDB.Ping()
}

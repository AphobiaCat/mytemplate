package database

import (
	"context"
	"mytemplate/internal/config"

	logc "mytemplate/pkg/log"

	"gorm.io/gorm"
)

var (
	DBMysqlExample *gorm.DB
)

type Option func(*gorm.DB) *gorm.DB

func GetDBMysqlExample() *gorm.DB {
	if DBMysqlExample == nil {
		logc.DebugError(context.Background(), "example mysql db is nil")
		return nil
	}
	return DBMysqlExample
}

func Setup(c config.Config) {
	dbExample, err := InitMySQLEngine(context.Background(), c.MysqlExample.DSN, c.MysqlExample.ReplicasDSN)
	if err != nil || dbExample == nil {
		logc.DebugError(context.Background(), "init example mysql failed", err)
		return
	}
	DBMysqlExample = dbExample

	logc.DebugError(context.Background(), "init growth mysql success")
}

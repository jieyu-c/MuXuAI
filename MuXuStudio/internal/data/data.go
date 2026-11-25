package data

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"jieyuc.cn/muxu/studio/internal/conf"

	"github.com/go-kratos/kratos/v3/log"
	_ "github.com/go-sql-driver/mysql"
	"github.com/google/wire"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewUserRepo)

// Data holds the long-lived storage clients shared by repos.
type Data struct {
	db *gorm.DB
}

// NewData opens the database client and returns it with a cleanup function.
func NewData(c *conf.Data) (*Data, func(), error) {
	dc := c.GetDatabase()
	db, err := gorm.Open(postgres.Open(DSN(dc)), &gorm.Config{})
	if err != nil {
		return nil, nil, err
	}
	if dc.GetDebug() {
		db = db.Debug()
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, nil, err
	}
	// Auto migration is a convenience for local development. In production,
	// apply schema changes as a separate reviewed step instead.
	if dc.GetAutoMigrate() {
		if err := migrate(db); err != nil {
			defer func() {
				err := sqlDB.Close()
				if err != nil {
					log.Error("close database error, %s", err.Error())
				}
			}()

			return nil, nil, err
		}
	}
	cleanup := func() {
		log.Info("closing the data resources")
		if err := sqlDB.Close(); err != nil {
			log.Error("failed closing the database", "err", err)
		}
	}
	return &Data{db: db}, cleanup, nil
}

// DSN 整理数据库配置为PostgreSQL对应的格式
func DSN(database *conf.Data_Database) string {
	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=Asia/Shanghai",
		database.GetHost(), database.GetUser(), database.GetPassword(), database.GetDbname(), database.GetPort())
}

func migrate(db *gorm.DB) error {
	return db.AutoMigrate(&User{})
}
